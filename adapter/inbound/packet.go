package inbound

import (
	"net"
	"net/netip"

	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/transport/socks5"
)

// PacketAdapter is a UDP Packet adapter for socks/redir/tun
type PacketAdapter struct {
	constant.UDPPacket
	metadata *constant.Metadata
}

// Metadata returns destination metadata
func (s *PacketAdapter) Metadata() *constant.Metadata {
	return s.metadata
}

// NewPacket is PacketAdapter generator
func NewPacket(target socks5.Addr, originTarget net.Addr, packet constant.UDPPacket, source constant.Type) *PacketAdapter {
	metadata := parseSocksAddr(target)
	metadata.NetWork = constant.UDP
	metadata.Type = source
	if ip, port, err := parseAddr(packet.LocalAddr()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = constant.Port(port)
	}
	if originTarget != nil {
		if addrPort, err := netip.ParseAddrPort(originTarget.String()); err == nil {
			metadata.OriginDst = addrPort
		}
	}
	return &PacketAdapter{
		UDPPacket: packet,
		metadata:  metadata,
	}
}
