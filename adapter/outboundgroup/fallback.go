package outboundgroup

import (
	"context"
	"encoding/json"

	"github.com/lijinglin3/clash/adapter/outbound"
	"github.com/lijinglin3/clash/common/singledo"
	"github.com/lijinglin3/clash/component/dialer"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/constant/provider"
)

type Fallback struct {
	*outbound.Base
	disableUDP bool
	single     *singledo.Single
	providers  []provider.ProxyProvider
}

func (f *Fallback) Now() string {
	proxy := f.findAliveProxy(false)
	return proxy.Name()
}

// DialContext implements constant.ProxyAdapter
func (f *Fallback) DialContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.Conn, error) {
	proxy := f.findAliveProxy(true)
	c, err := proxy.DialContext(ctx, metadata, f.Base.DialOptions(opts...)...)
	if err == nil {
		c.AppendToChains(f)
	}
	return c, err
}

// ListenPacketContext implements constant.ProxyAdapter
func (f *Fallback) ListenPacketContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.PacketConn, error) {
	proxy := f.findAliveProxy(true)
	pc, err := proxy.ListenPacketContext(ctx, metadata, f.Base.DialOptions(opts...)...)
	if err == nil {
		pc.AppendToChains(f)
	}
	return pc, err
}

// SupportUDP implements constant.ProxyAdapter
func (f *Fallback) SupportUDP() bool {
	if f.disableUDP {
		return false
	}

	proxy := f.findAliveProxy(false)
	return proxy.SupportUDP()
}

// MarshalJSON implements constant.ProxyAdapter
func (f *Fallback) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range f.proxies(false) {
		all = append(all, proxy.Name())
	}
	return json.Marshal(map[string]any{
		"type": f.Type().String(),
		"now":  f.Now(),
		"all":  all,
	})
}

// Unwrap implements constant.ProxyAdapter
func (f *Fallback) Unwrap(metadata *constant.Metadata) constant.Proxy {
	proxy := f.findAliveProxy(true)
	return proxy
}

func (f *Fallback) proxies(touch bool) []constant.Proxy {
	_, elm, _ := f.single.Do(func() (any, error) {
		return getProvidersProxies(f.providers, touch), nil
	})

	return elm.([]constant.Proxy)
}

func (f *Fallback) findAliveProxy(touch bool) constant.Proxy {
	proxies := f.proxies(touch)
	for _, proxy := range proxies {
		if proxy.Alive() {
			return proxy
		}
	}

	return proxies[0]
}

func NewFallback(option *GroupCommonOption, providers []provider.ProxyProvider) *Fallback {
	return &Fallback{
		Base: outbound.NewBase(outbound.BaseOption{
			Name:        option.Name,
			Type:        constant.Fallback,
			Interface:   option.Interface,
			RoutingMark: option.RoutingMark,
		}),
		single:     singledo.NewSingle(defaultGetProxiesDuration),
		providers:  providers,
		disableUDP: option.DisableUDP,
	}
}
