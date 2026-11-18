package utils

import (
	crand "crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// TwofaChallenge represents an ongoing 2FA verification attempt during login or setup.
type TwofaChallenge struct {
	Nonce     string
	Email     string
	UserIDs   []string
	Method    string // "email" | "totp"
	Code      string // Expected OTP for email method or setup confirmation
	Secret    string // Base32 TOTP secret for setup confirmation or cached verification
	Attempts  int
	ExpiresAt time.Time
}

// TwofaStore is an in-memory, thread-safe store for 2FA challenges.
type TwofaStore struct {
	mu         sync.Mutex
	challenges map[string]*TwofaChallenge
	byEmail    map[string]string // email -> nonce for deduplication / retrieval by request ID
}

var globalTwofaStore = newTwofaStore()

func newTwofaStore() *TwofaStore {
	store := &TwofaStore{
		challenges: make(map[string]*TwofaChallenge),
		byEmail:    make(map[string]string),
	}
	go store.cleanupLoop()
	return store
}

// GetTwofaStore returns the singleton TwofaStore instance.
func GetTwofaStore() *TwofaStore {
	return globalTwofaStore
}

// CreateChallenge creates a new challenge valid for ttl duration.
func (s *TwofaStore) CreateChallenge(email string, userIDs []string, method, code, secret string, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := crand.Read(raw); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(raw)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean any previous nonce for this email if present
	if oldNonce, exists := s.byEmail[email]; exists {
		delete(s.challenges, oldNonce)
	}

	challenge := &TwofaChallenge{
		Nonce:     nonce,
		Email:     email,
		UserIDs:   userIDs,
		Method:    method,
		Code:      code,
		Secret:    secret,
		Attempts:  0,
		ExpiresAt: time.Now().Add(ttl),
	}

	s.challenges[nonce] = challenge
	s.byEmail[email] = nonce

	return nonce, nil
}

// GetChallenge retrieves an active, unexpired challenge.
func (s *TwofaStore) GetChallenge(nonce string) (*TwofaChallenge, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, ok := s.challenges[nonce]
	if !ok || time.Now().After(challenge.ExpiresAt) {
		return nil, false
	}
	return challenge, true
}

// GetChallengeByEmail retrieves an active, unexpired challenge by email.
func (s *TwofaStore) GetChallengeByEmail(email string) (*TwofaChallenge, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nonce, ok := s.byEmail[email]
	if !ok {
		return nil, false
	}
	challenge, ok := s.challenges[nonce]
	if !ok || time.Now().After(challenge.ExpiresAt) {
		return nil, false
	}
	return challenge, true
}

// GetCodeByEmail retrieves the current active OTP code for an email (used by IssueActionLink).
func (s *TwofaStore) GetCodeByEmail(email string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	nonce, ok := s.byEmail[email]
	if !ok {
		return ""
	}
	challenge, ok := s.challenges[nonce]
	if !ok || time.Now().After(challenge.ExpiresAt) {
		return ""
	}
	return challenge.Code
}

// RecordFailedAttempt increments the failed attempt count. If attempts >= 5, removes the challenge.
func (s *TwofaStore) RecordFailedAttempt(nonce string) (attemptsLeft int, locked bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, ok := s.challenges[nonce]
	if !ok {
		return 0, true
	}
	challenge.Attempts++
	if challenge.Attempts >= 5 {
		delete(s.byEmail, challenge.Email)
		delete(s.challenges, nonce)
		return 0, true
	}
	return 5 - challenge.Attempts, false
}

// ConsumeChallenge removes and returns the challenge on successful verification.
func (s *TwofaStore) ConsumeChallenge(nonce string) (*TwofaChallenge, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, ok := s.challenges[nonce]
	if !ok || time.Now().After(challenge.ExpiresAt) {
		return nil, false
	}

	delete(s.byEmail, challenge.Email)
	delete(s.challenges, nonce)
	return challenge, true
}

// UpdateCode updates the expected code (e.g. on OTP resend) and extends expiry.
func (s *TwofaStore) UpdateCode(nonce, newCode string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, ok := s.challenges[nonce]
	if !ok || time.Now().After(challenge.ExpiresAt) {
		return false
	}
	challenge.Code = newCode
	challenge.ExpiresAt = time.Now().Add(ttl)
	return true
}

func (s *TwofaStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for nonce, c := range s.challenges {
			if now.After(c.ExpiresAt) {
				delete(s.byEmail, c.Email)
				delete(s.challenges, nonce)
			}
		}
		s.mu.Unlock()
	}
}
