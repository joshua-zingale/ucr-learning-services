package yaml

import (
	"fmt"
	"regexp"
)

type combinator func(*parserState) (*production, *ParsingError)

type production struct {
	state parserState
	value YamlValue
}

type parserState struct {
	line      int
	char      int
	remaining string
}

type ParsingError struct {
	line    int
	char    int
	message string
}

func (pe *ParsingError) Error() string {
	return fmt.Sprintf("line %d: char %d: %s", pe.line+1, pe.char+1, pe.message)
}

func (pe *ParsingError) addContext(message string) *ParsingError {
	return &ParsingError{
		line:    pe.line,
		char:    pe.char,
		message: fmt.Sprintf("%s: %s", message, pe.message),
	}
}

func (pe *ParsingError) rewriteMessage(message string) *ParsingError {
	return &ParsingError{
		line:    pe.line,
		char:    pe.char,
		message: message,
	}
}

func NewparserState(source string) parserState {
	return parserState{
		line:      0,
		char:      0,
		remaining: source,
	}
}

func (ps *parserState) Advance(numBytes int) *parserState {
	line := ps.line
	char := ps.char
	for _, b := range ps.remaining[:numBytes] {
		if b == '\n' {
			line += 1
			char = 0
		} else {
			char += 1
		}
	}
	return &parserState{
		line:      line,
		char:      char,
		remaining: ps.remaining[numBytes:],
	}
}

func (ps *parserState) newError(message string) *ParsingError {
	return &ParsingError{
		line:    ps.line,
		char:    ps.char,
		message: message,
	}
}

type YamlValue interface {
	isYamlValue()
}

func ParseYaml(source string) (YamlValue, error) {
	initialState := NewparserState(source)
	prod, err := newDashListParser(0)(&initialState)
	if err != nil {
		return nil, err
	}
	return prod.value, nil
}

var parseEOLorEOF combinator = second(optional(parseInlineWhiteSpace), or("end of line or end of file", parseEOF, parseLineBreak))

var parseWhiteSpace combinator = repeatedAsList(or("whitespace", parseInlineWhiteSpace, parseEOLorEOF))

func newDashListParser(numLeadingSpaces int) combinator {
	return first(newListParserSeparatedBy(second(optional(parseInlineWhiteSpace), parseLineBreak), newSpaceIndentedParser(numLeadingSpaces, parseDashItem)), parseEOLorEOF)
}

var parseDashItem combinator = second(newStringParser("-"), second(parseInlineWhiteSpace, parseUserId))

func newSpaceIndentedParser(indentationLevel int, c combinator) combinator {
	return func(ps *parserState) (*production, *ParsingError) {

		observedIndentation := 0
		for observedIndentation < len(ps.remaining) && ps.remaining[observedIndentation] == ' ' {
			observedIndentation += 1
		}
		if observedIndentation != indentationLevel {
			return nil, ps.newError(fmt.Sprintf("invalid indentation: found %d but expected %d", observedIndentation, indentationLevel))
		}

		// " " is a one-byte character
		return c(ps.Advance(observedIndentation))
	}
}

var parseUserId combinator = newRegexParser("user ID", *regexp.MustCompile(`^([\w@\.\-]+)`))

var groupIdRegex combinator = newRegexParser("group ID", *regexp.MustCompile(`^([\w\-]+):`))

func first(firstc combinator, secondc combinator) combinator {
	return func(ps *parserState) (*production, *ParsingError) {
		prod1, err := firstc(ps)
		if err != nil {
			return nil, err
		}
		prod2, err := secondc(&prod1.state)
		if err != nil {
			return nil, err
		}
		return &production{
			state: prod2.state,
			value: prod1.value,
		}, nil
	}
}

func second(firstc combinator, secondc combinator) combinator {
	return func(ps *parserState) (*production, *ParsingError) {
		prod1, err := firstc(ps)
		if err != nil {
			return nil, err
		}
		prod2, err := secondc(&prod1.state)
		if err != nil {
			return nil, err
		}
		return &production{
			state: prod2.state,
			value: prod2.value,
		}, nil
	}
}

func optional(c combinator) combinator {
	return func(ps *parserState) (*production, *ParsingError) {
		prod, err := c(ps)
		if err != nil {
			return &production{
				value: nil,
				state: *ps,
			}, nil
		}
		return prod, nil
	}
}

func parseInlineWhiteSpace(state *parserState) (*production, *ParsingError) {
	for index, b := range state.remaining {
		if b != ' ' {
			if index == 0 {
				return nil, state.newError("expected inline whitespace")
			}
			fmt.Println(index, state.remaining)
			return &production{
				state: *state.Advance(index),
			}, nil
		}
	}
	return nil, state.newError("expected inline whitespace")
}

func newRegexParser(productionName string, re regexp.Regexp) combinator {

	return func(state *parserState) (*production, *ParsingError) {
		submatches := re.FindStringSubmatch(state.remaining)
		if len(submatches) == 0 {
			return nil, state.newError(fmt.Sprintf("invalid %s", productionName))
		}

		return &production{
			state: *state.Advance(len(submatches[0])),
			value: YamlString(submatches[1]),
		}, nil
	}

}

func newStringParser(s string) combinator {
	return func(ps *parserState) (*production, *ParsingError) {

		if len(ps.remaining) >= len(s) && ps.remaining[:len(s)] == s {
			return &production{
				state: *ps.Advance(len(s)),
				value: YamlString(s)}, nil
		}

		return nil, ps.newError(fmt.Sprintf("expected '%s'", s))
	}
}

func parseLineBreak(state *parserState) (*production, *ParsingError) {
	if len(state.remaining) == 0 {
		return nil, state.newError("expected line break")
	}

	if len(state.remaining) > 1 && state.remaining[0] == '\r' && state.remaining[1] == '\n' {
		return &production{
			state: *state.Advance(2),
		}, nil
	} else if state.remaining[0] == '\n' || state.remaining[0] == '\r' {
		return &production{
			state: *state.Advance(1),
		}, nil
	}
	return nil, state.newError("expected line break")
}

func parseEOF(state *parserState) (*production, *ParsingError) {
	if len(state.remaining) != 0 {
		return nil, state.newError("expected the end of the file")
	}
	return &production{
		state: *state,
		value: nil,
	}, nil
}

func or(name string, combinators ...combinator) combinator {

	var latestError *ParsingError = nil
	return func(ps *parserState) (*production, *ParsingError) {
		for _, c := range combinators {
			if prod, err := c(ps); err == nil {
				return prod, nil
			} else if latestError == nil || err.line > latestError.line || err.line == latestError.line && err.char > latestError.char {
				latestError = err
			}
		}
		return nil, latestError.rewriteMessage(fmt.Sprintf("expected %s", name))
	}
}

func repeatedAsList(c combinator) combinator {
	return func(ps *parserState) (*production, *ParsingError) {
		var list YamlList
		var state *parserState = ps
		for {
			prod, err := c(state)
			if err != nil {
				return &production{
					state: *state,
					value: list,
				}, nil
			}
			list = append(list, prod.value)
			state = &prod.state
		}
	}

}

func newListParserSeparatedBy(separator combinator, c combinator) combinator {
	return func(ps *parserState) (*production, *ParsingError) {
		var list YamlList
		var state *parserState = ps
		for {
			prod, err := c(state)
			if err != nil {
				return &production{
					state: *state,
					value: list,
				}, nil
			}
			list = append(list, prod.value)
			state = &prod.state
			prod, err = separator(state)
			if err != nil {
				return &production{
					state: *state,
					value: list,
				}, nil
			}
			state = &prod.state
		}
	}

}

type YamlString string
type YamlMap map[string]YamlValue
type YamlList []YamlValue

func (s YamlString) isYamlValue() {}

func (m YamlMap) isYamlValue()  {}
func (m YamlList) isYamlValue() {}
