# driftcheck

> Detect configuration drift between deployed services and their declared state.

---

## Installation

```bash
go install github.com/yourusername/driftcheck@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/driftcheck.git
cd driftcheck
go build -o driftcheck .
```

---

## Usage

Point `driftcheck` at your declared state file and a running service endpoint to compare them:

```bash
driftcheck --declared ./config/service.yaml --target https://api.example.com/config
```

**Example output:**

```
[DRIFT DETECTED] service: payment-api
  - expected replicas: 3, got: 2
  - expected env.LOG_LEVEL: "info", got: "debug"

[OK] service: auth-service
```

### Flags

| Flag          | Description                              | Default  |
|---------------|------------------------------------------|----------|
| `--declared`  | Path to declared state file (YAML/JSON)  | required |
| `--target`    | URL or path of the deployed service      | required |
| `--format`    | Output format: `text`, `json`            | `text`   |
| `--fail`      | Exit with non-zero code on drift         | `false`  |

---

## CI Integration

```bash
driftcheck --declared ./config/service.yaml --target $SERVICE_URL --fail
```

Use `--fail` to break pipelines when drift is detected.

---

## License

MIT © 2024 yourusername