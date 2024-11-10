package html

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const (
	OPEN_TAG    = '<'
	CLOSE_TAG   = '>'
	CLOSE_ELEM  = '/'
	HEAD_ELEM	= "head"
	BODY_ELEM	= "body"
	SCRIPT_ELEM = "script"
	TEXT_NODE_TAG   = "_txt"
)

var (
	selfClosingElems = map[string]bool{
		"input": true,
		"link": true,
		"meta": true,
	}
)

type Node struct {
	Tag           string
	Name		  string
	Attrs         []string
	Text		  string
	Line          int
	Column        int
	Parent        *Node
	Children      []*Node
	Closing       *Node
	IsClosing     bool
	IsSelfClosing bool
}

func NewNode() *Node {
	return &Node{
		Attrs: []string{},
		IsSelfClosing: false,
		IsClosing:     false,
		Children:      []*Node{},
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("Node(Tag:%s, Line:%d, Col:%d, IsClosing:%t, IsSelfClosing:%t, Text:%s)",
		n.Tag, n.Line, n.Column, n.IsClosing, n.IsSelfClosing, n.Text)
}

// ParseTag reads a tag and its attributes. 
// Returns the a string of the tag and its attributes and the length of text read.
func ParseTag(reader *bufio.Reader) (string, error) {
	text := strings.Builder{}
	count := 0
loop:
	for r, _, errRead := reader.ReadRune(); errRead == nil; r, _, errRead = reader.ReadRune() {
		switch {
		case count == 0 && unicode.IsSpace(r):
			reader.UnreadRune()
			err := fmt.Errorf("space '%c' encountered when parsing tag", r)
			return text.String(), err
		case count > 0 && r == OPEN_TAG:
			reader.UnreadRune()
			err := fmt.Errorf("open tag rune '<' encountered after parsing '%s'", text.String())
			return text.String(), err
		case r == unicode.ReplacementChar:
			err := fmt.Errorf("unrecognized rune: '%c'", r)
			return text.String(), err
		default:
			text.WriteRune(r)
			if r == CLOSE_TAG {
				break loop
			} 
		}
		count += 1
	}

	return strings.ToLower(text.String()), nil
}

// ParseText returns the parsed text, the nr of new lines and an error, if any.
// The text and nr of new lines are always returned.
func ParseText(reader *bufio.Reader) (string, int, error){
	text := strings.Builder{}
	var err error
	var prev rune
	nrLines := 0
loop:
	for r, _, errRead := reader.ReadRune(); errRead == nil; r, _, errRead = reader.ReadRune() {
		switch{
		case r == unicode.ReplacementChar:
			err = fmt.Errorf("unrecognized rune: '%c'", r)
			break loop
		case prev == OPEN_TAG && unicode.IsLetter(r):
			reader.UnreadRune()
			break loop
		case prev == OPEN_TAG && unicode.IsSpace(r):
		default:
			text.WriteRune(r)
			if r == '\n' {
				nrLines += 1
			}
		}
	}

	return text.String(), nrLines, err
}

// Parse reads the input stream and returns a list of all the html nodes.
// It returns an error if
func Parse(text io.ReadCloser) ([]*Node, error) {
	defer text.Close()
	reader := bufio.NewReader(text)
	res := []*Node{}
	lineNr, colNr := 0, 0
	accu := strings.Builder{}
	saveAccu := false
	txtNode := NewNode()
	for r, _, errRead := reader.ReadRune(); errRead == nil; r, _, errRead = reader.ReadRune() {
		if r == '\n' {
			lineNr += 1
			colNr = 0
		} else if r == OPEN_TAG {
			if saveAccu {
				txtNode.Tag = TEXT_NODE_TAG
				txtNode.Text = accu.String()
				res = append(res, txtNode)
				txtNode = NewNode()
			}
			//reset accumulator
			saveAccu = false
			accu = strings.Builder{}

			cur := NewNode()
			reader.UnreadRune()
			tagText, errTag := ParseTag(reader)
			colNr += len(tagText)
			if errTag == nil {
				cur.Tag = tagText
				attrs := strings.Split(strings.Trim(tagText, "</>"), " ")
				cur.Name = attrs[0]
				if len(attrs) > 1 {
					cur.Attrs = append(cur.Attrs, attrs[1:]...)
				}
				cur.Line = lineNr
				cur.Column = colNr

				alwaysClosing, found := selfClosingElems[cur.Name] 
				cur.IsSelfClosing = strings.HasSuffix(cur.Tag, "/>") || (found && alwaysClosing)
				cur.IsClosing = strings.HasPrefix(cur.Tag, "</")
			}
			res = append(res, cur)
		} else {
			accu.WriteRune(r)
			saveAccu = saveAccu || !unicode.IsSpace(r)
			if txtNode.Line == 0 && saveAccu {
				txtNode.Line = lineNr
				txtNode.Column = colNr
			}
		}
	}

	return res, nil
}


