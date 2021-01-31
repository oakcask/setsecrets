# setsecrets

`setsecrets` is an wrapper command that maps secrets to environment variables then execve.

## Usage

```
$ setsecrets <options> [<env>=<secret name>] [--] command...
```

### Examples

```
$ setsecrets --gcp POSTGRES_PASSWORD=secrets/pgpass/versions/latest -- env
```

### Options

* `--gcp` enables provider gcp.

## Providers

Currently supported only gcp.

### gcp

`gcp` provider utilizes [Secret Manager](https://cloud.google.com/secret-manager)
on Google Cloud Platform. In this provider, secret name is resource ID for
the secret version. It's string like `projects/*/secrets/*/versions/*`.

The command automatically detects current project and supply project ID to
the resource ID for the secret version,
when secret name does not have prefix `projects/`.

## Build

```
go build ./cmd/...
```
