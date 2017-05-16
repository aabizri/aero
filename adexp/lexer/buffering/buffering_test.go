package buffering

import (
	"io"
	"strings"
	"testing"
	"workspace/aero/adexp/lexer/ondemand"
)

const testString = " -TITLE SAM -   ARCID AFR 456 -IFPLID XX11111111 -ADEP LFPG -ADES EGLL -EOBD 140110 -EOBT 0900 -CTOT 0930 -REGUL XXXXXXX -REGCAUSE XXXX -TAXITIME XXXXX -GEO -GEOID 01 -LATTD 520000N -LONGTD 0150000W -BEGIN ADDR -FAC LLEVZPZX -FAC LFFFZQZX -END ADDR"

func TestLexer(t *testing.T) {
	buf := strings.NewReader(testString)
	embedded := ondemand.New(buf)
	lexer := New(embedded, 10)
	t.Log("entering loop")
Loop:
	for i := 0; ; i++ {
		t.Logf("iteration %d starting...", i)
		expr, err := lexer.Lex()
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
