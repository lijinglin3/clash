package outbound

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/lijinglin3/clash/common/structure"
	"github.com/lijinglin3/clash/component/dialer"
	"github.com/lijinglin3/clash/constant"
	obfs "github.com/lijinglin3/clash/transport/simple-obfs"
	"github.com/lijinglin3/clash/transport/snell"
)

type Snell struct {
	*Base
	psk        []byte
	pool       *snell.Pool
	obfsOption *simpleObfsOption
	version    int
}

type SnellOption struct {
	BasicOption
	Name     string         `proxy:"name"`
	Server   string         `proxy:"server"`
	Port     int            `proxy:"port"`
	Psk      string         `proxy:"psk"`
	UDP      bool           `proxy:"udp,omitempty"`
	Version  int            `proxy:"version,omitempty"`
	ObfsOpts map[string]any `proxy:"obfs-opts,omitempty"`
}

type streamOption struct {
	psk        []byte
	version    int
	addr       string
	obfsOption *simpleObfsOption
}

func streamConn(c net.Conn, option streamOption) *snell.Snell {
	switch option.obfsOption.Mode {
	case "tls":
		c = obfs.NewTLSObfs(c, option.obfsOption.Host)
	case "http":
		_, port, _ := net.SplitHostPort(option.addr)
		c = obfs.NewHTTPObfs(c, option.obfsOption.Host, port)
	}
	return snell.StreamConn(c, option.psk, option.version)
}

// StreamConn implements constant.ProxyAdapter
func (s *Snell) StreamConn(c net.Conn, metadata *constant.Metadata) (net.Conn, error) {
	c = streamConn(c, streamOption{s.psk, s.version, s.addr, s.obfsOption})
	err := snell.WriteHeader(c, metadata.String(), uint(metadata.DstPort), s.version)
	return c, err
}

// DialContext implements constant.ProxyAdapter
func (s *Snell) DialContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (_ constant.Conn, err error) {
	if s.version == snell.Version2 && len(opts) == 0 {
		c, err := s.pool.Get()
		if err != nil {
			return nil, err
		}

		if err = snell.WriteHeader(c, metadata.String(), uint(metadata.DstPort), s.version); err != nil {
			c.Close()
			return nil, err
		}
		return NewConn(c, s), err
	}

	c, err := dialer.DialContext(ctx, "tcp", s.addr, s.Base.DialOptions(opts...)...)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", s.addr, err)
	}
	tcpKeepAlive(c)

	defer func(c net.Conn) {
		safeConnClose(c, err)
	}(c)

	c, err = s.StreamConn(c, metadata)
	return NewConn(c, s), err
}

// ListenPacketContext implements constant.ProxyAdapter
func (s *Snell) ListenPacketContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.PacketConn, error) {
	c, err := dialer.DialContext(ctx, "tcp", s.addr, s.Base.DialOptions(opts...)...)
	if err != nil {
		return nil, err
	}
	tcpKeepAlive(c)
	c = streamConn(c, streamOption{s.psk, s.version, s.addr, s.obfsOption})

	err = snell.WriteUDPHeader(c, s.version)
	if err != nil {
		return nil, err
	}

	pc := snell.PacketConn(c)
	return newPacketConn(pc, s), nil
}

func NewSnell(option SnellOption) (*Snell, error) {
	addr := net.JoinHostPort(option.Server, strconv.Itoa(option.Port))
	psk := []byte(option.Psk)

	decoder := structure.NewDecoder(structure.Option{TagName: "obfs", WeaklyTypedInput: true})
	obfsOption := &simpleObfsOption{Host: "bing.com"}
	if err := decoder.Decode(option.ObfsOpts, obfsOption); err != nil {
		return nil, fmt.Errorf("snell %s initialize obfs error: %w", addr, err)
	}

	switch obfsOption.Mode {
	case "tls", "http", "":
	default:
		return nil, fmt.Errorf("snell %s obfs mode error: %s", addr, obfsOption.Mode)
	}

	// backward compatible
	if option.Version == 0 {
		option.Version = snell.DefaultSnellVersion
	}
	switch option.Version {
	case snell.Version1, snell.Version2:
		if option.UDP {
			return nil, fmt.Errorf("snell version %d not support UDP", option.Version)
		}
	case snell.Version3:
	default:
		return nil, fmt.Errorf("snell version error: %d", option.Version)
	}

	s := &Snell{
		Base: &Base{
			name:  option.Name,
			addr:  addr,
			tp:    constant.Snell,
			udp:   option.UDP,
			iface: option.Interface,
			rmark: option.RoutingMark,
		},
		psk:        psk,
		obfsOption: obfsOption,
		version:    option.Version,
	}

	if option.Version == snell.Version2 {
		s.pool = snell.NewPool(func(ctx context.Context) (*snell.Snell, error) {
			c, err := dialer.DialContext(ctx, "tcp", addr, s.Base.DialOptions()...)
			if err != nil {
				return nil, err
			}

			tcpKeepAlive(c)
			return streamConn(c, streamOption{psk, option.Version, addr, obfsOption}), nil
		})
	}
	return s, nil
}
