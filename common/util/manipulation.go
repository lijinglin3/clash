package util

import "github.com/samber/lo"

func EmptyOr[T comparable](v, def T) T {
	ret, _ := lo.Coalesce(v, def)
	return ret
}
