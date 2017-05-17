package ondemand

import (
	"io"
	"strings"
	"testing"

	"github.com/aabizri/aero/adexp/lexer/ondemand"
	"github.com/aabizri/aero/adexp/lexer/scannify"
)

const testString = " -TITLE SAM -ARCID AFR 456 -IFPLID XX11111111 -ADEP LFPG -ADES EGLL -EOBD 140110 -EOBT 0900 -CTOT 0930 -REGUL XXXXXXX -REGCAUSE XXXX -TAXITIME XXXXX -GEOID 01 -LATTD 520000N -LONGTD 0150000W -BEGIN ADDR -FAC LLEVZPZX -FAC LFFFZQZX -END ADDR"

func TestLexer(t *testing.T) {
	buf := strings.NewReader(testString)
	lexScanner := scannify.New(ondemand.NewLexReader(buf))
	parser := New(lexScanner)
Loop:
	for i := 0; ; i++ {
		expr, err := parser.Parse()
		switch err {
		case io.EOF:
			t.Logf("got EOF for expression %d", i)
			break Loop
		case nil:
		default:
			t.Fatalf("got error: %v", err)
		}

		if expr == nil {
			t.Fatalf("got nil expression for #%d", i)
		}

		t.Logf("Got for expression %d:\nKind: \t\t%s\nKeyword: \t%s (len %d)\nValue: \t\t%v", i, expr.Kind, expr.Keyword, len(expr.Keyword), expr.Value)
	}
}
