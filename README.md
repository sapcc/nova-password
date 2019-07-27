# Synopsis
```
$ nova-password -h
Get the admin password for an OpenStack server

Usage:
  nova-password <server-name>|<server-id> [<server-name>|<server-id>...] [flags]

Flags:
  -d, --debug                     print out request and response objects
  -h, --help                      help for nova-password
  -i, --private-key-path string   a path to the RSA private key (PuTTY and OpenSSH formats) (default "/Users/uuu/.ssh/id_rsa")
  -v, --version                   print the nova-password version

```
# Usage

- source the `openrc` of the project where the server is running in the current shell to load the relevant ENV vars


## Prerequisites
- The user's (admin) password is provisioned when the server is created. We assume this hasn't been altered.
- During creation, the password is encrypted using your **public** SSH key, which is in your User Profile. 
- The **private** key corresponding to that public key, is used to decrypt the password.
  It is assumed to be your default private key, but you can specify another using `-i`. 
- Only RSA PKCS #1 v1.5 is supported by OpenStack (**elaborate**)

## Windows

```sh
. .\openrc.ps1
# obtain password for server "my-server" by name
.\nova-password.exe --private-key-path C:\Users\user\key.pem my-server
# obtain password for server by ID
.\nova-password.exe 717433dc-4c2e-4d62-9467-6dd3715b2c6c server-name
# using a PuTTY ppk
.\nova-password.exe my-server -i C:\Users\user\.ssh\putty.ppk
```
