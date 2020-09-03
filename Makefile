.PHONY := clean test tools agency

CANIS_ROOT=$(abspath .)
DIDCOMM_LB_FILES = $(wildcard pkg/didcomm/loadbalancer/*.go pkg/didcomm/loadbalancer/**/*.go cmd/canis-didcomm-lb/*.go)
DIDCOMM_ISSUER_FILES = $(wildcard pkg/didcomm/issuer/*.go pkg/didcomm/issuer/**/*.go cmd/canis-didcomm-issuer/*.go)

all: clean tools build

commit: cover build

# Cleanup files (used in Jenkinsfile)
clean:
	rm -f bin/*

tools:
	go get bou.ke/staticfiles
	go get github.com/vektra/mockery/.../
	go get golang.org/x/tools/cmd/cover
	go get -u github.com/golang/protobuf/protoc-gen-go

swagger_pack: pkg/static/canis-apiserver_swagger.go
pkg/static/canis-apiserver_swagger.go: canis-apiserver-pb pkg/apiserver/api/spec/canis-apiserver.swagger.json
	staticfiles -o pkg/static/canis-apiserver_swagger.go --package static pkg/apiserver/api/spec


build: bin/canis-apiserver bin/agent bin/sirius bin/canis-didcomm-issuer bin/canis-didcomm-lb
build-canis-apiserver: bin/canis-apiserver
build-canis-scheduler: bin/canis-scheduler
build-canis-didcomm-issuer: bin/canis-didcomm-issuer
build-canis-didcomm-lb: bin/canis-didcomm-lb

canis-apiserver: bin/canis-apiserver
bin/canis-apiserver: canis-apiserver-pb swagger_pack
	@. ./canis.sh; cd cmd/canis-apiserver && go build -o $(CANIS_ROOT)/bin/canis-apiserver

canis-didcomm-issuer: bin/canis-didcomm-issuer
bin/canis-didcomm-issuer: $(DIDCOMM_ISSUER_FILES)
	@. ./canis.sh; cd cmd/canis-didcomm-issuer && go build -o $(CANIS_ROOT)/bin/canis-didcomm-issuer

canis-didcomm-lb: bin/canis-didcomm-lb
bin/canis-didcomm-lb: $(DIDCOMM_LB_FILES)
	@. ./canis.sh; cd cmd/canis-didcomm-lb && go build -o $(CANIS_ROOT)/bin/canis-didcomm-lb

sirius: bin/sirius
bin/sirius:
	@. ./canis.sh; cd cmd/sirius && go build -o $(CANIS_ROOT)/bin/sirius

.PHONY: canis-docker
package: canis-docker

build-agent: bin/agent
build-router: bin/router

agent: bin/agent
bin/agent: canis-apiserver-pb
	@. ./canis.sh; cd cmd/agent && go build -o $(CANIS_ROOT)/bin/agent

agency: bin/agency bin/router
bin/agency:
	@. ./canis.sh; cd cmd/agency && go build -o $(CANIS_ROOT)/bin/agency

router: bin/router
bin/router:
	@. ./canis.sh; cd cmd/router && go build -o $(CANIS_ROOT)/bin/router

canis-docker: build
	@echo "Building canis docker image"
	@docker build -f ./docker/canis/Dockerfile --no-cache -t canis/canis:latest .

canis-apiserver-pb: pkg/apiserver/api/canis-apiserver.pb.go
pkg/apiserver/api/canis-apiserver.pb.go:pkg/apiserver/api/canis-apiserver.proto
	cd pkg && protoc -I/home/pfeairheller/opt/protoc-3.6.1/include -I . -I apiserver/api/ apiserver/api/canis-apiserver.proto --go_out=plugins=grpc:.
	cd pkg && protoc -I/home/pfeairheller/opt/protoc-3.6.1/include -I . -I apiserver/api/ apiserver/api/canis-apiserver.proto --grpc-gateway_out=logtostderr=true:.
	cd pkg && protoc -I/home/pfeairheller/opt/protoc-3.6.1/include -I . -I apiserver/api/ apiserver/api/canis-apiserver.proto --swagger_out=logtostderr=true:.
	mv pkg/apiserver/api/canis-apiserver.swagger.json pkg/apiserver/api/spec
demo-web:
	cd demo && npm run build

# Development Local Run Shortcuts
test: clean tools
	@. ./canis.sh; ./scripts/test.sh

cover:
	go test -coverprofile cover.out ./pkg/...
	go tool cover -html=cover.out

install:
	@helm install canis ./canis-chart --values ./k8s/config/local/values.yaml --kubeconfig ./k8s/config/local/kubeconfig.yaml

uninstall:
	@helm uninstall canis && ([ $$? -eq 0 ] && echo "") || echo "nothing to uninstall!"

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