package keel

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/openapi"

	dapr "github.com/dapr/go-sdk/client"
)

var (
	daprClient dapr.Client
	once       sync.Once
	log        = logger.NewLogger("keel.keel")
	K8S        = func() bool {
		if ok := os.Getenv("KUBERNETES_PORT"); ok != "" {
			return true
		}
		return false
	}()
	daprAddr = func() string {
		if K8S {
			if port := os.Getenv("DAPR_HTTP_PORT"); port != "" {
				return fmt.Sprintf("localhost:%s", port)
			}
		}
		return "localhost:3500"
	}()
)

type Tenant struct {
	TenantID   string      `json:"tenant_id"`
	TenantName string      `json:"tenant_name"`
	ExpireTime int64       `json:"expire_time"`
	Extra      interface{} `json:"extra,omitempty"`
}

type Plugin struct {
	*openapi.IdentifyResp `json:",inline"`
	Secret                string    `json:"secret"`
	RegisterTime          int64     `json:"register_time,omitempty"`
	ActiveTenant          []*Tenant `json:"active_tenant,omitempty"`
}

type PluginRoute struct {
	Status         openapi.PluginStatus `json:"status"`
	TkeelVersion   string               `json:"tkeel_version"`
	RegisterAddons map[string]string    `json:"register_addons,omitempty"`
}

type CallReq struct {
	Header   http.Header
	URLValue url.Values
	Body     []byte
}

func GetClient() dapr.Client {
	once.Do(func() {
		cli, err := dapr.NewClient()
		if err != nil {
			panic(err)
		}
		daprClient = cli
	})
	return daprClient
}

func SetDaprAddr(addr string) {
	daprAddr = addr
}

func WaitDaprSidecarReady(retry int) bool {
	if !K8S {
		return true
	}

	health := func() bool {
		resp, err := http.DefaultClient.Get(K8SDaprSidecarProbe)
		if err != nil || (resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK) {
			log.Debugf("dapr sidecar not ready: %s", func() string {
				if err != nil {
					return err.Error()
				}
				return resp.Status
			}())
			return false
		}
		return true
	}

	retryFunc := func() bool {
		for i := 0; i < retry; i++ {
			if !health() {
				time.Sleep(10 * time.Second)
			} else {
				return true
			}
		}
		return false
	}

	if !health() {
		time.Sleep(5 * time.Second)
		return retryFunc()
	}

	return false
}
