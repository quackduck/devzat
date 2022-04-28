package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
)

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
			Log.Fatal("Error reading tokens file: " + err.Error())
		}
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&Tokens)
	if err != nil {
		error := "Error decoding tokens file: " + err.Error()
		MainRoom.broadcast(Devbot, error)
		Log.Fatal(error)
	}
}

func saveTokens() {
	f, err := os.Create(Config.DataDir + string(os.PathSeparator) + "tokens.json")
	if err != nil {
		Log.Fatal("Error saving tokens file: " + err.Error())
	}
	defer f.Close()
	data, err := json.Marshal(Tokens)
	if err != nil {
		error := "Error encoding tokens file: " + err.Error()
		MainRoom.broadcast(Devbot, error)
		Log.Fatal(error)
	}
	_, err = f.Write(data)
	if err != nil {
		error := "Error writing tokens file: " + err.Error()
		MainRoom.broadcast(Devbot, error)
		Log.Fatal(error)
	}
}

func checkToken(token string) bool {
	if token == Integrations.RPC.Key {
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
		u.writeln(Devbot, "No tokens found.")
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
		u.writeln(Devbot, "Please provide a sha256 hash of a token to revoke.")
		return
	}
	for i, t := range Tokens {
		if shasum(t.Token) == rest {
			Tokens = append(Tokens[:i], Tokens[i+1:]...)
			saveTokens()
			u.writeln(Devbot, "Token revoked!")
			return
		}
	}
	u.writeln(Devbot, "Token not found.")
}

func grantTokenCMD(rest string, u *User) {
	if !auth(u) {
		u.room.broadcast(Devbot, "Not authorized")
		return
	}

	toUser, ok := findUserByName(u.room, rest)
	if !ok {
		// fallback to sending the token to the admin running the cmd
		toUser = u
	}
	token := generateToken()
	Tokens = append(Tokens, tokensDbEntry{token, rest})
	if toUser != u {
		toUser.writeln(Devbot, "You have been granted a token: "+token)
	}
	u.writeln(Devbot, "Granted token: "+token)
	saveTokens()
}

func generateToken() string {
	// https://stackoverflow.com/a/59457748
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		Log.Fatal("Error generating token: " + err.Error())
	}
	token := "dvz@" + hex.EncodeToString(b)
	// check if it's already in use
	for _, t := range Tokens {
		if t.Token == token {
			return generateToken()
		}
	}
	return token
}
