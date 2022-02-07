package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommonGetQueryItemsStartAndEnd(t *testing.T) {
	s, e := getQueryItemsStartAndEnd(1, 10, 23)
	assert.Equal(t, s, 0)
	assert.Equal(t, e, 9)
}
