package hmac

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"testing"
)

func BenchmarkSigner_SignVerifySha1(b *testing.B) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	ctx := context.Background()

	signer := New(WithKey(key), WithHash(sha1.New))

	for range b.N {
		signature, err := signer.Sign(ctx, payload, 0)
		if err != nil {
			panic(err)
		}

		ok, err := signer.Verify(ctx, signature, payload)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("verification failed")
		}
	}
}

func BenchmarkSigner_SignVerifySha256(b *testing.B) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	ctx := context.Background()

	signer := New(WithKey(key), WithHash(sha256.New))

	for range b.N {
		signature, err := signer.Sign(ctx, payload, 0)
		if err != nil {
			panic(err)
		}

		ok, err := signer.Verify(ctx, signature, payload)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("verification failed")
		}
	}
}

func BenchmarkSigner_SignVerifySha384(b *testing.B) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	ctx := context.Background()

	signer := New(WithKey(key), WithHash(sha512.New384))

	for range b.N {
		signature, err := signer.Sign(ctx, payload, 0)
		if err != nil {
			panic(err)
		}

		ok, err := signer.Verify(ctx, signature, payload)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("verification failed")
		}
	}
}

func BenchmarkSigner_SignVerifySha512(b *testing.B) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	ctx := context.Background()

	signer := New(WithKey(key), WithHash(sha512.New))

	for range b.N {
		signature, err := signer.Sign(ctx, payload, 0)
		if err != nil {
			panic(err)
		}

		ok, err := signer.Verify(ctx, signature, payload)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("verification failed")
		}
	}
}

func BenchmarkSigner_SignVerifySha512_224(b *testing.B) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	ctx := context.Background()

	signer := New(WithKey(key), WithHash(sha512.New512_224))

	for range b.N {
		signature, err := signer.Sign(ctx, payload, 0)
		if err != nil {
			panic(err)
		}

		ok, err := signer.Verify(ctx, signature, payload)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("verification failed")
		}
	}
}

func BenchmarkSigner_SignVerifySha512_256(b *testing.B) {
	key := []byte("onvIzKsW6Ec2Q5VqS49zrNlmvrvibh8e")
	payload := []byte("hello, world!")
	ctx := context.Background()

	signer := New(WithKey(key), WithHash(sha512.New512_256))

	for range b.N {
		signature, err := signer.Sign(ctx, payload, 0)
		if err != nil {
			panic(err)
		}

		ok, err := signer.Verify(ctx, signature, payload)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("verification failed")
		}
	}
}
