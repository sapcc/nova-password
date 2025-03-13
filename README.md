# Usage

```sh
Get the admin password for an OpenStack server

Usage:
  nova-password <server-name>|<server-id> [<server-name>|<server-id>...] [flags]

Flags:
  -d, --debug                     print out request and response objects
  -h, --help                      help for nova-password
  -i, --private-key-path string   a path to the RSA private key (PuTTY and OpenSSH formats) (default "~/.ssh/id_rsa")
  -q, --quiet                     quiet (no extra output)
  -v, --version                   version for nova-password
  -w, --wait uint                 wait for the password timeout in seconds
```

## Prerequisites

Download and unzip the latest release for your operating system from the [releases](../../releases/latest) page.

* The private key corresponding to the public key used to create a compute instance is required.
* Only RSA PKCS #1 v1.5 is supported by OpenStack.
* **OpenStack environment variables for authentication must be set.** These are typically sourced from an `openrc` file, which includes credentials like your OpenStack username, project, and authentication endpoint. Without these environment variables, the tool will not be able to authenticate with OpenStack.

For reference, you can find simple examples of `openrc` files for Linux/macOS and Windows below:

- [Example `openrc.sh` for Linux/macOS](openrc.sh)
- [Example `openrc.ps1` for Windows](openrc.ps1)

## TLS options

* `OS_CACERT` - environment variable with a path to a custom CA certificate.
* `OS_INSECURE` - skip endpoint TLS certificate validation. Set to `true` **only if you are otherwise convinced of the OpenStack endpoint's authenticity**.

## Windows

Before using `nova-password` on Windows, make sure to source the OpenStack environment variables by running the `openrc.ps1` script.

```sh
.\openrc.ps1
.\nova-password.exe --private-key-path C:\Users\user\key.pem my-server
# or
.\nova-password.exe 717433dc-4c2e-4d62-9467-6dd3715b2c6c server-name
# or
.\nova-password.exe my-server -i C:\Users\user\.ssh\putty.ppk
```

## Linux / macOS

Before using `nova-password` on Linux or macOS, ensure that the OpenStack environment variables are sourced by running the `openrc.sh` script.

```sh
source ./openrc.sh
./nova-password --private-key-path ~/.ssh/id_rsa my-server
# or
./nova-password 717433dc-4c2e-4d62-9467-6dd3715b2c6c server-name
# or
./nova-password my-server -i ~/.ssh/putty.ppk
```
