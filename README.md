# Go-Based Implementation of the `which` Command

This is an attempt to recreate a `which` utility in Go across multiple platforms and architectures with the intention that the Go version could be a drop in replacement for the native version.

### Not attempting

Windows was attempted, however there are myriad ways in which to determine if a binary file is executable, so for now, this cross platform effort will be restriced to *nix based systems.

GNU which is quite a different animal from the version found on macOS, OpenBSD, and Ubuntu. Some of the additional options look good, but I'm not to try and mix two different versions of which into one program.

## Platforms

| Platform            | Builds|Tests|Architecture|
| :-------------------| :----:|:---:|:----------:|
| `darwin`            | ✅    |✅   |`arm64`          |
| `linux - Fedora`    | ✅    |❌   |`amd64`, `arm64` |
| `linux - Ubuntu`    | ✅    |✅   |`amd64`, `arm64` |
| `openbsd`           | ✅    |✅   |`amd64`          |

## `which` Implementations

### Darwin

|Flags               |Implemented|
|:-------------------|:----:|
|`-a flag`           |✅    |
|`-s flag`           |✅    |

### OpenBSD

|Flags              |Implemented|
|:------------------|:----:|
|`-a flag`          |✅    |
| 

### Ubuntu

|Flags               |Implemented|
|:-------------------|:----:|
|`-a flag`           |✅    |
|`-s flag`           |✅    |