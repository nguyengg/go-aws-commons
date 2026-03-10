package ddb

import "encoding/hex"

// Encoding abstracts [base64.Encoding] and [HexEncoding].
type Encoding interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}

// HexEncoding implements [Encoding] using [hex.EncodeToString] and [hex.DecodeString].
type HexEncoding struct {
}

func (h HexEncoding) EncodeToString(src []byte) string {
	return hex.EncodeToString(src)
}

func (h HexEncoding) DecodeString(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
