package rule

import (
	"path/filepath"
	"strings"

	"github.com/lijinglin3/clash/constant"
)

// Implements constant.Rule
var _ constant.Rule = (*Process)(nil)

type Process struct {
	adapter  string
	process  string
	nameOnly bool
}

func (ps *Process) RuleType() constant.RuleType {
	if ps.nameOnly {
		return constant.Process
	}

	return constant.ProcessPath
}

func (ps *Process) Match(metadata *constant.Metadata) bool {
	if ps.nameOnly {
		return strings.EqualFold(filepath.Base(metadata.ProcessPath), ps.process)
	}

	return strings.EqualFold(metadata.ProcessPath, ps.process)
}

func (ps *Process) Adapter() string {
	return ps.adapter
}

func (ps *Process) Payload() string {
	return ps.process
}

func (ps *Process) ShouldResolveIP() bool {
	return false
}

func (ps *Process) ShouldFindProcess() bool {
	return true
}

func NewProcess(process, adapter string, nameOnly bool) (*Process, error) {
	return &Process{
		adapter:  adapter,
		process:  process,
		nameOnly: nameOnly,
	}, nil
}
