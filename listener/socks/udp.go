package socks

import (
	"net"

	"github.com/lijinglin3/clash/adapter/inbound"
	"github.com/lijinglin3/clash/common/pool"
	"github.com/lijinglin3/clash/common/sockopt"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/log"
	"github.com/lijinglin3/clash/transport/socks5"
)

type UDPListener struct {
	packetConn net.PacketConn
	addr       string
	closed     bool
}

// RawAddress implements constant.Listener
func (l *UDPListener) RawAddress() string {
	return l.addr
}

// Address implements constant.Listener
func (l *UDPListener) Address() string {
	return l.packetConn.LocalAddr().String()
}

// Close implements constant.Listener
func (l *UDPListener) Close() error {
	l.closed = true
	return l.packetConn.Close()
}

func NewUDP(addr string, in chan<- *inbound.PacketAdapter) (constant.Listener, error) {
	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}

	if err := sockopt.UDPReuseaddr(l.(*net.UDPConn)); err != nil {
		log.Warnln("Failed to Reuse UDP Address: %s", err)
	}

	sl := &UDPListener{
		packetConn: l,
		addr:       addr,
	}
	go func() {
		for {
			buf := pool.Get(pool.UDPBufferSize)
			n, remoteAddr, err := l.ReadFrom(buf)
			if err != nil {
				pool.Put(buf)
				if sl.closed {
					break
				}
				continue
			}
			handleSocksUDP(l, in, buf[:n], remoteAddr)
		}
	}()

	return sl, nil
}

func handleSocksUDP(pc net.PacketConn, in chan<- *inbound.PacketAdapter, buf []byte, addr net.Addr) {
	target, payload, err := socks5.DecodeUDPPacket(buf)
	if err != nil {
		// Unresolved UDP packet, return buffer to the pool
		pool.Put(buf)
		return
	}
	packet := &packet{
		pc:      pc,
		rAddr:   addr,
		payload: payload,
		bufRef:  buf,
	}
	select {
	case in <- inbound.NewPacket(target, pc.LocalAddr(), packet, constant.SOCKS5):
	default:
	}
}
