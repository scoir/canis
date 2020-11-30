.PHONY := all clean test cover install uninstall expose tools build

CANIS_ROOT=$(abspath .)
APISERVER_FILES = $(wildcard pkg/apiserver/*.go pkg/apiserver/**/*.go cmd/canis-apiserver/*.go) pkg/protogen/common/messages.pb.go pkg/apiserver/api/protogen/canis-apiserver.pb.go
SIRIUS_FILES = $(wildcard pkg/sirius/**/*.go cmd/sirius/*.go)
DIDCOMM_LB_FILES = $(wildcard pkg/didcomm/loadbalancer/*.go pkg/didcomm/loadbalancer/**/*.go cmd/canis-didcomm-lb/*.go)
DIDCOMM_ISSUER_FILES = $(wildcard pkg/didcomm/issuer/*.go pkg/didcomm/issuer/**/*.go cmd/canis-didcomm-issuer/*.go)
DIDCOMM_VERIFIER_FILES = $(wildcard pkg/didcomm/verifier/*.go pkg/didcomm/verifier/**/*.go cmd/canis-didcomm-verifier/*.go)
DIDCOMM_DOORMAN_FILES = $(wildcard pkg/didcomm/doorman/*.go pkg/didcomm/doorman/**/*.go cmd/canis-didcomm-doorman/*.go)
HTTP_INDY_RESOLVER_FILES = $(wildcard pkg/resolver/*.go pkg/resolver/**/*.go cmd/http-indy-resolver/*.go)
WEBHOOK_NOTIFIER_FILES = $(wildcard pkg/notifier/*.go pkg/notifier/**/*.go cmd/canis-webhook-notifier/*.go)

all: clean tools build

commit: cover build

# Cleanup files (used in Jenkinsfile)
clean:
	rm -f bin/*

tools:
	@echo "Loading golang tools..."
	@go get "bou.ke/staticfiles@v0.0.0-20190225145250-827d7f6389cd"
	@go get "github.com/vektra/mockery/.../@v1.1.2"
	@go get "golang.org/x/tools/cmd/cover@v0.0.0-20200904140424-93eecc3576be"
	@go get "github.com/golang/protobuf/protoc-gen-go@v1.4.2"
	@go get "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.15.2"
	@go get "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.15.2"

build: bin/canis-apiserver bin/sirius bin/canis-didcomm-issuer bin/canis-didcomm-verifier bin/canis-didcomm-lb bin/canis-didcomm-doorman bin/http-indy-resolver bin/canis-webhook-notifier

.PHONY: canis-apiserver
canis-apiserver: bin/canis-apiserver
bin/canis-apiserver: $(APISERVER_FILES)
	@echo 'Building canis-apiserver...'
	@. ./canis.sh; cd cmd/canis-apiserver && go build -o $(CANIS_ROOT)/bin/canis-apiserver

.PHONY: canis-didcomm-issuer
canis-didcomm-issuer: bin/canis-didcomm-issuer
bin/canis-didcomm-issuer: $(DIDCOMM_ISSUER_FILES)
	@echo 'Building canis-didcomm-issuer...'
	@. ./canis.sh; cd cmd/canis-didcomm-issuer && go build -o $(CANIS_ROOT)/bin/canis-didcomm-issuer

.PHONY: canis-didcomm-verifier
canis-didcomm-verifier: bin/canis-didcomm-verifier
bin/canis-didcomm-verifier: $(DIDCOMM_VERIFIER_FILES)
	@echo 'Building canis-didcomm-verifier...'
	@. ./canis.sh; cd cmd/canis-didcomm-verifier && go build -o $(CANIS_ROOT)/bin/canis-didcomm-verifier

.PHONY: canis-didcomm-doorman
canis-didcomm-doorman: bin/canis-didcomm-doorman
bin/canis-didcomm-doorman: canis-didcomm-doorman-pb $(DIDCOMM_DOORMAN_FILES)
	@echo 'Building canis-didcomm-doorman...'
	@. ./canis.sh; cd cmd/canis-didcomm-doorman && go build -o $(CANIS_ROOT)/bin/canis-didcomm-doorman

.PHONY: canis-didcomm-lb
canis-didcomm-lb: bin/canis-didcomm-lb
bin/canis-didcomm-lb: $(DIDCOMM_LB_FILES)
	@echo 'Building canis-didcomm-lb...'
	@. ./canis.sh; cd cmd/canis-didcomm-lb && go build -o $(CANIS_ROOT)/bin/canis-didcomm-lb

.PHONY: http-indy-resolver
http-indy-resolver: bin/http-indy-resolver
bin/http-indy-resolver: $(HTTP_INDY_RESOLVER_FILES)
	@echo 'Building http-indy-resolver...'
	@. ./canis.sh; cd cmd/http-indy-resolver && go build -o $(CANIS_ROOT)/bin/http-indy-resolver

.PHONY: canis-webhook-notifier
canis-webhook-notifier: bin/canis-webhook-notifier
bin/canis-webhook-notifier: $(WEBHOOK_NOTIFIER_FILES)
	@echo 'Building canis-webhook-notifier...'
	@. ./canis.sh; cd cmd/canis-webhook-notifier && go build -o $(CANIS_ROOT)/bin/canis-webhook-notifier

.PHONY: sirius
sirius: bin/sirius
bin/sirius: $(SIRIUS_FILES)
	@echo 'Building sirius...'
	@. ./canis.sh; cd cmd/sirius && go build -o $(CANIS_ROOT)/bin/sirius

.PHONY: canis-docker
canis-docker:
	@echo "Building canis docker image..."
	@docker build -f docker/local/Dockerfile.build -t canislabs/canis:latest .

.PHONY: canis-docker-publish
canis-docker-publish: canis-docker
	@echo "publishing canis to registry..."
	@docker tag canislabs/canis:latest registry.hyades.svc.cluster.local:5000/canis:latest
	@docker push registry.hyades.svc.cluster.local:5000/canis:latest

.PHONY: canis-build-docker
canis-build-docker:
	@echo "Building canis building docker image..."
	@cd docker/local && docker build -f Dockerfile.golang -t canislabs/canisbuild:golang-1.15 .
	@cd docker/local && docker build -f Dockerfile.bionic -t canislabs/canisbase:latest .
	@docker push canislabs/canisbuild:golang-1.15
	@docker push canislabs/canisbase:latest

.PHONY: canis-docker-ubuntu
canis-docker-ubuntu: build
	@docker build -f ./docker/canis/Dockerfile --no-cache -t canislabs/canis:latest .

.PHONY: all-pb
all-pb: canis-common-pb canis-apiserver-pb canis-didcomm-doorman-pb canis-didcomm-issuer-pb canis-didcomm-verifier-pb canis-didcomm-lb-pb

.PHONY: canis-common-pb
canis-common-pb: pkg/protogen/common/messages.pb.go
pkg/protogen/common/messages.pb.go: pkg/proto/common/messages.proto
	@echo "Generating common messages protobuf files..."
	@cd pkg && protoc -I proto/include/ -I proto/common/ proto/common/messages.proto --go_out=plugins=grpc:.
	@mv pkg/github.com/scoir/canis/pkg/protogen/common/messages.pb.go pkg/protogen/common/messages.pb.go
	@rm -rf pkg/github.com

.PHONY: swagger_pack
swagger_pack: pkg/static/canis-apiserver_swagger.go
pkg/static/canis-apiserver_swagger.go: pkg/apiserver/api/spec/canis-apiserver.swagger.json
	@staticfiles -o pkg/static/canis-apiserver_swagger.go --package static pkg/apiserver/api/spec

.PHONY: canis-apiserver-pb
canis-apiserver-pb: pkg/apiserver/api/protogen/canis-apiserver.pb.go
pkg/apiserver/api/protogen/canis-apiserver.pb.go: pkg/protogen/common/messages.pb.go pkg/proto/canis-apiserver.proto
	@echo "Generating apiserver protobuf files..."
	@cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I apiserver/api/ proto/canis-apiserver.proto --go_out=plugins=grpc:.
	@mv pkg/apiserver/api/canis-apiserver.pb.go pkg/apiserver/api/protogen/canis-apiserver.pb.go
	@cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I apiserver/api/ proto/canis-apiserver.proto --grpc-gateway_out=logtostderr=true:.
	@mv pkg/apiserver/api/canis-apiserver.pb.gw.go pkg/apiserver/api/protogen/canis-apiserver.pb.gw.go
	@cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I apiserver/api/ proto/canis-apiserver.proto --swagger_out=logtostderr=true:.
	@mv pkg/canis-apiserver.swagger.json pkg/apiserver/api/spec

.PHONY: canis-didcomm-doorman-pb
canis-didcomm-doorman-pb: pkg/didcomm/doorman/api/protogen/canis-didcomm-doorman.pb.go
pkg/didcomm/doorman/api/protogen/canis-didcomm-doorman.pb.go: pkg/protogen/common/messages.pb.go pkg/proto/canis-didcomm-doorman.proto
	@echo "Generating doorman protobuf files..."
	@cd pkg && protoc -I proto -I proto/include/ -I proto/common -I didcomm/doorman/api/ proto/canis-didcomm-doorman.proto --go_out=plugins=grpc:.
	@mv pkg/didcomm/doorman/api/canis-didcomm-doorman.pb.go pkg/didcomm/doorman/api/protogen/canis-didcomm-doorman.pb.go

.PHONY: canis-didcomm-issuer-pb
canis-didcomm-issuer-pb: pkg/didcomm/issuer/api/protogen/canis-didcomm-issuer.pb.go
pkg/didcomm/issuer/api/protogen/canis-didcomm-issuer.pb.go: pkg/protogen/common/messages.pb.go pkg/proto/canis-didcomm-issuer.proto
	@echo "Generating issuer protobuf files..."
	@cd pkg && protoc -I proto -I proto/include/ -I proto/common  -I didcomm/issuer/api/ proto/canis-didcomm-issuer.proto --go_out=plugins=grpc:.
	@mv pkg/didcomm/issuer/api/canis-didcomm-issuer.pb.go pkg/didcomm/issuer/api/protogen/canis-didcomm-issuer.pb.go

.PHONY: canis-didcomm-verifier-pb
canis-didcomm-verifier-pb: pkg/didcomm/verifier/api/protogen/canis-didcomm-verifier.pb.go
pkg/didcomm/verifier/api/protogen/canis-didcomm-verifier.pb.go: pkg/protogen/common/messages.pb.go pkg/proto/canis-didcomm-verifier.proto
	@echo "Generating verifier protobuf files..."
	@cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I didcomm/verifier/api/ proto/canis-didcomm-verifier.proto --go_out=plugins=grpc:.
	@mv pkg/didcomm/verifier/api/canis-didcomm-verifier.pb.go pkg/didcomm/verifier/api/protogen/canis-didcomm-verifier.pb.go

.PHONY: canis-didcomm-lb-pb
canis-didcomm-lb-pb: pkg/didcomm/loadbalancer/api/protogen/canis-didcomm-loadbalancer.pb.go
pkg/didcomm/loadbalancer/api/protogen/canis-didcomm-loadbalancer.pb.go: pkg/protogen/common/messages.pb.go pkg/proto/canis-didcomm-loadbalancer.proto
	@echo "Generating loadbalancer protobuf files..."
	@cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I didcomm/loadbalancer/api/ proto/canis-didcomm-loadbalancer.proto --go_out=plugins=grpc:.
	@mv pkg/didcomm/loadbalancer/api/canis-didcomm-loadbalancer.pb.go pkg/didcomm/loadbalancer/api/protogen/canis-didcomm-loadbalancer.pb.go

# Development Local Run Shortcuts
test: clean tools swagger_pack
	@. ./canis.sh; ./scripts/test.sh

cover:
	@echo "Generating coverage report..."
	@go test -coverprofile cover.out ./pkg/...
	@go tool cover -html=cover.out

install: canis-docker-publish
	@echo "Installing canis to minikube files..."
	@helm install canis ./deploy/canis-chart --set image.repository=registry.hyades.svc.cluster.local:5000/canis --kubeconfig ./config/kubeconfig.yaml --namespace=hyades --create-namespace
	@./scripts/endpoint.sh

uninstall:
	@echo "Uninstalling canis..."
	@helm uninstall canis --namespace hyades && ([ $$? -eq 0 ] && echo "") || echo "nothing to uninstall!"

expose:
	@minikube service -n hyades canis-apiserver-loadbalancer --url

.PHONY: copy-config-defaults
copy-config-defaults:
	@cp -f ./deploy/compose/genesis-file.yaml.default ./deploy/compose/genesis-file.yaml
	@cp -f ./deploy/compose/indy-registry.yaml.default ./deploy/compose/indy-registry.yaml
	@cp -f ./deploy/compose/aries-indy-vdri-config.yaml.default ./deploy/compose/aries-indy-vdri-config.yaml
