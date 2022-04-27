package main

import (
	"encoding/json"
	"math/rand"
	"os"
)

type tokensDbEntry struct {
	token string
	data  string
}
type tokensDb []tokensDbEntry

var Tokens = make(tokensDb, 0)

func readTokens() {
	f, err := os.Open(Config.DataDir + string(os.PathSeparator) + "tokens.json")
	if err != nil && !os.IsNotExist(err) {
		Log.Fatal("Error reading tokens file: " + err.Error())
	}
	defer f.Close()
	if err == nil {
		return
	} // if it doesn't exist, it will be created on next save
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
	err = json.NewEncoder(f).Encode(Tokens)
	if err != nil {
		error := "Error encoding tokens file: " + err.Error()
		MainRoom.broadcast(Devbot, error)
		Log.Fatal(error)
	}
}

func checkToken(token string) bool {
	for _, t := range Tokens {
		if t.token == token {
			return true
		}
	}
	return false
}

func lsTokensCMD(_ string, u *User) {
	if len(Tokens) == 0 {
		u.writeln(Devbot, "No tokens found.")
		return
	}
	u.writeln(Devbot, "Tokens:")
	for _, t := range Tokens {
		u.writeln(Devbot, t.token+"    "+t.data)
	}
}

func revokeTokenCMD(rest string, u *User) {
	if len(rest) == 0 {
		u.writeln(Devbot, "Please provide a token to revoke.")
		return
	}
	for i, t := range Tokens {
		if t.token == rest {
			Tokens = append(Tokens[:i], Tokens[i+1:]...)
			saveTokens()
			u.writeln(Devbot, "Token revoked!")
			return
		}
	}
	u.writeln(Devbot, "Token not found.")
}

func grantTokenCMD(rest string, u *User) {
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
	// get a random token
	token := ""
	for i := 0; i < 32; i++ {
		token += string(rune(65 + rand.Intn(25)))
	}
	// check if it's already in use
	for _, t := range Tokens {
		if t.token == token {
			return generateToken()
		}
	}
	return token
}
