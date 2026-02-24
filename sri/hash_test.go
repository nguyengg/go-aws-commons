package sri

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	data := []byte("hello, world!")

	tests := []struct {
		name        string
		fn          func() Hash
		expected    string
		expectedHex string
	}{
		{
			name:        "sha1",
			fn:          NewSha1,
			expected:    "sha1-HwnTDHB9U/PRbFMN1z1wps51lqk",
			expectedHex: "sha1:1f09d30c707d53f3d16c530dd73d70a6ce7596a9",
		},
		{
			name:        "sha224",
			fn:          NewSha224,
			expected:    "sha224-VE4THAfnVPg51auGzUqG0CViujIkCu/DTJpDAg",
			expectedHex: "sha224:544e131c07e754f839d5ab86cd4a86d02562ba32240aefc34c9a4302",
		},
		{
			name:        "sha256",
			fn:          NewSha256,
			expected:    "sha256-aOZWslHmfoNYvvhIOrDVHGYZ8+ehqfDnWDjUH/No9yg",
			expectedHex: "sha256:68e656b251e67e8358bef8483ab0d51c6619f3e7a1a9f0e75838d41ff368f728",
		},
		{
			name:        "sha384",
			fn:          NewSha384,
			expected:    "sha384-b58jhCXsokOe1Fgawf20X8djeef7qUvAp2JPo+erHsNwG0v83aN2ynVRkub0XypO",
			expectedHex: "sha384:6f9f238425eca2439ed4581ac1fdb45fc76379e7fba94bc0a7624fa3e7ab1ec3701b4bfcdda376ca755192e6f45f2a4e",
		},
		{
			name:        "sha512",
			fn:          NewSha512,
			expected:    "sha512-bCYYNY2gfIMLiMWvjDU1CA6OYDyIuJECiiWczbmsgC0PwBcMmdWK/88AeGzhiPxddT6MZiivIHHDJw1QRFxLHA",
			expectedHex: "sha512:6c2618358da07c830b88c5af8c3535080e8e603c88b891028a259ccdb9ac802d0fc0170c99d58affcf00786ce188fc5d753e8c6628af2071c3270d50445c4b1c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.fn()
			assert.Equal(t, tt.name, h.Name())

			_, _ = h.Write(data)
			assert.Equal(t, tt.expected, h.SumToString(nil))
			assert.Equal(t, tt.expectedHex, h.SumToCustomString(nil, ":", hex.EncodeToString))
		})
	}
}
