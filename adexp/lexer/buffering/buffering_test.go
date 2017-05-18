package buffering

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aabizri/aero/adexp/lexer/ondemand"
	"github.com/aabizri/aero/internal/repeating"
)

const testString = " -TITLE   SAM -   ARCID AFR 456 -IFPLID XX11111111 -ADEP   LFPG  -ADES  EGLL -EOBD  140110   -EOBT 0900 -CTOT 0930 -REGUL XXXXXXX -REGCAUSE XXXX -TAXITIME XXXXX -GEO -GEOID 01 -LATTD 520000N -LONGTD 0150000W -BEGIN ADDR -FAC LLEVZPZX -FAC LFFFZQZX -END ADDR "

func TestLexer(t *testing.T) {
	buf := strings.NewReader(testString)
	embedded := ondemand.New(buf)
	defer embedded.Close()
	lexer := New(embedded, 10)
	defer lexer.Close()

Loop:
	for i := 0; ; i++ {
		t.Logf("iteration %d starting...", i)
		// We unread thrice per expression
		if i != 0 && i%4 != 0 {
			t.Log("unreading ...")
			err := lexer.UnreadLex()
			if err != nil {
				t.Errorf("error when unreading: %v", err)
			}
		}
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

func BenchmarkLexer_BufferSize(b *testing.B) {
	gen := func(s int) func(*testing.B) {
		return func(b *testing.B) {
			buf := repeating.NewStringReader(testString)
			embedded := ondemand.New(buf)
			lexer := New(embedded, s)

			for i := 0; i < b.N; i++ {
				lexer.ReadLex()
			}

			lexer.Close()
			embedded.Close()
		}
	}
	for s := 0; s <= 300; s += 10 {
		b.Run(fmt.Sprintf("%d_elements", s), gen(s))
	}
}

func BenchmarkLexer_CallerSleep(b *testing.B) {
	gen := func(d time.Duration) func(*testing.B) {
		return func(b *testing.B) {
			buf := repeating.NewStringReader(testString)
			embedded := ondemand.New(buf)
			lexer := New(embedded, 100)

			for i := 0; i < b.N; i++ {
				lexer.ReadLex()
				time.Sleep(d)
			}

			lexer.Close()
			embedded.Close()
		}
	}
	for d := time.Nanosecond; d < time.Second; d *= 10 {
		if d == time.Nanosecond {
			b.Run(time.Duration(0).String(), gen(0))
		}
		b.Run(d.String(), gen(d))
	}
}
