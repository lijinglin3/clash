package tunnel

import (
	"fmt"
	"net"

	"github.com/lijinglin3/clash/adapter/inbound"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/transport/socks5"
)

type Listener struct {
	listener net.Listener
	addr     string
	target   socks5.Addr
	proxy    string
	closed   bool
}

// RawAddress implements constant.Listener
func (l *Listener) RawAddress() string {
	return l.addr
}

// Address implements constant.Listener
func (l *Listener) Address() string {
	return l.listener.Addr().String()
}

// Close implements constant.Listener
func (l *Listener) Close() error {
	l.closed = true
	return l.listener.Close()
}

func (l *Listener) handleTCP(conn net.Conn, in chan<- constant.ConnContext) {
	conn.(*net.TCPConn).SetKeepAlive(true)
	ctx := inbound.NewSocket(l.target, conn, constant.TUNNEL)
	ctx.Metadata().SpecialProxy = l.proxy
	in <- ctx
}

func New(addr, target, proxy string, in chan<- constant.ConnContext) (*Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	targetAddr := socks5.ParseAddr(target)
	if targetAddr == nil {
		return nil, fmt.Errorf("invalid target address %s", target)
	}

	rl := &Listener{
		listener: l,
		target:   targetAddr,
		proxy:    proxy,
		addr:     addr,
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				if rl.closed {
					break
				}
				continue
			}
			go rl.handleTCP(c, in)
		}
	}()

	return rl, nil
}
