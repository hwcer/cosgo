package cosgo

import (
	"context"
	"github.com/hwcer/cosgo/utils"
)

var SCC = utils.NewSCC(nil)

func GO(fn func()) {
	SCC.GO(fn)
}

func CGO(fn func(ctx context.Context)) {
	SCC.CGO(fn)
}
