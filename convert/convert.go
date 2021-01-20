package convert

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/tmccombs/hcl2json/convert"
)

// Options for the hcl2json converter.
type Options struct {
	Simplify bool
}

// Bytes takes the contents of an HCL file, as bytes, and converts
// them into a JSON representation of the HCL file.
func Bytes(bytes []byte, filename string, options Options) ([]byte, error) {
	return convert.Bytes(bytes, filename, convert.Options{
		Simplify: options.Simplify,
	})
}

// File takes the path to an HCL file and converts its contents to its JSON
// representation.
func File(path string, options Options) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return Reader(f, path, options)
}

// Reader reads HCL file contents from r and converts it to its JSON
// representation.
func Reader(r io.Reader, filename string, options Options) ([]byte, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Bytes(buf, filename, options)
}
