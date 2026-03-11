IMG ?= quay.io/redhat-best-practices-for-k8s/bps-operator:latest

.PHONY: build
build:
	go build -o bin/manager ./cmd/

.PHONY: test
test:
	go test ./... -coverprofile cover.out

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: generate
generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: manifests
manifests:
	controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: build-image
build-image:
	docker build -t $(IMG) .

.PHONY: push-image
push-image:
	docker push $(IMG)

.PHONY: install
install: manifests
	kubectl apply -f config/crd/bases/

.PHONY: uninstall
uninstall:
	kubectl delete -f config/crd/bases/

.PHONY: deploy
deploy: manifests
	kubectl apply -f config/crd/bases/
	kubectl apply -f config/rbac/
	kubectl apply -f config/manager/

.PHONY: undeploy
undeploy:
	kubectl delete -f config/manager/
	kubectl delete -f config/rbac/
	kubectl delete -f config/crd/bases/

.PHONY: run
run:
	go run ./cmd/ --operator-namespace=$(shell kubectl config view --minify -o jsonpath='{..namespace}')

.PHONY: coverage-html
coverage-html: test
	go tool cover -html=cover.out
