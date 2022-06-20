package values

import (
	"fmt"
	"testing"
)

func TestBitSet_Set(t *testing.T) {
	b := BitSet{}
	b.Set(2)
	t.Logf("%v", fmt.Sprintf("%08b", b))
	b.Set(12)
	t.Logf("%v", fmt.Sprintf("%08b", b))

	t.Logf("Has %v", b.Has(12))
	b.Delete(12)
	t.Logf("%v", fmt.Sprintf("%08b", b))
	t.Logf("Has %v", b.Has(12))
}
