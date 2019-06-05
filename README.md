# Usage

## Prerequisites

* The private key corresponding to the public key, used to create a compute instance, is required
* Only RSA PKCS #1 v1.5 is supported by OpenStack

## Windows

```sh
.\openrc.ps1
.\nova-password.exe --private-key-path C:\Users\user\key.pem my-server
# or
.\nova-password.exe 717433dc-4c2e-4d62-9467-6dd3715b2c6c server-name
# or
.\nova-password.exe my-server -i C:\Users\user\.ssh\putty.ppk
```
