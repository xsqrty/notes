package jwtsafe

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrJWTKeysUnavailable = errors.New("jwt keys unavailable")
	ErrJWTCreateToken     = errors.New("token could not be created")
	ErrJWTInvalid         = errors.New("token is invalid")
	ErrJWTExpired         = errors.New("token is expired")
)

var (
	// hs265Alg defines the JSON header message encoding algorithm and type for HS256 JWT.
	hs265Alg = []byte(`{"alg":"HS256","typ":"JWT"}`)
	// hs265AlgHeader is the base64 URL-encoded representation of hs265Alg.
	hs265AlgHeader = base64.RawURLEncoding.EncodeToString(hs265Alg)
)

// JWTSafe is an interface for encoding and decoding JWTs securely.
type JWTSafe interface {
	Encode(claims MapClaims) (string, error)
	Decode(token string) (MapClaims, error)
	Close() error
}

// MapClaims represents a map of string keys to interface{} values, commonly used to define JWT claims.
type MapClaims map[string]interface{}

// keyPair represents a structure holding the current and previous cryptographic keys as byte slices.
type keyPair struct {
	cur  []byte
	prev []byte
}

// jwtSafe manages JSON Web Token (JWT) operations with rotating secret keys for enhanced security.
// It periodically rotates the secret key, encodes claims to tokens, decodes tokens, and validates signatures.
type jwtSafe struct {
	ticker     *time.Ticker
	keyPair    atomic.Pointer[keyPair]
	secretSize int
	expires    time.Duration
	closeOnce  sync.Once
}

// New creates and returns a new jwtSafe instance initialized with the specified expiration duration and secret size.
func New(expires time.Duration, secretSize int) *jwtSafe {
	js := &jwtSafe{
		secretSize: secretSize,
		expires:    expires,
		ticker:     time.NewTicker(expires),
	}

	js.rotateKey()
	go func() {
		for range js.ticker.C {
			js.rotateKey()
		}
	}()

	return js
}

// Encode generates a signed JWT token from the provided claims and returns it as a string or an error if creation fails.
func (js *jwtSafe) Encode(claims MapClaims) (string, error) {
	claims["exp"] = time.Now().Add(js.expires).Unix()

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", ErrJWTCreateToken
	}

	pair := js.keyPair.Load()
	if pair == nil {
		return "", ErrJWTKeysUnavailable
	}

	unsigned := hs265AlgHeader + "." + base64.RawURLEncoding.EncodeToString(payload)
	return unsigned + "." + js.sign(unsigned, pair.cur), nil
}

// Decode validates and decodes a JWT token, returning its claims or an error if the token is invalid or expired.
func (js *jwtSafe) Decode(token string) (MapClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrJWTInvalid
	}

	pair := js.keyPair.Load()
	if pair == nil {
		return nil, ErrJWTKeysUnavailable
	}

	keys := [2][]byte{pair.cur, pair.prev}
	if !js.verify(parts[0]+"."+parts[1], parts[2], keys) {
		return nil, ErrJWTInvalid
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrJWTInvalid
	}

	var claims MapClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, ErrJWTInvalid
	}

	if exp, ok := claims["exp"].(int64); ok {
		if time.Now().Unix() > exp {
			return nil, ErrJWTExpired
		}
	}

	return claims, nil
}

// Close stops the ticker. It ensures the operation runs only once.
func (js *jwtSafe) Close() error {
	js.closeOnce.Do(func() {
		js.ticker.Stop()
	})

	return nil
}

// sign generates a base64-encoded HMAC-SHA256 hash of the provided unsigned string using the specified key.
func (js *jwtSafe) sign(unsigned string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// Verify checks if the provided signature matches any of the HMAC signatures generated using the given keys.
// Returns true if a match is found, otherwise returns false.
func (js *jwtSafe) verify(unsigned string, signature string, keys [2][]byte) bool {
	for _, key := range keys {
		if key == nil {
			continue
		}

		if hmac.Equal([]byte(signature), []byte(js.sign(unsigned, key))) {
			return true
		}
	}

	return false
}

// rotateKey generates a new cryptographic key and updates the current and previous key pair for signing and verification.
func (js *jwtSafe) rotateKey() {
	key := make([]byte, js.secretSize)
	if _, err := rand.Read(key); err != nil {
		panic(fmt.Sprintf("jwtSafe: failed to generate random key: %v", err))
	}

	newPair := &keyPair{
		cur: key,
	}

	pair := js.keyPair.Load()
	if pair != nil {
		newPair.prev = pair.cur
	}

	js.keyPair.Store(newPair)
}
