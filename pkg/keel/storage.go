package keel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dapr/go-sdk/client"
)

var (
	PrivateStore = func() string {
		if e := os.Getenv("KEEL_PRIVATE_STORE"); e != "" {
			return e
		}
		return DefaultPrivateStore
	}()

	PublicStore = func() string {
		if e := os.Getenv("KEEL_PUBLIC_STORE"); e != "" {
			return e
		}
		return DefaultPublicStore
	}()
)

func GetStoreKey(prifix, id string) string {
	return fmt.Sprintf("%s%s", prifix, id)
}

func GetAllRegisteredPlugin(ctx context.Context) (allPlugin map[string]string, etag string, err error) {
	allPinItem, err := GetClient().GetState(ctx, PrivateStore,
		KeyAllRegisteredPlugin)
	if err != nil {
		return nil, "", fmt.Errorf("error get state: %w", err)
	}
	if allPinItem.Etag == "" {
		log.Debugf("all plugins is not registered")
		return nil, "", nil
	}
	retP := make(map[string]string)
	err = json.Unmarshal(allPinItem.Value, &retP)
	if err != nil {
		log.Errorf("error json Unmarshal(%v): %s", allPinItem.Value, err)
		return nil, "", fmt.Errorf("error json unmarshal: %w", err)
	}
	return retP, allPinItem.Etag, nil
}

func SaveAllRegisteredPlugin(ctx context.Context, allPlugins map[string]string, etag string) error {
	allpByte, err := json.Marshal(allPlugins)
	if err != nil {
		log.Errorf("error json marshal all resigtered plugin map(%v): %s",
			allPlugins, err)
		return fmt.Errorf("error json marshal: %w", err)
	}
	err = GetClient().SaveBulkState(ctx, PrivateStore,
		&client.SetStateItem{
			Key:   KeyAllRegisteredPlugin,
			Value: allpByte,
			Etag: &client.ETag{
				Value: etag,
			},
			Options: &client.StateOptions{
				Concurrency: client.StateConcurrencyFirstWrite,
				Consistency: client.StateConsistencyStrong,
			},
		})
	if err != nil {
		log.Errorf("error save all resigtered plugin map(%v): %s",
			allPlugins, err)
		return fmt.Errorf("error save state: %w", err)
	}
	return nil
}

func GetScrapeFlag(ctx context.Context) (flag, etag string, err error) {
	flagItem, err := GetClient().GetState(ctx, PrivateStore,
		KeyScrapeFlag)
	if err != nil {
		log.Errorf("error get scrape flag(%s): %s", KeyScrapeFlag, err)
		return "", "", fmt.Errorf("error get state: %w", err)
	}
	return string(flagItem.Value), flagItem.Etag, nil
}

func SaveScrapeFlag(ctx context.Context, etag string, ttlSecond int64) error {
	err := GetClient().SaveBulkState(ctx, PrivateStore,
		&client.SetStateItem{
			Key:   KeyScrapeFlag,
			Value: []byte("true"),
			Etag: &client.ETag{
				Value: func() string {
					if etag == "" {
						return "-1"
					}
					return etag
				}(),
			},
			Metadata: map[string]string{"ttlInSeconds": fmt.Sprintf("%d", ttlSecond)},
			Options: &client.StateOptions{
				Concurrency: client.StateConcurrencyFirstWrite,
				Consistency: client.StateConsistencyStrong,
			},
		})
	if err != nil {
		log.Errorf("error save scrape flag: %s",
			err)
		return fmt.Errorf("error save state: %w", err)
	}
	return nil
}

func GetPlugin(ctx context.Context,
	pID string) (p *Plugin, etag string, err error) {
	pluginItem, err := GetClient().GetState(ctx, PrivateStore,
		GetStoreKey(KeyPrefixPlugin, pID))
	if err != nil {
		log.Errorf("error get plugin(%s): %s", pID, err)
		return nil, "", fmt.Errorf("error get state: %w", err)
	}
	if pluginItem.Etag == "" {
		log.Debugf("plugin(%s) is not registered", pID)
		return nil, "", nil
	}
	retP := &Plugin{}
	err = json.Unmarshal(pluginItem.Value, retP)
	if err != nil {
		log.Errorf("error json Unmarshal(%v): %s", pluginItem, err)
		return nil, "", fmt.Errorf("error json unmarshal: %w", err)
	}
	return retP, pluginItem.Etag, nil
}

func SavePlugin(ctx context.Context, pin *Plugin, etag string) error {
	npByte, err := json.Marshal(pin)
	if err != nil {
		log.Errorf("error json marshal plugin(%s): %s",
			pin.PluginID, err)
		return fmt.Errorf("error json marshal: %w", err)
	}
	err = GetClient().SaveBulkState(ctx, PrivateStore,
		&client.SetStateItem{
			Key:   GetStoreKey(KeyPrefixPlugin, pin.PluginID),
			Value: npByte,
			Etag: &client.ETag{
				Value: etag,
			},
			Options: &client.StateOptions{
				Concurrency: client.StateConcurrencyFirstWrite,
				Consistency: client.StateConsistencyStrong,
			},
		})
	if err != nil {
		log.Errorf("error save plugin(%s): %s",
			pin.PluginID, err)
		return fmt.Errorf("error save state: %w", err)
	}
	return nil
}

func GetPluginRoute(ctx context.Context,
	pID string) (p *PluginRoute, etag string, err error) {
	routeItem, err := GetClient().GetState(ctx, PublicStore,
		GetStoreKey(KeyPrefixPluginRoute, pID))
	if err != nil {
		log.Errorf("error get plugin_route(%s): %s", pID, err)
		return nil, "", fmt.Errorf("error get state: %w", err)
	}
	if routeItem.Etag == "" {
		log.Debugf("plugin route(%s) is not registered", pID)
		return nil, "", nil
	}
	retP := &PluginRoute{}
	err = json.Unmarshal(routeItem.Value, retP)
	if err != nil {
		log.Errorf("error json Unmarshal(%v): %s", routeItem, err)
		return nil, "", fmt.Errorf("error json unmarshal: %w", err)
	}
	return retP, routeItem.Etag, nil
}

func SavePluginRoute(ctx context.Context, pID string, pRoute *PluginRoute, etag string) error {
	pRouteByte, err := json.Marshal(pRoute)
	if err != nil {
		log.Errorf("error json marshal plugin_route(%s): %s",
			pID, err)
		return fmt.Errorf("error json marshal: %w", err)
	}
	err = GetClient().SaveBulkState(ctx, PublicStore,
		&client.SetStateItem{
			Key:   GetStoreKey(KeyPrefixPluginRoute, pID),
			Value: pRouteByte,
			Etag: &client.ETag{
				Value: etag,
			},
			Options: &client.StateOptions{
				Concurrency: client.StateConcurrencyFirstWrite,
				Consistency: client.StateConsistencyStrong,
			},
		})
	if err != nil {
		log.Errorf("error save plugin_route(%s): %s",
			pID, err)
		return fmt.Errorf("error save state: %w", err)
	}
	return nil
}

func DeletePluginRoute(ctx context.Context, pID string) error {
	err := GetClient().DeleteState(ctx, PublicStore,
		GetStoreKey(KeyPrefixPluginRoute, pID))
	if err != nil {
		log.Errorf("error delete plugin route(%s): %s", pID, err)
		return fmt.Errorf("error delete state: %w", err)
	}
	return nil
}

func DeletePlugin(ctx context.Context, pID string) error {
	err := GetClient().DeleteState(ctx, PrivateStore,
		GetStoreKey(KeyPrefixPlugin, pID))
	if err != nil {
		log.Errorf("error delete plugin(%s): %s", pID, err)
		return fmt.Errorf("error delete state: %w", err)
	}
	return nil
}
