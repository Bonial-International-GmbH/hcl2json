package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	opts := &options{
		pretty:      true,
		pattern:     "**/*.tf",
		parallelism: 10,
	}

	cmd := &cobra.Command{
		Use:           "hcl2json [paths...]",
		Short:         "Converts HCL files to JSON.",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return opts.run(cmd.InOrStdin(), cmd.OutOrStdout(), args)
		},
	}

	cmd.Flags().BoolVar(&opts.simplify, "simplify", opts.simplify, "If true attempt to simply expressions which don't contain any variables or unknown functions")
	cmd.Flags().BoolVar(&opts.pretty, "pretty", opts.pretty, "If true the resulting JSON is pretty-printed")
	cmd.Flags().StringVar(&opts.pattern, "pattern", opts.pattern, "Glob pattern to match files in directory args")
	cmd.Flags().IntVar(&opts.parallelism, "parallelism", opts.parallelism, "Number of files to convert in parallel")

	return cmd
}

func main() {
	cmd := newRootCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
