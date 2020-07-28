module github.com/scoir/canis

replace github.com/hyperledger/aries-framework-go => ../aries-framework-go

replace github.com/pfeairheller/indy-vdr/wrappers/golang => ../indy-vdr/wrappers/golang

go 1.14

require (
	bou.ke/staticfiles v0.0.0-20190225145250-827d7f6389cd // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/hyperledger/aries-framework-go v0.1.3
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.3
	github.com/mr-tron/base58 v1.2.0
	github.com/mwitkow/grpc-proxy v0.0.0-20181017164139-0f1106ef9c76
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pfeairheller/indy-vdr/wrappers/golang v0.0.0-20200721120153-a6a48010aad0
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/vektra/mockery v1.1.2 // indirect
	go.mongodb.org/mongo-driver v1.3.4
	goji.io v2.0.2+incompatible
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/tools v0.0.0-20200803225502-5a22b632c5de // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
