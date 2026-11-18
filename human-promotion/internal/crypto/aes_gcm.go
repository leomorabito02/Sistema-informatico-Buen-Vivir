package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// KeySize es el largo de clave que exige AES-256.
const KeySize = 32

// ciphertextPrefix marca la versión del formato guardado en Firestore
// ("v1:" + base64(nonce || ciphertext || tag)). Si algún día se rota la
// clave o cambia el algoritmo, el prefijo permite distinguir qué
// registros quedan por re-cifrar.
const ciphertextPrefix = "v1:"

// ErrInvalidCiphertext señala un valor guardado que no tiene el formato
// esperado o que no pasa la verificación de autenticidad de GCM (dato
// alterado o cifrado con otra clave).
var ErrInvalidCiphertext = errors.New("dato cifrado inválido")

// FieldCipher cifra y descifra los campos confidenciales de la ficha y
// de los hitos antes de persistirlos (cifrado a nivel de aplicación:
// Firestore nunca ve el texto plano).
type FieldCipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(stored string) (string, error)
}

type aesGCMCipher struct {
	aead cipher.AEAD
}

// NewAESGCM arma el cifrador AES-256-GCM a partir de una clave de 32 bytes.
func NewAESGCM(key []byte) (FieldCipher, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("la clave de cifrado debe tener %d bytes, tiene %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &aesGCMCipher{aead: aead}, nil
}

// Encrypt usa un nonce aleatorio por llamada (nunca se reutiliza con la
// misma clave), que viaja delante del ciphertext.
func (c *aesGCMCipher) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("no se pudo generar el nonce: %w", err)
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertextPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func (c *aesGCMCipher) Decrypt(stored string) (string, error) {
	encoded, ok := strings.CutPrefix(stored, ciphertextPrefix)
	if !ok {
		return "", ErrInvalidCiphertext
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	nonceSize := c.aead.NonceSize()
	if len(raw) < nonceSize {
		return "", ErrInvalidCiphertext
	}
	plaintext, err := c.aead.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	return string(plaintext), nil
}
