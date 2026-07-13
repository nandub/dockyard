# Build and Commands

## Authoritative Local Commands

The Makefile is the local command wrapper source.

Observed targets:

- `make tidy`
- `make tidy-check`
- `make fmt`
- `make fmt-check`
- `make test`
- `make build`
- `make dev-build`
- `make verify`
- `make clean`

## Build

```sh
make build
```

Underlying build form:

```sh
go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

## Verify

```sh
make verify
```

Observed behavior: tidy check, format check, and tests.

## CI Comparison

CI is defined in `.github/workflows/ci.yml`. Release artifact behavior is defined in `.github/workflows/release.yml`.

Historical detail: `.ai/onboarding/reports/build.md` and `.ai/onboarding/reports/ci.md`.
