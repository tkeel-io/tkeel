package helm

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/output"
)

// TODO: Make the call mock.
func TestListInstallable(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		updateNow bool
		want      struct {
			contain string
			err     error
		}
	}{
		{"lower json and not update test success case", "json", false, struct {
			contain string
			err     error
		}{"Chart", nil}},
		{"lower yaml and not update test success case", "yaml", false, struct {
			contain string
			err     error
		}{contain: "Chart", err: nil}},
		{"upper json and update test success case", "JSON", true, struct {
			contain string
			err     error
		}{contain: "Chart", err: nil}},
		{"invalid format case", "TAML", true, struct {
			contain string
			err     error
		}{contain: "", err: output.ErrInvalidFormatType}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := ListInstallable(test.format, test.updateNow)
			assert.True(t, strings.Contains(string(b), test.want.contain))
			assert.Equal(t, test.want.err, err)
		})
	}
}

func TestListRepo(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   error
	}{
		{"test repo list is work with json format", "json", nil},
		{"test repo list is work with yaml format", "yaml", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := ListRepo(test.format)
			assert.Equal(t, test.want, err)
			if err != nil {
				return
			}
			fmt.Println(string(c))
			switch strings.ToLower(test.format) {
			case "json":
				assert.Equal(t, true, json.Valid(c))
			case "yaml":
				assert.NotEqual(t, c[0], '"')
				assert.NotEqual(t, c[0], '{')
			default:
				assert.Equal(t, test.want, err)
			}
		})
	}
}

func TestAddRepo(t *testing.T) {
	d, _ := ListRepo("json")
	assert.False(t, strings.Contains(string(d), privateRepoName))

	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{"test for default private repo name", "https://charts.bitnami.com/bitnami", nil},
		{"test for default private repo name", "https://tkeel-io.github.io/helm-charts", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// TODO: stub http request.
			err := AddRepo(test.url)
			assert.Equal(t, test.wantErr, err)
			d, _ := ListRepo("json")
			assert.True(t, strings.Contains(string(d), test.url))
			assert.True(t, strings.Contains(string(d), privateRepoName))
		})
	}

	cmd := exec.Command("helm", "repo", "remove", privateRepoName)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		log.Warn("please remove the test repo manual")
	}
}
