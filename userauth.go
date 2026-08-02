package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type UserRecord struct {
	PasswordHash string `json:"password_hash"`
}

var (
	UserAuthStore = map[string]*UserRecord{}
	userAuthMu    sync.RWMutex
)

func userAuthPath() string {
	return filepath.Join(Config.DataDir, "user-auth.json")
}

func loadUserAuth() {
	data, err := os.ReadFile(userAuthPath())
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		Log.Println("Could not load user auth:", err)
		return
	}
	userAuthMu.Lock()
	defer userAuthMu.Unlock()
	if err = json.Unmarshal(data, &UserAuthStore); err != nil {
		Log.Println("Could not parse user auth:", err)
	}
}

func saveUserAuth() error {
	userAuthMu.RLock()
	data, err := json.MarshalIndent(UserAuthStore, "", "  ")
	userAuthMu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(userAuthPath(), data, 0600)
}

func registerUser(username, password string) error {
	username = strings.ToLower(cleanName(username))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	userAuthMu.Lock()
	UserAuthStore[username] = &UserRecord{PasswordHash: string(hash)}
	userAuthMu.Unlock()
	return saveUserAuth()
}

func removeRegisteredUser(username string) bool {
	username = strings.ToLower(cleanName(username))
	userAuthMu.Lock()
	_, ok := UserAuthStore[username]
	if ok {
		delete(UserAuthStore, username)
	}
	userAuthMu.Unlock()
	if ok {
		saveUserAuth() //nolint:errcheck
	}
	return ok
}

func isUserRegistered(username string) bool {
	userAuthMu.RLock()
	_, ok := UserAuthStore[strings.ToLower(username)]
	userAuthMu.RUnlock()
	return ok
}

func verifyPassword(username, password string) bool {
	userAuthMu.RLock()
	record, ok := UserAuthStore[strings.ToLower(username)]
	userAuthMu.RUnlock()
	if !ok {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(password)) == nil
}
