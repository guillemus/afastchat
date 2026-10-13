package openrouter

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const (
	keyBytes   = 32
	keyIDBytes = 8
)

var ErrUnknownSealingKey = errors.New("openrouter key was sealed with a key that is no longer configured")

type sealingKey struct {
	id   []byte
	aead cipher.AEAD
}

type KeyCipher struct {
	current  sealingKey
	previous []sealingKey
}

func NewKeyCipher(encodedKeys string) (*KeyCipher, error) {
	encoded := strings.Fields(strings.ReplaceAll(encodedKeys, ",", " "))
	if len(encoded) == 0 {
		return nil, errors.New("openrouter token key is missing")
	}

	keys := make([]sealingKey, 0, len(encoded))
	for _, value := range encoded {
		key, err := newSealingKey(value)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	return &KeyCipher{current: keys[0], previous: keys[1:]}, nil
}

func newSealingKey(encoded string) (sealingKey, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return sealingKey{}, fmt.Errorf("decode openrouter token key: %w", err)
	}

	if len(raw) != keyBytes {
		return sealingKey{}, fmt.Errorf("openrouter token key must be %d bytes", keyBytes)
	}

	block, err := aes.NewCipher(raw)
	if err != nil {
		return sealingKey{}, fmt.Errorf("create openrouter key cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return sealingKey{}, fmt.Errorf("create openrouter key cipher: %w", err)
	}

	digest := sha256.Sum256(raw)
	return sealingKey{id: digest[:keyIDBytes], aead: aead}, nil
}

func (c *KeyCipher) Seal(plaintext string) []byte {
	nonce := make([]byte, c.current.aead.NonceSize())
	_, _ = rand.Read(nonce)

	sealed := append(bytes.Clone(c.current.id), nonce...)
	return c.current.aead.Seal(sealed, nonce, []byte(plaintext), c.current.id)
}

func (c *KeyCipher) Open(sealed []byte) (string, error) {
	if len(sealed) < keyIDBytes {
		return "", errors.New("openrouter key ciphertext is too short")
	}

	keyID := sealed[:keyIDBytes]
	key, ok := c.key(keyID)
	if !ok {
		return "", ErrUnknownSealingKey
	}

	rest := sealed[keyIDBytes:]
	if len(rest) < key.aead.NonceSize() {
		return "", errors.New("openrouter key ciphertext is too short")
	}

	nonce, ciphertext := rest[:key.aead.NonceSize()], rest[key.aead.NonceSize():]
	plaintext, err := key.aead.Open(nil, nonce, ciphertext, keyID)
	if err != nil {
		return "", fmt.Errorf("open openrouter key: %w", err)
	}

	return string(plaintext), nil
}

func (c *KeyCipher) key(id []byte) (sealingKey, bool) {
	if bytes.Equal(id, c.current.id) {
		return c.current, true
	}

	for _, key := range c.previous {
		if bytes.Equal(id, key.id) {
			return key, true
		}
	}

	return sealingKey{}, false
}
