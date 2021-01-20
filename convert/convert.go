package convert

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/tmccombs/hcl2json/convert"
)

type Options struct {
	Simplify bool
}

func Bytes(bytes []byte, filename string, options Options) ([]byte, error) {
	return convert.Bytes(bytes, filename, convert.Options{
		Simplify: options.Simplify,
	})
}

func File(path string, options Options) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return Reader(f, path, options)
}

func Reader(r io.Reader, filename string, options Options) ([]byte, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Bytes(buf, filename, options)
}
