package handler

import (
	"encoding/json"
	"fmt"
	"testing"
)

type myclass struct {
	Name string
}



func TestHandle(t *testing.T) {
	a := &myclass{Name:"hwc"}

	b,_:= json.Marshal(a)
	fmt.Printf("json:%+v\n",string(b))

}
