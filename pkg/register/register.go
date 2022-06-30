package register

import (
	"sync"
	"time"

	"github.com/tkeel-io/kit/log"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	apps_v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var once sync.Once
var _pluginRegistry *PluginRegistry

type PluginInfo struct {
	ID        string
	Status    openapi_v1.PluginStatus
	IsUpgrade bool
	Callback  func() bool
}

type PluginRegistry struct {
	sync.RWMutex
	Plugins map[string]*PluginInfo
	stopCh  chan struct{}
}

func Init() {
	once.Do(func() {
		_pluginRegistry = &PluginRegistry{
			Plugins: make(map[string]*PluginInfo),
			stopCh:  make(chan struct{}),
		}
	})
}

func Instance() *PluginRegistry {
	return _pluginRegistry
}

func (pr *PluginRegistry) Register(pluginID string, isUpgrade bool, callback func() bool) {
	pr.Lock()
	defer pr.Unlock()
	log.Debugf("register new plugin: %s, upgrade: %v", pluginID, isUpgrade)
	if plugin, ok := pr.Plugins[pluginID]; ok {
		if isUpgrade {
			plugin.Status = openapi_v1.PluginStatus_WAIT_RUNNING
			plugin.Callback = callback
		} else {

		}
	} else {
		pr.Plugins[pluginID] = &PluginInfo{
			ID:        pluginID,
			Status:    openapi_v1.PluginStatus_WAIT_RUNNING,
			IsUpgrade: isUpgrade,
			Callback:  callback,
		}
	}
}

func (pr *PluginRegistry) Run() {
	log.Info("plugin registry is running")
	config, err := rest.InClusterConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	sharedInformers := informers.NewSharedInformerFactoryWithOptions(clientset, time.Minute, informers.WithNamespace(""))

	deployInformer := sharedInformers.Apps().V1().Deployments().Informer()
	deployInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oDep := newObj.(*apps_v1.Deployment)
			if oDep.Status.ReadyReplicas == oDep.Status.Replicas {
				pr.Lock()
				defer pr.Unlock()
				plugin, ok := pr.Plugins[oDep.Name]
				if !ok {
					return
				}
				log.Debugf("pod %s status updated, ready replicas : %d/%d", oDep.Name, oDep.Status.ReadyReplicas, oDep.Status.Replicas)
				if plugin.Status == openapi_v1.PluginStatus_RUNNING {
					log.Debugf("plugin %s registered, skip", oDep.Name)
					return
				}
				if plugin.Callback() {
					log.Debugf("plugin %s registered successfully", oDep.Name)
					plugin.Status = openapi_v1.PluginStatus_RUNNING
				} else {
					log.Debugf("plugin %s fail to register", oDep.Name)
					plugin.Status = openapi_v1.PluginStatus_ERR_REGISTER
				}
			}
		},
		DeleteFunc: func(obj interface{}) {},
	})
	go deployInformer.Run(pr.stopCh)

	statefulInformer := sharedInformers.Apps().V1().StatefulSets().Informer()
	statefulInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oSta := newObj.(*apps_v1.StatefulSet)
			if oSta.Status.ReadyReplicas == oSta.Status.Replicas {
				pr.Lock()
				defer pr.Unlock()
				plugin, ok := pr.Plugins[oSta.Name]
				if !ok {
					return
				}
				log.Debugf("pod %s status updated, ready replicas : %d/%d", oSta.Name, oSta.Status.ReadyReplicas, oSta.Status.Replicas)
				if plugin.Status == openapi_v1.PluginStatus_RUNNING {
					log.Debugf("plugin %s registered, skip", oSta.Name)
					return
				}
				if plugin.Callback() {
					log.Debugf("plugin %s registered successfully", oSta.Name)
					plugin.Status = openapi_v1.PluginStatus_RUNNING
				} else {
					log.Debugf("plugin %s fail to register", oSta.Name)
					plugin.Status = openapi_v1.PluginStatus_ERR_REGISTER
				}
			}
		},
		DeleteFunc: func(obj interface{}) {},
	})
	go statefulInformer.Run(pr.stopCh)
}

func (pr *PluginRegistry) Stop() {
	close(pr.stopCh)
}
