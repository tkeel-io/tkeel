package keel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	ok, err := CheckRegisterPluginTkeelVersion("v1.0", "v0.1.0")
	assert.NoError(t, err)
	assert.False(t, ok)
}
