#!/usr/bin/env bash
set -euo pipefail

crd_name=${1:?usage: wait-for-crd-established.sh <crd-name>}
timeout_seconds=30

for ((attempt = 1; attempt <= timeout_seconds; attempt++)); do
	status=$(oc get "crd/${crd_name}" -o go-template='{{range .status.conditions}}{{if eq .type "Established"}}{{.status}}{{end}}{{end}}' 2>/dev/null || true)
	if [[ "$status" == "True" ]]; then
		echo "CRD ${crd_name} is established"
		exit 0
	fi
	sleep 1
done

echo "Timed out waiting for CRD ${crd_name} to become Established" >&2
oc get "crd/${crd_name}" -o yaml || true
exit 1
