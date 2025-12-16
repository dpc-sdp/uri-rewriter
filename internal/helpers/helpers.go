package helpers

import (
	"net"
	"strings"

	"github.com/miekg/dns"
)

// cnameChain returns the ordered CNAME targets starting from the given name,
// following CNAME records until none is present or a safety limit is reached.
func CnameChain(name string) ([]string, error) {
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
