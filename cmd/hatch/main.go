package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jazicorn/hatch/internal/config"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "hatch",
		Short: "hatch — local quiz and kata engine",
		Long:  "hatch is a local-first quiz and kata engine powered by LLMs and vector search.",
	}

	root.AddCommand(newConfigCmd())
	root.AddCommand(newIngestCmd())
	root.AddCommand(newSourcesCmd())
	return root
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage hatch configuration",
	}
	cmd.AddCommand(newConfigInitCmd())
	return cmd
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Write default config to ~/.hatch/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.Init()
		},
	}
}
