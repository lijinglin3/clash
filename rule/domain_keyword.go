package rule

import (
	"strings"

	"github.com/lijinglin3/clash/constant"
)

// Implements constant.Rule
var _ constant.Rule = (*DomainKeyword)(nil)

type DomainKeyword struct {
	keyword string
	adapter string
}

func (dk *DomainKeyword) RuleType() constant.RuleType {
	return constant.DomainKeyword
}

func (dk *DomainKeyword) Match(metadata *constant.Metadata) bool {
	return strings.Contains(metadata.Host, dk.keyword)
}

func (dk *DomainKeyword) Adapter() string {
	return dk.adapter
}

func (dk *DomainKeyword) Payload() string {
	return dk.keyword
}

func (dk *DomainKeyword) ShouldResolveIP() bool {
	return false
}

func (dk *DomainKeyword) ShouldFindProcess() bool {
	return false
}

func NewDomainKeyword(keyword, adapter string) *DomainKeyword {
	return &DomainKeyword{
		keyword: strings.ToLower(keyword),
		adapter: adapter,
	}
}
