package scannify

import (
	"errors"
	"io"
)

type ByteScanner struct {
	reader   io.ByteReader
	previous byte
	unreaded bool
}

func (bs ByteScanner) ReadByte() (byte, error) {
	switch bs.unreaded {
	case false:
		b, err := bs.reader.ReadByte()
		bs.previous = b
		if err != nil {
			return b, err
		}
		return b, nil
	case true:
		bs.unreaded = false
		return bs.previous, nil
	}

	return 0, nil
}

func (bs ByteScanner) UnreadByte() error {
	if bs.unreaded {
		return errors.New("Scanner: cannot unread more than once")
	}

	bs.unreaded = true
	return nil
}
