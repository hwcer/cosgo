package binder

import (
	"testing"
)

// TestBytes_SignedIntRoundTrip 确认 Int16/Int32/Int64 负数可 round-trip。
// 修复前: Int16(-1) → Uint16(0xFFFF) → int64(0xFFFF)=65535,丢失符号位。
func TestBytes_SignedIntRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		write any
		read  func() any
		want  int64
	}{
		{
			name:  "int16 negative",
			write: int16(-1),
			read:  func() any { var v int16; return &v },
			want:  -1,
		},
		{
			name:  "int16 min",
			write: int16(-32768),
			read:  func() any { var v int16; return &v },
			want:  -32768,
		},
		{
			name:  "int32 negative",
			write: int32(-12345),
			read:  func() any { var v int32; return &v },
			want:  -12345,
		},
		{
			name:  "int32 min",
			write: int32(-2147483648),
			read:  func() any { var v int32; return &v },
			want:  -2147483648,
		},
		{
			name:  "int64 negative",
			write: int64(-1),
			read:  func() any { var v int64; return &v },
			want:  -1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Bytes.Marshal(tc.write)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			dst := tc.read()
			if err := Bytes.Unmarshal(data, dst); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			var got int64
			switch v := dst.(type) {
			case *int16:
				got = int64(*v)
			case *int32:
				got = int64(*v)
			case *int64:
				got = *v
			}
			if got != tc.want {
				t.Errorf("round-trip: got %d, want %d", got, tc.want)
			}
		})
	}
}
