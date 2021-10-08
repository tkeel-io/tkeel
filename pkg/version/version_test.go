package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Run("test version", func(t *testing.T) {
		// act
		ver := Version()
		// assert
		assert.Equal(t, ver, "edge")
	})
}

func TestCommit(t *testing.T) {
	t.Run("test git commit", func(t *testing.T) {
		// act
		hash := Commit()
		// assert
		assert.Equal(t, hash, "")
	})
}

func TestGitVersion(t *testing.T) {
	t.Run("test Git version", func(t *testing.T) {
		// act
		ver := GitVersion()
		// assert
		assert.Equal(t, ver, "")
	})
}
