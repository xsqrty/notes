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
	hs265Alg       = []byte(`{"alg":"HS256","typ":"JWT"}`)
	hs265AlgHeader = base64.RawURLEncoding.EncodeToString(hs265Alg)
)

type JWTSafe interface {
	Encode(claims MapClaims) (string, error)
	Decode(token string) (MapClaims, error)
	Close() error
}

type MapClaims map[string]interface{}

type keyPair struct {
	cur  []byte
	prev []byte
}

type jwtSafe struct {
	ticker     *time.Ticker
	keyPair    atomic.Pointer[keyPair]
	secretSize int
	expires    time.Duration
	done       chan struct{}
}

func New(expires time.Duration, secretSize int) *jwtSafe {
	js := &jwtSafe{
		secretSize: secretSize,
		expires:    expires,
		ticker:     time.NewTicker(expires),
		done:       make(chan struct{}),
	}

	js.rotateKey()

	go func() {
		for {
			select {
			case <-js.ticker.C:
				js.rotateKey()
			case <-js.done:
				return
			}
		}
	}()

	return js
}

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

func (js *jwtSafe) Close() error {
	js.ticker.Stop()
	close(js.done)
	return nil
}

func (js *jwtSafe) sign(unsigned string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

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
