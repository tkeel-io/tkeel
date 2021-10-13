package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaultConfiguration(t *testing.T) {
	t.Run("test load default configuration", func(t *testing.T) {
		// act.
		conf := LoadDefaultConfiguration()
		// assert.
		assert.Equal(t, conf.Plugin.Port, DefaultHTTPPort)
		assert.Equal(t, conf.Plugin.ID, "")
		assert.Equal(t, conf.Plugin.Version, "")
	})
}

func TestLoadStandaloneConfiguration(t *testing.T) {
	testCases := []struct {
		name          string
		path          string
		errorExpected bool
	}{
		{
			name:          "Valid config file",
			path:          "./testdata/config.yaml",
			errorExpected: false,
		},
		{
			name:          "Invalid file path",
			path:          "invalid_file.yaml",
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := LoadStandaloneConfiguration(tc.path)
			if tc.errorExpected {
				assert.Error(t, err, "Expected an error")
				assert.Nil(t, config, "Config should not be loaded")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.NotNil(t, config, "Config not loaded as expected")
			}
		})
	}

	t.Run("Parse environment variables", func(t *testing.T) {
		os.Setenv("PLUGIN_SECRET_KEY", "zhu88jie22")
		os.Setenv("PLUGIN_HTTP_PORT", "10010")
		config, err := LoadStandaloneConfiguration("./testdata/env_variables_config.yaml")
		assert.NoError(t, err, "Unexpected error")
		assert.NotNil(t, config, "Config not loaded as expected")
		assert.Equal(t, 10010, config.Plugin.Port)
	})
}
