package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Bonial-International-GmbH/hcl2json/convert"
	"github.com/spf13/cobra"
)

type options struct {
	simplify    bool
	pretty      bool
	ext         string
	concurrency int
}

func newRootCommand() *cobra.Command {
	opts := &options{
		pretty:      true,
		ext:         ".tf",
		concurrency: 10,
	}

	cmd := &cobra.Command{
		Use:           "hcl2json [paths...]",
		Short:         "Converts HCL files to JSON.",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			return run(cmd.InOrStdin(), cmd.OutOrStdout(), args, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.simplify, "simplify", opts.simplify, "If true attempt to simply expressions which don't contain any variables or unknown functions")
	cmd.Flags().BoolVar(&opts.pretty, "pretty", opts.pretty, "If true the resulting JSON is pretty-printed")
	cmd.Flags().StringVar(&opts.ext, "extension", opts.ext, "Convert files with extension when recursively scanning directories")
	cmd.Flags().IntVar(&opts.concurrency, "concurrency", opts.concurrency, "Number of files to convert concurrently")

	return cmd
}

func main() {
	cmd := newRootCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(in io.Reader, out io.Writer, args []string, opts *options) error {
	converted, err := convertArgs(in, args, opts)
	if err != nil {
		return err
	}

	if !opts.pretty {
		_, err = out.Write(converted)
	} else {
		var indented bytes.Buffer

		err = json.Indent(&indented, converted, "", "    ")
		if err != nil {
			return fmt.Errorf("failed to indent file: %w", err)
		}

		_, err = indented.WriteTo(out)
	}

	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}

func convertArgs(in io.Reader, args []string, opts *options) ([]byte, error) {
	convOpts := convert.Options{
		Simplify: opts.simplify,
	}

	if len(args) == 0 {
		return convert.Reader(in, "", convOpts)
	}

	if len(args) == 1 {
		path := args[0]

		if path == "" || path == "-" {
			return convert.Reader(in, "", convOpts)
		}

		isDir, err := isDirectory(path)
		if err != nil {
			return nil, err
		} else if !isDir {
			return convert.File(path, convOpts)
		}
	}

	files, err := collectFiles(args, opts.ext)
	if err != nil {
		return nil, err
	}

	return convert.Bulk(opts.concurrency, files, convOpts)
}

func isDirectory(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}
	return fi.IsDir(), nil
}

func collectFiles(paths []string, ext string) ([]string, error) {
	var filePaths []string

	for _, path := range paths {
		isDir, err := isDirectory(path)
		if err != nil {
			return nil, err
		} else if isDir {
			matches, err := glob(path, ext)
			if err != nil {
				return nil, err
			}

			filePaths = append(filePaths, matches...)
		} else {
			filePaths = append(filePaths, path)
		}
	}

	return filePaths, nil
}

// glob recursively scans dir for files with the given extension and returns
// them. The extension must start with a dot.
func glob(dir string, ext string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() && filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
