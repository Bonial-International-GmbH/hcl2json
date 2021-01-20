package convert

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBulk(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	tests := []struct {
		name        string
		files       []string
		expected    []byte
		expectedErr error
	}{
		{
			name: "simple module",
			files: []string{
				"testdata/module1/main.tf",
				"testdata/module1/variables.tf",
			},
			expected: readFile(t, "testdata/module1.golden"),
		},
		{
			name: "empty module",
			files: []string{
				"testdata/module2/main.tf",
			},
			expected: readFile(t, "testdata/module2.golden"),
		},
		{
			name:     "no files",
			files:    []string{},
			expected: []byte(`{}`),
		},
		{
			name: "nonexistent file",
			files: []string{
				"testdata/module1/main.tf",
				"testdata/module1/nonexistent.tf",
			},
			expectedErr: errors.New("failed to open file: open testdata/module1/nonexistent.tf: no such file or directory"),
		},
		{
			name: "invalid file",
			files: []string{
				"testdata/module1/main.tf",
				"testdata/module1/variables.tf",
				"testdata/module2/main.tf",
				"testdata/module3/invalid.tf",
			},
			expectedErr: errors.New(`parse config: [testdata/module3/invalid.tf:1,1-8: Argument or block definition required; An argument or block definition is required here. To set an argument, use the equals sign "=" to introduce the argument value.]`),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			converted, err := Bulk(10, test.files, Options{})
			if test.expectedErr != nil {
				require.Error(err)
				require.Equal(test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(err)
				require.JSONEq(string(test.expected), string(converted))
			}
		})
	}
}

func readFile(t *testing.T, path string) []byte {
	buf, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return buf
}
