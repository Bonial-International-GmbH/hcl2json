package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/tmccombs/hcl2json/convert"
)

type options struct {
	simplify    bool
	pretty      bool
	pattern     string
	parallelism int
}

type converter struct {
	opts options
	in   io.Reader
	out  io.Writer
}

func newConverter(in io.Reader, out io.Writer, opts options) *converter {
	return &converter{
		in:   in,
		out:  out,
		opts: opts,
	}
}

func (c *converter) run(paths []string) error {
	converted, err := c.convert(paths)
	if err != nil {
		return err
	}

	if !c.opts.pretty {
		_, err = c.out.Write(converted)
	} else {
		var indented bytes.Buffer

		err := json.Indent(&indented, converted, "", "    ")
		if err != nil {
			return fmt.Errorf("failed to indent file: %w", err)
		}

		_, err = indented.WriteTo(c.out)
	}

	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}

// convert converts HCL files to JSON.
func (c *converter) convert(paths []string) ([]byte, error) {
	opts := &convert.Options{
		Simplify: c.opts.simplify,
	}

	if len(paths) == 0 {
		return convertReader(c.in, "", opts)
	}

	if len(paths) == 1 {
		path := paths[0]

		if path == "" || path == "-" {
			return convertReader(c.in, "", opts)
		}

		isDir, err := isDirectory(path)
		if err != nil {
			return nil, err
		} else if !isDir {
			return convertFile(path, opts)
		}
	}

	filePaths, err := c.resolveFilePaths(paths)
	if err != nil {
		return nil, err
	}

	return c.convertFiles(filePaths, opts)
}

// convertReader reads HCL from r and converts it to JSON using opts. Path
// hints the origin file path to the converter for display in errors.
func convertReader(r io.Reader, path string, opts *convert.Options) ([]byte, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	converted, err := convert.Bytes(buf, path, *opts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert file: %w", err)
	}

	return converted, nil
}

// convertFile converts a single file using opts and returns the JSON bytes.
func convertFile(path string, opts *convert.Options) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return convertReader(f, path, opts)
}

func (c *converter) resolveFilePaths(paths []string) ([]string, error) {
	var filePaths []string

	for _, path := range paths {
		isDir, err := isDirectory(path)
		if err != nil {
			return nil, err
		} else if isDir {
			globPattern := filepath.Join(path, c.opts.pattern)

			matches, err := filepath.Glob(globPattern)
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

type result struct {
	filePath string
	raw      []byte
}

// convertFiles converts multiple files in parallel and returns the JSON
// representation of the files keyed by file path.
func (c *converter) convertFiles(filePaths []string, opts *convert.Options) ([]byte, error) {
	if len(filePaths) == 0 {
		return nil, nil
	}

	numWorkers := max(1, min(c.opts.parallelism, len(filePaths)))

	errCh := make(chan error, numWorkers)
	defer close(errCh)

	workCh := make(chan string)
	outCh := make(chan *result, numWorkers)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(&wg, workCh, errCh, outCh, opts)
	}

	go assignWork(&wg, workCh, outCh, filePaths)

	fileMap := make(map[string]json.RawMessage)

	// Collect conversion results.
	for {
		select {
		case res, ok := <-outCh:
			if !ok {
				// outCh was closed, we are done.
				return json.Marshal(fileMap)
			}

			// Treat result bytes as raw JSON to prevent double encoding.
			fileMap[res.filePath] = json.RawMessage(res.raw)
		case err := <-errCh:
			return nil, err
		}
	}
}

// assignWork places file paths into workCh and waits for completion of the
// wait group. Ultimately closes outCh to indicate that processing is finished.
func assignWork(wg *sync.WaitGroup, workCh chan<- string, outCh chan<- *result, filePaths []string) {
	defer close(outCh)

	for _, filePath := range filePaths {
		workCh <- filePath
	}

	close(workCh)

	wg.Wait()
}

// worker reads file paths from workCh, converts the files and writes the
// result to outCh. The workers stops on the first error it encounters and
// write it to errCh.
func worker(wg *sync.WaitGroup, workCh <-chan string, errCh chan<- error, outCh chan<- *result, opts *convert.Options) {
	defer wg.Done()

	for filePath := range workCh {
		converted, err := convertFile(filePath, opts)
		if err != nil {
			errCh <- err
			return
		}

		outCh <- &result{
			filePath: filePath,
			raw:      converted,
		}
	}
}
