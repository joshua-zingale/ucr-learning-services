package yaml

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

type Position struct {
	line int
	char int
}

type ParsingError struct {
	Position
	message string
	parent  *ParsingError
}

func (pe *ParsingError) Error() string {
	return fmt.Sprintf("line %d: char %d: %s", pe.line+1, pe.char+1, pe.message)
}

func (pe *ParsingError) addContext(message string) *ParsingError {
	return &ParsingError{
		Position: Position{
			line: pe.line,
			char: pe.char,
		},
		message: fmt.Sprintf("%s: %s", message, pe.message),
	}
}

type Parser struct {
	Position
	offset int
	source string
}

func NewYamlParser(source string) *Parser {
	return &Parser{
		source: source,
	}
}

func ParseYaml(source string) (any, *ParsingError) {
	parser := NewYamlParser(source)
	return parser.parseValue(0)
}

// func (p *Parser) advanceToNextContent() int {
// 	p.skipWhitespace()

// }

// Returns -1 if there is no next line that is indented
func (p *Parser) getNextLineIndentation() int {
	p2 := *p

	for !p2.consumeLineBreak() {
		r, ok := p2.peek()
		if !ok {
			return -1
		}
		p2.advance(r)
	}

	p2.skipWhitespace()

	if p2.isEOF() {
		return -1
	}

	return p2.char
}

func (p *Parser) parseValue(minimumIndentation int) (any, *ParsingError) {

	isBeginningOfDocument := p.offset == 0
	lookahead := *p
	lookahead.skipWhitespace()

	if lookahead.line == p.line && !isBeginningOfDocument && !lookahead.isEOF() {
		r, ok := p.peek()
		if !ok {
			panic("unreachable! 76b76trxs")
		}
		return nil, p.newUnexpectedCharacterError("line break", r)
	}

	contentIndentation := lookahead.char
	if contentIndentation < minimumIndentation {
		return []string{}, nil
	}

	*p = lookahead

	if firstKey, err := p.parseKey(); err == nil {
		v, err := p.parseMap(firstKey, contentIndentation)
		if err != nil {
			return nil, err.addContext("parsing map")
		}
		return v, nil
	}

	if firstItem, err := p.parseDashItem(); err == nil {
		v, err := p.parseDashList(firstItem, contentIndentation)
		if err != nil {
			return nil, err.addContext("parsing dash list")
		}
		return v, nil
	}

	return nil, p.newError("expected mapping or list")

}

var keyRegex = regexp.MustCompile(`^([a-zA-Z]\w*):`)

func (p *Parser) parseKey() (string, *ParsingError) {
	matches, err := p.parseRegex(keyRegex)

	if err != nil {
		return "", err
	}
	return matches[1], err

}

func (p *Parser) parseMap(firstKey string, indentation int) (map[string]any, *ParsingError) {

	outputMap := make(map[string]any)

	key := firstKey
	for {
		value, err := p.parseValue(indentation + 1)
		if err != nil {
			return nil, err.addContext("parsing map value")
		}

		outputMap[key] = value

		lookahead := *p

		lookahead.skipWhitespace()
		if lookahead.isEOF() {
			break
		}
		if lookahead.line == p.line {
			return nil, lookahead.newError("expected line break")
		}

		if lookahead.char > indentation {
			return nil, lookahead.newError("invalid indentation")
		}

		if lookahead.char < indentation {
			break
		}

		*p = lookahead

		nextKey, err := p.parseKey()
		if err != nil {
			return nil, err
		}

		key = nextKey
	}

	return outputMap, nil
}

var dashItemRegex = regexp.MustCompile(`^- +([\w-@\.]+)`)

func (p *Parser) parseDashItem() (string, *ParsingError) {
	matches, err := p.parseRegex(dashItemRegex)
	if err != nil {
		return "", nil
	}
	return matches[1], nil
}

func (p *Parser) parseDashList(firstItem string, indentation int) ([]string, *ParsingError) {

	outList := []string{firstItem}

	for {
		lookahead := *p
		lookahead.skipWhitespace()

		if lookahead.isEOF() {
			break
		}

		if lookahead.line == p.line {
			return nil, lookahead.newError("expected line break")
		}

		if lookahead.char > indentation {
			return nil, lookahead.newError(fmt.Sprintf("expected indentation of %d space(s) but encountered an indentation of %d", indentation, lookahead.char))
		}

		if lookahead.char < indentation {
			break
		}

		*p = lookahead

		nextItem, err := p.parseDashItem()
		if err != nil {
			return nil, err
		}

		outList = append(outList, nextItem)
	}

	return outList, nil
}

func (p *Parser) newError(message string) *ParsingError {
	return &ParsingError{
		Position: p.Position,
		message:  message,
	}
}

func (p *Parser) newUnexpectedCharacterError(expected string, r rune) *ParsingError {
	return p.newError(fmt.Sprintf("expected %s but found '%c'", expected, r))
}

func (p *Parser) advance(r rune) {

	length := utf8.RuneLen(r)

	if p.offset+length > len(p.source) {
		panic("internal error: attempted to advance parser beyond source boundry")
	}

	if r == '\n' {
		p.line += 1
		p.char = 0
	} else {
		p.char += 1
	}

	p.offset += length
}

func (p *Parser) peek() (rune, bool) {

	if p.offset >= len(p.source) {
		return 0, false
	}
	for _, r := range p.source[p.offset:] {
		return r, true
	}
	panic("unreachable! 3h8fuen")
}

func (p *Parser) next() (rune, bool) {
	if r, ok := p.peek(); ok {
		p.advance(r)
		return r, ok
	}
	return 0, false
}

var lineBreakingSequences []string = []string{
	"\n",
	"\r\n",
	"\r",
}

func (p *Parser) isLineBreak() bool {
	for _, sequence := range lineBreakingSequences {
		if len(sequence) > len(p.source)-p.offset {
			continue
		}
		if p.source[p.offset:p.offset+len(sequence)] == sequence {
			return true
		}
	}
	return false
}

func (p *Parser) isEOF() bool {
	return len(p.source) <= p.offset
}

func (p *Parser) consumeLineBreak() bool {
	for _, sequence := range lineBreakingSequences {
		if len(sequence) > len(p.source)-p.offset {
			continue
		}
		sourceSequence := p.source[p.offset : p.offset+len(sequence)]
		if sourceSequence == sequence {
			for _, r := range sourceSequence {
				p.advance(r)
			}
			return true
		}
	}
	return false
}

func (p *Parser) skipInlineWhitespace() {
	for {
		r, ok := p.peek()
		if !ok || r != ' ' {
			break
		}
		p.advance(r)
	}
}

func (p *Parser) skipWhitespace() {
	for {
		p.skipInlineWhitespace()
		if !p.consumeLineBreak() {
			return
		}
	}
}

func (p *Parser) parseRegex(re *regexp.Regexp) ([]string, *ParsingError) {
	matches := re.FindStringSubmatch(p.source[p.offset:])

	if len(matches) == 0 {
		return nil, p.newError(fmt.Sprintf("failed to match %s", re.String()))
	}
	for _, r := range matches[0] {
		p.advance(r)
	}

	return matches, nil
}

// type YamlValue interface {
// 	isYamlValue()
// }

// type YamlMapEntry struct {
// 	key   string
// 	value YamlValue
// }

// type YamlMap []YamlMapEntry
// type YamlList []string

// func (m YamlMap) isYamlValue()  {}
// func (m YamlList) isYamlValue() {}
