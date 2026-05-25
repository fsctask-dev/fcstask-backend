package app

import (
	"fcstask-backend/internal/controller"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtectedPaths(t *testing.T) {
	assert.Equal(t, controller.ProtectedPaths(), protectedPaths())
}
