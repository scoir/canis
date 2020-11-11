.PHONY := clean test tools agency

CANIS_ROOT=$(abspath .)
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
	go get "bou.ke/staticfiles@v0.0.0-20190225145250-827d7f6389cd"
	go get "github.com/vektra/mockery/.../@v1.1.2"
	go get "golang.org/x/tools/cmd/cover@v0.0.0-20200904140424-93eecc3576be"
	go get "github.com/golang/protobuf/protoc-gen-go@v1.4.2"
	go get "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.15.2"
	go get "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.15.2"

swagger_pack: pkg/static/canis-apiserver_swagger.go
pkg/static/canis-apiserver_swagger.go: canis-apiserver-pb pkg/apiserver/api/spec/canis-apiserver.swagger.json
	@staticfiles -o pkg/static/canis-apiserver_swagger.go --package static pkg/apiserver/api/spec

build: bin/canis-apiserver bin/sirius bin/canis-didcomm-issuer bin/canis-didcomm-verifier bin/canis-didcomm-lb bin/canis-didcomm-doorman bin/http-indy-resolver bin/canis-webhook-notifier
build-canis-apiserver: bin/canis-apiserver
build-canis-didcomm-issuer: bin/canis-didcomm-issuer
build-canis-didcomm-verifier: bin/canis-didcomm-verifier
build-canis-didcomm-lb: bin/canis-didcomm-lb
build-http-indy-resolver: bin/http-indy-resolver
build-canis-webhook-notifier: bin/canis-webhook-notifier

canis-apiserver: bin/canis-apiserver
bin/canis-apiserver: canis-apiserver-pb swagger_pack
	@echo 'building canis-apiserver...'
	@. ./canis.sh; cd cmd/canis-apiserver && go build -o $(CANIS_ROOT)/bin/canis-apiserver

canis-didcomm-issuer: bin/canis-didcomm-issuer
bin/canis-didcomm-issuer: $(DIDCOMM_ISSUER_FILES)
	@echo 'building canis-didcomm-issuer...'
	@. ./canis.sh; cd cmd/canis-didcomm-issuer && go build -o $(CANIS_ROOT)/bin/canis-didcomm-issuer

canis-didcomm-verifier: bin/canis-didcomm-verifier
bin/canis-didcomm-verifier: $(DIDCOMM_VERIFIER_FILES)
	@echo 'building canis-didcomm-verifier...'
	@. ./canis.sh; cd cmd/canis-didcomm-verifier && go build -o $(CANIS_ROOT)/bin/canis-didcomm-verifier

canis-didcomm-doorman: bin/canis-didcomm-doorman
bin/canis-didcomm-doorman: $(DIDCOMM_DOORMAN_FILES)
	@echo 'building canis-didcomm-doorman...'
	@. ./canis.sh; cd cmd/canis-didcomm-doorman && go build -o $(CANIS_ROOT)/bin/canis-didcomm-doorman

canis-didcomm-lb: bin/canis-didcomm-lb
bin/canis-didcomm-lb: $(DIDCOMM_LB_FILES)
	@echo 'building canis-didcomm-lb...'
	@. ./canis.sh; cd cmd/canis-didcomm-lb && go build -o $(CANIS_ROOT)/bin/canis-didcomm-lb

http-indy-resolver: bin/http-indy-resolver
bin/http-indy-resolver: $(HTTP_INDY_RESOLVER_FILES)
	@echo 'building http-indy-resolver...'
	@. ./canis.sh; cd cmd/http-indy-resolver && go build -o $(CANIS_ROOT)/bin/http-indy-resolver

canis-webhook-notifier: bin/canis-webhook-notifier
bin/canis-webhook-notifier: $(WEBHOOK_NOTIFIER_FILES)
	@echo 'building canis-webhook-notifier...'
	@. ./canis.sh; cd cmd/canis-webhook-notifier && go build -o $(CANIS_ROOT)/bin/canis-webhook-notifier

sirius: bin/sirius
bin/sirius: $(SIRIUS_FILES)
	@echo 'building sirius...'
	@. ./canis.sh; cd cmd/sirius && go build -o $(CANIS_ROOT)/bin/sirius

.PHONY: canis-docker
package: canis-docker

canis-docker:
	@echo "building canis docker image..."
	@docker build -f docker/local/Dockerfile.build -t canislabs/canis:latest .

canis-docker-publish: canis-docker
	@echo "publishing canis to registry..."
	@docker tag canislabs/canis:latest registry.hyades.svc.cluster.local:5000/canis:latest
	@docker push registry.hyades.svc.cluster.local:5000/canis:latest

canis-build-docker:
	@echo "building canis build docker image..."
	@cd docker/local && docker build -f Dockerfile.golang -t canislabs/canisbuild:golang-1.15 .
	@cd docker/local && docker build -f Dockerfile.bionic -t canislabs/canisbase:latest .
	@docker push canislabs/canisbuild:golang-1.15
	@docker push canislabs/canisbase:latest

canis-docker-ubuntu: build
	@docker build -f ./docker/canis/Dockerfile --no-cache -t canislabs/canis:latest .


all-pb: canis-common-pb canis-apiserver-pb canis-didcomm-doorman-pb canis-didcomm-issuer-pb canis-didcomm-verifier-pb canis-didcomm-lb-pb

canis-common-pb: pkg/proto/canis-common.pb.go
pkg/proto/canis-common.pb.go:pkg/proto
	cd pkg && protoc -I proto/include/ -I proto/common/ proto/common/messages.proto --go_out=plugins=grpc:.
	mv pkg/github.com/scoir/canis/pkg/protogen/common/messages.pb.go pkg/protogen/common/messages.pb.go
	rm -rf pkg/github.com

canis-apiserver-pb: canis-common-pb pkg/apiserver/api/canis-apiserver.pb.go
pkg/apiserver/api/canis-apiserver.pb.go:pkg/proto/canis-apiserver.proto
	cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I apiserver/api/ proto/canis-apiserver.proto --go_out=plugins=grpc:.
	mv pkg/apiserver/api/canis-apiserver.pb.go pkg/apiserver/api/protogen/canis-apiserver.pb.go
	cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I apiserver/api/ proto/canis-apiserver.proto --grpc-gateway_out=logtostderr=true:.
	mv pkg/apiserver/api/canis-apiserver.pb.gw.go pkg/apiserver/api/protogen/canis-apiserver.pb.gw.go
	cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I apiserver/api/ proto/canis-apiserver.proto --swagger_out=logtostderr=true:.
	mv pkg/canis-apiserver.swagger.json pkg/apiserver/api/spec

canis-didcomm-doorman-pb: canis-common-pb pkg/didcomm/doorman/api/protogen/canis-didcomm-doorman.pb.go
pkg/didcomm/doorman/api/protogen/canis-didcomm-doorman.pb.go:pkg/proto/canis-didcomm-doorman.proto
	cd pkg && protoc -I proto -I proto/include/ -I proto/common -I didcomm/doorman/api/ proto/canis-didcomm-doorman.proto --go_out=plugins=grpc:.
	mv pkg/didcomm/doorman/api/canis-didcomm-doorman.pb.go pkg/didcomm/doorman/api/protogen/canis-didcomm-doorman.pb.go

canis-didcomm-issuer-pb: canis-common-pb pkg/didcomm/issuer/api/protogen/canis-didcomm-issuer.pb.go
pkg/didcomm/issuer/api/protogen/canis-didcomm-issuer.pb.go:pkg/proto/canis-didcomm-issuer.proto
	cd pkg && protoc -I proto -I proto/include/ -I proto/common  -I didcomm/issuer/api/ proto/canis-didcomm-issuer.proto --go_out=plugins=grpc:.
	mv pkg/didcomm/issuer/api/canis-didcomm-issuer.pb.go pkg/didcomm/issuer/api/protogen/canis-didcomm-issuer.pb.go

canis-didcomm-verifier-pb: canis-common-pb pkg/didcomm/verifier/api/canis-didcomm-verifier.pb.go
pkg/didcomm/verifier/api/canis-didcomm-verifier.pb.go:pkg/proto/canis-didcomm-verifier.proto
	cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I didcomm/verifier/api/ proto/canis-didcomm-verifier.proto --go_out=plugins=grpc:.
	mv pkg/didcomm/verifier/api/canis-didcomm-verifier.pb.go pkg/didcomm/verifier/api/protogen/canis-didcomm-verifier.pb.go

canis-didcomm-lb-pb: canis-common-pb pkg/didcomm/loadbalancer/api/protogen/canis-didcomm-loadbalancer.pb.go
pkg/didcomm/loadbalancer/api/protogen/canis-didcomm-loadbalancer.pb.go:pkg/proto/canis-didcomm-loadbalancer.proto
	cd pkg && protoc -I proto -I proto/include/ -I proto/common/ -I didcomm/loadbalancer/api/ proto/canis-didcomm-loadbalancer.proto --go_out=plugins=grpc:.
	mv pkg/didcomm/loadbalancer/api/canis-didcomm-loadbalancer.pb.go pkg/didcomm/loadbalancer/api/protogen/canis-didcomm-loadbalancer.pb.go

demo-web:
	cd demo && npm run build

# Development Local Run Shortcuts
test: clean tools swagger_pack
	@. ./canis.sh; ./scripts/test.sh

cover:
	go test -coverprofile cover.out ./pkg/...
	go tool cover -html=cover.out

install: canis-docker-publish
	@helm install canis ./deploy/canis-chart --set image.repository=registry.hyades.svc.cluster.local:5000/canis --kubeconfig ./config/kubeconfig.yaml --namespace=hyades --create-namespace
	@./scripts/endpoint.sh

uninstall:
	@helm uninstall canis --namespace hyades && ([ $$? -eq 0 ] && echo "") || echo "nothing to uninstall!"

expose:
	minikube service -n hyades canis-apiserver-loadbalancer --url

von-ip:
	@docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' von_webserver_1

ledger:
	@helm upgrade --install ledger ./ledger-chart --values ./k8s/config/local/values.yaml --kubeconfig ./k8s/config/local/kubeconfig.yaml

prune:
	@echo
	@echo "These might be overly aggressive but they work and I just reclaimed 7gb of docker images sooooooooooo"
	@echo
	@echo "Deletes dangling data"
	@echo -e "\t \U0001F92F  #docker system prune"
	@echo
	@echo "Deletes any image not referenced by any container"
	@echo -e "\t \U0001F92F  #docker image prune -a"

copy-config-defaults:
	@cp -f ./deploy/compose/genesis-file.yaml.default ./deploy/compose/genesis-file.yaml
	@cp -f ./deploy/compose/indy-registry.yaml.default ./deploy/compose/indy-registry.yaml
	@cp -f ./deploy/compose/aries-indy-vdri-config.yaml.default ./deploy/compose/aries-indy-vdri-config.yaml
