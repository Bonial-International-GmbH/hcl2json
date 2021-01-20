# hcl2json

This is a modified version of [hcl2json](https://github.com/tmccombs/hcl2json)
with an added support for converting multiple files and directories
concurrently.

Directories are scanned recursively for files having the extension configured
via the `--extension` flag.

## Compatibility notices

This project is based on [hcl2json](https://github.com/tmccombs/hcl2json)
v0.3.1 which does not include
[#20](https://github.com/tmccombs/hcl2json/pull/20). Our usecase is parsing and
processing lots of `*.tf` files where the added array wrapping of block values
is tedious to work with. If you depend on the block wrapping, please use the
original project instead.
