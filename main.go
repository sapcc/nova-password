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
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/utils/client"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/howeyc/gopass"
	"github.com/kayrus/putty"
	env "github.com/sapcc/cloud-env"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

const MaxKeySize = 10240

var Version string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          "nova-password <server-name>|<server-id> [<server-name>|<server-id>...]",
	Short:        "Get the admin password for an OpenStack server",
	SilenceUsage: true,
	Version:      Version,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Usage()
			return fmt.Errorf("server name has to be provided")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// get the wait timeout
		wait := viper.GetUint("wait")
		// get the quiet flag
		quiet := viper.GetBool("quiet")
		// Convert Unix path type to Windows path type, when necessary
		keyPath := filepath.FromSlash(viper.GetString("private-key-path"))

		if quiet == false {
			log.Printf("private-key-path: %s\n", keyPath)
		}

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
				pass, err := getKeyPass(quiet)
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
					pass, err := getKeyPass(quiet)
					if err != nil {
						return err
					}

					privateKey, err = ssh.ParseRawPrivateKeyWithPassphrase(key, pass)
					if err != nil {
						return err
					}
				} else {
					if pkerr != nil {
						// if there was an error in putty format, print it as well
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
				err = processServer(client, server, v, wait, quiet)
				if err != nil {
					log.Printf("Error getting the password for the %q server: %s", server, err)
					errors = append(errors, fmt.Errorf("Error getting the password for the %q server: %s", server, err))
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
	RootCmd.PersistentFlags().UintP("wait", "w", 0, "wait for the password timeout in seconds")
	RootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet (no extra output)")
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("private-key-path", RootCmd.PersistentFlags().Lookup("private-key-path"))
	viper.BindPFlag("wait", RootCmd.PersistentFlags().Lookup("wait"))
	viper.BindPFlag("quiet", RootCmd.PersistentFlags().Lookup("quiet"))

}

// newComputeV2 creates a ServiceClient that may be used with the v2 compute
// package.
func newComputeV2() (*gophercloud.ServiceClient, error) {
	ao, err := clientconfig.AuthOptions(nil)
	if err != nil {
		return nil, err
	}

	provider, err := openstack.NewClient(ao.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	if viper.GetBool("debug") {
		provider.HTTPClient = http.Client{
			Transport: &client.RoundTripper{
				Rt:     &http.Transport{},
				Logger: &client.DefaultLogger{},
			},
		}
	}

	err = openstack.Authenticate(provider, *ao)
	if err != nil {
		return nil, err
	}

	return openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
}

func processServer(client *gophercloud.ServiceClient, server string, privateKey *rsa.PrivateKey, wait uint, quiet bool) error {
	tmp := server
	// Verify whether UUID was provided. If the name was provided, resolve the server name
	_, err := uuid.Parse(server)
	if err != nil {
		server, err = servers.IDFromName(client, server)
		if err != nil {
			return err
		}
		if quiet == false {
			log.Printf("Resolved %q server name to the %q uuid", tmp, server)
		}
	}

	var res servers.GetPasswordResult
	if wait > 0 {
		if quiet == false {
			log.Printf("Waiting for %d seconds to get the password", wait)
		}
		// Wait for the encrypted server password
		res, err = waitForPassword(client, server, wait)
		if err != nil {
			return err
		}
	} else {
		// Get the encrypted server password
		res = servers.GetPassword(client, server)
	}

	if res.Err != nil {
		return res.Err
	}

	// Decrypt the password
	pwd, err := res.ExtractPassword(privateKey)
	if err != nil {
		return err
	}

	if quiet == false {
		fmt.Printf("%q instance password: %s\n", tmp, pwd)
	} else {
		fmt.Printf("%s\n", pwd)
	}

	return nil
}

func waitForPassword(c *gophercloud.ServiceClient, id string, secs uint) (servers.GetPasswordResult, error) {
	var res servers.GetPasswordResult
	err := gophercloud.WaitFor(int(secs), func() (bool, error) {
		var err error
		res = servers.GetPassword(c, id)
		if res.Err != nil {
			return false, res.Err
		}

		pass, err := res.ExtractPassword(nil)
		if err != nil {
			return false, err
		}

		if pass == "" {
			return false, nil
		}

		return true, nil
	})

	return res, err
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

func getKeyPass(quiet bool) ([]byte, error) {
	pass := env.Get("NOVA_PASSWORD_KEY_PASSWORD")

	if pass == "" {
		if quiet == true {
			return nil, fmt.Errorf(`Private key is encrypted with the password, please set the "NOVA_PASSWORD_KEY_PASSWORD" environment variable`)
		}

		log.Print("Private key is encrypted with the password")
		fmt.Print("Enter the password: ")
		return gopass.GetPasswd()
	}

	return []byte(pass), nil
}
