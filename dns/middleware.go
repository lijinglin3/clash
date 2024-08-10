package dns

import (
	"net"
	"strings"
	"time"

	"github.com/lijinglin3/clash/common/cache"
	"github.com/lijinglin3/clash/component/fakeip"
	"github.com/lijinglin3/clash/component/trie"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/context"
	"github.com/lijinglin3/clash/log"

	"github.com/miekg/dns"
)

type (
	handler    func(ctx *context.DNSContext, r *dns.Msg) (*dns.Msg, error)
	middleware func(next handler) handler
)

func withHosts(hosts *trie.DomainTrie) middleware {
	return func(next handler) handler {
		return func(ctx *context.DNSContext, r *dns.Msg) (*dns.Msg, error) {
			q := r.Question[0]

			if !isIPRequest(q) {
				return next(ctx, r)
			}

			record := hosts.Search(strings.TrimRight(q.Name, "."))
			if record == nil {
				return next(ctx, r)
			}

			ip := record.Data.(net.IP)
			msg := r.Copy()

			if v4 := ip.To4(); v4 != nil && q.Qtype == dns.TypeA {
				rr := &dns.A{}
				rr.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: dnsDefaultTTL}
				rr.A = v4

				msg.Answer = []dns.RR{rr}
			} else if v6 := ip.To16(); v6 != nil && q.Qtype == dns.TypeAAAA {
				rr := &dns.AAAA{}
				rr.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: dnsDefaultTTL}
				rr.AAAA = v6

				msg.Answer = []dns.RR{rr}
			} else {
				return next(ctx, r)
			}

			ctx.SetType(context.DNSTypeHost)
			msg.SetRcode(r, dns.RcodeSuccess)
			msg.Authoritative = true
			msg.RecursionAvailable = true

			return msg, nil
		}
	}
}

func withMapping(mapping *cache.LruCache) middleware {
	return func(next handler) handler {
		return func(ctx *context.DNSContext, r *dns.Msg) (*dns.Msg, error) {
			q := r.Question[0]

			if !isIPRequest(q) {
				return next(ctx, r)
			}

			msg, err := next(ctx, r)
			if err != nil {
				return nil, err
			}

			host := strings.TrimRight(q.Name, ".")

			for _, ans := range msg.Answer {
				var ip net.IP
				var ttl uint32

				switch a := ans.(type) {
				case *dns.A:
					ip = a.A
					ttl = a.Hdr.Ttl
					if !ip.IsGlobalUnicast() {
						continue
					}
				case *dns.AAAA:
					ip = a.AAAA
					ttl = a.Hdr.Ttl
					if !ip.IsGlobalUnicast() {
						continue
					}
				default:
					continue
				}

				if ttl < 1 {
					ttl = 1
				}
				mapping.SetWithExpire(ip.String(), host, time.Now().Add(time.Second*time.Duration(ttl)))
			}

			return msg, nil
		}
	}
}

func withFakeIP(fakePool *fakeip.Pool) middleware {
	return func(next handler) handler {
		return func(ctx *context.DNSContext, r *dns.Msg) (*dns.Msg, error) {
			q := r.Question[0]

			host := strings.TrimRight(q.Name, ".")
			if fakePool.ShouldSkipped(host) {
				return next(ctx, r)
			}

			switch q.Qtype {
			case dns.TypeAAAA, dns.TypeSVCB, dns.TypeHTTPS:
				return handleMsgWithEmptyAnswer(r), nil
			}

			if q.Qtype != dns.TypeA {
				return next(ctx, r)
			}

			rr := &dns.A{}
			rr.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: dnsDefaultTTL}
			ip := fakePool.Lookup(host)
			rr.A = ip
			msg := r.Copy()
			msg.Answer = []dns.RR{rr}

			ctx.SetType(context.DNSTypeFakeIP)
			setMsgTTL(msg, 1)
			msg.SetRcode(r, dns.RcodeSuccess)
			msg.Authoritative = true
			msg.RecursionAvailable = true

			return msg, nil
		}
	}
}

func withResolver(resolver *Resolver) handler {
	return func(ctx *context.DNSContext, r *dns.Msg) (*dns.Msg, error) {
		ctx.SetType(context.DNSTypeRaw)
		q := r.Question[0]

		// return a empty AAAA msg when ipv6 disabled
		if !resolver.ipv6 && q.Qtype == dns.TypeAAAA {
			return handleMsgWithEmptyAnswer(r), nil
		}

		msg, err := resolver.Exchange(r)
		if err != nil {
			log.Debugln("[DNS Server] Exchange %s failed: %v", q.String(), err)
			return msg, err
		}
		msg.SetRcode(r, msg.Rcode)
		msg.Authoritative = true

		return msg, nil
	}
}

func compose(middlewares []middleware, endpoint handler) handler {
	length := len(middlewares)
	h := endpoint
	for i := length - 1; i >= 0; i-- {
		middleware := middlewares[i]
		h = middleware(h)
	}

	return h
}

func newHandler(resolver *Resolver, mapper *ResolverEnhancer) handler {
	middlewares := []middleware{}

	if resolver.hosts != nil {
		middlewares = append(middlewares, withHosts(resolver.hosts))
	}

	if mapper.mode == constant.DNSFakeIP {
		middlewares = append(middlewares, withFakeIP(mapper.fakePool))
		middlewares = append(middlewares, withMapping(mapper.mapping))
	}

	return compose(middlewares, withResolver(resolver))
}
