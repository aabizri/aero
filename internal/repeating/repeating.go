package repeating

import (
	"io"
	"strings"
)

const repeatInitCount = 1000

// StringReader is a reapeating string reader
type StringReader struct {
	*strings.Reader
	firstRune rune
	firstSize int
}

// NewStringReader returns a repeating string reader
func NewStringReader(str string) *StringReader {
	return &StringReader{
		Reader: strings.NewReader(strings.Repeat(str, repeatInitCount)),
	}
}

// ReadRune reads one rune from a StringReader
func (sr *StringReader) ReadRune() (rune, int, error) {
	r, s, err := sr.Reader.ReadRune()
	if err == io.EOF {
		if sr.firstSize == 0 {
			sr.Reader.Seek(0, io.SeekStart)
			r, s, err = sr.Reader.ReadRune()
			sr.firstRune = r
			sr.firstSize = s
		} else {
			sr.Reader.Seek(int64(sr.firstSize), io.SeekStart)
			r, s, err = sr.firstRune, sr.firstSize, nil
		}
	}
	return r, s, err
}
