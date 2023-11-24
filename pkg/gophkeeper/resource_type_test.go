package gophkeeper_test

import (
	"testing"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/stretchr/testify/assert"
)

func TestResourceType(t *testing.T) {
	t.Run("Piece", func(t *testing.T) {
		assert.Equal(
			t,
			"Piece",
			gophkeeper.ResourceTypePiece.String(),
			"invalid resource type string representation",
		)
	})
	t.Run("Blob", func(t *testing.T) {
		assert.Equal(
			t,
			"Blob",
			gophkeeper.ResourceTypeBlob.String(),
			"invalid resource type string representation",
		)
	})
	t.Run("invalid", func(t *testing.T) {
		assert.Panics(
			t,
			func() {
				_ = (gophkeeper.ResourceType)(-1).String()
			},
			"expected to panic",
		)
	})
}
