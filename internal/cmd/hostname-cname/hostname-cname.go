package hostname_cname

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/dpc-sdp/uri-rewriter/internal/helpers"
	"github.com/urfave/cli/v3"
)

func RunHostnameCNAME(ctx context.Context, c *cli.Command) error {
	if c.NArg() != 1 {
		return cli.Exit("exactly one URI argument is required", 2)
	}

	raw := strings.TrimSpace(c.Args().Get(0))
	u, err := url.Parse(raw)
	if err != nil {
		return cli.Exit(err.Error(), 2)
	}
	if u.Host == "" {
		return cli.Exit("URI must include a hostname", 2)
	}

	origHost := u.Hostname()
	origPort := u.Port()

	level := c.Int("cname-level")
	if level < 1 {
		level = 1
	}

	// Resolve CNAME chain for the hostname
	chain, err := helpers.CnameChain(origHost)
	if err != nil {
		// If DNS fails entirely, leave unchanged but signal failure
		return cli.Exit(err.Error(), 3)
	}

	// If there is no CNAME, do not alter
	newHost := origHost
	if len(chain) > 0 {
		idx := level - 1
		if idx >= len(chain) {
			idx = len(chain) - 1
		}
		newHost = strings.TrimSuffix(chain[idx], ".") // normalize trailing dot
	}

	// Rebuild host (preserve original port if any)
	if origPort != "" {
		u.Host = net.JoinHostPort(newHost, origPort)
	} else {
		u.Host = newHost
	}

	// Output only the rewritten URI
	fmt.Fprintln(c.Writer, u.String())
	return nil
}
