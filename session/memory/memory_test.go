package memory

import (
	"strings"
	"testing"
	"time"
)

func TestMemory(t *testing.T) {
	a := "1&2&3&4"
	t.Logf("%v", strings.SplitN(a, "&", 2))

	var expires time.Time
	t.Logf("%v", expires.Unix())
}
