package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name        string
		in          io.Reader
		opts        *options
		args        []string
		expected    []byte
		exactOutput bool
		expectedErr error
	}{
		{
			name: "convert module",
			args: []string{
				"convert/testdata/module1",
			},
			expected: []byte(`{"convert/testdata/module1/main.tf":{"resource":{"aws_security_group":{"sg":{"name":"${var.name}","vpc_id":"${var.vpc_id}"}},"aws_security_group_rule":{"egress-all":{"cidr_blocks":["0.0.0.0/0"],"description":"Allow all egress traffic","from_port":-1,"protocol":"all","security_group_id":"${aws_security_group.sg.id}","to_port":-1,"type":"egress"}}}},"convert/testdata/module1/variables.tf":{"variable":{"name":{"default":"sg","description":"Name of the security group"},"vpc_id":{"description":"ID of the VPC"}}},"convert/testdata/module1/nested/main.tf":{"resource":{"null_resource":{"null":{}}}}}`),
		},
		{
			name: "convert single file",
			args: []string{
				"convert/testdata/module1/variables.tf",
			},
			expected: []byte(`{"variable":{"name":{"default":"sg","description":"Name of the security group"},"vpc_id":{"description":"ID of the VPC"}}}`),
		},
		{
			name:     "single file from stdin (1)",
			in:       bytes.NewBuffer([]byte("variable \"name\" {}\n")),
			expected: []byte(`{"variable":{"name":{}}}`),
		},
		{
			name:     "single file from stdin (2)",
			args:     []string{"-"},
			in:       bytes.NewBuffer([]byte("variable \"name\" {}\n")),
			expected: []byte(`{"variable":{"name":{}}}`),
		},
		{
			name: "different extension",
			args: []string{
				"convert/testdata",
			},
			opts:     &options{ext: ".hcl"},
			expected: []byte(`{}`),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			in := test.in
			if in == nil {
				in = os.Stdout
			}

			opts := test.opts
			if opts == nil {
				opts = &options{
					pretty:      true,
					ext:         ".tf",
					concurrency: 10,
				}
			}

			var out bytes.Buffer
			err := run(in, &out, test.args, opts)
			if test.expectedErr != nil {
				require.Error(err)
				require.Equal(test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(err)
				if test.exactOutput {
					require.Equal(string(test.expected), out.String())
				} else {
					require.JSONEq(string(test.expected), out.String())
				}
			}
		})
	}
}
