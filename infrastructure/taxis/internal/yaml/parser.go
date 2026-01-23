package yaml

import (
	"fmt"
	"regexp"
)

type combinator[T any] func(*parserState) (*production[T], *ParsingError)

type production[T any] struct {
	state parserState
	value T
}

type parserState struct {
	line        int
	char        int
	remaining   string
	indentation int
	eofConsumed bool
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
		line:        0,
		char:        0,
		remaining:   source,
		eofConsumed: false,
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
		line:        line,
		char:        char,
		remaining:   ps.remaining[numBytes:],
		indentation: ps.indentation,
		eofConsumed: ps.eofConsumed,
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
	prod, err := first[YamlList, string](first(newDashListParser(0), repeated(parseEmptyLine)), first[string, string](optional(parseInlineWhiteSpace), parseEOF))(&initialState)
	if err != nil {
		return nil, err
	}
	return prod.value, nil
}

var parseEOLorEOF = or("end of line or end of file", parseEOF, parseLineBreak)

var parseWhiteSpace = repeated(or("whitespace", parseInlineWhiteSpace, parseEOLorEOF))

// func parseDashList(state *parserState) (production[YamlList], error) {
// 	return
// }

var parseEmptyLine = second(optional(parseInlineWhiteSpace), parseLineBreak)

func newDashListParser(numLeadingSpaces int) combinator[YamlList] {
	return mapOut(func(l []string) YamlList { return YamlList(l) },
		separatedBy(second(optional(parseInlineWhiteSpace), first(parseLineBreak, repeated(parseEmptyLine))), newSpaceIndentedParser(numLeadingSpaces, parseDashItem)))
}

var parseDashItem = second(newStringParser("-"), second(parseInlineWhiteSpace, parseUserId))

func newSpaceIndentedParser[T any](indentationLevel int, c combinator[T]) combinator[T] {
	return func(ps *parserState) (*production[T], *ParsingError) {

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

var parseUserId = newRegexParser("user ID", *regexp.MustCompile(`^([\w@\.\-]+)`))

var groupIdRegex = newRegexParser("group ID", *regexp.MustCompile(`^([\w\-]+):`))

func first[T, U any](firstc combinator[T], secondc combinator[U]) combinator[T] {
	return func(ps *parserState) (*production[T], *ParsingError) {
		prod1, err := firstc(ps)
		if err != nil {
			return nil, err
		}
		prod2, err := secondc(&prod1.state)
		if err != nil {
			return nil, err
		}
		return &production[T]{
			state: prod2.state,
			value: prod1.value,
		}, nil
	}
}

func second[T, U any](firstc combinator[T], secondc combinator[U]) combinator[U] {
	return func(ps *parserState) (*production[U], *ParsingError) {
		prod1, err := firstc(ps)
		if err != nil {
			return nil, err
		}
		prod2, err := secondc(&prod1.state)
		if err != nil {
			return nil, err
		}
		return &production[U]{
			state: prod2.state,
			value: prod2.value,
		}, nil
	}
}

func optional[T any](c combinator[T]) combinator[T] {
	return func(ps *parserState) (*production[T], *ParsingError) {
		prod, err := c(ps)
		if err != nil {
			return &production[T]{
				state: *ps,
			}, nil
		}
		return prod, nil
	}
}

func mapOut[T, U any](f func(T) U, c combinator[T]) combinator[U] {
	return func(ps *parserState) (*production[U], *ParsingError) {
		prod, err := c(ps)
		if err != nil {
			return nil, err
		}
		return &production[U]{
			state: prod.state,
			value: f(prod.value),
		}, nil
	}
}

func parseInlineWhiteSpace(state *parserState) (*production[string], *ParsingError) {
	for index, b := range state.remaining {
		if b != ' ' {
			if index == 0 {
				return nil, state.newError("expected inline whitespace")
			}
			return &production[string]{
				state: *state.Advance(index),
			}, nil
		}
	}
	if len(state.remaining) > 0 {
		return &production[string]{
			state: *state.Advance(len(state.remaining)),
		}, nil
	}
	return nil, state.newError("expected inline whitespace")
}

func newRegexParser(productionName string, re regexp.Regexp) combinator[string] {

	return func(state *parserState) (*production[string], *ParsingError) {
		submatches := re.FindStringSubmatch(state.remaining)
		if len(submatches) == 0 {
			return nil, state.newError(fmt.Sprintf("invalid %s", productionName))
		}

		return &production[string]{
			state: *state.Advance(len(submatches[0])),
			value: submatches[1],
		}, nil
	}

}

func newStringParser(s string) combinator[string] {
	return func(ps *parserState) (*production[string], *ParsingError) {

		if len(ps.remaining) >= len(s) && ps.remaining[:len(s)] == s {
			return &production[string]{
				state: *ps.Advance(len(s)),
				value: s}, nil
		}

		return nil, ps.newError(fmt.Sprintf("expected '%s'", s))
	}
}

func parseLineBreak(state *parserState) (*production[string], *ParsingError) {
	if len(state.remaining) == 0 {
		return nil, state.newError("expected line break")
	}

	if len(state.remaining) > 1 && state.remaining[0] == '\r' && state.remaining[1] == '\n' {
		return &production[string]{
			state: *state.Advance(2),
		}, nil
	} else if state.remaining[0] == '\n' || state.remaining[0] == '\r' {
		return &production[string]{
			state: *state.Advance(1),
		}, nil
	}
	return nil, state.newError("expected line break")
}

func parseEOF[T any](state *parserState) (*production[T], *ParsingError) {

	if len(state.remaining) != 0 {
		fmt.Printf("'%s'\n", state.remaining)
		return nil, state.newError("expected the end of the file")
	}
	if state.eofConsumed {
		return nil, state.newError("nothing left to parse")
	}
	return &production[T]{
		state: parserState{
			line:        state.line,
			char:        state.char,
			remaining:   state.remaining,
			indentation: state.indentation,
			eofConsumed: true,
		},
	}, nil
}

func or[T any](name string, combinators ...combinator[T]) combinator[T] {

	var latestError *ParsingError = nil
	return func(ps *parserState) (*production[T], *ParsingError) {
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

func atLeastOne[T any](name string, c combinator[[]T]) combinator[[]T] {
	return func(ps *parserState) (*production[[]T], *ParsingError) {
		prod, err := c(ps)
		if err != nil {
			if len(prod.value) == 0 {
				return nil, ps.newError(fmt.Sprintf("expected at least one %s", name))
			}
		}
		return prod, nil
	}
}

func repeated[T any](c combinator[T]) combinator[[]T] {
	return func(ps *parserState) (*production[[]T], *ParsingError) {
		var list []T
		var state *parserState = ps
		for {
			prod, err := c(state)
			if err != nil {
				return &production[[]T]{
					state: *state,
					value: list,
				}, nil
			}
			list = append(list, prod.value)
			state = &prod.state
		}
	}

}

func separatedBy[T, U any](separator combinator[T], c combinator[U]) combinator[[]U] {
	return func(ps *parserState) (*production[[]U], *ParsingError) {
		var list []U
		var afterItemState parserState = *ps
		var afterSeparatorState parserState = *ps
		for {
			prod1, err := c(&afterSeparatorState)
			if err != nil {
				return &production[[]U]{
					state: afterItemState,
					value: list,
				}, nil
			}
			list = append(list, prod1.value)
			afterItemState = prod1.state
			prod2, err := separator(&afterItemState)
			if err != nil {
				return &production[[]U]{
					state: afterItemState,
					value: list,
				}, nil
			}
			afterSeparatorState = prod2.state
		}
	}
}

type YamlMap map[string]YamlValue
type YamlList []string

func (m YamlMap) isYamlValue()  {}
func (m YamlList) isYamlValue() {}
