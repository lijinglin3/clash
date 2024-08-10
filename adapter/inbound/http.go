package inbound

import (
	"net"
	"net/netip"

	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/context"
	"github.com/lijinglin3/clash/transport/socks5"
)

// NewHTTP receive normal http request and return HTTPContext
func NewHTTP(target socks5.Addr, source, originTarget net.Addr, conn net.Conn) *context.ConnContext {
	metadata := parseSocksAddr(target)
	metadata.NetWork = constant.TCP
	metadata.Type = constant.HTTP
	if ip, port, err := parseAddr(source); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = constant.Port(port)
	}
	if originTarget != nil {
		if addrPort, err := netip.ParseAddrPort(originTarget.String()); err == nil {
			metadata.OriginDst = addrPort
		}
	}
	return context.NewConnContext(conn, metadata)
}
