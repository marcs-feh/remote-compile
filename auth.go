package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"slices"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const ArgonMem = 64 * 1024 /* Memory, in KiB, so this is 64MiB */
const SaltLength = 16
const SessionKeyLength = 128

type Session struct {
	userId Id
	created time.Time
	timeToLive time.Duration
}

type SessionKey string


type SessionStore struct {
	sessions map[SessionKey]Session
	mutex *sync.Mutex
}

func NewSessionStore() SessionStore {
	return SessionStore{
		sessions: make(map[SessionKey]Session),
		mutex: new(sync.Mutex),
	}
}

func (store *SessionStore) BeginSession(db *sql.DB, username, password string, ttl time.Duration) (sessionKey SessionKey, err error) {
	user, found, err := GetUserByName(db, username)
	if !found || err != nil {
		err = fmt.Errorf("failed to begin session, could not fetch user %q", username)
		return
	}

	authorized := VerifyPassword(password, user.authHash, user.authSalt)
	if !authorized {
		err = fmt.Errorf("failed to verify password")
		return
	}

	var keyBytes []byte
	keyBytes, err = generateRandomBytes(SessionKeyLength)
	if err != nil { return }
	sessionKey = SessionKey(base64Encode(keyBytes))

	session := Session {
		userId: user.id,
		created: time.Now(),
		timeToLive: ttl,
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.sessions[sessionKey] = session

	return
}

func (store *SessionStore) EndSession(key SessionKey){
	store.mutex.Lock()
	defer store.mutex.Unlock()
	delete(store.sessions, key)
}


func generateRandomBytes(nbytes int) ([]byte, error) {
	buf := make([]byte, nbytes)
	n, err := rand.Reader.Read(buf)
	if err != nil {
		return buf, err
	}
	if n != nbytes {
		return buf, fmt.Errorf("failed to read enough bytes")
	}
	return buf, err
}

func generateSalt() ([]byte, error) {
	return generateRandomBytes(SaltLength)
}

func VerifyPassword(password string, key, salt []byte) bool {
	key2 := argon2.IDKey([]byte(password), salt, 1, ArgonMem, 4, 64)
	return slices.Equal(key, key2)
}

func CreatePassword(password string) (key []byte, salt []byte, err error) {
	salt, err = generateSalt()
	if err != nil { return }
	key = argon2.IDKey([]byte(password), salt, 1, ArgonMem, 4, 64)
	return
}

