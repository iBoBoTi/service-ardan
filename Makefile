# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)


# http://sales-service.sales-system.svc.cluster.local:4000/debug/pprof/
# Kind
# 	For full Kind v0.20 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.20.0

# ==============================================================================
# Define dependencies

GOLANG          := golang:1.20
ALPINE          := alpine:3.18
KIND            := kindest/node:v1.27.3
POSTGRES        := postgres:15.3
VAULT           := hashicorp/vault:1.14
GRAFANA         := grafana/grafana:9.5.3
PROMETHEUS      := prom/prometheus:v2.44.0
TEMPO           := grafana/tempo:2.1.1
LOKI            := grafana/loki:2.8.2
PROMTAIL        := grafana/promtail:2.8.2
TELEPRESENCE    := datawire/ambassador-telepresence-manager:2.14.0

KIND_CLUSTER    := boboti-starter-cluster
NAMESPACE       := sales-system
APP             := sales
BASE_IMAGE_NAME := boboti/ardan/service
SERVICE_NAME    := sales-api
VERSION         := 0.0.1
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)
METRICS_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME)-metrics:$(VERSION)

run:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

run-help:
	go run app/services/sales-api/main.go --help
# ==============================================================================
# Building containers

all: sales

sales:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

# sets up cluster and loads telepresence
dev-kind:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)

	telepresence --context=kind-$(KIND_CLUSTER) helm install

dev-up: dev-kind
	telepresence --context=kind-$(KIND_CLUSTER) connect

# quits telepresence and deletes cluster
dev-down:
	telepresence quit -s
	kind delete cluster --name $(KIND_CLUSTER)

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

# loads service image into the cluster
dev-load:
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=$(NAMESPACE) --for=condition=Available deployment/sales

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

dev-describe:
	kubectl describe nodes
	kubectl describe svc

dev-describe-deployment:
	kubectl describe deployment --namespace=$(NAMESPACE) $(APP)

# describe the pod
dev-describe-sales:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(APP)

dev-update: all dev-load dev-restart  # when you update the source code of the image

dev-update-apply: all dev-load dev-apply # when you update the configuration of the cluster

metrics-local:
	expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view:
	expvarmon -ports="$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"
	
dev-tel:
	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-tel-connect:
	telepresence --context=kind-$(KIND_CLUSTER) connect

tidy:
	rm -rf vendor
	go mod tidy
	go mod vendor