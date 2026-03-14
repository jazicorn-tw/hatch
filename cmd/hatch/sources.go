package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jazicorn/hatch/internal/config"
	"github.com/jazicorn/hatch/internal/store/sqlite"
)

func newSourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "Manage ingestion sources",
	}
	cmd.AddCommand(newSourcesListCmd())
	cmd.AddCommand(newSourcesRemoveCmd())
	return cmd
}

func newSourcesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured ingestion sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("sources list: %w", err)
			}
			if len(cfg.Sources) == 0 {
				fmt.Println("No sources configured. Add one to ~/.hatch/config.yaml under 'sources:'.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tTYPE\tPATH")
			for _, s := range cfg.Sources {
				fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Type, s.Path)
			}
			return w.Flush()
		},
	}
}

func newSourcesRemoveCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a source from config and delete its indexed data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSourcesRemove(cmd.Context(), name)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Name of the source to remove (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runSourcesRemove(ctx context.Context, name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("sources remove: load config: %w", err)
	}

	// Find and remove the source from the slice.
	found := false
	filtered := cfg.Sources[:0]
	for _, s := range cfg.Sources {
		if s.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, s)
	}
	if !found {
		return fmt.Errorf("sources remove: source %q not found", name)
	}

	// Delete indexed data from the store.
	dbPath := cfg.DBPath
	if dbPath == "" {
		home, _ := os.UserHomeDir()
		dbPath = filepath.Join(home, ".hatch", "hatch.db")
	}
	st, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("sources remove: open store: %w", err)
	}
	defer st.Close()

	if err := st.DeleteBySource(ctx, name); err != nil {
		return fmt.Errorf("sources remove: delete from store: %w", err)
	}

	// Rewrite config without the removed source.
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	home, _ := os.UserHomeDir()
	v.AddConfigPath(filepath.Join(home, ".hatch"))
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("sources remove: read config: %w", err)
	}
	v.Set("sources", filtered)
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("sources remove: write config: %w", err)
	}

	fmt.Printf("Removed source %q and deleted its indexed data.\n", name)
	return nil
}
