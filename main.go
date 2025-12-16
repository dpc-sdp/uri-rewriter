package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/miekg/dns"
	cli "github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "uri-rewriter",
		Usage: "Rewrite URI hostnames using CNAME records",
		Commands: []*cli.Command{
			{
				Name:  "hostname-cname",
				Usage: "Replace the URI's hostname with its CNAME at a given level",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "cname-level",
						Aliases: []string{"l"},
						Value:   1,
						Usage:   "CNAME level to use (1 = first CNAME in the chain)",
					},
				},
				ArgsUsage: "<uri>",
				Action:    runHostnameCNAME,
			},
		},
	}

	ctx := context.Background()
	if err := app.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

func runHostnameCNAME(ctx context.Context, c *cli.Command) error {
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
	chain, err := cnameChain(origHost)
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

// cnameChain returns the ordered CNAME targets starting from the given name,
// following CNAME records until none is present or a safety limit is reached.
func cnameChain(name string) ([]string, error) {
	cfg, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		// Fallback to common public resolvers if resolv.conf isn't available
		cfg = &dns.ClientConfig{Servers: []string{"1.1.1.1:53", "8.8.8.8:53"}}
	} else {
		// Append :port if missing
		for i, s := range cfg.Servers {
			if !strings.Contains(s, ":") {
				cfg.Servers[i] = net.JoinHostPort(s, cfg.Port)
			}
		}
	}

	client := &dns.Client{}
	current := name
	var chain []string
	// Prevent infinite loops
	for i := 0; i < 20; i++ {
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(current), dns.TypeCNAME)

		var resp *dns.Msg
		var lastErr error
		for _, server := range cfg.Servers {
			r, _, err := client.Exchange(m, server)
			if err != nil {
				lastErr = err
				continue
			}
			resp = r
			lastErr = nil
			break
		}
		if lastErr != nil {
			return chain, lastErr
		}
		if resp == nil || resp.Rcode != dns.RcodeSuccess {
			// No useful response
			return chain, nil
		}

		// Look for CNAME in answers
		found := false
		for _, ans := range resp.Answer {
			if cname, ok := ans.(*dns.CNAME); ok {
				target := cname.Target
				chain = append(chain, target)
				current = target
				found = true
				break
			}
		}
		if !found {
			// No CNAME for this name
			return chain, nil
		}
	}
	return chain, nil
}
