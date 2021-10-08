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
		// act
		p, err = FromFlags()
		conf := p.Conf()
		// assert
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, conf.Plugin.ID, "com-keel-hello")
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
					switch a.HttpReq.Method {
					case http.MethodGet:
						req := utils.GetURLValue(a.HttpReq.URL, "data")
						a.Write([]byte(req))
					case http.MethodPost:
						resp := &struct {
							Data string `json:"data"`
						}{}
						err = utils.ReadBody2Json(a.HttpReq.Body, resp)
						assert.NoError(t, err)
					default:
						http.Error(a, "method not allow", http.StatusMethodNotAllowed)
						assert.NotEqualValues(t, a.HttpReq.Method, http.MethodGet, http.MethodPost)
					}
				},
			})
		}()
		time.Sleep(2 * time.Second)
	})

}
