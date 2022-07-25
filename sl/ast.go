package sl

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type (
	Node interface {
		addChild(Node)
		parse(*Parser) error

		Children() []Node
		String() string
	}

	node struct {
		children []Node

		Tok Token
		Lit string
	}
	List struct {
		*node
	}
	IntLiteral struct {
		*node

		Int int64
	}
	StringLiteral struct {
		*node
	}
	Symbol struct {
		*node
	}
	Basic struct {
		*node
	}
	Comment struct {
		*node
	}
)

func (n *node) Children() []Node {
	return n.children
}
func (n *node) parse(p *Parser) error {
	n.Tok, n.Lit = p.scanIgnoreWhitespace()
	return nil
}
func (n *node) String() string {
	return n.Lit
}

func (n *node) addChild(child Node) {
	if n.children == nil {
		n.children = make([]Node, 0, 3)
	}
	n.children = append(n.children, child)
}

func basic() *Basic {
	return &Basic{node: &node{}}
}
func (k *Basic) String() string {
	return strings.ToUpper(k.node.Lit)
}

func symbol() *Symbol {
	return &Symbol{node: &node{}}
}

func intLiteral() *IntLiteral {
	return &IntLiteral{node: &node{}}
}
func (i *IntLiteral) parse(p *Parser) error {
	i.node.Tok, i.node.Lit = p.scanIgnoreWhitespace()
	var err error
	i.Int, err = strconv.ParseInt(i.node.Lit, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid INT literal: %v (%d:%d)", err, p.s.Line, p.s.Offset)
	}
	return nil
}

func stringLiteral() *StringLiteral {
	return &StringLiteral{node: &node{}}
}
func stringLiteralEx(tok Token, lit string) *StringLiteral {
	return &StringLiteral{node: &node{
		Tok: tok,
		Lit: lit,
	}}
}
func (s *StringLiteral) String() string {
	return fmt.Sprintf("%q", s.node.Lit)
}
func (s *StringLiteral) parse(p *Parser) error {
	var buf bytes.Buffer
	// skip QUOTE
	p.scanIgnoreWhitespace()

loop:
	for {
		ch := p.s.read()
		switch ch {
		case eof:
			return fmt.Errorf("Invalid STRING. Missing closing '\"' (%d:%d)", p.s.Line, p.s.Offset)

		case '\\':
			ch = p.s.read()
			switch ch {
			case 'n':
				buf.WriteRune('\n')

			case '"':
				buf.WriteRune('"')

			case '\\':
				buf.WriteRune('\\')

			default:
				return fmt.Errorf("Invalid escaped STRING character \"%c\" (%d:%d). Only \\n, \\\", and \\\\ are allowed.", ch, p.s.Line, p.s.Offset)
			}
			continue

		case '"':
			break loop
		}
		buf.WriteRune(ch)
	}

	s.node.Lit = buf.String()
	return nil
}

func comment() *Comment {
	return &Comment{node: &node{}}
}

func list() *List {
	return &List{node: &node{}}
}
func (l *List) String() string {
	var b strings.Builder
	b.WriteString("( ")
	strs := make([]string, len(l.Children()))
	for i, child := range l.Children() {
		strs[i] = child.String()
	}
	b.WriteString(strings.Join(strs, " "))
	b.WriteString(" )")
	return b.String()
}
func (l *List) parse(p *Parser) error {
	p.scanIgnoreWhitespace()
	for {
		child, err := p.Parse()
		if err != nil {
			return err
		}
		if child == nil {
			tok, _ := p.scanIgnoreWhitespace()
			if tok == PAREN_CLOSE {
				l.node.Tok = LIST
				l.node.Lit = l.String()
				return nil
			}
			return fmt.Errorf("Invalid list. Missing closing parens \")\" (%d:%d)", p.s.Line, p.s.Offset)
		}
		l.addChild(child)
	}
}
