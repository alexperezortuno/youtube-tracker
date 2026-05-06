package youtube

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

const (
	MaxErrors      = 3
	CooldownTime   = 15 * time.Minute
	MaxUnitsPerSec = 10
	MaxRequests    = 50
)

var (
	ErrQuotaExceeded  = errors.New("quota exceeded")
	ErrAllKeysBlocked = errors.New("all API keys are blocked")
)

type apiKey struct {
	Value        string
	ErrorCount   int
	BlockedUntil time.Time
	UnitsUsed    int
	RequestsUsed int
	LastReset    time.Time
}

type KeyManager struct {
	keys        []apiKey
	mu          sync.Mutex
	idx         int
	unitsPerSec int
	maxRequests int
	logger      *slog.Logger
}

func NewKeyManager(keys []string) *KeyManager {
	return NewKeyManagerWithConfig(keys, MaxUnitsPerSec, MaxRequests, nil)
}

func NewKeyManagerWithConfig(keys []string, unitsPerSec, maxRequests int, logger *slog.Logger) *KeyManager {
	if len(keys) == 0 {
		panic("no API keys provided")
	}

	k := make([]apiKey, len(keys))
	for i, key := range keys {
		k[i] = apiKey{
			Value:     key,
			LastReset: time.Now(),
		}
	}

	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(&nullWriter{}, nil))
	}

	return &KeyManager{
		keys:        k,
		unitsPerSec: unitsPerSec,
		maxRequests: maxRequests,
		logger:      logger,
	}
}

type nullWriter struct{}

func (n *nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (km *KeyManager) nextKey() *apiKey {
	now := time.Now()

	for i := 0; i < len(km.keys); i++ {
		km.idx = (km.idx + 1) % len(km.keys)
		key := &km.keys[km.idx]

		if key.BlockedUntil.After(now) {
			continue
		}

		if now.Sub(key.LastReset) > time.Second {
			key.UnitsUsed = 0
			key.RequestsUsed = 0
			key.LastReset = now
		}

		return key
	}

	return nil
}

func (km *KeyManager) NextKey() (string, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	key := km.nextKey()
	if key != nil {
		return key.Value, nil
	}

	km.logger.Error("all API keys are blocked")
	return "", ErrAllKeysBlocked
}

func (km *KeyManager) Take(keyValue string, units, requests int) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	for i := range km.keys {
		if km.keys[i].Value != keyValue {
			continue
		}

		now := time.Now()
		if now.Sub(km.keys[i].LastReset) > time.Second {
			km.keys[i].UnitsUsed = 0
			km.keys[i].RequestsUsed = 0
			km.keys[i].LastReset = now
		}

		km.keys[i].UnitsUsed += units
		km.keys[i].RequestsUsed += requests

		km.logger.Debug("API key quota used",
			"key_idx", i,
			"units", km.keys[i].UnitsUsed,
			"requests", km.keys[i].RequestsUsed,
			"limit_units", km.unitsPerSec,
			"limit_requests", km.maxRequests,
		)

		if km.keys[i].UnitsUsed >= km.unitsPerSec || km.keys[i].RequestsUsed >= km.maxRequests {
			key := &km.keys[i]
			key.BlockedUntil = now.Add(time.Second - now.Sub(key.LastReset))
			km.logger.Warn("API key rate limit reached",
				"key_idx", i,
				"blocked_until", key.BlockedUntil,
			)
		}

		return nil
	}

	return errors.New("key not found")
}

func (km *KeyManager) CheckAvailable(keyValue string) bool {
	km.mu.Lock()
	defer km.mu.Unlock()

	for i := range km.keys {
		if km.keys[i].Value != keyValue {
			continue
		}

		key := &km.keys[i]
		now := time.Now()

		if now.Sub(key.LastReset) > time.Second {
			key.UnitsUsed = 0
			key.RequestsUsed = 0
			key.LastReset = now
			return true
		}

		return key.UnitsUsed < km.unitsPerSec && key.RequestsUsed < km.maxRequests
	}

	return false
}

func (km *KeyManager) MarkError(keyValue string, statusCode int) {
	km.mu.Lock()
	defer km.mu.Unlock()

	for i := range km.keys {
		if km.keys[i].Value != keyValue {
			continue
		}

		km.keys[i].ErrorCount++

		if statusCode == 403 || statusCode == 429 || km.keys[i].ErrorCount >= MaxErrors {
			km.keys[i].BlockedUntil = time.Now().Add(CooldownTime)
			km.keys[i].ErrorCount = 0
			km.logger.Warn("API key blocked",
				"key_idx", i,
				"reason", map[int]string{
					403: "quota exceeded",
					429: "rate limited",
				}[statusCode],
				"blocked_until", km.keys[i].BlockedUntil,
			)
			return
		}

		km.logger.Warn("API key error",
			"key_idx", i,
			"error_count", km.keys[i].ErrorCount,
			"status_code", statusCode,
		)
		return
	}
}

func (km *KeyManager) MarkSuccess(keyValue string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	for i := range km.keys {
		if km.keys[i].Value != keyValue {
			continue
		}

		km.keys[i].ErrorCount = 0
		km.logger.Debug("API key success", "key_idx", i)
		return
	}
}

func (km *KeyManager) Stats() []KeyStats {
	km.mu.Lock()
	defer km.mu.Unlock()

	stats := make([]KeyStats, len(km.keys))
	now := time.Now()

	for i, k := range km.keys {
		stats[i] = KeyStats{
			Idx:          i,
			ErrorCount:   k.ErrorCount,
			BlockedUntil: k.BlockedUntil,
			UnitsUsed:    k.UnitsUsed,
			RequestsUsed: k.RequestsUsed,
			IsBlocked:    k.BlockedUntil.After(now),
			IsAtLimit:    k.UnitsUsed >= km.unitsPerSec || k.RequestsUsed >= km.maxRequests,
		}
	}

	return stats
}

type KeyStats struct {
	Idx          int
	ErrorCount   int
	BlockedUntil time.Time
	UnitsUsed    int
	RequestsUsed int
	IsBlocked    bool
	IsAtLimit    bool
}

func (km *KeyManager) Count() int {
	km.mu.Lock()
	defer km.mu.Unlock()
	return len(km.keys)
}

func (km *KeyManager) UnitsPerSec() int {
	km.mu.Lock()
	defer km.mu.Unlock()
	return km.unitsPerSec
}

func (km *KeyManager) MaxRequests() int {
	km.mu.Lock()
	defer km.mu.Unlock()
	return km.maxRequests
}
