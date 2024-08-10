package outboundgroup

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lijinglin3/clash/adapter/outbound"
	"github.com/lijinglin3/clash/common/singledo"
	"github.com/lijinglin3/clash/component/dialer"
	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/constant/provider"
)

type urlTestOption func(*URLTest)

func urlTestWithTolerance(tolerance uint16) urlTestOption {
	return func(u *URLTest) {
		u.tolerance = tolerance
	}
}

type URLTest struct {
	*outbound.Base
	tolerance  uint16
	disableUDP bool
	fastNode   constant.Proxy
	single     *singledo.Single
	fastSingle *singledo.Single
	providers  []provider.ProxyProvider
}

func (u *URLTest) Now() string {
	return u.fast(false).Name()
}

// DialContext implements constant.ProxyAdapter
func (u *URLTest) DialContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (c constant.Conn, err error) {
	c, err = u.fast(true).DialContext(ctx, metadata, u.Base.DialOptions(opts...)...)
	if err == nil {
		c.AppendToChains(u)
	}
	return c, err
}

// ListenPacketContext implements constant.ProxyAdapter
func (u *URLTest) ListenPacketContext(ctx context.Context, metadata *constant.Metadata, opts ...dialer.Option) (constant.PacketConn, error) {
	pc, err := u.fast(true).ListenPacketContext(ctx, metadata, u.Base.DialOptions(opts...)...)
	if err == nil {
		pc.AppendToChains(u)
	}
	return pc, err
}

// Unwrap implements constant.ProxyAdapter
func (u *URLTest) Unwrap(metadata *constant.Metadata) constant.Proxy {
	return u.fast(true)
}

func (u *URLTest) proxies(touch bool) []constant.Proxy {
	_, elm, _ := u.single.Do(func() (any, error) {
		return getProvidersProxies(u.providers, touch), nil
	})

	return elm.([]constant.Proxy)
}

func (u *URLTest) fast(touch bool) constant.Proxy {
	shared, elm, _ := u.fastSingle.Do(func() (any, error) {
		proxies := u.proxies(touch)
		fast := proxies[0]
		min := fast.LastDelay()
		fastNotExist := true

		for _, proxy := range proxies[1:] {
			if u.fastNode != nil && proxy.Name() == u.fastNode.Name() {
				fastNotExist = false
			}

			if !proxy.Alive() {
				continue
			}

			delay := proxy.LastDelay()
			if delay < min {
				fast = proxy
				min = delay
			}
		}

		// tolerance
		if u.fastNode == nil || fastNotExist || !u.fastNode.Alive() || u.fastNode.LastDelay() > fast.LastDelay()+u.tolerance {
			u.fastNode = fast
		}

		return u.fastNode, nil
	})
	if shared && touch { // a shared fastSingle.Do() may cause providers untouched, so we touch them again
		touchProviders(u.providers)
	}

	return elm.(constant.Proxy)
}

// SupportUDP implements constant.ProxyAdapter
func (u *URLTest) SupportUDP() bool {
	if u.disableUDP {
		return false
	}

	return u.fast(false).SupportUDP()
}

// MarshalJSON implements constant.ProxyAdapter
func (u *URLTest) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range u.proxies(false) {
		all = append(all, proxy.Name())
	}
	return json.Marshal(map[string]any{
		"type": u.Type().String(),
		"now":  u.Now(),
		"all":  all,
	})
}

func parseURLTestOption(config map[string]any) []urlTestOption {
	opts := []urlTestOption{}

	// tolerance
	if tolerance, ok := config["tolerance"].(int); ok {
		opts = append(opts, urlTestWithTolerance(uint16(tolerance)))
	}

	return opts
}

func NewURLTest(option *GroupCommonOption, providers []provider.ProxyProvider, options ...urlTestOption) *URLTest {
	urlTest := &URLTest{
		Base: outbound.NewBase(outbound.BaseOption{
			Name:        option.Name,
			Type:        constant.URLTest,
			Interface:   option.Interface,
			RoutingMark: option.RoutingMark,
		}),
		single:     singledo.NewSingle(defaultGetProxiesDuration),
		fastSingle: singledo.NewSingle(time.Second * 10),
		providers:  providers,
		disableUDP: option.DisableUDP,
	}

	for _, option := range options {
		option(urlTest)
	}

	return urlTest
}
