package postgres

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithBlobsWith(t *testing.T) {
	g := New(nil, nil, WithBlobsDir("test"))
	assert.Equal(t, "test", g.blobsDir, "blobs dirs do not match")
}

func TestWithPasswordEncoding(t *testing.T) {
	encoding := *base64.RawURLEncoding
	g := New(nil, nil, WithPasswordEncoding(&encoding))
	assert.Equal(t, encoding, g.passwordEncoding, "password encodings do not match")

	t.Run("Panics with nil encoding", func(t *testing.T) {
		assert.Panics(
			t,
			func() {
				_ = New(nil, nil, WithPasswordEncoding(nil))
			},
		)
	})
}
