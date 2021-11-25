package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallChart(t *testing.T) {
	_ = SetNamespace("default")
	tests := []struct {
		name        string
		releaseName string
		chart       string
		version     string
		want        error
	}{
		{"install bitnami/drupal", "bd", "bitnami/drupal", "", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, installChart(test.releaseName, test.chart, test.version))
		})
	}
}
