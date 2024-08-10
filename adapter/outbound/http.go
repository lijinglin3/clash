package outbound

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/lijinglin3/clash/component/dialer"
	"github.com/lijinglin3/clash/constant"
)

type HTTP struct {
	*Base
	user      string
	pass      string
	tlsConfig *tls.Config
	Headers   http.Header
}

type HTTPOption struct {
	BasicOption
	Name           string            `proxy:"name"`
	Server         string            `proxy:"server"`
	Port           int               `proxy:"port"`
	UserName       string            `proxy:"username,omitempty"`
	Password       string            `proxy:"password,omitempty"`
	TLS            bool              `proxy:"tls,omitempty"`
	SNI            string            `proxy:"sni,omitempty"`
	SkipCertVerify bool              `proxy:"skip-cert-verify,omitempty"`
	Headers        map[string]string `proxy:"headers,omitempty"`
}

// StreamConn implements constant.ProxyAdapter
func (h *HTTP) StreamConn(c net.Conn, metadata *constant.Metadata) (net.Conn, error) {
	if h.tlsConfig != nil {
		cc := tls.Client(c, h.tlsConfig)
		ctx, cancel := context.WithTimeout(context.Background(), constant.DefaultTLSTimeout)
		defer cancel()
		err := cc.HandshakeContext(ctx)
		c = cc
		if err != nil {
			return nil, fmt.Errorf("%s connect error: %w", h.addr, err)
		}
	}

	if err := h.shakeHand(metadata, c); err != nil {
		return nil, err
	}
	return c, nil
}

// DialContext implements constant.ProxyAdapter
func (h *HTTP) DialContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (_ constant.Conn, err error) {
	c, err := dialer.DialContext(ctx, "tcp", h.addr, h.Base.DialOptions(opts...)...)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", h.addr, err)
	}
	tcpKeepAlive(c)

	defer func(c net.Conn) {
		safeConnClose(c, err)
	}(c)

	c, err = h.StreamConn(c, metadata)
	if err != nil {
		return nil, err
	}

	return NewConn(c, h), nil
}

func (h *HTTP) shakeHand(metadata *constant.Metadata, rw io.ReadWriter) error {
	addr := metadata.RemoteAddress()
	req := &http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Host: addr,
		},
		Host:   addr,
		Header: h.Headers.Clone(),
	}

	req.Header.Add("Proxy-Connection", "Keep-Alive")

	if h.user != "" && h.pass != "" {
		auth := h.user + ":" + h.pass
		req.Header.Add("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}

	if err := req.Write(rw); err != nil {
		return err
	}

	resp, err := http.ReadResponse(bufio.NewReader(rw), req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusProxyAuthRequired {
		return errors.New("HTTP need auth")
	}

	if resp.StatusCode == http.StatusMethodNotAllowed {
		return errors.New("CONNECT method not allowed by proxy")
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return errors.New(resp.Status)
	}

	return fmt.Errorf("can not connect remote err code: %d", resp.StatusCode)
}

func NewHTTP(option HTTPOption) *HTTP {
	var tlsConfig *tls.Config
	if option.TLS {
		sni := option.Server
		if option.SNI != "" {
			sni = option.SNI
		}
		tlsConfig = &tls.Config{
			InsecureSkipVerify: option.SkipCertVerify,
			ServerName:         sni,
		}
	}

	headers := http.Header{}
	for name, value := range option.Headers {
		headers.Add(name, value)
	}

	return &HTTP{
		Base: &Base{
			name:  option.Name,
			addr:  net.JoinHostPort(option.Server, strconv.Itoa(option.Port)),
			tp:    constant.Http,
			iface: option.Interface,
			rmark: option.RoutingMark,
		},
		user:      option.UserName,
		pass:      option.Password,
		tlsConfig: tlsConfig,
		Headers:   headers,
	}
}
