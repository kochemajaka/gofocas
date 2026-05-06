# Contributing

## Getting started

1. Fork and clone the repository.
2. Install Go 1.22 or later.
3. Run tests in stub mode (no FANUC SDK required):

```sh
make test
```

## Branches and commits

- Use [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `docs:`, `refactor:`, `test:`.
- One logical change per commit.

## Adding a new reader

1. Add the domain type to the appropriate file (or create a new one if it is a new concern).
2. Add the method to `readers.go`.
3. Add the corresponding method to `internal/fwlib32/types.go` (`Binder` interface + stub implementation).
4. Add the CGo implementation in `internal/fwlib32/fwlib32.go`.
5. Add a unit test in `client_test.go` using `fakeBinder`.

## Running integration tests

Integration tests require a real FANUC controller accessible over Ethernet:

```sh
FOCAS_TEST_ADDR=192.168.1.1 go test -tags='focas_cgo,integration' ./...
```

## Do not commit

- `fwlib32.h`, `*.so`, `*.dll`, `*.lib` — see `.gitignore`.
- Any file containing FANUC proprietary content.
