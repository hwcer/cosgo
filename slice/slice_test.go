package slice

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	fmt.Println(String(arr))

	fmt.Println(Last(arr, 0))
	fmt.Println(Last(arr, 1))
	fmt.Println(Last(arr, 2))
	fmt.Println(Last(arr, 3))
	fmt.Println(Last(arr, 4))
	fmt.Println(Last(arr, 5))
	fmt.Println(Last(arr, 6))
}
