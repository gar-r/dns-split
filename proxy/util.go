package proxy

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

func withPort(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) < 2 {
		return fmt.Sprintf("%s:%d", addr, defaultDnsPort)
	}
	return addr
}

func fail(w dns.ResponseWriter, req *dns.Msg) {
	m := &dns.Msg{}
	m.SetRcode(req, dns.RcodeServerFailure)
	err := w.WriteMsg(m)
	if err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func warnQSize(r *dns.Msg) {
	log.Warn().
		Int("size", len(r.Question)).
		Stringer("msg", r).
		Msg("question size")
}
