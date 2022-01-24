module github.com/tkeel-io/tkeel

go 1.16

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/casbin/casbin/v2 v2.39.0
	github.com/dapr/go-sdk v1.3.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/emicklei/go-restful v2.15.0+incompatible
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-oauth2/oauth2/v4 v4.4.2
	github.com/go-oauth2/redis/v4 v4.1.1
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-redis/redis/v8 v8.11.4
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tkeel-io/kit v0.0.0-20211223050802-7dfccfe43fdb
	github.com/tkeel-io/security v0.0.0-20220124084748-5e5d7f65fb21
	github.com/tkeel-io/tkeel-interface/openapi v0.0.0-20211223081012-25aaa61491ab
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20211124211545-fe61309f8881 // indirect
	google.golang.org/genproto v0.0.0-20211129164237-f09f9a12af12
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/gorm v1.22.3
	helm.sh/helm/v3 v3.7.2
	k8s.io/cli-runtime v0.23.1
	sigs.k8s.io/kustomize/api v0.10.2-0.20220110233228-13e26004fd4e
	sigs.k8s.io/kustomize/kyaml v0.13.0
	sigs.k8s.io/yaml v1.2.0
)

exclude sigs.k8s.io/kustomize/api v0.2.0

replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.6.0

replace sigs.k8s.io/kustomize/api => github.com/kubernetes-sigs/kustomize/api v0.10.2-0.20220110233228-13e26004fd4e

replace sigs.k8s.io/kustomize/kyaml => github.com/kubernetes-sigs/kustomize/kyaml v0.10.2-0.20220110233228-13e26004fd4e
