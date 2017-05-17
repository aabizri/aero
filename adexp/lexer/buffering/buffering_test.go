package buffering

import (
	"io"
	"strings"
	"testing"

	"github.com/aabizri/aero/adexp/lexer/ondemand"
)

const testString = " -TITLE   SAM -   ARCID AFR 456 -IFPLID XX11111111 -ADEP   LFPG  -ADES  EGLL -EOBD  140110   -EOBT 0900 -CTOT 0930 -REGUL XXXXXXX -REGCAUSE XXXX -TAXITIME XXXXX -GEO -GEOID 01 -LATTD 520000N -LONGTD 0150000W -BEGIN ADDR -FAC LLEVZPZX -FAC LFFFZQZX -END ADDR "

func TestLexer(t *testing.T) {
	buf := strings.NewReader(testString)
	embedded := ondemand.NewLexReader(buf)
	lexer := NewLexScanner(embedded, 10)
	t.Log("entering loop")
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

type unlimitedRR struct {
	*strings.Reader
	firstRune rune
	firstSize int
}

func (urr *unlimitedRR) ReadRune() (rune, int, error) {
	r, s, err := urr.Reader.ReadRune()
	if err == io.EOF {
		if urr.firstSize == 0 {
			urr.Reader.Seek(0, io.SeekStart)
			r, s, err = urr.Reader.ReadRune()
			urr.firstRune = r
			urr.firstSize = s
		} else {
			urr.Reader.Seek(int64(urr.firstSize), io.SeekStart)
			r, s, err = urr.firstRune, urr.firstSize, nil
		}
	}
	return r, s, err
}

func BenchmarkLexer(b *testing.B) {
	buf := &unlimitedRR{Reader: strings.NewReader(strings.Repeat(testString, 1000))}
	embedded := ondemand.NewLexReader(buf)
	lexer := NewLexScanner(embedded, 100)
	b.Log("entering loop")

	for i := 0; i < b.N; i++ {
		lexer.ReadLex()
	}
}
