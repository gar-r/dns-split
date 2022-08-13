package proxy

import (
	"fmt"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

func query(server string, msg *dns.Msg) (*dns.Msg, error) {
	client := &dns.Client{Net: defaultDnsProto}
	res, _, err := client.Exchange(msg, withPort(server))
	if err == nil {
		logAnswer(res, server)
	}
	return res, err
}

func logAnswer(msg *dns.Msg, server string) {
	for _, ans := range msg.Answer {
		contentStr := content(ans)
		log.Info().
			Str("server", server).
			Str("hostname", ans.Header().Name).
			Str("content", contentStr).
			Msg("answer")
	}
}

func content(ans dns.RR) string {
	switch ans.Header().Rrtype {
	case dns.TypeA:
		rec := ans.(*dns.A)
		return rec.A.String()
	case dns.TypeAAAA:
		rec := ans.(*dns.AAAA)
		return rec.AAAA.String()
	case dns.TypeMX:
		rec := ans.(*dns.MX)
		return rec.Mx
	case dns.TypeTXT:
		rec := ans.(*dns.TXT)
		return fmt.Sprintf("%v", rec.Txt)
	case dns.TypeSRV:
		rec := ans.(*dns.SRV)
		return rec.Target
	default:
		return ans.String()
	}
}
