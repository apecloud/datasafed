repocli
=======

`repocli` is a companion tool to KubeBlocks' DataProtection module, which can operate various kinds of storage backends for uploading and downloading backup data.

## Build

`repocli` requires Go 1.21 or higher.

```bash
# Compile repocli binary, output to build/
make build
```

### Build Docker Image

```bash
# Build image
make build-docker-image
# Push image
make push-docker-image
```

If you want to build a multi-arch image, you need to set BUILDX_ENABLED=true. The default archs to build are `linux/amd64` and `linux/arm64`, and you can specify other archs through `BUILDX_PLATFORMS`.

```bash
export BUILDX_ENABLED=true
export BUILDX_PLATFORMS=darwin/arm64,linux/arm64
make build-docker-image
make push-docker-image
```

## Usage

### Configuration

Before using `repocli`, you need to prepare a configuration file, mainly the storage backend configuration information. `repocli` integrates [rclone](https://rclone.org/) as the driver for accessing remote storage. Most configuration items are passed through to `rclone`, and a few configuration items are handled by `repocli` itself. The configuration file is in the .ini format.

```ini
[storage]

#############################
# handled by rclone
#############################
# Refers to https://rclone.org/docs/#configure
type = s3
provider = AWS
env_auth = false
access_key_id = XXX
secret_access_key = YYY
region = us-east-1

#############################
# handled by repocli
#############################
# All directory parameters are relative to root
root = mybucket/
```

`repocli` loads the configuration from `/etc/repocli/repocli.conf` by default, but you can override this with the `-c/--conf` parameter.

#### Special Environment Variables

`REPOCLI_LOCAL_BACKEND_PATH`: If the user sets this variable to a local directory, `repocli` will ignore the backend in the configuration file and use a [local backend](https://rclone.org/local/) pointing to the specified path.

`REPOCLI_BACKEND_BASE_PATH`: This variable is a path, and `repocli`` will prepend this path to all remote path parameters, meaning that all operations will be restricted to the specified directory.

```bash
# Without REPOCLI_BACKEND_BASE_PATH:
# It creates the file at /hello.txt
echo "hello" | repocli push - hello.txt

# With REPOCLI_BACKEND_BASE_PATH:
# It creates the file at a/b/c/hello.txt
export REPOCLI_BACKEND_BASE_PATH=a/b/c
echo "hello" | repocli push - hello.txt
```

### Commands

See [docs/repocli.md](docs/repocli.md) for details and examples.
