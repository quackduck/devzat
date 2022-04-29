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
	PluginCMDs = map[string]cmdInst{}
	RPCCMDs    = []CMD{
		{"plugins", pluginsCMD, "", "List plugin commands"},
	}
	RPCCMDsRest = []CMD{
		{"lstokens", lsTokensCMD, "", "List plugin token hashes and their data (admin)"},
		{"revoke", revokeTokenCMD, "<token hash>", "Revoke a plugin token (admin)"},
		{"grant", grantTokenCMD, "[user] [data]", "Grant a token and optionally send it to a user (admin)"},
	}
)

func addRPCCMDs() {
	if Integrations.RPC != nil {
		MainCMDs = append(MainCMDs, RPCCMDs...)
		RestCMDs = append(RestCMDs, RPCCMDsRest...)
	}
}

type pluginServer struct {
	pb.UnimplementedPluginServer
}

// Since pb.MiddlewareChannelMessage includes pb.IsEvent_Event, we'll just use that
type listenerCollection struct {
	nonMiddleware []chan pb.MiddlewareChannelMessage
	middleware    []chan pb.MiddlewareChannelMessage
}

var listeners = listenerCollection{
	nonMiddleware: make([]chan pb.MiddlewareChannelMessage, 0, 4),
	middleware:    make([]chan pb.MiddlewareChannelMessage, 0, 4),
}

func (s *pluginServer) RegisterListener(stream pb.Plugin_RegisterListenerServer) error {
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

	var c *chan pb.MiddlewareChannelMessage
	isMiddleware := listener.Middleware != nil && *listener.Middleware
	isOnce := listener.Once != nil && *listener.Once

	var regex *regexp.Regexp
	if listener.Regex != nil {
		regex, err = regexp.Compile(*listener.Regex)
		if err != nil {
			return status.Error(codes.InvalidArgument, "Invalid regex")
		}
	}

	var channelCollection *[]chan pb.MiddlewareChannelMessage

	if isMiddleware {
		channelCollection = &listeners.middleware
	} else {
		channelCollection = &listeners.nonMiddleware
	}

	*channelCollection = append(*channelCollection, make(chan pb.MiddlewareChannelMessage))
	c = &(*channelCollection)[len(*channelCollection)-1]
	defer func() {
		// Remove the channel from the channelCollection where the channel is equal to c
		for i, channel := range *channelCollection {
			if channel == *c {
				*channelCollection = append((*channelCollection)[:i], (*channelCollection)[i+1:]...)
				break
			}
		}
	}()

	for {
		message := <-*c

		// If something goes wrong, make sure the goroutine sending the message doesn't block on waiting for a response
		sendNilResponse := func() {
			*c <- &pb.ListenerClientData_Response{
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

		err := stream.Send(message.(*pb.Event))

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
				*c <- data
			}
		}

		if isOnce {
			break
		}
	}

	return nil
}

type cmdInst struct {
	argsInfo string
	info     string
	c        chan *pb.CmdInvocation
}

func (s *pluginServer) RegisterCmd(def *pb.CmdDef, stream pb.Plugin_RegisterCmdServer) error {
	Log.Print("[gRPC] Registering command with name " + def.Name)
	PluginCMDs[def.Name] = cmdInst{
		argsInfo: def.ArgsInfo,
		info:     def.Info,
		c:        make(chan *pb.CmdInvocation),
	}

	defer delete(PluginCMDs, def.Name)

	for {
		invocation := <-PluginCMDs[def.Name].c
		err := stream.Send(invocation)
		if err != nil {
			return err
		}
	}
}

func (s *pluginServer) SendMessage(ctx context.Context, msg *pb.Message) (*pb.MessageRes, error) {
	if msg.GetEphemeralTo() != "" {
		u, success := findUserByName(Rooms[msg.Room], *msg.EphemeralTo)
		if !success {
			return nil, status.Error(codes.NotFound, "Could not find user "+*msg.EphemeralTo)
		}
		u.writeln(msg.GetFrom(), msg.Msg)
	} else {
		r := Rooms[msg.Room]
		if r == nil {
			return nil, status.Error(codes.InvalidArgument, "Room does not exist")
		}
		r.broadcast(msg.GetFrom(), msg.Msg)
	}
	return &pb.MessageRes{}, nil
}

func newPluginServer() *pluginServer {
	s := &pluginServer{}
	return s
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

	token := values[0]
	token = strings.TrimPrefix(token, "Bearer ")
	if !checkToken(token) {
		return status.Error(codes.Unauthenticated, "Invalid authorization header")
	}
	return nil
}

func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	err := authorize(ctx)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func streamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := authorize(stream.Context())
	if err != nil {
		return err
	}

	return handler(srv, stream)
}

func rpcInit() {
	if Integrations.RPC == nil {
		return
	}
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", Integrations.RPC.Port))
		if err != nil {
			fmt.Println("[gRPC] Failed to listen for plugin server:", err)
			return
		}
		// TODO: add TLS if configured
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(unaryInterceptor),
			grpc.StreamInterceptor(streamInterceptor),
			grpc.KeepaliveParams(keepalive.ServerParameters{Time: time.Second * 10}),
		)
		pb.RegisterPluginServer(grpcServer, newPluginServer())
		fmt.Printf("[gRPC] Plugin server started on port %d\n", Integrations.RPC.Port)
		err = grpcServer.Serve(lis)
		if err != nil {
			fmt.Println("[gRPC] Failed to serve:", err)
		}
	}()
}

func runPluginCMDs(u *User, currCmd string, args string) (found bool) {
	if pluginCmd, ok := PluginCMDs[currCmd]; ok {
		pluginCmd.c <- &pb.CmdInvocation{
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
	if len(listeners.nonMiddleware) > 0 {
		for _, l := range listeners.nonMiddleware {
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
	if len(listeners.middleware) > 0 {
		for _, m := range listeners.middleware {
			m <- &pb.Event{
				Room: u.room.name,
				From: stripansi.Strip(u.Name),
				Msg:  line,
			}
			res := (<-m).(*pb.ListenerClientData_Response).Response
			if res.Msg != nil {
				line = *res.Msg
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

type tokensDbEntry struct {
	Token string `json:"token"`
	Data  string `json:"data"`
}

var Tokens = make([]tokensDbEntry, 0)

func initTokens() {
	if Integrations.RPC == nil {
		return
	}

	f, err := os.Open(Config.DataDir + string(os.PathSeparator) + "tokens.json")
	if err != nil {
		if !os.IsNotExist(err) {
			Log.Println("Error reading tokens file:", err)
		}
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&Tokens)
	if err != nil {
		MainRoom.broadcast(Devbot, "Error decoding tokens file: "+err.Error())
		Log.Println(err)
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
		MainRoom.broadcast(Devbot, "Error encoding tokens file: "+err.Error())
		Log.Println(err)
	}
	_, err = f.Write(data)
	if err != nil {
		MainRoom.broadcast(Devbot, "Error writing tokens file: "+err.Error())
		Log.Println(err)
	}
}

func checkToken(token string) bool {
	if Integrations.RPC.Key != "" && token == Integrations.RPC.Key {
		return true
	}
	for _, t := range Tokens {
		if t.Token == token {
			return true
		}
	}
	return false
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
	u.writeln(Devbot, "Tokens:")
	for _, t := range Tokens {
		u.writeln(Devbot, shasum(t.Token)+"    "+t.Data)
	}
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
	for i, t := range Tokens {
		if shasum(t.Token) == rest {
			Tokens = append(Tokens[:i], Tokens[i+1:]...)
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

	Tokens = append(Tokens, tokensDbEntry{token, rest})
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
	for _, t := range Tokens {
		if t.Token == token {
			return generateToken()
		}
	}
	return token, nil
}
