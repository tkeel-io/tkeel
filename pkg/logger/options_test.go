package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		o := DefaultOptions()
		assert.Equal(t, defaultJSONOutput, o.JSONFormatEnabled)
		assert.Equal(t, undefinedID, o.ID)
		assert.Equal(t, defaultOutputLevel, o.OutputLevel)
	})

	t.Run("set plugin ID", func(t *testing.T) {
		o := DefaultOptions()
		assert.Equal(t, undefinedID, o.ID)

		o.SetID("plugin-test")
		assert.Equal(t, "plugin-test", o.ID)
	})

	t.Run("set output level", func(t *testing.T) {
		o := DefaultOptions()
		assert.Equal(t, defaultOutputLevel, o.OutputLevel)

		o.SetOutputLevel("debug")
		assert.Equal(t, "debug", o.OutputLevel)
	})

	t.Run("set undefined output level", func(t *testing.T) {
		o := DefaultOptions()
		assert.Equal(t, defaultOutputLevel, o.OutputLevel)

		o.SetOutputLevel("high")
		assert.Equal(t, defaultOutputLevel, o.OutputLevel)
	})

	t.Run("attaching log related cmd flags", func(t *testing.T) {
		o := DefaultOptions()

		logLevelAsserted := false
		testStringVarFn := func(p *string, name string, value string, usage string) {
			if name == "log-level" && value == defaultOutputLevel {
				logLevelAsserted = true
			}
		}

		logAsJSONAsserted := false
		testBoolVarFn := func(p *bool, name string, value bool, usage string) {
			if name == "log-as-json" && value == defaultJSONOutput {
				logAsJSONAsserted = true
			}
		}

		o.AttachCmdFlags(testStringVarFn, testBoolVarFn)

		// assert.
		assert.True(t, logLevelAsserted)
		assert.True(t, logAsJSONAsserted)
	})
}

func TestApplyOptionsToLoggers(t *testing.T) {
	testOptions := Options{
		JSONFormatEnabled: true,
		ID:                "plugin-app",
		OutputLevel:       "debug",
	}

	// Create two loggers.
	testLoggers := []Logger{
		NewLogger("testLogger0"),
		NewLogger("testLogger1"),
	}

	for _, l := range testLoggers {
		l.EnableJSONOutput(false)
		l.SetOutputLevel(InfoLevel)
	}

	assert.NoError(t, ApplyOptionsToLoggers(&testOptions))

	for _, l := range testLoggers {
		assert.Equal(
			t,
			"plugin-app",
			(l.(*pluginLogger)).logger.Data[logFieldID])
		assert.Equal(
			t,
			toLogrusLevel(DebugLevel),
			(l.(*pluginLogger)).logger.Logger.GetLevel())
	}
}
