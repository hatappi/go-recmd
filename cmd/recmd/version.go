package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
)

type versionOption struct {
	short bool
}

func newVersionCmd() *cobra.Command {
	opts := &versionOption{
		short: false,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			short, err := cmd.Flags().GetBool("short")
			if err != nil {
				return err
			}

			if version == "" {
				version = "None"
			}

			if short {
				fmt.Println(version)
			} else {
				fmt.Printf("Version %s (git-%s)\n", version, commit)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&opts.short, "short", "s", opts.short, "show short version")

	return cmd
}
