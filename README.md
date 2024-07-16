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

* The private key corresponding to the public key, used to create a compute instance, is required
* Only RSA PKCS #1 v1.5 is supported by OpenStack

## TLS options

* `OS_CACERT` - environment variable with a path to custom CA certificate.
* `OS_INSECURE` - skip endpoint TLS certificate validation. Set to `true` **only if you are otherwise convinced of the OpenStack endpoint's authenticity**.

## Windows

```sh
.\openrc.ps1
.\nova-password.exe --private-key-path C:\Users\user\key.pem my-server
# or
.\nova-password.exe 717433dc-4c2e-4d62-9467-6dd3715b2c6c server-name
# or
.\nova-password.exe my-server -i C:\Users\user\.ssh\putty.ppk
```
