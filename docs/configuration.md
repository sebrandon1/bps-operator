# Configuration

All CRs use API group **`bps.openshift.io/v1alpha1`**.

Create a `BestPracticeScanner` CR to configure scanning:

```yaml
apiVersion: bps.openshift.io/v1alpha1
kind: BestPracticeScanner
metadata:
  name: my-scanner
  namespace: my-app
spec:
  targetNamespace: my-app
  labelSelector:
    matchLabels:
      app: my-workload
  scanInterval: "10m"
  checks:
    - access-control-sys-admin-capability-check
    - lifecycle-liveness-probe
  suspend: false
```

## Sample CRs

| File | Purpose |
|---|---|
| `config/samples/scanner_sample.yaml` | Periodic scan of all pods in a namespace |
| `config/samples/scanner_label_selector.yaml` | Scan only pods matching a label selector |
| `config/samples/scanner_networking.yaml` | Run networking checks only |

## Field Reference

- **targetNamespace**: Which namespace to scan. Defaults to the CR's own namespace.
- **labelSelector**: Filter pods by labels. Omit to scan all pods in the namespace.
- **scanInterval**: How often to re-scan (e.g., "5m", "1h30m", "10s"). Must be a valid Go duration string. Omit for a one-shot scan. The API validates this format at admission time, rejecting invalid durations like "5mins" or "1 hour".
- **checks**: Run only specific checks by name. Omit to run all checks.
- **suspend**: Set to `true` to pause periodic scanning.

## Admission Validation

The operator includes an optional validating webhook for `BestPracticeScanner` resources. It is **disabled by default**: the default `install.yaml` deploys no webhook Service, certificate, or `ValidatingWebhookConfiguration`, so the default install works without cert-manager. When enabled, the webhook runs on every `CREATE` or `UPDATE` and enforces:

- **`spec.scanInterval`**: Must be a valid Go duration string (e.g. `5m`, `1h30m`). Empty (one-shot) is allowed. The CRD schema also enforces this format, so invalid durations are rejected even without the webhook.
- **`spec.checks`**: Every named check must be a known check identifier. Run `make list-checks` to see valid names. Without the webhook, unknown names are accepted by the API.

Invalid resources are rejected at admission time with a descriptive error message.

### Enabling the webhook

Deploy the operator first (`make deploy` or `install.yaml`), then opt in with one of the certificate providers:

```bash
# cert-manager (must already be installed)
make deploy-webhook

# OpenShift (service-ca operator issues the certificate)
make deploy-webhook WEBHOOK_VARIANT=openshift
```

This applies the webhook Service, certificate resources and `ValidatingWebhookConfiguration` from `config/webhook/<variant>/`, waits for the `webhook-server-cert` Secret, and adds `--enable-webhook` to the operator Deployment. The operator reports ready only once the webhook server is serving. `make install-webhook-yaml` renders the same resources to `install-webhook.yaml` (apply it, then run `make enable-webhook`).

The webhook uses `failurePolicy: Ignore`, so scanner create/update keeps working if the webhook Service is unavailable (invalid specs are not rejected during that time). To remove it, run `make undeploy-webhook`.
