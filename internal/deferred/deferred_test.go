package deferred

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeferred(t *testing.T) {
	t.Run("Returns excepted value", func(t *testing.T) {
		var d Deferred[string]
		go d.Set("Hello, World!")
		var value, err = d.Get(context.Background())
		assert.Nil(t, err, "expected to get the value, not an error")
		assert.Equal(t, value, "Hello, World!")
	})
	t.Run("Fails with context cancellation", func(t *testing.T) {
		var d Deferred[any]
		var ctx, cancel = context.WithCancel(context.Background())
		go cancel()
		var _, err = d.Get(ctx)
		assert.NotNil(t, err, "expected to get an error due to context cancellation")
	})
}
