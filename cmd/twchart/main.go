package main

import (
	"fmt"

	"github.com/calvinmclean/twchart/api"

	"github.com/spf13/cobra"
)

func main() {
	api := api.New()
	cmd := api.Command()

	// Enable data loading and storage setup for serve command
	cmd.PersistentPreRunE = func(c *cobra.Command, _ []string) error {
		if c.Name() != "serve" {
			return nil
		}

		storeFlag := c.Flag("store")
		if storeFlag != nil && storeFlag.Value.String() != "" {
			err := api.Setup(storeFlag.Value.String())
			if err != nil {
				return fmt.Errorf("error setting up storage: %w", err)
			}
		}

		dirFlag := c.Flag("dir")
		if dirFlag == nil || dirFlag.Value.String() == "" {
			return nil
		}

		return api.Load(dirFlag.Value.String())
	}

	// Add custom flags to serve command
	for _, c := range cmd.Commands() {
		if c.Name() != "serve" {
			continue
		}

		c.Flags().String("dir", "", "directory to read data from")
		c.Flags().String("store", "", "filename for JSON KV store")
	}

	err := cmd.Execute()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
