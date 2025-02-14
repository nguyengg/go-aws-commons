package endec

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// GetParameterAPIClient abstracts the AWS Systems Manager API GetParameter which is used by ParameterStoreEndec.
type GetParameterAPIClient interface {
	GetParameter(context.Context, *ssm.GetParameterInput, ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// ParameterStoreEndec is an Endec with key from AWS Systems Manager's Parameter Store.
//
// To make sure the same key that was used during encryption will also be used for decryption, the key's version will be
// affixed (in plaintext) to the ciphertext in TLV-encoded format.
type ParameterStoreEndec interface {
	// GetValueBinary returns the parameter value from AWS Systems Manager as binary.
	GetValueBinary(ctx context.Context, version int64) ([]byte, int64, error)
	Endec
}

// ParameterStoreEndecOptions customises NewParameterStoreEndec.
type ParameterStoreEndecOptions struct {
	// Endec controls the encryption/decryption algorithm.
	//
	// By default, [NewChaCha20Poly1305] is used which requires a 256-bit key from AWS Secrets Manager.
	Endec EndecWithKey

	// WithDecryption overrides [ssm.GetParameterInput.WithDecryption].
	WithDecryption *bool

	// Label suffixes the label [ssm.GetParameterInput.Name].
	//
	// If version is available, the label will not be suffixed since Parameter Store (unlike Secrets Manager) only
	// allows specifying one or the other.
	Label *string

	// ValueDecoder can be used to control how the parameter value is decoded into a key.
	//
	// If not given, the default function will cycle through this list of decoders:
	//  1. [base64.RawStdEncoding.DecodeString]
	//  2. [hex.DecodeString]
	//  3. `[]byte(string)`
	ValueDecoder func(string) ([]byte, error)
}

// NewParameterStoreEndec returns a new SecretsManagerEndec.
//
// See ParameterStoreEndecOptions for customisation options.
func NewParameterStoreEndec(client GetParameterAPIClient, name string, optFns ...func(*ParameterStoreEndecOptions)) ParameterStoreEndec {
	opts := &ParameterStoreEndecOptions{
		Endec:        NewChaCha20Poly1305(),
		ValueDecoder: defaultSecretStringDecoder,
	}
	for _, fn := range optFns {
		fn(opts)
	}

	return &parameterStoreEndec{
		client:         client,
		name:           name,
		endec:          opts.Endec,
		withDecryption: opts.WithDecryption,
		label:          opts.Label,
		valueDecoder:   opts.ValueDecoder,
	}
}

type parameterStoreEndec struct {
	client         GetParameterAPIClient
	name           string
	endec          EndecWithKey
	withDecryption *bool
	label          *string
	valueDecoder   func(string) ([]byte, error)
}

func (p parameterStoreEndec) GetValueBinary(ctx context.Context, version int64) ([]byte, int64, error) {
	var name string
	if version != 0 {
		name = fmt.Sprintf("%s:%d", p.name, version)
	} else if p.label != nil {
		name = fmt.Sprintf("%s:%s", p.name, *p.label)
	} else {
		name = p.name
	}

	getParameterOutput, err := p.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: p.withDecryption,
	})
	if err != nil {
		return nil, 0, err
	}

	param := getParameterOutput.Parameter
	key, err := p.valueDecoder(aws.ToString(param.Value))
	return key, param.Version, nil
}

func (p parameterStoreEndec) Encode(ctx context.Context, plaintext []byte) ([]byte, error) {
	key, version, err := p.GetValueBinary(ctx, 0)
	if err != nil {
		return nil, err
	}

	ciphertext, err := p.endec.EncodeWithKey(ctx, key, plaintext)
	if err != nil {
		return nil, err
	}

	// returned ciphertext is TLV-encoded.
	// 0x01 is version which is always 8 byte long.
	// 0x00 indicates payload with arbitrary length. this must be the last component.
	var b bytes.Buffer

	if version != 0 {
		b.WriteByte(0x01)
		b.WriteByte(byte(8))
		if err = binary.Write(&b, binary.LittleEndian, version); err != nil {
			return nil, err
		}
	}

	b.WriteByte(0x00)
	b.Write(ciphertext)
	return b.Bytes(), nil
}

func (p parameterStoreEndec) Decode(ctx context.Context, ciphertext []byte) ([]byte, error) {
	var (
		version    int64
		code, size byte
		err        error
	)

	// incoming ciphertext is TLV-encoded.
	// 0x01 is version which is always 8 byte long.
	// 0x00 indicates payload with all remaining bytes; this must be the last component.

ingBad:
	for b := bytes.NewBuffer(ciphertext); err == nil; {
		if code, err = b.ReadByte(); err != nil {
			break
		}

		switch code {
		case 0x00:
			ciphertext = b.Bytes()
			break ingBad
		case 0x01:
			if size, err = b.ReadByte(); err == nil {
				if size != 8 {
					return nil, fmt.Errorf("invalid version format")
				}
				err = binary.Read(b, binary.LittleEndian, &version)
			}
		}
	}

	switch {
	case err == io.EOF:
		return nil, fmt.Errorf("ciphertext ends too soon")
	case err != nil:
		return nil, err
	}

	key, _, err := p.GetValueBinary(ctx, version)
	if err != nil {
		return nil, err
	}

	return p.endec.DecodeWithKey(ctx, key, ciphertext)
}
