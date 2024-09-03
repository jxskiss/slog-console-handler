package terminal

import (
	"fmt"
	"io"
)

const NoColor Color = 0

const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	Gray
)

type Color uint8

// Format adds the coloring to the given string.
func (c Color) Format(s string) string {
	if c != NoColor {
		s = fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
	}
	return s
}

func (c Color) Append(b []byte, ss ...[]byte) []byte {
	for _, x := range ss {
		if c == NoColor {
			b = append(b, x...)
		} else {
			b = fmt.Appendf(b, "\x1b[%dm%s\x1b[0m", uint8(c), x)
		}
	}
	return b
}

func CheckIsTerminal(w io.Writer) bool {
	return checkIfTerminal(w)
}
