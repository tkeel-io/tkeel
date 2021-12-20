/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/tkeel/pkg/repository"
	helmAction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/getter"
)

func TestDriver_String(t *testing.T) {
	tests := []struct {
		name string
		d    Driver
		want string
	}{
		{"secret", Secret, "secret"},
		{"configMap", ConfigMap, "configmap"},
		{"Mem", Mem, "memory"},
		{"SQL", SQL, "sql"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelmRepo_Close(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"helm repo close", fields{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			assert.Panics(t, func() {
				_ = r.Close()
			})
		})
	}
}

func TestHelmRepo_GetDriver(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}
	tests := []struct {
		name   string
		fields fields
		want   Driver
	}{
		{"get driver", fields{driver: Secret}, Secret},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			if got := r.GetDriver(); got != tt.want {
				t.Errorf("GetDriver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelmRepo_Info(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}

	i := repository.Info{
		Name:        "test",
		URL:         "url",
		Annotations: nil,
	}
	tests := []struct {
		name   string
		fields fields
		want   *repository.Info
	}{
		{"get info", fields{info: &i}, &i},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			if got := r.Info(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Info() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelmRepo_Namespace(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"get namespace", fields{namespace: "namespace"}, "namespace"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			if got := r.Namespace(); got != tt.want {
				t.Errorf("Namespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelmRepo_SetDriver(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}
	type args struct {
		driver Driver
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"set driver", fields{driver: Secret}, args{Mem}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			if err := r.SetDriver(tt.args.driver); (err != nil) != tt.wantErr {
				t.Errorf("SetDriver() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, Mem, r.driver)
		})
	}
}

func TestHelmRepo_SetInfo(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}
	type args struct {
		info repository.Info
	}
	i := repository.Info{
		Name:        "name",
		URL:         "url",
		Annotations: nil,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"set info", fields{info: nil}, args{info: i}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			r.SetInfo(tt.args.info)

			assert.Equal(t, i, *r.info)
		})
	}
}

func TestHelmRepo_SetNamespace(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"set namespace", fields{namespace: "namespace"}, args{namespace: "test"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			if err := r.SetNamespace(tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("SetNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHelmRepo_setActionConfig(t *testing.T) {
	type fields struct {
		info         *repository.Info
		actionConfig *helmAction.Configuration
		httpGetter   getter.Getter
		driver       Driver
		namespace    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"set action configSetup", fields{actionConfig: nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRepo{
				info:         tt.fields.info,
				actionConfig: tt.fields.actionConfig,
				httpGetter:   tt.fields.httpGetter,
				driver:       tt.fields.driver,
				namespace:    tt.fields.namespace,
			}
			if err := r.configSetup(); (err != nil) != tt.wantErr {
				t.Errorf("configSetup() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.NotNil(t, r.actionConfig)
		})
	}
}

func TestNewHelmRepo(t *testing.T) {
	type args struct {
		info      repository.Info
		driver    Driver
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		want    *HelmRepo
		wantErr bool
	}{
		{
			"new helm repo",
			args{
				info: repository.Info{
					Name:        "name",
					URL:         "url",
					Annotations: nil,
				},
				driver:    Mem,
				namespace: "namespace",
			},
			&HelmRepo{info: &repository.Info{
				Name:        "name",
				URL:         "url",
				Annotations: nil,
			}, driver: Mem, namespace: "namespace"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHelmRepo(tt.args.info, tt.args.driver, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHelmRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.info, got.info)
			assert.Equal(t, tt.want.driver, got.driver)
			assert.Equal(t, tt.want.namespace, got.namespace)
		})
	}
}

func Test_initActionConfig(t *testing.T) {
	type args struct {
		namespace string
		driver    Driver
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"init helm action configSetup",
			args{
				namespace: "namespace",
				driver:    Mem,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := initActionConfig(tt.args.namespace, tt.args.driver)
			if (err != nil) != tt.wantErr {
				t.Errorf("initActionConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

/*
func TestSearch(t *testing.T) {
	info := NewInfo("tkeel", _tkeelRepo, nil)
	repo, err := NewHelmRepo(*info, Secret, "default")
	assert.Nil(t, err)

	ibs, err := repo.Search("*")
	assert.Nil(t, err)

	fmt.Printf("%+v\n", ibs)
	fmt.Println()

	i, err := repo.Get("iothub", "0.2.0")
	assert.Nil(t, err)
	fmt.Printf("%+v\n", i)

	fmt.Println("=== Run Install === ")
	i.SetPluginID("test")
	err = i.Install()
	assert.Nil(t, err)

	list, err := repo.Search("iothub")
	if err != nil {
		assert.True(t, list[0].Installed)
	}

	ti := NewHelmInstallerQuick("test", repo.Namespace(), repo.actionConfig)
	i = &ti
	err = i.Uninstall()
	assert.Nil(t, err)
}

//Test Search / Install / Uninstall / Get.
*/
