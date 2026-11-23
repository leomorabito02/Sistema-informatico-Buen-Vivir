package utils

import (
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"auth/internal/models"
	"gorm.io/gorm"
)

const (
	whereNonceActive = "nonce = ? AND expires_at > ?"
	whereNonce       = "nonce = ?"
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

// TwofaStore is a thread-safe store for 2FA challenges with optional PostgreSQL persistence.
type TwofaStore struct {
	mu         sync.Mutex
	db         *gorm.DB
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

// SetDB configures optional PostgreSQL backing so active challenges survive Cloud Run scaling.
func (s *TwofaStore) SetDB(db *gorm.DB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db = db
}

func (s *TwofaStore) getDB() *gorm.DB {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db
}

func normalizeEmailKey(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func recordToChallenge(rec *models.TwofaChallengeRecord) *TwofaChallenge {
	var userIDs []string
	if rec.UserIDsJSON != "" {
		_ = json.Unmarshal([]byte(rec.UserIDsJSON), &userIDs)
	}
	return &TwofaChallenge{
		Nonce:     rec.Nonce,
		Email:     rec.Email,
		UserIDs:   userIDs,
		Method:    rec.Method,
		Code:      rec.Code,
		Secret:    rec.Secret,
		Attempts:  rec.Attempts,
		ExpiresAt: rec.ExpiresAt,
	}
}

// CreateChallenge creates a new challenge valid for ttl duration.
func (s *TwofaStore) CreateChallenge(email string, userIDs []string, method, code, secret string, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := crand.Read(raw); err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(raw)
	emailKey := normalizeEmailKey(email)
	expiresAt := time.Now().Add(ttl)

	if db := s.getDB(); db != nil {
		if userIDs == nil {
			userIDs = []string{}
		}
		userIDsBytes, err := json.Marshal(userIDs)
		if err != nil {
			return "", err
		}
		rec := &models.TwofaChallengeRecord{
			Nonce:       nonce,
			Email:       emailKey,
			UserIDsJSON: string(userIDsBytes),
			Method:      method,
			Code:        code,
			Secret:      secret,
			Attempts:    0,
			ExpiresAt:   expiresAt,
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("email = ?", emailKey).Delete(&models.TwofaChallengeRecord{}).Error; err != nil {
				return err
			}
			return tx.Create(rec).Error
		})
		if err != nil {
			slog.Error("twofa_store: failed to persist challenge", "error", err)
			return "", err
		}
		return nonce, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if oldNonce, exists := s.byEmail[emailKey]; exists {
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
		ExpiresAt: expiresAt,
	}

	s.challenges[nonce] = challenge
	s.byEmail[emailKey] = nonce

	return nonce, nil
}

// GetChallenge retrieves an active, unexpired challenge.
func (s *TwofaStore) GetChallenge(nonce string) (*TwofaChallenge, bool) {
	if db := s.getDB(); db != nil {
		var rec models.TwofaChallengeRecord
		err := db.Where(whereNonceActive, nonce, time.Now()).Take(&rec).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("twofa_store: failed to get challenge", "error", err)
			}
			return nil, false
		}
		return recordToChallenge(&rec), true
	}

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
	emailKey := normalizeEmailKey(email)
	if db := s.getDB(); db != nil {
		var rec models.TwofaChallengeRecord
		err := db.Where("email = ? AND expires_at > ?", emailKey, time.Now()).Order("created_at DESC").Take(&rec).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("twofa_store: failed to get challenge by email", "error", err)
			}
			return nil, false
		}
		return recordToChallenge(&rec), true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	nonce, ok := s.byEmail[emailKey]
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
	challenge, ok := s.GetChallengeByEmail(email)
	if !ok {
		return ""
	}
	return challenge.Code
}

// RecordFailedAttempt increments the failed attempt count. If attempts >= 5, removes the challenge.
func (s *TwofaStore) RecordFailedAttempt(nonce string) (attemptsLeft int, locked bool) {
	if db := s.getDB(); db != nil {
		locked = true
		err := db.Transaction(func(tx *gorm.DB) error {
			var rec models.TwofaChallengeRecord
			if err := tx.Where(whereNonce, nonce).Take(&rec).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					attemptsLeft = 0
					locked = true
					return nil
				}
				return err
			}
			rec.Attempts++
			if rec.Attempts >= 5 {
				if err := tx.Where(whereNonce, nonce).Delete(&models.TwofaChallengeRecord{}).Error; err != nil {
					return err
				}
				attemptsLeft = 0
				locked = true
				return nil
			}
			if err := tx.Model(&models.TwofaChallengeRecord{}).Where(whereNonce, nonce).Update("attempts", rec.Attempts).Error; err != nil {
				return err
			}
			attemptsLeft = 5 - rec.Attempts
			locked = false
			return nil
		})
		if err != nil {
			slog.Error("twofa_store: failed to record failed attempt", "error", err)
			return 0, true
		}
		return attemptsLeft, locked
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, ok := s.challenges[nonce]
	if !ok {
		return 0, true
	}
	challenge.Attempts++
	if challenge.Attempts >= 5 {
		delete(s.byEmail, normalizeEmailKey(challenge.Email))
		delete(s.challenges, nonce)
		return 0, true
	}
	return 5 - challenge.Attempts, false
}

// ConsumeChallenge removes and returns the challenge on successful verification.
func (s *TwofaStore) ConsumeChallenge(nonce string) (*TwofaChallenge, bool) {
	if db := s.getDB(); db != nil {
		var consumed *TwofaChallenge
		var ok bool
		err := db.Transaction(func(tx *gorm.DB) error {
			var rec models.TwofaChallengeRecord
			if err := tx.Where(whereNonceActive, nonce, time.Now()).Take(&rec).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return err
			}
			if err := tx.Where(whereNonce, nonce).Delete(&models.TwofaChallengeRecord{}).Error; err != nil {
				return err
			}
			consumed = recordToChallenge(&rec)
			ok = true
			return nil
		})
		if err != nil {
			slog.Error("twofa_store: failed to consume challenge", "error", err)
			return nil, false
		}
		return consumed, ok
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, ok := s.challenges[nonce]
	if !ok || time.Now().After(challenge.ExpiresAt) {
		return nil, false
	}

	delete(s.byEmail, normalizeEmailKey(challenge.Email))
	delete(s.challenges, nonce)
	return challenge, true
}

// UpdateCode updates the expected code (e.g. on OTP resend) and extends expiry.
func (s *TwofaStore) UpdateCode(nonce, newCode string, ttl time.Duration) bool {
	if db := s.getDB(); db != nil {
		now := time.Now()
		res := db.Model(&models.TwofaChallengeRecord{}).
			Where(whereNonceActive, nonce, now).
			Updates(map[string]any{
				"code":       newCode,
				"expires_at": now.Add(ttl),
			})
		if res.Error != nil {
			slog.Error("twofa_store: failed to update code", "error", res.Error)
			return false
		}
		return res.RowsAffected > 0
	}

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
				delete(s.byEmail, normalizeEmailKey(c.Email))
				delete(s.challenges, nonce)
			}
		}
		db := s.db
		s.mu.Unlock()

		if db != nil {
			if err := db.Where("expires_at <= ?", now).Delete(&models.TwofaChallengeRecord{}).Error; err != nil {
				slog.Warn("twofa_store: cleanup expired db rows failed", "error", err)
			}
		}
	}
}
