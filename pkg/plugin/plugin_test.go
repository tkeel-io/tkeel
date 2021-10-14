package plugin

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tkeel/pkg/openapi"
	"github.com/tkeel-io/tkeel/pkg/utils"
	"github.com/tkeel-io/tkeel/pkg/version"
)

var (
	p   *Plugin
	err error
)

func TestNewPluginFromFlag(t *testing.T) {
	t.Run("test create plugin from flag", func(t *testing.T) {
		// act.
		p, err = FromFlags()
		conf := p.Conf()
		// assert.
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, conf.Plugin.ID, "keel-hello")
		assert.Equal(t, conf.Plugin.Version, version.Version())
		assert.Equal(t, conf.Plugin.Port, 8080)
	})
}

func TestRun(t *testing.T) {
	t.Run("test run plugin", func(t *testing.T) {
		go func() {
			p.Run(&openapi.API{
				Endpoint: "/echo",
				H: func(a *openapi.APIEvent) {
					switch a.HTTPReq.Method {
					case http.MethodGet:
						req := utils.GetURLValue(a.HTTPReq.URL, "data")
						a.Write([]byte(req))
					case http.MethodPost:
						resp := &struct {
							Data string `json:"data"`
						}{}
						err = utils.ReadBody2Json(a.HTTPReq.Body, resp)
						assert.NoError(t, err)
					default:
						http.Error(a, "method not allow", http.StatusMethodNotAllowed)
						assert.NotEqualValues(t, a.HTTPReq.Method, http.MethodGet, http.MethodPost)
					}
				},
			})
		}()
		time.Sleep(2 * time.Second)
	})
}
