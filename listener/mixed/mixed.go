package mixed

import (
	"net"

	"github.com/lijinglin3/clash/common/cache"
	cnet "github.com/lijinglin3/clash/common/net"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/listener/http"
	"github.com/lijinglin3/clash/listener/socks"
	"github.com/lijinglin3/clash/transport/socks4"
	"github.com/lijinglin3/clash/transport/socks5"
)

type Listener struct {
	listener net.Listener
	addr     string
	cache    *cache.LruCache
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

func New(addr string, in chan<- constant.ConnContext) (constant.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	ml := &Listener{
		listener: l,
		addr:     addr,
		cache:    cache.New(cache.WithAge(30)),
	}
	go func() {
		for {
			c, err := ml.listener.Accept()
			if err != nil {
				if ml.closed {
					break
				}
				continue
			}
			go handleConn(c, in, ml.cache)
		}
	}()

	return ml, nil
}

func handleConn(conn net.Conn, in chan<- constant.ConnContext, lru *cache.LruCache) {
	conn.(*net.TCPConn).SetKeepAlive(true)

	bufConn := cnet.NewBufferedConn(conn)
	head, err := bufConn.Peek(1)
	if err != nil {
		return
	}

	switch head[0] {
	case socks4.Version:
		socks.HandleSocks4(bufConn, in)
	case socks5.Version:
		socks.HandleSocks5(bufConn, in)
	default:
		http.HandleConn(bufConn, in, lru)
	}
}
