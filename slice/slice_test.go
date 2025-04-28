package slice

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	fmt.Println(String(arr))
}
