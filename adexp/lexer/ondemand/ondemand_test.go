//TODO: Add logarithmic testing, focusing on the 100ns-1ms range.

package ondemand

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aabizri/aero/internal/repeating"
)

const testString = " -TITLE SAM -ARCID AFR 456 -IFPLID XX11111111 -ADEP LFPG -ADES EGLL -EOBD 140110 -EOBT 0900 -CTOT 0930 -REGUL XXXXXXX -REGCAUSE XXXX -TAXITIME XXXXX -GEO -GEOID 01 -LATTD 520000N -LONGTD 0150000W -BEGIN ADDR -FAC LLEVZPZX -FAC LFFFZQZX -END ADDR"

func TestLexer(t *testing.T) {
	buf := strings.NewReader(testString)
	lexer := New(buf)
Loop:
	for i := 0; ; i++ {
		expr, err := lexer.ReadLex()
		switch err {
		case io.EOF:
			t.Logf("got EOF for expression %d", i)
			break Loop
		case nil:
		default:
			t.Fatalf("got error: %v", err)
		}

		if expr == nil {
			t.Fatalf("got nil lexeme for lexeme %d", i)
		}

		t.Logf("Got for expression %d:\nKind: \t%s\nValue: \t%s (len %d)", i, expr.Kind, expr.Value, len(expr.Value))
	}
}

func BenchmarkLexer(b *testing.B) {
	buf := repeating.NewStringReader(testString)
	lexer := New(buf)

	for i := 0; i < b.N; i++ {
		lexer.ReadLex()
	}
}

func BenchmarkLexer_SlowReader(b *testing.B) {
	gen := func(d time.Duration) func(*testing.B) {
		return func(b *testing.B) {
			buf := repeating.NewStringReader(testString)
			lexer := NewLexReader(buf)

			for i := 0; i < b.N; i++ {
				lexer.ReadLex()
				time.Sleep(d)
			}
		}
	}
	for d := time.Nanosecond; d < time.Second; d *= 10 {
		if d == time.Nanosecond {
			b.Run(time.Duration(0).String(), gen(0))
		}
		b.Run(d.String(), gen(d))
	}
}
