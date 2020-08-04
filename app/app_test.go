package app

import (
	"testing"
)

type sConf struct {
	AppBaseConf

	Pindex int `json:"pindex"`
}

type sFlag struct {
	AppBaseFlag

	Test string `arg:"-t, help:test falg"`
}

func TestApp(t *testing.T) {
}
