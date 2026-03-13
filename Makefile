IMG ?= quay.io/redhat-best-practices-for-k8s/bps-operator:latest
OPERATOR_NAMESPACE ?= bps-operator-system
TEST_NAMESPACE ?= bps-test
KIND_CLUSTER_NAME ?= kind

# Use oc if available, fall back to kubectl
KUBECTL ?= $(shell command -v oc 2>/dev/null || echo kubectl)

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

.PHONY: test-e2e
test-e2e: build-image ## Run e2e tests against a Kind cluster
	@echo "Loading image into Kind cluster..."
	kind load docker-image $(IMG) --name $(KIND_CLUSTER_NAME)
	$(MAKE) deploy
	$(KUBECTL) set image deployment/bps-operator-controller-manager manager=$(IMG) -n $(OPERATOR_NAMESPACE)
	$(KUBECTL) patch deployment bps-operator-controller-manager -n $(OPERATOR_NAMESPACE) \
		-p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","imagePullPolicy":"IfNotPresent"}]}}}}'
	@echo "Waiting for operator to be ready..."
	$(KUBECTL) rollout status deployment/bps-operator-controller-manager -n $(OPERATOR_NAMESPACE) --timeout=120s
	$(KUBECTL) apply -f config/samples/test-workloads.yaml
	$(KUBECTL) wait --for=jsonpath='{.status.phase}'=Active namespace/$(TEST_NAMESPACE) --timeout=10s
	$(KUBECTL) apply -f config/samples/scanner_bps_test.yaml
	@echo "Waiting for scan to complete..."
	@for i in $$(seq 1 90); do \
		PHASE=$$($(KUBECTL) get bestpracticescanners test-scanner -n $(TEST_NAMESPACE) -o jsonpath='{.status.phase}' 2>/dev/null); \
		if [ "$$PHASE" = "Completed" ]; then echo "Scan completed."; break; fi; \
		if [ $$i -eq 90 ]; then \
			echo "Timed out waiting for scan (phase: $$PHASE)"; \
			echo "=== Operator Logs ==="; \
			$(KUBECTL) logs deployment/bps-operator-controller-manager -n $(OPERATOR_NAMESPACE) --tail=50 2>/dev/null || true; \
			echo "=== Scanner Status ==="; \
			$(KUBECTL) get bestpracticescanners test-scanner -n $(TEST_NAMESPACE) -o yaml 2>/dev/null || true; \
			echo "=== Pod Status ==="; \
			$(KUBECTL) get pods -A 2>/dev/null || true; \
			exit 1; \
		fi; \
		sleep 2; \
	done
	@echo "=== E2E Results ==="
	@$(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE)

CHECKS_DIR ?= $(shell cd ../../redhat-best-practices-for-k8s/checks 2>/dev/null && pwd)

.PHONY: build-image
build-image:
	@if [ -z "$(CHECKS_DIR)" ] || [ ! -d "$(CHECKS_DIR)" ]; then \
		echo "ERROR: checks library not found. Set CHECKS_DIR to the path of the checks module."; \
		exit 1; \
	fi
	rm -rf .checks-vendor && cp -r $(CHECKS_DIR) .checks-vendor
	docker build -t $(IMG) .
	rm -rf .checks-vendor

.PHONY: push-image
push-image:
	docker push $(IMG)

##@ Installation

.PHONY: install
install: manifests ## Install CRDs onto the cluster
	$(KUBECTL) apply -f config/crd/bases/

.PHONY: uninstall
uninstall: ## Remove CRDs from the cluster
	$(KUBECTL) delete --ignore-not-found -f config/crd/bases/

.PHONY: deploy
deploy: manifests ## Deploy the operator (CRDs + RBAC + manager Deployment)
	$(KUBECTL) apply -f config/crd/bases/
	$(KUBECTL) apply -f config/rbac/
	$(KUBECTL) apply -f config/manager/

.PHONY: undeploy
undeploy: ## Remove the operator from the cluster
	$(KUBECTL) delete --ignore-not-found -f config/manager/
	$(KUBECTL) delete --ignore-not-found -f config/rbac/
	$(KUBECTL) delete --ignore-not-found -f config/crd/bases/

##@ Local Development

.PHONY: run
run: install ## Run the operator locally against the current cluster
	go run ./cmd/ --operator-namespace=$(OPERATOR_NAMESPACE)

##@ Test Workloads

# Helper: deploy test workloads and prepare namespace (no scanner CR)
.PHONY: _deploy-test-workloads
_deploy-test-workloads: install
	@$(KUBECTL) delete pods -n $(TEST_NAMESPACE) --all --ignore-not-found 2>/dev/null || true
	@$(KUBECTL) delete --ignore-not-found -f config/samples/scanner_bps_test.yaml 2>/dev/null || true
	@$(KUBECTL) delete --ignore-not-found -f config/samples/scanner_bps_test_periodic.yaml 2>/dev/null || true
	$(KUBECTL) apply -f config/samples/test-workloads.yaml
	@if $(KUBECTL) api-resources 2>/dev/null | grep -q securitycontextconstraints; then \
		echo "OpenShift detected — granting privileged SCC to default SA in $(TEST_NAMESPACE)"; \
		$(KUBECTL) adm policy add-scc-to-user privileged -z default -n $(TEST_NAMESPACE); \
	fi
	@echo "Waiting for namespace to be ready..."
	@$(KUBECTL) wait --for=jsonpath='{.status.phase}'=Active namespace/$(TEST_NAMESPACE) --timeout=10s

.PHONY: deploy-test
deploy-test: _deploy-test-workloads ## Deploy test workloads only (no scanner CR)
	@echo ""
	@echo "Test workloads deployed to $(TEST_NAMESPACE)."
	@echo "Use 'make deploy-scan' or 'make deploy-periodic-scan' to add a scanner."

.PHONY: deploy-scan
deploy-scan: _deploy-test-workloads ## Deploy test workloads + one-shot scanner into bps-test namespace
	$(KUBECTL) apply -f config/samples/scanner_bps_test.yaml
	@echo ""
	@echo "One-shot scanner deployed to $(TEST_NAMESPACE)."
	@echo "Run 'make run' in another terminal to start the operator."
	@echo "Then: $(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE)"

.PHONY: deploy-periodic-scan
deploy-periodic-scan: _deploy-test-workloads ## Deploy test workloads + periodic scanner (5m interval) into bps-test namespace
	$(KUBECTL) apply -f config/samples/scanner_bps_test_periodic.yaml
	@echo ""
	@echo "Periodic scanner deployed to $(TEST_NAMESPACE) (interval: 5m)."
	@echo "Run 'make run' in another terminal to start the operator."
	@echo "Then: $(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE)"

.PHONY: undeploy-test
undeploy-test: ## Remove test workloads, scanner, and bps-test namespace
	$(KUBECTL) delete --ignore-not-found -f config/samples/scanner_bps_test_periodic.yaml
	$(KUBECTL) delete --ignore-not-found -f config/samples/scanner_bps_test.yaml
	$(KUBECTL) delete --ignore-not-found -f config/samples/test-workloads.yaml
	@if $(KUBECTL) api-resources 2>/dev/null | grep -q securitycontextconstraints; then \
		$(KUBECTL) adm policy remove-scc-from-user privileged -z default -n $(TEST_NAMESPACE) 2>/dev/null || true; \
	fi

.PHONY: scan
scan: install ## One-shot: deploy test workloads, run operator, show results, then stop
	@$(MAKE) deploy-scan
	@echo ""
	@echo "Starting operator..."
	@go run ./cmd/ --operator-namespace=$(OPERATOR_NAMESPACE) &>/tmp/bps-operator.log & \
		PID=$$!; \
		echo "Operator PID: $$PID"; \
		echo "Waiting for scan to complete..."; \
		for i in $$(seq 1 30); do \
			PHASE=$$($(KUBECTL) get bestpracticescanners test-scanner -n $(TEST_NAMESPACE) -o jsonpath='{.status.phase}' 2>/dev/null); \
			if [ "$$PHASE" = "Completed" ]; then break; fi; \
			sleep 1; \
		done; \
		echo ""; \
		$(MAKE) show-results; \
		kill $$PID 2>/dev/null; \
		wait $$PID 2>/dev/null; \
		echo ""; \
		echo "Operator stopped. Full logs at /tmp/bps-operator.log"

.PHONY: show-results
show-results: ## Show scan results from the cluster
	@echo "=== Scanner Status ==="
	@$(KUBECTL) get bestpracticescanners -n $(TEST_NAMESPACE) 2>/dev/null || echo "No scanners found"
	@echo ""
	@echo "=== Results ==="
	@$(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE) 2>/dev/null || echo "No results found"
	@echo ""
	@echo "=== Summary ==="
	@COMPLIANT=$$($(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE) -o jsonpath='{.items[?(@.spec.complianceStatus=="Compliant")].metadata.name}' 2>/dev/null | wc -w | tr -d ' '); \
	NONCOMPLIANT=$$($(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE) -o jsonpath='{.items[?(@.spec.complianceStatus=="NonCompliant")].metadata.name}' 2>/dev/null | wc -w | tr -d ' '); \
	SKIPPED=$$($(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE) -o jsonpath='{.items[?(@.spec.complianceStatus=="Skipped")].metadata.name}' 2>/dev/null | wc -w | tr -d ' '); \
	echo "  Compliant:     $$COMPLIANT"; \
	echo "  NonCompliant:  $$NONCOMPLIANT"; \
	echo "  Skipped:       $$SKIPPED"

.PHONY: show-failures
show-failures: ## Show details for all non-compliant results
	@echo "=== Non-Compliant Checks ==="
	@$(KUBECTL) get bestpracticeresults -n $(TEST_NAMESPACE) \
		-o jsonpath='{range .items[?(@.spec.complianceStatus=="NonCompliant")]}{.spec.checkName}{"\t"}{.spec.reason}{"\n"}{end}' 2>/dev/null \
		| column -t -s $$'\t' || echo "No results found"

.PHONY: show-scan-yaml
show-scan-yaml: ## Print the one-shot scanner CR YAML
	@cat config/samples/scanner_bps_test.yaml

.PHONY: show-periodic-scan-yaml
show-periodic-scan-yaml: ## Print the periodic scanner CR YAML
	@cat config/samples/scanner_bps_test_periodic.yaml

.PHONY: clean
clean: undeploy-test uninstall ## Remove everything: test workloads, CRDs, namespace
	$(KUBECTL) delete namespace $(TEST_NAMESPACE) --ignore-not-found
	@echo "Cleaned up."

##@ Coverage

.PHONY: coverage-html
coverage-html: test
	go tool cover -html=cover.out

##@ Help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
