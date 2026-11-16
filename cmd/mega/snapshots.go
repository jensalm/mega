package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/michaelquigley/mega/internal/config"
	"github.com/michaelquigley/mega/internal/snapshot"
)

func init() {
	rootCmd.AddCommand(snapshotsCmd())
}

// snapshotsCmd lists the configuration snapshots, and with diff, names what
// differs between one and the live configuration and prints the commands
// that would put its copies back. restoring is the operator's: it writes
// nothing into a harness's configuration, and nothing at all.
func snapshotsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "List the configuration snapshots mega keeps of each harness",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			store, err := openSnapshots()
			if err != nil {
				return err
			}
			return listSnapshots(c.OutOrStdout(), store)
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "diff NAME",
		Short: "Name the files that differ between a snapshot (or newest) and the live configuration, and print the copy commands",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			store, err := openSnapshots()
			if err != nil {
				return err
			}
			return diffSnapshot(c.OutOrStdout(), store, args[0])
		},
	})
	return cmd
}

// openSnapshots is the store under the data directory the config names, to
// read; it never takes a snapshot.
func openSnapshots() (*snapshot.Store, error) {
	cfg, err := config.Load(configPath, dataPath, listenAddr)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return snapshot.NewStore(snapshot.Options{Dir: cfg.SnapshotsPath()}), nil
}

func listSnapshots(w io.Writer, store *snapshot.Store) error {
	all, err := store.List()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		fmt.Fprintf(w, "no configuration snapshots in '%s' yet\n", store.Dir())
		return nil
	}
	fmt.Fprintf(w, "configuration snapshots in '%s', newest first:\n", store.Dir())
	for _, m := range all {
		fmt.Fprintf(w, "  %s  %s  %s, %d files%s\n", m.Name, m.Taken.Local().Format(time.DateTime), m.Reason, len(m.Files), versions(m))
	}
	fmt.Fprintln(w, "'mega snapshots diff NAME' (or newest) names what differs from the live configuration")
	return nil
}

// versions is a snapshot's recorded harness versions, as ", claude 2.1.325".
func versions(m *snapshot.Manifest) string {
	var out []string
	for name, v := range m.Versions {
		out = append(out, name+" "+v)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return ""
	}
	return ", " + strings.Join(out, ", ")
}

func diffSnapshot(w io.Writer, store *snapshot.Store, name string) error {
	m, err := store.Find(name)
	if err != nil {
		return err
	}
	diff, err := store.Diff(m)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "snapshot '%s', taken '%s' at '%s', against the live configuration:\n", m.Name, m.Taken.Local().Format(time.DateTime), m.Reason)
	if len(diff) == 0 {
		fmt.Fprintln(w, "  nothing differs")
		return nil
	}
	var commands []string
	for _, d := range diff {
		fmt.Fprintf(w, "  %-22s %s\n", d.Change, d.Live)
		if d.Restore != "" {
			commands = append(commands, d.Restore)
		}
	}
	if len(commands) > 0 {
		fmt.Fprintln(w, "\nto put the snapshot's copies back, run what you choose of these; mega runs none of them:")
		for _, c := range commands {
			fmt.Fprintln(w, "  "+c)
		}
	}
	return nil
}

// printSnapshots says where the configuration snapshots are and when the
// newest was taken, for 'mega version': names and times only.
func printSnapshots(cfg *config.Config) {
	if cfg.Snapshots.Keep == 0 {
		fmt.Println("configuration snapshots: none are taken (snapshots.keep is 0)")
		return
	}
	store := snapshot.NewStore(snapshot.Options{Dir: cfg.SnapshotsPath()})
	all, err := store.List()
	switch {
	case err != nil:
		fmt.Printf("configuration snapshots: in '%s', which could not be read: %v\n", store.Dir(), err)
	case len(all) == 0:
		fmt.Printf("configuration snapshots: none yet, in '%s'\n", store.Dir())
	default:
		fmt.Printf("configuration snapshots: newest '%s' at '%s', %d kept, in '%s'\n", all[0].Taken.Local().Format(time.DateTime), all[0].Reason, len(all), store.Dir())
	}
}
