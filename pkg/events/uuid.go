package events

import (
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // SHA-1 solo para derivar ids determinísticos (UUID v5), no como función criptográfica
	"encoding/hex"
	"fmt"
)

// NewUUID genera un UUID v4 aleatorio (RFC 4122).
func NewUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand no falla en las plataformas soportadas; si pasara,
		// no hay forma segura de seguir generando ids.
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return formatUUID(b)
}

// NameUUID genera un UUID v5 determinístico a partir de un texto: el
// mismo texto da siempre el mismo id. Se usa cuando un mismo hecho
// puede llegar repetido al productor y queremos que el consumidor lo
// descarte como duplicado.
func NameUUID(name string) string {
	// Namespace propio del sistema (UUID v4 fijo, generado una vez).
	namespace := [16]byte{0x6b, 0x1c, 0x2f, 0x4a, 0x93, 0x0e, 0x4d, 0x58, 0x8a, 0x3e, 0x51, 0x77, 0xc2, 0x09, 0xd4, 0x6f}
	h := sha1.New() //nolint:gosec // ver import
	h.Write(namespace[:])
	h.Write([]byte(name))
	sum := h.Sum(nil)
	var b [16]byte
	copy(b[:], sum[:16])
	b[6] = (b[6] & 0x0f) | 0x50
	b[8] = (b[8] & 0x3f) | 0x80
	return formatUUID(b)
}

// IsUUID dice si s tiene la forma canónica de un UUID
// (8-4-4-4-12 dígitos hexadecimales).
func IsUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !isHex(c) {
				return false
			}
		}
	}
	return true
}

func isHex(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func formatUUID(b [16]byte) string {
	var buf [36]byte
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], b[10:])
	return string(buf[:])
}
