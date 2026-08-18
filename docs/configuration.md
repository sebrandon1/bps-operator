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
