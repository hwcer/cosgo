package app

import (
	"context"
	"github.com/hwcer/cosgo/utils"
)

var (
	SCC *utils.SCC
)

func init() {
	SCC = utils.NewSCC(nil)

}

func GO(fn func()) {
	SCC.GO(fn)
}

func CGO(fn func(ctx context.Context)) {
	SCC.CGO(fn)
}
