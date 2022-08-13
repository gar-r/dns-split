package hook

import (
	"net"

	"git.okki.hu/garric/dns-split/config"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vishvananda/netlink"
)

func ExecNetlinkHook(res *dns.Msg, dnsServer string, cfg *config.Config) {
	if skipNetlink(dnsServer, cfg) {
		return
	}

	link, err := netlink.LinkByName(cfg.Tunnel.Netlink.Dev)
	if err != nil {
		log.Error().Err(err).
			Str("dev", cfg.Tunnel.Netlink.Dev).
			Msg("failed to get link")
		return
	}

	for _, ans := range res.Answer {
		processAnswer(ans, link)
	}
}

func skipNetlink(dnsServer string, cfg *config.Config) bool {
	return !cfg.Tunnel.Netlink.Enabled || dnsServer != cfg.Tunnel.Dns
}

func processAnswer(ans dns.RR, link netlink.Link) {
	ip := extractIP(ans)
	if ip == nil {
		return
	}

	err := setRoute(ip, link)
	if err != nil {
		log.Error().Err(err).
			Str("hostname", ans.Header().Name).
			Str("link", link.Attrs().Name).
			Stringer("ip", ip).
			Msg("failed to set link")
	}
}

func extractIP(ans dns.RR) net.IP {
	switch ans.Header().Rrtype {
	case dns.TypeA:
		rec := ans.(*dns.A)
		return rec.A
	case dns.TypeAAAA:
		rec := ans.(*dns.AAAA)
		return rec.AAAA
	}
	return nil
}

func setRoute(ip net.IP, link netlink.Link) error {
	dst := netlink.NewIPNet(ip)
	route := &netlink.Route{
		Dst:       dst,
		LinkIndex: link.Attrs().Index,
		Scope:     netlink.SCOPE_HOST,
	}

	routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		return err
	}

	for _, r := range routes {
		if dst.IP.Equal(r.Dst.IP) {
			return replaceRoute(route)
		}
	}

	return addRoute(route)
}

func addRoute(route *netlink.Route) error {
	routeEvent(route).Msg("adding new route")
	return netlink.RouteAdd(route)
}

func replaceRoute(route *netlink.Route) error {
	routeEvent(route).Msg("replacing route")
	return netlink.RouteReplace(route)
}

func routeEvent(route *netlink.Route) *zerolog.Event {
	return log.Info().
		Stringer("dst", route.Dst).
		Int("scope", int(route.Scope)).
		Int("link", route.LinkIndex)
}
