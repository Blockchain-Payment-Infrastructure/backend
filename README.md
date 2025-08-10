# Project backend

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

Create a file `.env` with the following contents
```
DB_DATABASE=blockchainpay
DB_PASSWORD=testing123
DB_USERNAME=postgres
DB_PORT=5432
DB_HOST=localhost
GIN_MODE=debug
PORT=8080
```

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
