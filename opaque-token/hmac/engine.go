package hmac

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/subtle"
	"fmt"
	"hash"
	"io"

	"github.com/nguyengg/go-aws-commons/opaque-token/keys"
)

// engine implements [Engine].
type engine struct {
	keyProvider  keys.Provider
	hashProvider func() hash.Hash
	rand         func([]byte) error
}

var _ Engine = engine{}

func (e engine) Sign(ctx context.Context, payload []byte, nonceSize byte) ([]byte, error) {
	var nonce []byte
	if nonceSize > 0 {
		nonce = make([]byte, nonceSize)
		if err := e.rand(nonce); err != nil {
			return nil, err
		}

	}

	return e.hash(ctx, bytes.NewReader(payload), nonce, nil)
}

func (e engine) hash(ctx context.Context, payload io.Reader, nonce []byte, versionId *string) ([]byte, error) {
	secret, versionId, err := e.keyProvider.Provide(ctx, versionId)
	if err != nil {
		return nil, err
	}

	// token is essentially TLV (https://en.wikipedia.org/wiki/Type%E2%80%93length%E2%80%93value):
	// 0x01 is versionId
	// 0x02 is nonce
	// 0x00 indicates payload with arbitrary length. this must be the last component.
	var b bytes.Buffer

	if versionId != nil {
		v := *versionId
		b.WriteByte(0x01)
		b.WriteByte(byte(len(v)))
		b.Write([]byte(v))
	}

	if n := len(nonce); n > 0 {
		b.WriteByte(0x02)
		b.WriteByte(byte(n))
		b.Write(nonce)
	}

	w := hmac.New(e.hashProvider, secret)
	if _, err = io.Copy(w, payload); err != nil {
		return nil, err
	}
	if _, err = w.Write(nonce); err != nil {
		return nil, err
	}

	b.WriteByte(0x00)
	b.Write(w.Sum(nil))
	return b.Bytes(), nil
}

func (e engine) Verify(ctx context.Context, signature, payload []byte) (ok bool, err error) {
	secret, nonce, expected, err := e.unpack(ctx, signature)
	if err != nil {
		return false, err
	}

	w := hmac.New(e.hashProvider, secret)
	w.Write(payload)
	w.Write(nonce)
	actual := w.Sum(nil)
	return subtle.ConstantTimeCompare(expected, actual) == 1, nil
}

func (e engine) unpack(ctx context.Context, rawPayload []byte) (key, nonce, payload []byte, err error) {
	var (
		versionId  *string
		code, size byte
	)

	// token is TLV-encoded (https://en.wikipedia.org/wiki/Type%E2%80%93length%E2%80%93value):
	// 0x01 is versionId
	// 0x02 is nonce
	// 0x00 indicates payload with all remaining bytes; this must be the last component.

ingBad:
	for b := bytes.NewBuffer(rawPayload); err == nil; {
		if code, err = b.ReadByte(); err != nil {
			break
		}

		switch code {
		case 0x00:
			payload = b.Bytes()
			break ingBad
		case 0x01:
			if size, err = b.ReadByte(); err == nil {
				data := make([]byte, size)
				if _, err = b.Read(data); err == nil {
					v := string(data)
					versionId = &v
				}
			}
		case 0x02:
			if size, err = b.ReadByte(); err == nil {
				nonce = make([]byte, size)
				_, err = b.Read(nonce)
			}
		}
	}

	switch {
	case err == io.EOF:
		err = fmt.Errorf("token ends too soon")
	case err != nil:
	default:
		key, _, err = e.keyProvider.Provide(ctx, versionId)
	}

	return
}
