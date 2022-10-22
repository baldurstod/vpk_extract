[![Go Report Card](https://goreportcard.com/badge/github.com/baldurstod/vpk_extract)](https://goreportcard.com/badge/github.com/baldurstod/vpk_extract)

# vpk_extract
A `.vpk` (Valve Pak) incremental extractor. Check the CRCs inside the `.vpk` against a list of know CRCs.

# Build

```
go build vpk_extract
```

or

```
SET GOOS=linux&&SET GOARCH=amd64&&go build vpk_extract
```

# Command line

```
vpk_extract [-c COMMAND] -i INPUT_VPK.vpk -o OUTPUT_FOLDER [GLOB_PATTERNS]
```
COMMAND is either `extract` or `crc` and default to `extract` if omitted

GLOB_PATTERNS default to * if omitted

`-i` is required is COMMAND is `extract`

# Examples

Extract everything from a .vpk
```
vpk_extract -i INPUT_VPK.vpk -o OUTPUT_FOLDER
```

Extract only the models and materials subdirectories
```
vpk_extract -i INPUT_VPK.vpk -o OUTPUT_FOLDER models/* materials/*
```

Compute the CRC of existing files in the output folder
```
vpk_extract -c crc -o OUTPUT_FOLDER
```
