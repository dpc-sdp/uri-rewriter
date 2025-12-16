# uri-rewriter

A small Go command-line utility to rewrite the hostname component of a URI using DNS CNAME records.

## Features

- Uses `github.com/urfave/cli/v2` for command definition
- Subcommand: `hostname-cname`
- Flag: `--cname-level, -l` (default: 1)
- Accepts a single URI argument
- Performs a DNS lookup of the URI's hostname and follows its CNAME chain
- Replaces the hostname with the CNAME at the requested level
- Writes only the rewritten URI to stdout (no extra output)

## Install

```bash
git clone https://github.com/nicksantamaria/uri-rewriter.git
cd uri-rewriter
go build .
# The binary will be at ./uri-rewriter
```

Or install directly with Go:

```bash
go install github.com/nicksantamaria/uri-rewriter@latest
```

## Usage

```bash
uri-rewriter hostname-cname [--cname-level N] <uri>
```

- `--cname-level, -l`: The step in the CNAME chain to use for replacement.
  - `1` = the first CNAME target (closest to the original hostname)
  - If the requested level exceeds the available CNAME chain length, the last available CNAME target is used.
  - If there is no CNAME for the hostname, the URI is output unchanged.

### Examples

Assume DNS has the following CNAME chain:

```
www.example.com -> www.example.net -> origin.example.org
```

- Level 1 (first CNAME):

```bash
uri-rewriter hostname-cname -l 1 "https://www.example.com/path?x=1"
# https://www.example.net/path?x=1
```

- Level 2 (second CNAME):

```bash
uri-rewriter hostname-cname -l 2 "https://www.example.com/path?x=1"
# https://origin.example.org/path?x=1
```

- With port preservation:

```bash
uri-rewriter hostname-cname -l 1 "https://www.example.com:8443/path"
# https://www.example.net:8443/path
```

## Notes

- Only the rewritten URI is printed to stdout. Errors (e.g., invalid URI, DNS failures) are written to stderr with a non-zero exit code.
- The tool follows the system resolver configuration (`/etc/resolv.conf`) to query CNAME records. If it is unavailable, it falls back to public resolvers (Cloudflare `1.1.1.1` and Google `8.8.8.8`).
- Trailing dots in DNS names (e.g., `example.com.`) are removed in the output for URI compatibility.

## License

MIT
