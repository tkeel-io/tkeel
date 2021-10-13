package plugins

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tkeel/pkg/plugin"
)

var pp *plugin.Plugin

func TestMain(m *testing.M) {
	os.Setenv("PLUGIN_ID", "keel-manager")
	pp, _ = plugin.FromFlags()
	m.Run()
	os.Exit(0)
}

func TestNewManager(t *testing.T) {
	t.Run("test load default configuration", func(t *testing.T) {
		// act.
		_, err := New(pp)
		// assert.
		assert.NoError(t, err)
	})
}

func TestManagerRun(t *testing.T) {
	t.Run("test manager run", func(t *testing.T) {
		p, err := New(pp)
		assert.NoError(t, err)
		p.Run()
		st, err := time.ParseDuration("30s")
		assert.NoError(t, err)
		time.Sleep(st)
	})
}
