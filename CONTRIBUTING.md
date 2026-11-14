# Contributing to asbuilt

Issues and pull requests are welcome. asbuilt is one static Go binary with no third-party dependencies, so keep changes focused on the collectors, the rule engine and the evidence every finding carries.

## Checks

CI runs on every push and on every pull request against `main` (`.github/workflows/ci.yml`), on Linux, Windows and macOS, with Go 1.26 and Go 1.27:

```bash
go vet ./...
go test ./...
go build -o asbuilt ./cmd/asbuilt
```

The module path is `github.com/Balanov/asbuilt` and `go.mod` requires Go 1.26 or newer. The module has no third-party dependencies, so there is no `go.sum` to update.

The Linux job also builds the released binary and runs it without root:

```bash
go build -ldflags "-s -w" -o asbuilt ./cmd/asbuilt
./asbuilt --version
./asbuilt rules | head -20
./asbuilt check --json | head -40
```

Keep every step working on a runner without root: the collectors read only what the kernel exposes unprivileged, and no finding about an unreadable SMBIOS is asserted.

CI renders the brand assets and checks the result:

```bash
go run ./tools/brandgen -out /tmp/brand
test -s /tmp/brand/logo-512.png
```

## Build

```bash
go build -trimpath -ldflags "-s -w" -o asbuilt ./cmd/asbuilt
```

or install the tagged module version:

```bash
go install github.com/Balanov/asbuilt/cmd/asbuilt@latest
```

The release workflow builds the same binary with `CGO_ENABLED=0` for linux-amd64, linux-arm64, darwin-amd64, darwin-arm64 and windows-amd64, and publishes `checksums.txt` next to every binary.

The landing page under `site/` is a separate Vite application with its own `package.json` and lockfile; it is not part of the Go build:

```bash
cd site
npm ci
npm run build
```

## Run it without a server

The repository ships a generated fixture that reproduces a misconfigured dual-socket server:

```bash
go run ./tools/fixturegen -out tests/fixtures/server
asbuilt --root tests/fixtures/server check
```

## Propose a change

- Open an issue first when the rule, the judgement or the measured impact is in question; open a pull request against `main` when it is not.
- Keep a pull request focused on one change and keep unrelated formatting out of it.
- Add or update tests for every behavioural change. A failing regression test that comes with the fix is the fastest way to get a change merged.
- Keep the collectors read-only and keep them separate from the rule engine: a collector fix must not change a verdict, and a new rule must not touch the parsers.
- Every rule carries the evidence it used and a source that justifies it. Keep `asbuilt explain <rule-id>` in step with the rule it describes.
- Keep `README.md` and `README_RU.md` consistent when you change documented behaviour.
- Keep provider and vendor mentions factual. No paid placements, and no links or documentation changes in exchange for test keys or credits.
- Do not commit secrets, API keys or credentials. Use environment variables and local fixtures instead.

## License

MIT, see `LICENSE` (Copyright (c) 2026 Anton Balanov). There is no separate contributor license agreement and no requirement to assign copyright.

## Security

Do not open a public issue for an unpatched vulnerability. Report it privately through a GitHub security advisory for `Balanov/asbuilt` (Security -> Report a vulnerability) instead, and include the affected version or commit and the steps that reproduce it.