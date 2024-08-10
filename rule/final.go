package rule

import (
	"github.com/lijinglin3/clash/constant"
)

// Implements constant.Rule
var _ constant.Rule = (*Match)(nil)

type Match struct {
	adapter string
}

func (f *Match) RuleType() constant.RuleType {
	return constant.MATCH
}

func (f *Match) Match(metadata *constant.Metadata) bool {
	return true
}

func (f *Match) Adapter() string {
	return f.adapter
}

func (f *Match) Payload() string {
	return ""
}

func (f *Match) ShouldResolveIP() bool {
	return false
}

func (f *Match) ShouldFindProcess() bool {
	return false
}

func NewMatch(adapter string) *Match {
	return &Match{
		adapter: adapter,
	}
}
