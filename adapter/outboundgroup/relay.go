package outboundgroup

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lijinglin3/clash/adapter/outbound"
	"github.com/lijinglin3/clash/common/singledo"
	"github.com/lijinglin3/clash/component/dialer"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/constant/provider"
)

type Relay struct {
	*outbound.Base
	single    *singledo.Single
	providers []provider.ProxyProvider
}

// DialContext implements constant.ProxyAdapter
func (r *Relay) DialContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.Conn, error) {
	var proxies []constant.Proxy
	for _, proxy := range r.proxies(metadata, true) {
		if proxy.Type() != constant.Direct {
			proxies = append(proxies, proxy)
		}
	}

	switch len(proxies) {
	case 0:
		return outbound.NewDirect().DialContext(ctx, metadata, r.Base.DialOptions(opts...)...)
	case 1:
		return proxies[0].DialContext(ctx, metadata, r.Base.DialOptions(opts...)...)
	}

	first := proxies[0]
	last := proxies[len(proxies)-1]

	c, err := dialer.DialContext(ctx, "tcp", first.Addr(), r.Base.DialOptions(opts...)...)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", first.Addr(), err)
	}
	tcpKeepAlive(c)

	var currentMeta *constant.Metadata
	for _, proxy := range proxies[1:] {
		currentMeta, err = addrToMetadata(proxy.Addr())
		if err != nil {
			return nil, err
		}

		c, err = first.StreamConn(c, currentMeta)
		if err != nil {
			return nil, fmt.Errorf("%s connect error: %w", first.Addr(), err)
		}

		first = proxy
	}

	c, err = last.StreamConn(c, metadata)
	if err != nil {
		return nil, fmt.Errorf("%s connect error: %w", last.Addr(), err)
	}

	return outbound.NewConn(c, r), nil
}

// MarshalJSON implements constant.ProxyAdapter
func (r *Relay) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range r.rawProxies(false) {
		all = append(all, proxy.Name())
	}
	return json.Marshal(map[string]any{
		"type": r.Type().String(),
		"all":  all,
	})
}

func (r *Relay) rawProxies(touch bool) []constant.Proxy {
	_, elm, _ := r.single.Do(func() (any, error) {
		return getProvidersProxies(r.providers, touch), nil
	})

	return elm.([]constant.Proxy)
}

func (r *Relay) proxies(metadata *constant.Metadata, touch bool) []constant.Proxy {
	proxies := r.rawProxies(touch)

	for n, proxy := range proxies {
		subproxy := proxy.Unwrap(metadata)
		for subproxy != nil {
			proxies[n] = subproxy
			subproxy = subproxy.Unwrap(metadata)
		}
	}

	return proxies
}

func NewRelay(option *GroupCommonOption, providers []provider.ProxyProvider) *Relay {
	return &Relay{
		Base: outbound.NewBase(outbound.BaseOption{
			Name:        option.Name,
			Type:        constant.Relay,
			Interface:   option.Interface,
			RoutingMark: option.RoutingMark,
		}),
		single:    singledo.NewSingle(defaultGetProxiesDuration),
		providers: providers,
	}
}
