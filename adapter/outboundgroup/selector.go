package outboundgroup

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lijinglin3/clash/adapter/outbound"
	"github.com/lijinglin3/clash/common/singledo"
	"github.com/lijinglin3/clash/component/dialer"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/constant/provider"
)

type Selector struct {
	*outbound.Base
	disableUDP bool
	single     *singledo.Single
	selected   string
	providers  []provider.ProxyProvider
}

// DialContext implements constant.ProxyAdapter
func (s *Selector) DialContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.Conn, error) {
	c, err := s.selectedProxy(true).DialContext(ctx, metadata, s.Base.DialOptions(opts...)...)
	if err == nil {
		c.AppendToChains(s)
	}
	return c, err
}

// ListenPacketContext implements constant.ProxyAdapter
func (s *Selector) ListenPacketContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.PacketConn, error) {
	pc, err := s.selectedProxy(true).ListenPacketContext(ctx, metadata, s.Base.DialOptions(opts...)...)
	if err == nil {
		pc.AppendToChains(s)
	}
	return pc, err
}

// SupportUDP implements constant.ProxyAdapter
func (s *Selector) SupportUDP() bool {
	if s.disableUDP {
		return false
	}

	return s.selectedProxy(false).SupportUDP()
}

// MarshalJSON implements constant.ProxyAdapter
func (s *Selector) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range getProvidersProxies(s.providers, false) {
		all = append(all, proxy.Name())
	}

	return json.Marshal(map[string]any{
		"type": s.Type().String(),
		"now":  s.Now(),
		"all":  all,
	})
}

func (s *Selector) Now() string {
	return s.selectedProxy(false).Name()
}

func (s *Selector) Set(name string) error {
	for _, proxy := range getProvidersProxies(s.providers, false) {
		if proxy.Name() == name {
			s.selected = name
			s.single.Reset()
			return nil
		}
	}

	return errors.New("proxy not exist")
}

// Unwrap implements constant.ProxyAdapter
func (s *Selector) Unwrap(metadata *constant.Metadata) constant.Proxy {
	return s.selectedProxy(true)
}

func (s *Selector) selectedProxy(touch bool) constant.Proxy {
	_, elm, _ := s.single.Do(func() (any, error) {
		proxies := getProvidersProxies(s.providers, touch)
		for _, proxy := range proxies {
			if proxy.Name() == s.selected {
				return proxy, nil
			}
		}

		return proxies[0], nil
	})

	return elm.(constant.Proxy)
}

func NewSelector(option *GroupCommonOption, providers []provider.ProxyProvider) *Selector {
	selected := providers[0].Proxies()[0].Name()
	return &Selector{
		Base: outbound.NewBase(outbound.BaseOption{
			Name:        option.Name,
			Type:        constant.Selector,
			Interface:   option.Interface,
			RoutingMark: option.RoutingMark,
		}),
		single:     singledo.NewSingle(defaultGetProxiesDuration),
		providers:  providers,
		selected:   selected,
		disableUDP: option.DisableUDP,
	}
}
