package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/howeyc/gopass"
	"github.com/kayrus/putty"
	env "github.com/sapcc/cloud-env"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

const MaxKeySize = 10240

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "nova-password <server-name>|<server-id> [<server-name>|<server-id>...]",
	Short:        "Get the admin password for an OpenStack server",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Usage()
			return fmt.Errorf("server name has to be provided")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Convert Unix path type to Windows path type, when necessary
		keyPath := filepath.FromSlash(viper.GetString("private-key-path"))

		fmt.Printf("private-key-path: %s\n", keyPath)

		// Read the key
		key, err := readKey(keyPath)
		if err != nil {
			return err
		}

		var privateKey interface{}

		puttyKey, pkerr := putty.New(key)
		if pkerr == nil {
			if puttyKey.Algo != "ssh-rsa" {
				return fmt.Errorf("unsupported key type %s\nOnly RSA PKCS #1 v1.5 is supported by OpenStack", puttyKey.Algo)
			}
			// parse putty key
			if puttyKey.Encryption != "none" {
				// If the key is encrypted, decrypt it
				log.Print("Private key is encrypted with the password")
				fmt.Print("Enter the password: ")
				pass, err := gopass.GetPasswd()
				if err != nil {
					return err
				}
				privateKey, err = puttyKey.ParseRawPrivateKey(pass)
				if err != nil {
					return err
				}
			} else {
				privateKey, err = puttyKey.ParseRawPrivateKey(nil)
				if err != nil {
					return err
				}
			}
		} else {
			// Parse the pem key
			privateKey, err = ssh.ParseRawPrivateKey(key)
			if err != nil {
				if err.Error() != "ssh: no key found" {
					// If the key is encrypted, decrypt it
					log.Print("Private key is encrypted with the password")
					fmt.Print("Enter the password: ")
					pass, err := gopass.GetPasswd()
					if err != nil {
						return err
					}

					privateKey, err = ssh.ParseRawPrivateKeyWithPassphrase(key, pass)
					if err != nil {
						return err
					}
				} else {
					if pkerr != nil {
						log.Print(pkerr)
					}
					return err
				}
			}
		}

		var errors []error

		switch v := privateKey.(type) {
		// Only RSA PKCS #1 v1.5 is supported by OpenStack
		case *rsa.PrivateKey:
			// Initialize the compute client
			client, err := newComputeV2()
			if err != nil {
				return err
			}

			for _, server := range args {
				err = processServer(client, server, v)
				if err != nil {
					log.Printf("%s", err)
					errors = append(errors, err)
				}
			}
		default:
			return fmt.Errorf("unsupported key type %T\nOnly RSA PKCS #1 v1.5 is supported by OpenStack", v)
		}

		if len(errors) > 0 {
			return fmt.Errorf("%v", errors)
		}

		return nil
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func main() {
	initRootCmdFlags()
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initRootCmdFlags() {
	// Get the current user home dir
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Set default key path
	defaultKeyPath := filepath.FromSlash(usr.HomeDir + "/.ssh/id_rsa")

	// debug flag
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "print out request and response objects")
	RootCmd.PersistentFlags().StringP("private-key-path", "i", defaultKeyPath, "a path to the RSA private key (PuTTY and OpenSSH formats)")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("private-key-path", RootCmd.PersistentFlags().Lookup("private-key-path"))
}

// newComputeV2 creates a ServiceClient that may be used with the v2 compute
// package.
func newComputeV2() (*gophercloud.ServiceClient, error) {
	ao, err := env.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	if viper.GetBool("debug") {
		client.HTTPClient = http.Client{
			Transport: &clients.LogRoundTripper{
				Rt: &http.Transport{},
			},
		}
	}

	err = openstack.Authenticate(client, ao)
	if err != nil {
		return nil, err
	}

	return openstack.NewComputeV2(client, gophercloud.EndpointOpts{
		Region: env.Get("OS_REGION_NAME"),
	})
}

func processServer(client *gophercloud.ServiceClient, server string, privateKey *rsa.PrivateKey) error {
	// Verify whether UUID was provided. If the name was provided, resolve the server name
	_, err := uuid.Parse(server)
	if err != nil {
		fmt.Printf("server: %s", server)
		server, err = servers.IDFromName(client, server)
		if err != nil {
			fmt.Println()
			return err
		}
		fmt.Printf(" (%s)\n", server)
	} else {
		fmt.Printf("server: %s\n", server)
	}

	// Get the encrypted server password
	req := servers.GetPassword(client, server)
	if req.Err != nil {
		return req.Err
	}

	// Decrypt the password
	pwd, err := req.ExtractPassword(privateKey)
	if err != nil {
		return err
	}
	fmt.Printf("Decrypted compute instance password: %s\n", pwd)

	return nil
}

func readKey(path string) ([]byte, error) {
	f, err := os.Open(filepath.FromSlash(path))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	if size > MaxKeySize {
		return nil, fmt.Errorf("Invalid key size: %d bytes", size)
	}

	if size == 0 {
		// force to use "MaxKeySize", when detected file size is 0 (e.g. /dev/stdin)
		size = MaxKeySize
	}

	key := make([]byte, size)

	// Read the key
	_, err = f.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}
