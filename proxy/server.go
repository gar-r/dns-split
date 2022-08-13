package proxy

import (
	"git.okki.hu/garric/dns-split/config"
	"git.okki.hu/garric/dns-split/hook"
	"git.okki.hu/garric/dns-split/router"

	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

const defaultDnsProto = "udp"
const defaultDnsPort = 53

type Server struct {
	Config *config.Config
	Router *router.Router
}

func (s *Server) ListenAndServe() error {
	dns.HandleFunc(".", s.HandleFunc)
	server := dns.Server{
		Addr: withPort(s.Config.Addr),
		Net:  defaultDnsProto,
	}
	return server.ListenAndServe()
}

func (s *Server) HandleFunc(w dns.ResponseWriter, r *dns.Msg) {
	question := s.extractQuestion(r)
	if question == nil {
		return
	}

	dnsServer := s.Router.Route(question.Name)
	res, err := query(dnsServer, r)
	if err != nil {
		log.Error().Err(err).
			Str("question", question.Name).
			Str("server", dnsServer).
			Msg("query failed")
		fail(w, r)
		return
	}
	s.afterResolve(res, dnsServer)

	writeResponse(w, res)

}

func writeResponse(w dns.ResponseWriter, res *dns.Msg) {
	err := w.WriteMsg(res)
	if err != nil {
		log.Error().Err(err).
			Stringer("msg", res).
			Msg("unable to write response")
	}
}

func (s *Server) afterResolve(res *dns.Msg, dnsServer string) {
	hook.ExecNetlinkHook(res, dnsServer, s.Config)
}

func (s *Server) extractQuestion(r *dns.Msg) *dns.Question {
	switch len(r.Question) {
	case 0:
		warnQSize(r)
		return nil
	default:
		warnQSize(r)
		fallthrough
	case 1:
		return &r.Question[0]
	}
}
