package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"os"
	"os/exec"
	"path/filepath"
)

func ProtoGo() {
	logger.Info("======================开始生成GO Message======================")
	out := fmt.Sprintf("--go_out=%v", cosgo.Config.GetString(FlagsNameGo))
	path := fmt.Sprintf("--proto_path=%v", cosgo.Config.GetString(FlagsNameOut))
	file := filepath.Join(cosgo.Config.GetString(FlagsNameOut), "*.proto")

	if err := os.Chdir(cosgo.Dir()); err != nil {
		logger.Fatal(err)
	}

	cmd := exec.Command("./protoc", out, path, file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Proto GO Path:%v", cosgo.Config.GetString(FlagsNameGo))
}
