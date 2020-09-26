package main

import (
	"context"
	"cosgo/app"
	"cosgo/logger"
	"github.com/spf13/pflag"
	"sync"
)

func init() {
	pflag.String("hwc", "", "test pflag")
	app.Flag.SetDefault("proAddr", "0.0.0.0:8080") //开启性能分析工具
	logger.Debug("test main init")
}

type test struct {
	name string
}

func (this *test) ID() string {
	return this.name
}

func (this *test) Init() error {

	return nil
}

func (this *test) Start(ctx context.Context, wgp *sync.WaitGroup) error {
	return nil
}

func (this *test) Stop() error {
	return nil
}

func main() {

	app.SetMain(func() {
		logo := `
	.----------------.   .----------------. 
	| .--------------. | | .--------------. |
	| | _____  _____ | | | | _____  _____ | |
	| ||_   _||_   _|| | | ||_   _||_   _|| |
	| |  | |    | |  | | | |  | | /\ | |  | |
	| |  | '    ' |  | | | |  | |/  \| |  | |
	| |   \ '--' /   | | | |  |   /\   |  | |
	| |    '.__.'    | | | |  |__/  \__|  | |
	| |              | | | |              | |
	| '--------------' | | '--------------' |
	'----------------'   '----------------' 
 `
		logger.Debug(logo)
	})

	app.Start(&test{name: "testMod"})
}
