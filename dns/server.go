package dns

import (
	"errors"
	"net"

	"github.com/lijinglin3/clash/common/sockopt"
	"github.com/lijinglin3/clash/context"
	"github.com/lijinglin3/clash/log"

	"github.com/miekg/dns"
)

var (
	address string
	server  = &Server{}

	dnsDefaultTTL uint32 = 600
)

type Server struct {
	*dns.Server
	handler handler
}

// ServeDNS implement D.Handler ServeDNS
func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg, err := handlerWithContext(s.handler, r)
	if err != nil {
		dns.HandleFailed(w, r)
		return
	}
	msg.Compress = true
	w.WriteMsg(msg)
}

func handlerWithContext(handler handler, msg *dns.Msg) (*dns.Msg, error) {
	if len(msg.Question) == 0 {
		return nil, errors.New("at least one question is required")
	}

	ctx := context.NewDNSContext(msg)
	return handler(ctx, msg)
}

func (s *Server) setHandler(handler handler) {
	s.handler = handler
}

func ReCreateServer(addr string, resolver *Resolver, mapper *ResolverEnhancer) {
	if addr == address && resolver != nil {
		handler := newHandler(resolver, mapper)
		server.setHandler(handler)
		return
	}

	if server.Server != nil {
		server.Shutdown()
		server = &Server{}
		address = ""
	}

	if addr == "" {
		return
	}

	var err error
	defer func() {
		if err != nil {
			log.Errorln("Start DNS server error: %s", err.Error())
		}
	}()

	_, port, err := net.SplitHostPort(addr)
	if port == "0" || port == "" || err != nil {
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return
	}

	p, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return
	}

	err = sockopt.UDPReuseaddr(p)
	if err != nil {
		log.Warnln("Failed to Reuse UDP Address: %s", err)

		err = nil
	}

	address = addr
	handler := newHandler(resolver, mapper)
	server = &Server{handler: handler}
	server.Server = &dns.Server{Addr: addr, PacketConn: p, Handler: server}

	go func() {
		server.ActivateAndServe()
	}()

	log.Infoln("DNS server listening at: %s", p.LocalAddr().String())
}
