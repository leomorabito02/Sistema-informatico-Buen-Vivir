package utils

import (
	crand "crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)


// nonceEntry holds the set of valid user IDs for a login nonce and its expiry time.
type nonceEntry struct {
	allowedUserIDs map[string]struct{}
	expiresAt      time.Time
}

// NonceStore is an in-memory, thread-safe store for single-use login nonces.
type NonceStore struct {
	mu      sync.Mutex
	entries map[string]nonceEntry
}

var globalNonceStore = newNonceStore()

func newNonceStore() *NonceStore {
	ns := &NonceStore{entries: make(map[string]nonceEntry)}
	go ns.cleanupLoop()
	return ns
}

// GetNonceStore returns the singleton NonceStore instance.
func GetNonceStore() *NonceStore {
	return globalNonceStore
}

// Generate creates a new single-use nonce bound to the given user IDs, valid for ttl duration.
// Returns the opaque nonce string.
func (ns *NonceStore) Generate(userIDs []string, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := crand.Read(raw); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(raw)

	allowed := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		allowed[id] = struct{}{}
	}

	ns.mu.Lock()
	ns.entries[nonce] = nonceEntry{
		allowedUserIDs: allowed,
		expiresAt:      time.Now().Add(ttl),
	}
	ns.mu.Unlock()

	return nonce, nil
}

// Consume validates that the nonce exists, has not expired, and that userID is in its allowed set.
// On success it deletes the nonce (single-use guarantee) and returns true.
func (ns *NonceStore) Consume(nonce, userID string) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	entry, ok := ns.entries[nonce]
	if !ok || time.Now().After(entry.expiresAt) {
		delete(ns.entries, nonce)
		return false
	}
	if _, allowed := entry.allowedUserIDs[userID]; !allowed {
		return false
	}

	delete(ns.entries, nonce)
	return true
}

// cleanupLoop removes expired nonces every 5 minutes to prevent memory growth.
func (ns *NonceStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		ns.mu.Lock()
		for k, v := range ns.entries {
			if now.After(v.expiresAt) {
				delete(ns.entries, k)
			}
		}
		ns.mu.Unlock()
	}
}
