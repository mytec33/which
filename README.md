# Go-Based Implementation of the `which` Command

This is an attempt to recreate a *nix utility across multiple platforms with the intention the Go version could be a drop in replacement for the native version.

Windows was attempted, however there myriad ways in which to determine if a binary file is executable, so for now, this cross platform will be restriced to *nix based systems.

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

### GNU version I've encountered on Fedora.

|Flags              |Implemented|
|:------------------|:----:|
|`--all -a`         |❌ ✅ |
|`--read-alias, i`  |❌    |
|`--skip-alias`     |❌    |
|`--read-functions` |❌    |
|`--skip-functions` |❌    |
|`--skip-dot`       |❌    |
|`--skip-tilde`     |❌    |
|`--show-dot`       |❌    |
|`--show-tilde`     |❌    |
|`--tty-only  `     |❌    |
|`--version, -v, -V`|❌    |
|`--help`           |❌    |

### OpenBSD

|Flags              |Implemented|
|:------------------|:----:|
|`-a flag`          |✅    |
| 

