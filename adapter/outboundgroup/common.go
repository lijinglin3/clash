package outboundgroup

import (
	"time"

	"github.com/lijinglin3/clash/constant"
	"github.com/lijinglin3/clash/constant/provider"
)

const (
	defaultGetProxiesDuration = time.Second * 5
)

func touchProviders(providers []provider.ProxyProvider) {
	for _, p := range providers {
		p.Touch()
	}
}

func getProvidersProxies(providers []provider.ProxyProvider, touch bool) []constant.Proxy {
	proxies := []constant.Proxy{}
	for _, provider := range providers {
		if touch {
			provider.Touch()
		}
		proxies = append(proxies, provider.Proxies()...)
	}
	return proxies
}
