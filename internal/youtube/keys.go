package youtube

import (
	"log"
	"sync"
	"time"
)

const (
	maxErrors    = 3                // errores antes de bloquear
	cooldownTime = 15 * time.Minute // tiempo bloqueada
)

type apiKey struct {
	Value        string
	ErrorCount   int
	BlockedUntil time.Time
}

type KeyManager struct {
	keys []apiKey
	mu   sync.Mutex
	idx  int
}

// CONSTRUCTOR

func NewKeyManager(keys []string) *KeyManager {
	if len(keys) == 0 {
		log.Fatal("[KEYS] no API keys provided")
	}

	k := make([]apiKey, len(keys))

	for i, key := range keys {
		k[i] = apiKey{
			Value: key,
		}
	}

	return &KeyManager{
		keys: k,
	}
}

// GET NEXT KEY (ROUND ROBIN)

func (km *KeyManager) NextKey() string {
	km.mu.Lock()
	defer km.mu.Unlock()

	now := time.Now()

	// loop keys
	for i := 0; i < len(km.keys); i++ {

		km.idx = (km.idx + 1) % len(km.keys)
		key := &km.keys[km.idx]

		// if key is blocked, skip
		if key.BlockedUntil.After(now) {
			continue
		}

		return key.Value
	}

	// 🔥 fallback: all keys are blocked
	log.Println("[KEYS] all keys are blocked, using fallback")

	return km.keys[0].Value
}

// MARK ERROR

func (km *KeyManager) MarkError(keyValue string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	for i := range km.keys {

		if km.keys[i].Value != keyValue {
			continue
		}

		km.keys[i].ErrorCount++

		log.Printf("[KEYS] error on key (count=%d)", km.keys[i].ErrorCount)

		if km.keys[i].ErrorCount >= maxErrors {

			km.keys[i].BlockedUntil = time.Now().Add(cooldownTime)
			km.keys[i].ErrorCount = 0

			log.Printf("[KEYS] key blocked for %s", cooldownTime)
		}

		return
	}
}

// MARK SUCCESS

func (km *KeyManager) MarkSuccess(keyValue string) {
	km.mu.Lock()
	defer km.mu.Unlock()

	for i := range km.keys {

		if km.keys[i].Value != keyValue {
			continue
		}

		// reset error count
		km.keys[i].ErrorCount = 0

		return
	}
}

// DEBUG (OPTIONAL)

func (km *KeyManager) Stats() {
	km.mu.Lock()
	defer km.mu.Unlock()

	for i, k := range km.keys {
		log.Printf("[KEY %d] errors=%d blockedUntil=%v",
			i,
			k.ErrorCount,
			k.BlockedUntil,
		)
	}
}

// Hey count
func (km *KeyManager) Count() int {
	km.mu.Lock()
	defer km.mu.Unlock()
	return len(km.keys)
}
