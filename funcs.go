package cosgo

import (
	"path/filepath"
)

//func GO(fn func()) {
//	SCC.GO(fn)
//}
//
//func CGO(fn func(ctx context.Context)) {
//	SCC.CGO(fn)
//}
//
//// SGO 带崩溃保护
//func SGO(fn func(context.Context), handles ...utils.TryHandle) {
//	SCC.SGO(fn, handles...)
//}

// Abs 获取以工作路径为起点的绝对路径
func Abs(dir ...string) string {
	path := filepath.Join(dir...)
	if !filepath.IsAbs(path) {
		path = filepath.Join(Dir(), path)
	}
	return path
}

func Debug() bool {
	return Config.GetBool(AppDebug)
}

// Dir 程序工作目录
func Dir() string {
	return Config.GetString(AppDir)
}

// Name 项目内部获取appName
func Name() string {
	return Config.GetString(AppName)
}

// WorkDir 程序工作目录 废弃
func WorkDir() string {
	return Config.GetString(AppDir)
}
