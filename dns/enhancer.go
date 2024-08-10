package dns

import (
	"net"

	"github.com/lijinglin3/clash/common/cache"
	"github.com/lijinglin3/clash/component/fakeip"
	"github.com/lijinglin3/clash/constant"
)

type ResolverEnhancer struct {
	mode     constant.DNSMode
	fakePool *fakeip.Pool
	mapping  *cache.LruCache
}

func (h *ResolverEnhancer) FakeIPEnabled() bool {
	return h.mode == constant.DNSFakeIP
}

func (h *ResolverEnhancer) MappingEnabled() bool {
	return h.mode == constant.DNSFakeIP || h.mode == constant.DNSMapping
}

func (h *ResolverEnhancer) IsExistFakeIP(ip net.IP) bool {
	if !h.FakeIPEnabled() {
		return false
	}

	if pool := h.fakePool; pool != nil {
		return pool.Exist(ip)
	}

	return false
}

func (h *ResolverEnhancer) IsFakeIP(ip net.IP) bool {
	if !h.FakeIPEnabled() {
		return false
	}

	if pool := h.fakePool; pool != nil {
		return pool.IPNet().Contains(ip) && !pool.Gateway().Equal(ip)
	}

	return false
}

func (h *ResolverEnhancer) FindHostByIP(ip net.IP) (string, bool) {
	if pool := h.fakePool; pool != nil {
		if host, existed := pool.LookBack(ip); existed {
			return host, true
		}
	}

	if mapping := h.mapping; mapping != nil {
		if host, existed := h.mapping.Get(ip.String()); existed {
			return host.(string), true
		}
	}

	return "", false
}

func (h *ResolverEnhancer) PatchFrom(o *ResolverEnhancer) {
	if h.mapping != nil && o.mapping != nil {
		o.mapping.CloneTo(h.mapping)
	}

	if h.fakePool != nil && o.fakePool != nil {
		h.fakePool.CloneFrom(o.fakePool)
	}
}

func NewEnhancer(cfg Config) *ResolverEnhancer {
	var fakePool *fakeip.Pool
	var mapping *cache.LruCache

	if cfg.EnhancedMode != constant.DNSNormal {
		fakePool = cfg.Pool
		mapping = cache.New(cache.WithSize(4096))
	}

	return &ResolverEnhancer{
		mode:     cfg.EnhancedMode,
		fakePool: fakePool,
		mapping:  mapping,
	}
}
