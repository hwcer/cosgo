package cosgo

import (
	"os"
	"path/filepath"
)

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

func Exist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
