package main

import (
	"fmt"
	"os"

	"code.google.com/p/gcfg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var config Config

type Config struct {
	Siad struct {
		ApiPort        uint16
		RpcPort        uint16
		NoBootstrap    bool
		ConfigFilename string
	}
}

// confilgFilenameDefault checks multiple directories for a config file and
// loads the first one it finds. "" is returned if no config file is found.
func configFilenameDefault() string {
	// Try home folder.
	home, err := homedir.Dir()
	if err == nil {
		// Check home/.config/config
		filename := home + "/.config/sia/config"
		if _, err := os.Stat(filename); err == nil {
			return filename
		}

		// Check home/.sia/config
		filename = home + "/.sia/config"
		if _, err := os.Stat(filename); err == nil {
			return filename
		}

		// Check home/.sia.conf
		filename = home + "/.sia.conf"
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	// Try /etc/sia.conf
	filename := "etc/sia.conf"
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	return ""
}

// startEnvironment calls createEnvironment(), which will handle everything
// else.
func startEnvironment(cmd *cobra.Command, args []string) {
	_, err := CreateEnvironment(config.Siad.RpcPort, config.Siad.ApiPort, config.Siad.NoBootstrap)
	if err != nil {
		fmt.Println(err)
	}
}

// homeFolder displays a user's home directory to them, which is nice for
// windows users since they might not know which directory is their home
// directory.
func homeFolder(cmd *cobra.Command, args []string) {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(home)
	}
}

// Prints version information about Sia Daemon.
func version(cmd *cobra.Command, args []string) {
	fmt.Println("Sia Daemon v0.1.0")
}

func main() {
	root := &cobra.Command{
		Use:   os.Args[0],
		Short: "Sia Daemon v0.1.0",
		Long:  "Sia Daemon v0.1.0",
		Run:   startEnvironment,
	}

	root.AddCommand(&cobra.Command{
		Use:   "home-folder",
		Short: "Print home folder",
		Long:  "Print the filepath of the home folder as seen by the binary.",
		Run:   homeFolder,
	})

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information about the Sia Daemon",
		Run:   version,
	})

	// Add flag defaults, which have the lowest priority. Every value will be
	// set.
	defaultConfigFile := configFilenameDefault()
	root.PersistentFlags().Uint16VarP(&config.Siad.ApiPort, "api-port", "a", 9980, "Which port is used to communicate with the user.")
	root.PersistentFlags().Uint16VarP(&config.Siad.RpcPort, "rpc-port", "r", 9988, "Which port is used when talking to other nodes on the network.")
	root.PersistentFlags().BoolVarP(&config.Siad.NoBootstrap, "no-bootstrap", "n", false, "Disable bootstrapping on this run.")
	root.PersistentFlags().StringVarP(&config.Siad.ConfigFilename, "config-file", "c", defaultConfigFile, "Tell siad where to load the config file.")

	// Load the config file, which has the middle priorty. Only values defined
	// in the config file will be set.
	if config.Siad.ConfigFilename != "" {
		err := gcfg.ReadFileInto(&config, config.Siad.ConfigFilename)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Execute wil over-write any flags set by the config file, but only if the
	// user specified them manually.
	root.Execute()
}
