package main

import (
	"maps"
	"bytes"
	"fmt"
)

type testcase struct {
	name string
	in string
	out map[string]any
}

func main() {
	tests := []testcase {
		testcase{"empty json", "{}", make(map[string]any)},
		testcase{"one field", "{\"name\":\"ivan\"}", map[string]any{"name":"ivan"}},
		testcase{"extra space", "{\"name\"  : \"ivan\"}", map[string]any{"name":"ivan"}},
	}

	for i := range tests {
		got, error := Parse(tests[i].in)
		if error != nil {
			panic(fmt.Sprintf("Test failed %s; expected %v but got %v", tests[i].name, tests[i].out, error))
		}
		if !maps.Equal(got, tests[i].out) {
			panic(fmt.Sprintf("Test failed %s; expected %v but got %v", tests[i].name, tests[i].out, got))
		}
	}
}

func Parse(json string) (map[string]any, error) {
	return parse(json, 0, len(json) - 1)
}

func isBalancedBrackets(json string, pos int, end int) bool {
	i := pos
	target := json[i]
	var balanced_c byte
	balanced := false

	switch target {
	case '{':
		balanced_c = '}'
	case '[':
		balanced_c = ']'
	}

	for end >= i {
		if (json[end] == balanced_c) {
			end--
			balanced = true
			break
		}
		end--
	}

	return balanced
}

func parse(json string, beg int, end int) (map[string]any, error) {
	builder := make(map[string]any)

	skipWhitespace(json, &beg, end)
	for beg < end {
		switch json[beg] {
		case '{':
			if !isBalancedBrackets(json, beg, end) {
				return nil, fmt.Errorf("Parse body error disbalanced pos: %d", beg)
			}
			skipWhitespace(json, &beg, end)
		case '}':
			return builder, nil
		case '"':
			err := readKeyValue(&builder, json, &beg, end)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Unexpected char %c pos %d", json[beg], beg)
		}
		beg++
	}

	return builder, nil
}

func skipWhitespace(json string, beg* int, end int) {
	i := *beg

	for i < end && (json[i] == ' ' || json[i] == '\t' || json[i] == '\n') {
		i++
	}

	*beg = i
}

func readStringLeteral(json string, beg* int, end int) (string, error) {
	var key bytes.Buffer
	i := *beg
	closed := false

	skipWhitespace(json, beg, end)

	if json[i] != '"' {
		return "", fmt.Errorf("Unexpected literal beggining pos: %d", i)
	}
	i++

	for i < end && !closed {
//		fmt.Printf("%s\n", key.String())

		switch json[i] {
		case '"':
			closed = true
		default:
			if !closed {
				key.WriteByte(json[i])
			}
		}

		i++
	}

	*beg = i
	return key.String(), nil
}

func readKeyValue(builder* map[string]any, json string, beg* int, end int) error {
	key, err := readStringLeteral(json, beg, end)
	if err != nil {
		return err
	}

	skipWhitespace(json, beg, end)

	if json[*beg] != ':' {
		return fmt.Errorf("Invalid key value pair")
	}
	*beg++
	skipWhitespace(json, beg, end)

	value, err := readStringLeteral(json, beg, end)
	if err != nil {
		return err
	}

	(*builder)[key] = value
	return nil
}
