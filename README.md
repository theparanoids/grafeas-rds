# grafeas-rds

> AWS RDS backend for Grafeas. This library can periodically refresh the IAM authentication token which is used as the password to connect to an AWS RDS service.

## Table of Contents

- [Background](#background)
- [Install](#install)
- [Usage](#usage)
- [Configuration](#configuration)
- [Contribute](#contribute)
- [License](#license)

## Background

[Grafeas](https://github.com/grafeas/grafeas) supports pluggable [storage backends](https://github.com/grafeas/grafeas#storage-backends),
and AWS RDS can be one of the options.
Furthermore, AWS RDS supports [IAM-based authentication](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/UsingWithRDS.IAMDBAuth.html),
which eliminates the needs to maintain a password,
including storing it, fetching it from the application, and rotating it periodically, etc.
However, the official documentation also states the following:

> Each token has a lifetime of 15 minutes.

As a result, we need a mechanism to refresh the token, hence this project.

## Install

This project is intended to be used as a library.

Import `github.com/theparanoids/grafeas-rds/rds` to use it.

Note that the Go version has to be >= `1.17` (see [go.mod](go.mod)).

## Usage

If the underlying database were PostgreSQL, the code would look like this:

```go
import (
    "log"

    "github.com/theparanoids/grafeas-rds/rds"
    "github.com/grafeas/grafeas/go/v1beta1/storage"
    "github.com/lib/pq"
)

func main() {
    provider := rds.NewGrafeasStorageProvider(
        &pq.Driver{},
        YourCredentialsCreator{},
        YourStorageCreator{},
    )
    if err := storage.RegisterStorageTypeProvider("rds_postgres", provider.Provide); err != nil {
        log.Fatalf("Error registering rds pgsql provider, %s", err)
    }
    // Set up and start the Grafeas server...
}
```

### Usage Notes

- Currently the configuration passed to `CredentialsCreator.Create` contains only
  [Athenz](https://github.com/AthenZ/athenz)-related fields;
  we welcome contributions to add support for any other mechanism.
- Regarding `StorageCreator`,
  we have an internal implementation to create a [grafeas-pqsql](https://github.com/grafeas/grafeas-pgsql) storage
  given a custom `driver.Connector`,
  and are actively working on upstreaming it.

## Configuration

A valid configuration file can be found [here](rds/config/testdata/valid.yaml);
it can be directly plugged into a configuration file for Grafeas server.

Some default values are also provided in [`config.go`](rds/config/config.go).

## Contribute

Please refer to [Contributing.md](Contributing.md) for information about how to get involved.
We welcome issues, questions, and pull requests.

## License

This project is licensed under the terms of the [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0) open source license. Please refer to [LICENSE](LICENSE) for the full terms.
