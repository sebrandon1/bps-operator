FROM golang:1.26 AS builder

WORKDIR /workspace

# Copy the checks library (local replace dependency)
COPY .checks-vendor/ /redhat-best-practices-for-k8s/checks/

COPY . .
# Adjust the replace directive for the container build context
RUN sed -i 's|=> .*|=> /redhat-best-practices-for-k8s/checks|' go.mod && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -o manager ./cmd/

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]
