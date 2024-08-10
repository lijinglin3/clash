package inbound

import (
	"net"
	"net/netip"

	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/context"
	"github.com/lijinglin3/clash/transport/socks5"
)

// NewSocket receive TCP inbound and return ConnContext
func NewSocket(target socks5.Addr, conn net.Conn, source constant.Type) *context.ConnContext {
	metadata := parseSocksAddr(target)
	metadata.NetWork = constant.TCP
	metadata.Type = source
	if ip, port, err := parseAddr(conn.RemoteAddr()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = constant.Port(port)
	}
	if addrPort, err := netip.ParseAddrPort(conn.LocalAddr().String()); err == nil {
		metadata.OriginDst = addrPort
	}
	return context.NewConnContext(conn, metadata)
}
