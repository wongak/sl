package sl

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

type (
	Token int

	// Scanner is the lexical scanner for sl
	Scanner struct {
		File   string
		Line   int
		Offset int

		r *bufio.Reader
	}
)

const (
	ILLEGAL Token = iota
	EOF
	WS

	// List
	PAREN_OPEN  // (
	PAREN_CLOSE // )

	// Symbols
	COLON // : signifying keyword
	PLUS  // +
	MINUS // -
	MULT  // *
	DIV   // /

	COMMENT // ;

	// quote signifying string
	QUOTE // "

	LIST

	INT
	FLOAT
	STRING

	NIL
	TRUE
	FALSE

	KEYWORD

	LITERAL
)

func (tok Token) String() string {
	switch tok {
	case EOF:
		return "EOF"
	case WS:
		return "WS"

	case PAREN_OPEN:
		return "PAREN_OPEN"
	case PAREN_CLOSE:
		return "PAREN_CLOSE"

	case NIL:
		return "NIL"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"

	default:
		return "ILLEGAL"
	}
}

var eof = rune(0)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == ','
}

func isLineBreak(ch rune) bool {
	return ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	s.Offset++
	return ch
}

func (s *Scanner) unread() {
	s.Offset--
	_ = s.r.UnreadRune()
}

func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			if isLineBreak(ch) {
				s.Line++
				s.Offset = 0
			}
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

func (s *Scanner) scanComment() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if ch == '\n' {
			s.Line++
			s.Offset = 0
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return COMMENT, buf.String()
}

func isSpecial(ch rune) bool {
	return ch == '(' || ch == ')'
}

func isIdent(ch rune) bool {
	if isWhitespace(ch) {
		return false
	}
	if isSpecial(ch) {
		return false
	}
	return true
}

func (s *Scanner) scanIdent(prefix ...rune) (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	if len(prefix) > 0 {
		for _, p := range prefix {
			buf.WriteRune(p)
		}
	} else {
		buf.WriteRune(s.read())
	}

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isIdent(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// If the string matches a keyword then return that keyword.
	switch strings.ToUpper(buf.String()) {
	case "NIL":
		return NIL, buf.String()
	case "TRUE":
		return TRUE, buf.String()
	case "FALSE":
		return FALSE, buf.String()
	}

	return LITERAL, buf.String()
}

func (s *Scanner) scanNumber() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isDigit(ch) && ch != '.' && ch != '-' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	if strings.IndexRune(buf.String(), '.') != -1 {
		return FLOAT, buf.String()
	}

	return INT, buf.String()
}

func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if isWhitespace(ch) {
		if isLineBreak(ch) {
			s.Line++
			s.Offset = 0
		}
		s.unread()
		return s.scanWhitespace()
	} else if isDigit(ch) {
		s.unread()
		return s.scanNumber()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return EOF, ""

	case '(':
		return PAREN_OPEN, string(ch)
	case ')':
		return PAREN_CLOSE, string(ch)

	case '+':
		return PLUS, string(ch)
	case '*':
		return MULT, string(ch)
	case '-':
		next := s.read()
		if isDigit(next) {
			s.unread()
			tok, lit := s.scanNumber()
			lit = "-" + lit
			return tok, lit
		} else if isWhitespace(next) {
			s.unread()
			return MINUS, string(ch)
		}
		s.unread()
		return s.scanIdent('-')
	case '/':
		return DIV, string(ch)

	case ';':
		s.unread()
		return s.scanComment()
	case '"':
		return QUOTE, string(ch)
	}

	return ILLEGAL, string(ch)
}
