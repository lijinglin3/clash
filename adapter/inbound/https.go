package inbound

import (
	"net"
	"net/http"
	"net/netip"

	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/context"
)

// NewHTTPS receive CONNECT request and return ConnContext
func NewHTTPS(request *http.Request, conn net.Conn) *context.ConnContext {
	metadata := parseHTTPAddr(request)
	metadata.Type = constant.HTTPCONNECT
	if ip, port, err := parseAddr(conn.RemoteAddr()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = constant.Port(port)
	}
	if addrPort, err := netip.ParseAddrPort(conn.LocalAddr().String()); err == nil {
		metadata.OriginDst = addrPort
	}
	return context.NewConnContext(conn, metadata)
}
