module github.com/scoir/canis

go 1.14

replace github.com/hyperledger/aries-framework-go => github.com/pfeairheller/aries-framework-go v0.1.4-0.20200918145734-b0d185a4e869

replace github.com/hyperledger/indy-vdr/wrappers/golang => github.com/pfeairheller/indy-vdr/wrappers/golang v0.0.1

require (
	bou.ke/staticfiles v0.0.0-20190225145250-827d7f6389cd // indirect
	github.com/btcsuite/btcutil v1.0.1
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/golangci/golangci-lint v1.31.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/tink/go v1.4.0-rc2.0.20200807212851-52ae9c6679b2
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.15.2
	github.com/hyperledger/aries-framework-go v0.1.4
	github.com/hyperledger/indy-vdr/wrappers/golang v0.0.0-20201031155907-5f437d26ed71
	github.com/hyperledger/ursa-wrapper-go v0.0.0-20201020141813-07eecfdbb801
	github.com/makiuchi-d/gozxing v0.0.0-20200903113411-25f730ed83da
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/mr-tron/base58 v1.2.0
	github.com/mwitkow/grpc-proxy v0.0.0-20181017164139-0f1106ef9c76
	github.com/piprate/json-gold v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/scoir/aries-storage-mongo v0.0.0-20200924155006-8a599355491e
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/unidoc/unipdf/v3 v3.10.1
	github.com/vektra/mockery v1.1.2 // indirect
	go.mongodb.org/mongo-driver v1.4.1
	goji.io v2.0.2+incompatible
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43 // indirect
	golang.org/x/sys v0.0.0-20200812155832-6a926be9bd1d // indirect
	golang.org/x/tools v0.0.0-20200904140424-93eecc3576be // indirect
	google.golang.org/genproto v0.0.0-20201007142714-5c0e72c5e71e
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	k8s.io/client-go v0.17.0
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
	nhooyr.io/websocket v1.8.3
	sigs.k8s.io/yaml v1.2.0 // indirect
)
