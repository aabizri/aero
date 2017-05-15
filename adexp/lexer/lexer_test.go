package lexer

import (
	"bytes"
	"context"
	"io"
	"testing"
)

const testString = " -TITLE SAM -ARCID AFR 456 -IFPLID XX11111111 -ADEP LFPG -ADES EGLL -EOBD 140110 -EOBT 0900 -CTOT 0930 -REGUL XXXXXXX -REGCAUSE XXXX -TAXITIME XXXXX"

func TestLexer_OnDemand(t *testing.T) {
	buf := bytes.NewBuffer([]byte(testString))
	ctx := context.Background()
	lexer := NewOnDemand(buf)
Loop:
	for i := 0; ; i++ {
		expr, err := lexer.ReadExpression(ctx)
		switch err {
		case io.EOF:
			t.Log("got EOF")
			break Loop
		case nil:
		default:
			t.Fatalf("got error: %v", err)
		}

		t.Logf("Got\nKeyword: \t%s (len %d)\nValue:  \t%s (len %d)", string(expr.Keyword), len(expr.Keyword), string(expr.Value), len(expr.Value))
	}
}
