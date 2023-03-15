package xlsx

import (
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
)

const (
	FlagsNameIn     string = "in"
	FlagsNameGo     string = "go"
	FlagsNameOut    string = "out"
	FlagsNameJson   string = "json"
	FlagsNameIgnore string = "ignore"
)

var mod *module

func init() {
	cosgo.Config.Flags(FlagsNameIn, "", "in", "需要解析的excel目录")
	cosgo.Config.Flags(FlagsNameOut, "", "out", "输出文件目录")
	cosgo.Config.Flags(FlagsNameGo, "", "out", "生成的GO文件")
	cosgo.Config.Flags(FlagsNameJson, "", "out", "是否导json格式")
	cosgo.Config.Flags(FlagsNameIgnore, "", "", "忽略的文件或者文件夹逗号分割多个")
}

func New() *module {
	if mod == nil {
		mod = &module{}
		mod.Id = "xlsx"
	}
	return mod
}

type module struct {
	cosgo.Module
}

func (this *module) Start() error {
	preparePath()
	LoadExcel(cosgo.Config.GetString(FlagsNameIn))
	logger.Info("\n========================恭喜大表哥导表成功========================\n")
	return nil
}
