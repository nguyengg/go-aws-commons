package endec

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// GetSecretValueAPIClient abstracts the Secrets Manager API GetSecretValue which is used by SecretsManagerEndec.
type GetSecretValueAPIClient interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// SecretsManagerEndec is an Endec with key from AWS Secrets Manager.
//
// To make sure the same key that was used during encryption will also be used for decryption, the key's version id will
// be affixed (in plaintext) to the ciphertext in TLV-encoded format.
type SecretsManagerEndec interface {
	// GetSecretBinary returns the secret from AWS Secrets Manager as binary.
	GetSecretBinary(ctx context.Context, versionId *string) ([]byte, *string, error)
	Endec
}

// SecretsManagerEndecOptions customises NewSecretsManagerEndec.
type SecretsManagerEndecOptions struct {
	// Endec controls the encryption/decryption algorithm.
	//
	// By default, [NewChaCha20Poly1305] is used which requires a 256-bit key from AWS Secrets Manager.
	Endec EndecWithKey

	// VersionStage overrides [secretsmanager.GetSecretValueInput.VersionStage].
	VersionStage *string

	// SecretStringDecoder can be used to control how the secret value is decoded into a key.
	//
	// By default, the [secretsmanager.GetSecretValueOutput.SecretBinary] is used as the secret key. If this is not
	// available because the secret was provided as a string instead, this function controls how the
	// [secretsmanager.GetSecretValueOutput.SecretString] is transformed into the secret key. If not given, the
	// default function will cycle through this list of decoders:
	//  1. [base64.RawStdEncoding.DecodeString]
	//  2. [hex.DecodeString]
	//  3. `[]byte(string)`
	SecretStringDecoder func(string) ([]byte, error)
}

// NewSecretsManagerEndec returns a new SecretsManagerEndec.
//
// See SecretsManagerEndecOptions for customisation options.
func NewSecretsManagerEndec(client GetSecretValueAPIClient, secretId string, optFns ...func(*SecretsManagerEndecOptions)) SecretsManagerEndec {
	opts := &SecretsManagerEndecOptions{
		Endec:               NewChaCha20Poly1305(),
		SecretStringDecoder: defaultSecretStringDecoder,
	}
	for _, fn := range optFns {
		fn(opts)
	}

	return &secretsManagerEndec{
		client:              client,
		secretId:            secretId,
		endec:               opts.Endec,
		versionStage:        opts.VersionStage,
		secretStringDecoder: opts.SecretStringDecoder,
	}
}

type secretsManagerEndec struct {
	client              GetSecretValueAPIClient
	secretId            string
	endec               EndecWithKey
	versionStage        *string
	secretStringDecoder func(string) ([]byte, error)
}

func (s secretsManagerEndec) GetSecretBinary(ctx context.Context, versionId *string) ([]byte, *string, error) {
	getSecretValueOutput, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     &s.secretId,
		VersionId:    versionId,
		VersionStage: s.versionStage,
	})
	if err != nil {
		return nil, nil, err
	}

	secretBinary, versionId := getSecretValueOutput.SecretBinary, getSecretValueOutput.VersionId
	if v := getSecretValueOutput.SecretString; v != nil {
		if secretBinary, err = s.secretStringDecoder(*v); err != nil {
			return nil, nil, err
		}
	}

	return secretBinary, versionId, nil
}

func (s secretsManagerEndec) Encode(ctx context.Context, plaintext []byte) ([]byte, error) {
	key, versionId, err := s.GetSecretBinary(ctx, nil)
	if err != nil {
		return nil, err
	}

	ciphertext, err := s.endec.EncodeWithKey(ctx, key, plaintext)
	if err != nil {
		return nil, err
	}

	// returned ciphertext is TLV-encoded.
	// 0x01 is versionId
	// 0x00 indicates payload with arbitrary length. this must be the last component.
	var b bytes.Buffer

	if versionId != nil {
		v := *versionId
		b.WriteByte(0x01)
		b.WriteByte(byte(len(v)))
		b.Write([]byte(v))
	}

	b.WriteByte(0x00)
	b.Write(ciphertext)
	return b.Bytes(), nil
}

func (s secretsManagerEndec) Decode(ctx context.Context, ciphertext []byte) ([]byte, error) {
	var (
		versionId  *string
		code, size byte
		err        error
	)

	// incoming ciphertext is TLV-encoded.
	// 0x01 is versionId
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
				data := make([]byte, size)
				if _, err = b.Read(data); err == nil {
					v := string(data)
					versionId = &v
				}
			}
		}
	}

	switch {
	case err == io.EOF:
		return nil, fmt.Errorf("ciphertext ends too soon")
	case err != nil:
		return nil, err
	}

	key, _, err := s.GetSecretBinary(ctx, versionId)
	if err != nil {
		return nil, err
	}

	return s.endec.DecodeWithKey(ctx, key, ciphertext)
}

func defaultSecretStringDecoder(v string) (data []byte, err error) {
	data, err = base64.RawStdEncoding.DecodeString(v)
	if err != nil {
		data, err = hex.DecodeString(v)
		if err != nil {
			data, err = []byte(v), nil
		}
	}

	return
}
