package openapi

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tkeel/pkg/utils"
)

var (
	o *Openapi
)

func newOprator() {
	o = NewOpenapi(8080, "keel-hello", "1.0")
}

func TestNewOprator(t *testing.T) {
	t.Run("test create oprator", func(t *testing.T) {
		o = NewOpenapi(8080, "keel-hello", "1.0")
		assert.NotNil(t, o)
		assert.NoError(t, o.Close())
	})
}

func TestOpratorAddOpenAPI(t *testing.T) {
	t.Run("test oprator add open api", func(t *testing.T) {
		newOprator()
		o.AddOpenAPI(&API{
			Endpoint: "/echo",
			H: func(a *APIEvent) {
				switch a.HTTPReq.Method {
				case http.MethodGet:
					req := utils.GetURLValue(a.HTTPReq.URL, "data")
					a.Write([]byte(req))
				case http.MethodPost:
					resp := &struct {
						Data string `json:"data"`
					}{}
					err := utils.ReadBody2Json(a.HTTPReq.Body, resp)
					assert.NoError(t, err)
				default:
					http.Error(a, "method not allow", http.StatusMethodNotAllowed)
					assert.NotEqualValues(t, a.HTTPReq.Method, http.MethodGet, http.MethodPost)
				}
			},
		})
		assert.NoError(t, o.Close())
	})
}

func TestOpratorIdentify(t *testing.T) {
	t.Run("test oprator default identify", func(t *testing.T) {
		newOprator()
		iresp, err := o.Identify()
		assert.NoError(t, err)
		assert.NotNil(t, iresp)
		assert.Equal(t, iresp.Ret, 0)
		assert.Equal(t, iresp.Msg, "ok")
		assert.Equal(t, iresp.PluginID, "keel-hello")
		assert.Equal(t, iresp.Version, "1.0")
		assert.Nil(t, iresp.AddonsPoints)
		assert.Nil(t, iresp.MainPlugins)
		assert.NoError(t, o.Close())
	})
}

func TestOpratorStatus(t *testing.T) {
	t.Run("test oprator default status", func(t *testing.T) {
		newOprator()
		iresp, err := o.Status()
		assert.NoError(t, err)
		assert.NotNil(t, iresp)
		assert.Equal(t, iresp.Ret, 0)
		assert.Equal(t, iresp.Msg, "ok")
	})
}

func TestOpratorListen(t *testing.T) {
	t.Run("test oprator listen", func(t *testing.T) {
		go func() {
			err := o.Listen()
			assert.NoError(t, err)
		}()
		time.Sleep(2 * time.Second)
	})
}

func TestDefaultOpratorHttpMethod(t *testing.T) {
	t.Run("test default oprator http method", func(t *testing.T) {
		newOprator()
		go func() {
			err := o.Listen()
			assert.NoError(t, err)
		}()
		time.Sleep(2 * time.Second)

		// test default identify.
		resp, err := http.DefaultClient.Get("http://127.0.0.1:8080/v1/identify")
		defer func() {
			assert.NoError(t, resp.Body.Close())
		}()
		assert.NoError(t, err)
		iresp := &IdentifyResp{}
		assert.NoError(t, utils.ReadBody2Json(resp.Body, iresp))
		assert.NotNil(t, iresp)
		assert.Equal(t, iresp.Ret, 0)
		assert.Equal(t, iresp.Msg, "ok")
		assert.Equal(t, iresp.PluginID, "keel-hello")
		assert.Equal(t, iresp.Version, "1.0")
		assert.Nil(t, iresp.AddonsPoints)
		assert.Nil(t, iresp.MainPlugins)

		// test default status.
		resp, err = http.DefaultClient.Get("http://127.0.0.1:8080/v1/status")
		defer func() {
			assert.NoError(t, resp.Body.Close())
		}()
		assert.NoError(t, err)
		sresp := &StatusResp{}
		assert.NoError(t, utils.ReadBody2Json(resp.Body, sresp))
		assert.NotNil(t, sresp)
		assert.Equal(t, sresp.Ret, 0)
		assert.Equal(t, sresp.Msg, "ok")
		assert.Equal(t, sresp.Status, "ACTIVE")

		// test default tenant bind.
		resp, err = http.DefaultClient.Get("http://127.0.0.1:8080/v1/tenant/bind")
		defer func() {
			assert.NoError(t, resp.Body.Close())
		}()
		assert.NoError(t, err)
		tresp := &TenantBindResp{}
		assert.NoError(t, utils.ReadBody2Json(resp.Body, tresp))
		assert.NotNil(t, tresp)
		assert.Equal(t, tresp.Ret, 0)
		assert.Equal(t, tresp.Msg, "ok")

		// test default addons identify.
		resp, err = http.DefaultClient.Get("http://127.0.0.1:8080/v1/addons/identify")
		defer func() {
			assert.NoError(t, resp.Body.Close())
		}()
		assert.NoError(t, err)
		aresp := &AddonsIdentifyResp{}
		assert.NoError(t, utils.ReadBody2Json(resp.Body, aresp))
		assert.NotNil(t, aresp)
		assert.Equal(t, aresp.Ret, 400)
		assert.Equal(t, aresp.Msg, "no extension point")

		assert.NoError(t, o.Close())
	})
}
