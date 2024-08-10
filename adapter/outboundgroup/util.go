package outboundgroup

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/lijinglin3/clash/constant"
)

func addrToMetadata(rawAddress string) (addr *constant.Metadata, err error) {
	host, port, err := net.SplitHostPort(rawAddress)
	if err != nil {
		err = fmt.Errorf("addrToMetadata failed: %w", err)
		return
	}

	ip := net.ParseIP(host)
	p, _ := strconv.ParseUint(port, 10, 16)
	if ip == nil {
		addr = &constant.Metadata{
			Host:    host,
			DstIP:   nil,
			DstPort: constant.Port(p),
		}
		return
	} else if ip4 := ip.To4(); ip4 != nil {
		addr = &constant.Metadata{
			Host:    "",
			DstIP:   ip4,
			DstPort: constant.Port(p),
		}
		return
	}

	addr = &constant.Metadata{
		Host:    "",
		DstIP:   ip,
		DstPort: constant.Port(p),
	}
	return
}

func tcpKeepAlive(c net.Conn) {
	if tcp, ok := c.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * time.Second)
	}
}
