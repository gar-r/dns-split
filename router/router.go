package router

import (
	"regexp"
	"strings"

	"git.okki.hu/garric/dns-split/config"
	"github.com/rs/zerolog/log"
)

type Router struct {
	Config *config.Config
}

func (r *Router) Route(hostname string) string {
	server := r.Config.Dns
	for _, domain := range r.Config.Tunnel.Domains {
		if matchDomain(hostname, domain) {
			server = r.Config.Tunnel.Dns
			break
		}
	}
	log.Info().
		Str("hostname", hostname).
		Str("server", server).
		Msg("routing")
	return server
}

var domainRe = regexp.MustCompile(`^(\*\.|)([\w.-]+)$`)

func matchDomain(hostname, domain string) bool {
	m := domainRe.FindAllStringSubmatch(domain, -1)
	if len(m) > 0 {
		wildcard := m[0][1]
		suffix := m[0][2]
		if wildcard == "*." {
			return strings.HasSuffix(hostname, suffix)
		}
		return hostname == suffix
	}
	return false
}
