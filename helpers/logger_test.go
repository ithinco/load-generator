package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewLogger(t *testing.T) {
	assert.NotNil(t, Logger)
}
