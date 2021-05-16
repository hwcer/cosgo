package app

import (
	"context"
	"github.com/hwclegend/cosgo/utils"
)

var (
	scc *utils.SCC
)

func init() {
	scc = utils.NewSCC(nil)
}

func GO(fn func()) {
	scc.GO(fn)
}

func CGO(fn func(ctx context.Context)) {
	scc.CGO(fn)
}
