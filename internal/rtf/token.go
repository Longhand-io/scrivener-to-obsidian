package rtf

import (
	"strconv"
)

type tokKind int

const (
	tokOpen tokKind = iota
	tokClose
	tokControl
	tokText
	tokEOF
)

type token struct {
	kind   tokKind
	word   string // control word or symbol
	param  int
	hasNum bool
	text   []byte // raw bytes for tokText (already decoded to UTF-8)
}

// cp1252 maps bytes 0x80..0x9F to runes; the rest of the range is Latin-1.
var cp1252 = [32]rune{
	0x20AC, 0xFFFD, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
	0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0xFFFD, 0x017D, 0xFFFD,
	0xFFFD, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
	0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0xFFFD, 0x017E, 0x0178,
}

func decodeByte(b byte) rune {
	if b >= 0x80 && b <= 0x9F {
		return cp1252[b-0x80]
	}
	return rune(b)
}

type lexer struct {
	data []byte
	pos  int
}

func isAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func (l *lexer) next() token {
	if l.pos >= len(l.data) {
		return token{kind: tokEOF}
	}
	c := l.data[l.pos]
	switch c {
	case '{':
		l.pos++
		return token{kind: tokOpen}
	case '}':
		l.pos++
		return token{kind: tokClose}
	case '\\':
		return l.control()
	case '\r', '\n':
		l.pos++
		return l.next()
	}
	start := l.pos
	for l.pos < len(l.data) {
		c := l.data[l.pos]
		if c == '{' || c == '}' || c == '\\' || c == '\r' || c == '\n' {
			break
		}
		l.pos++
	}
	raw := l.data[start:l.pos]
	buf := make([]byte, 0, len(raw))
	for _, b := range raw {
		if b < 0x80 {
			buf = append(buf, b)
		} else {
			buf = append(buf, string(decodeByte(b))...)
		}
	}
	return token{kind: tokText, text: buf}
}

func (l *lexer) control() token {
	l.pos++ // skip backslash
	if l.pos >= len(l.data) {
		return token{kind: tokEOF}
	}
	c := l.data[l.pos]
	if !isAlpha(c) {
		// control symbol
		l.pos++
		switch c {
		case '\'':
			if l.pos+2 <= len(l.data) {
				v, err := strconv.ParseUint(string(l.data[l.pos:l.pos+2]), 16, 8)
				l.pos += 2
				if err == nil {
					return token{kind: tokText, text: []byte(string(decodeByte(byte(v))))}
				}
			}
			return token{kind: tokText, text: nil}
		case '\r', '\n':
			// Cocoa writes "\<newline>" as a paragraph mark.
			return token{kind: tokControl, word: "par"}
		case '{', '}', '\\':
			return token{kind: tokText, text: []byte{c}}
		case '~':
			return token{kind: tokText, text: []byte(" ")}
		case '-':
			return token{kind: tokText, text: nil} // optional hyphen
		case '_':
			return token{kind: tokText, text: []byte("‑")}
		case '*':
			return token{kind: tokControl, word: "*"}
		}
		return token{kind: tokControl, word: string(c)}
	}
	start := l.pos
	for l.pos < len(l.data) && isAlpha(l.data[l.pos]) {
		l.pos++
	}
	word := string(l.data[start:l.pos])
	t := token{kind: tokControl, word: word}
	if l.pos < len(l.data) && (isDigit(l.data[l.pos]) || l.data[l.pos] == '-') {
		ns := l.pos
		l.pos++
		for l.pos < len(l.data) && isDigit(l.data[l.pos]) {
			l.pos++
		}
		if n, err := strconv.Atoi(string(l.data[ns:l.pos])); err == nil {
			t.param = n
			t.hasNum = true
		}
	}
	// a single space after a control word is a delimiter, not text
	if l.pos < len(l.data) && l.data[l.pos] == ' ' {
		l.pos++
	}
	return t
}
