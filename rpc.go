package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"

	pb "devzat/plugin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	PluginCMDs = map[string]PluginCMD{}
	RPCCMDs    = []CMD{
		{"plugins", pluginsCMD, "", "List plugin commands"},
	}
	RPCCMDsRest = []CMD{
		{"lstokens", lsTokensCMD, "", "List plugin token hashes and their data (admin)"},
		{"revoke", revokeTokenCMD, "<token hash>", "Revoke a plugin token (admin)"},
		{"grant", grantTokenCMD, "[user] [data]", "Grant a token and optionally send it to a user (admin)"},
	}
	ListenersNonMiddleware = make([]chan pb.MiddlewareChannelMessage, 0, 4)
	ListenersMiddleware    = make([]chan pb.MiddlewareChannelMessage, 0, 4)
	Tokens                 = make(map[string]string, 10)
)

type pluginServer struct {
	pb.UnimplementedPluginServer
	lock sync.Mutex
}

func (s *pluginServer) RegisterListener(stream pb.Plugin_RegisterListenerServer) error {
	s.lock.Lock()
	Log.Println("[gRPC] Registering event listener")
	initialData, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	listener := initialData.GetListener()
	if listener == nil {
		return status.Error(codes.InvalidArgument, "First message must be a listener")
	}

	isMiddleware := listener.Middleware != nil && *listener.Middleware
	isOnce := listener.Once != nil && *listener.Once

	var regex *regexp.Regexp
	if listener.Regex != nil {
		regex, err = regexp.Compile(*listener.Regex)
		if err != nil {
			return status.Error(codes.InvalidArgument, "Invalid regex")
		}
	}

	var listenerList *[]chan pb.MiddlewareChannelMessage

	if isMiddleware {
		listenerList = &ListenersMiddleware
	} else {
		listenerList = &ListenersNonMiddleware
	}

	c := make(chan pb.MiddlewareChannelMessage)
	*listenerList = append(*listenerList, c)

	s.lock.Unlock()
	defer func() {
		// Remove the channel from the list where the channel is equal to c
		for i := range *listenerList {
			if (*listenerList)[i] == c {
				*listenerList = append((*listenerList)[:i], (*listenerList)[i+1:]...)
				break
			}
		}
	}()

	for {
		message := <-c

		// If something goes wrong, make sure the goroutine sending the message doesn't block on waiting for a response
		sendNilResponse := func() {
			c <- &pb.ListenerClientData_Response{
				Response: &pb.MiddlewareResponse{
					Msg: nil,
				},
			}
		}

		// If there's a regex and it doesn't match, don't send the message to the plugin
		if listener.Regex != nil && !regex.MatchString(message.(*pb.Event).Msg) {
			if isMiddleware {
				sendNilResponse()
			}
			continue
		}

		err = stream.Send(message.(*pb.Event))
		if err != nil {
			if isMiddleware {
				sendNilResponse()
			}
			return err
		}

		if isMiddleware {
			mwRes, err := stream.Recv()
			if err != nil {
				sendNilResponse()
				return err
			}
			switch data := mwRes.Data.(type) {
			case *pb.ListenerClientData_Listener:
				sendNilResponse()
				return status.Error(codes.InvalidArgument, "Middleware returned a listener instead of a response")
			case *pb.ListenerClientData_Response:
				c <- data
			}
		}

		if isOnce {
			break
		}
	}
	return nil
}

type PluginCMD struct {
	argsInfo       string
	info           string
	invocationChan chan *pb.CmdInvocation
}

func (s *pluginServer) RegisterCmd(def *pb.CmdDef, stream pb.Plugin_RegisterCmdServer) error {
	s.lock.Lock()
	Log.Print("[gRPC] Registering command with name " + def.Name)
	PluginCMDs[def.Name] = PluginCMD{
		argsInfo:       def.ArgsInfo,
		info:           def.Info,
		invocationChan: make(chan *pb.CmdInvocation),
	}
	s.lock.Unlock()
	defer delete(PluginCMDs, def.Name)

	for {
		if err := stream.Send(<-PluginCMDs[def.Name].invocationChan); err != nil {
			return err
		}
	}
}

func (s *pluginServer) SendMessage(ctx context.Context, msg *pb.Message) (*pb.MessageRes, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if msg.GetEphemeralTo() != "" {
		u, success := findUserByName(Rooms[msg.Room], *msg.EphemeralTo)
		if !success {
			return nil, status.Error(codes.NotFound, "Could not find user "+*msg.EphemeralTo)
		}
		u.writeln(msg.GetFrom()+" -> ", msg.Msg)
	} else {
		r := Rooms[msg.Room]
		if r == nil {
			return nil, status.Error(codes.InvalidArgument, "Room does not exist")
		}
		r.broadcast(msg.GetFrom(), msg.Msg)
	}
	return &pb.MessageRes{}, nil
}

func authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "Missing metadata")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Error(codes.Unauthenticated, "Missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")

	if Integrations.RPC.Key != "" && token == Integrations.RPC.Key {
		return nil
	}
	if _, ok = Tokens[token]; !ok {
		return status.Error(codes.Unauthenticated, "Invalid authorization header")
	}
	return nil
}

func rpcInit() {
	if Integrations.RPC == nil {
		return
	}
	MainCMDs = append(MainCMDs, RPCCMDs...)
	RestCMDs = append(RestCMDs, RPCCMDsRest...)
	initTokens()
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", Integrations.RPC.Port))
		if err != nil {
			Log.Println("[gRPC] Failed to listen for plugin server:", err)
			return
		}
		// TODO: add TLS if configured
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				if err := authorize(ctx); err != nil {
					return nil, err
				}
				return handler(ctx, req)
			}),
			grpc.StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
				if err := authorize(stream.Context()); err != nil {
					return err
				}
				return handler(srv, stream)
			}),
			grpc.KeepaliveParams(keepalive.ServerParameters{Time: time.Second * 10}),
		)
		pb.RegisterPluginServer(grpcServer, &pluginServer{})
		Log.Printf("[gRPC] Plugin server started on port %d\n", Integrations.RPC.Port)
		if err = grpcServer.Serve(lis); err != nil {
			Log.Println("[gRPC] Failed to serve:", err)
		}
	}()
}

func runPluginCMDs(u *User, currCmd string, args string) (found bool) {
	if pluginCmd, ok := PluginCMDs[currCmd]; ok {
		pluginCmd.invocationChan <- &pb.CmdInvocation{
			Room: u.room.name,
			From: stripansi.Strip(u.Name),
			Args: args,
		}
		return true
	}
	return false
}

// Hook that is called when a user sends a message (not private DMs)
func sendMessageToPlugins(line string, u *User) {
	if len(ListenersNonMiddleware) > 0 {
		for _, l := range ListenersNonMiddleware {
			l <- &pb.Event{
				Room: u.room.name,
				From: stripansi.Strip(u.Name),
				Msg:  line,
			}
		}
	}
}

var middlewareLock = new(sync.Mutex)

func getMiddlewareResult(u *User, line string) string {
	if Integrations.RPC == nil {
		return line
	}

	middlewareLock.Lock()
	defer middlewareLock.Unlock()
	// Middleware hook
	if len(ListenersMiddleware) > 0 {
		for i := range ListenersMiddleware {
			ListenersMiddleware[i] <- &pb.Event{
				Room: u.room.name,
				From: stripansi.Strip(u.Name),
				Msg:  line,
			}
			if res := (<-ListenersMiddleware[i]).(*pb.ListenerClientData_Response).Response.Msg; res != nil {
				line = *res
			}
		}
	}
	return line
}

func pluginsCMD(_ string, u *User) {
	plugins := make([]CMD, 0, len(PluginCMDs))
	for n, c := range PluginCMDs {
		plugins = append(plugins, CMD{
			name:     n,
			info:     c.info,
			argsInfo: c.argsInfo,
		})
	}
	autogenerated := autogenCommands(plugins)
	if autogenerated == "" {
		autogenerated = "   (no plugin commands are loaded)"
	}
	u.room.broadcast("", "Plugin commands  \n"+autogenerated)
}

func initTokens() {
	f, err := os.Open(Config.DataDir + string(os.PathSeparator) + "tokens.json")
	if err != nil {
		if !os.IsNotExist(err) {
			Log.Println("Error reading tokens file:", err)
		}
		return
	}
	defer f.Close()
	j, err := io.ReadAll(f)
	if err != nil {
		Log.Println("Error reading tokens file:", err)
		return
	}

	err = json.Unmarshal(j, &Tokens)
	if err != nil {
		var s []struct { // old format
			Token string `json:"token"`
			Data  string `json:"data"`
		}
		err = json.Unmarshal(j, &s)
		if err != nil {
			Log.Println("Error decoding tokens file:", err)
			return
		}
		Log.Println("Changing token file format")
		for i := range s {
			Tokens[s[i].Token] = s[i].Data
		}
		f.Close()
		saveTokens()
	}
}

func saveTokens() {
	f, err := os.Create(Config.DataDir + string(os.PathSeparator) + "tokens.json")
	if err != nil {
		Log.Println(err)
	}
	defer f.Close()
	data, err := json.Marshal(Tokens)
	if err != nil {
		Log.Println("Error encoding tokens file:", err)
	}
	_, err = f.Write(data)
	if err != nil {
		Log.Println("Error writing tokens file:", err)
	}
}

func lsTokensCMD(_ string, u *User) {
	if !auth(u) {
		u.room.broadcast(Devbot, "Not authorized")
		return
	}

	if len(Tokens) == 0 {
		u.room.broadcast(Devbot, "No tokens found.")
		return
	}
	msg := "Tokens:  \n"
	i := 0
	for t := range Tokens {
		i++
		msg += Cyan.Cyan(strconv.Itoa(i+1)) + ". " + shasum(t) + "\t" + Tokens[t] + "  \n"
	}
	u.writeln(Devbot, msg)
}

func revokeTokenCMD(rest string, u *User) {
	if !auth(u) {
		u.room.broadcast(Devbot, "Not authorized")
		return
	}

	if len(rest) == 0 {
		u.room.broadcast(Devbot, "Please provide a sha256 hash of a token to revoke.")
		return
	}
	for token := range Tokens {
		if shasum(token) == rest {
			delete(Tokens, token)
			saveTokens()
			u.room.broadcast(Devbot, "Token revoked!")
			return
		}
	}
	u.room.broadcast(Devbot, "Token not found.")
}

func grantTokenCMD(rest string, u *User) {
	if !auth(u) {
		u.room.broadcast(Devbot, "Not authorized")
		return
	}

	token, err := generateToken()
	if err != nil {
		u.room.broadcast(Devbot, "Error generating token: "+err.Error())
		Log.Println(err)
		return
	}

	split := strings.Fields(rest)
	if len(split) > 0 && len(split[0]) > 0 && split[0][0] == '@' {
		toUser, ok := findUserByName(u.room, split[0][1:])
		if ok {
			toUser.writeln(Devbot, "You have been granted a token: "+token)
		} else {
			u.room.broadcast(Devbot, "Who's that?")
			return
		}
	}
	Tokens[token] = rest
	u.writeln(Devbot, "Granted token: "+token)
	saveTokens()
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := "dvz@" + base64.StdEncoding.EncodeToString(b)
	// check if it's already in use
	if _, ok := Tokens[token]; ok {
		return generateToken()
	}
	return token, nil
}
