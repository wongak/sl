package sl

import (
	"fmt"
	"io"
)

type (
	Parser struct {
		s   *Scanner
		buf struct {
			tok Token  // last read token
			lit string // last read literal
			n   int    // buffer size (max=1)
		}
	}
)

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) Parse() (Node, error) {
	var n Node
	var err error
	tok, lit := p.scanIgnoreWhitespace()
	if tok == EOF {
		return nil, nil
	}

	switch tok {
	case NIL, TRUE, FALSE:
		n = basic()

	case PLUS, MULT, MINUS, DIV:
		n = symbol()

	case INT:
		n = intLiteral()

	case COMMENT:
		n = comment()

	case QUOTE:
		n = stringLiteral()

	case LITERAL:
		n = stringLiteralEx(tok, lit)
		return n, nil

	case PAREN_OPEN:
		n = list()

		// me are parsing lists recursively
		// unscan, so the parent list deals with
		// closing parentheses
	case PAREN_CLOSE:
		p.unscan()
		return nil, nil

	default:
		return nil, fmt.Errorf("Invalid token \"%s\" (%d:%d).", lit, p.s.Line, p.s.Offset)
	}
	p.unscan()
	err = n.parse(p)
	if err != nil {
		return nil, err
	}
	return n, nil
}
