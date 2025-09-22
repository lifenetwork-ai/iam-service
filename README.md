[![codecov](https://codecov.io/gh/lifenetwork-ai/iam-service/branch/dev/graph/badge.svg?token=G3CKBzhKbk)](https://codecov.io/gh/lifenetwork-ai/iam-service)

## Setup development tools

### Install golangci-lint
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Swagger
- `make swagger`
- http://localhost:8080/swagger/index.html

### How to run linter
- `make lint`

### How to run service
- `make build`
- `make run`

### How to stop service
- `make stop`

### How to test

#### Unit test
```bash
make test
```

#### Integration test
```bash
make test-integration
```

#### Coverage
```bash
make cover-ucases-integration
```
### How to generate mock for unittest
1. Install `mockgen`:
```bash
go install go.uber.org/mock/mockgen@latest
```
2. Generate mock files:
```bash
make mocks
```
