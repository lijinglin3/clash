package tproxy

import (
	"net"

	"github.com/lijinglin3/clash/adapter/inbound"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/transport/socks5"
)

type Listener struct {
	listener net.Listener
	addr     string
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

func (l *Listener) handleTProxy(conn net.Conn, in chan<- constant.ConnContext) {
	target := socks5.ParseAddrToSocksAddr(conn.LocalAddr())
	conn.(*net.TCPConn).SetKeepAlive(true)
	in <- inbound.NewSocket(target, conn, constant.TPROXY)
}

func New(addr string, in chan<- constant.ConnContext) (constant.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	tl := l.(*net.TCPListener)
	rc, err := tl.SyscallConn()
	if err != nil {
		return nil, err
	}

	err = setsockopt(rc, addr)
	if err != nil {
		return nil, err
	}

	rl := &Listener{
		listener: l,
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
			go rl.handleTProxy(c, in)
		}
	}()

	return rl, nil
}
