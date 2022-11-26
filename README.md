### Prerequisites

- Go 1.18 (should still be backwards compatible with earlier versions)

### Running locally

1. From root of the repo
2. Run `docker-compose up` will start the dependencies and server on port 8080

### Running via docker

1. From root of the repo
2. Run `docker-compose up` will start the dependencies and server on port 8080

### Postman

The collections will need an environment setup with `scheme`, `port` and `host` variables setup with values of `http`, `8080` and `localhost` respectively.

### Run tests

Some of the integration tests use docker to spin up dependencies on demand (ie a postgres db) so just be aware that docker is needed to run the tests.

1. From root of the repo
2. Run `go test ./...`
