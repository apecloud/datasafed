datasafed
=========

`datasafed` is a companion tool to KubeBlocks' DataProtection module, which can operate various kinds of storage backends for uploading and downloading backup data.

## Build

`datasafed` requires Go 1.21 or higher.

```bash
# Compile datasafed binary, output to build/
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

Before using `datasafed`, you need to prepare a configuration file, mainly the storage backend configuration information. `datasafed` integrates [rclone](https://rclone.org/) as the driver for accessing remote storage. Most configuration items are passed through to `rclone`, and a few configuration items are handled by `datasafed` itself. The configuration file is in the .ini format.

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
# handled by datasafed
#############################
# All directory parameters are relative to root
root = mybucket/
```

`datasafed` loads the configuration from `/etc/datasafed/datasafed.conf` by default, but you can override this with the `-c/--conf` parameter.

#### Special Environment Variables

`DATASAFED_LOCAL_BACKEND_PATH`: If the user sets this variable to a local directory, `datasafed` will ignore the backend in the configuration file and use a [local backend](https://rclone.org/local/) pointing to the specified path.

`DATASAFED_BACKEND_BASE_PATH`: This variable is a path, and `datasafed` will prepend this path to all remote path parameters, meaning that all operations will be restricted to the specified directory.

```bash
# Without DATASAFED_BACKEND_BASE_PATH:
# It creates the file at /hello.txt
echo "hello" | datasafed push - hello.txt

# With DATASAFED_BACKEND_BASE_PATH:
# It creates the file at a/b/c/hello.txt
export DATASAFED_BACKEND_BASE_PATH=a/b/c
echo "hello" | datasafed push - hello.txt
```

### Commands

See [docs/datasafed.md](docs/datasafed.md) for details and examples.
