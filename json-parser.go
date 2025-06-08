package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

type testcase struct {
	name string
	in   string
	out  any
}

func main() {
	tests := []testcase{
		testcase{"empty json", "{}", make(map[string]any)},
		testcase{"one field", "{\"name\":\"ivan\"}", map[string]any{"name": "ivan"}},
		testcase{"a few fields", "{\"name\":\"ivan\",\"surname\":\"prikaznov\"}", map[string]any{"name": "ivan", "surname": "prikaznov"}},
		testcase{
			"a internal object",
			"{\"name\":\"ivan\",\"info\":{\"surname\":\"prikaznov\"}}",
			map[string]any{"info": map[string]any{"surname": "prikaznov"}, "name": "ivan"},
		},
		testcase{"an array", "[]", make([]any, 0)},
		testcase{"an array with string literals", "[\"Hello\",\"World\"]", []any{"Hello", "World"}},
		testcase{"an array with objects", "[{\"name\":\"ivan\"}]", []any{map[string]any{"name":"ivan"}}},
		testcase{"an array with empty object", "[{}]", []any{map[string]any{}}},
		testcase{"an array with an array with empty object", "[[{}]]", []any{[]any{map[string]any{}}}},
		testcase{"an array with objects and string literal", "[{\"name\":\"ivan\"},\"hello\"]", []any{map[string]any{"name":"ivan"}, "hello"}},
		testcase{"object with array field", "{\"names\":[\"Vasya\",\"Ivan\"]}", map[string]any{"names":[]any{"Vasya","Ivan"}}},
		testcase{"array of integers", "[1,2,3]", []any{1,2,3}},
		testcase{"object with array integers field", "{\"ages\":[21,23]}", map[string]any{"ages":[]any{21,23}}},
	}

	for i := range tests {
		got, error := Parse(tests[i].in)

		fmt.Printf("%v\n", got)

		if error != nil {
			panic(fmt.Sprintf("Test failed %s; expected %v but got %v", tests[i].name, tests[i].out, error))
		}
		if !reflect.DeepEqual(got, tests[i].out) {
			panic(fmt.Sprintf("Test failed %s; expected %v but got %v", tests[i].name, tests[i].out, got))
		}
	}
}

func Parse(json string) (any, error) {
	var pos = 0
	return parse(json, &pos, len(json)-1)
}

func parse(json string, pos *int, end int) (any, error) {
	switch json[*pos] {
	case '{':
		obj, err := parseObject(json, pos, end)
		if err != nil {
			return nil, err
		}
		return obj, nil
	case '[':
		arr, err := parseArray(json, pos, end)
		if err != nil {
			return nil, err
		}
		return arr, nil
	default:
		return nil, fmt.Errorf("Unexpected char %c pos %d", json[*pos], *pos)
	}
}

func parseArray(json string, pos *int, end int) ([]any, error) {
	if !isBalancedBrackets(json, *pos, end) {
		return nil, fmt.Errorf("Parse array error disbalanced pos: %d", *pos)
	}
	if json[*pos] != '[' {
		return nil, fmt.Errorf("Unexpected array beggining pos: %d", *pos)
	}

	builder := make([]any, 0)

	for *pos < end {
		*pos++

		switch json[*pos] {
		case '"':
			str, err := readStringLeteral(json, pos, end)
			if err != nil {
				return nil, err
			}
			builder = append(builder, str)
		case '[':
			arr, err := parseArray(json, pos, end)
			if err != nil {
				return nil, err
			}
			builder = append(builder, arr)
		case '{':
			obj, err := parseObject(json, pos, end)
			if err != nil {
				return nil, err
			}
			builder = append(builder, obj)
		case ']':
			return builder, nil
		case ',':
		default:
			if isNumber(json[*pos]) {
				i, err := parseNumber(json, pos, end)
				if err != nil {
					return nil, err
				}
				builder = append(builder, i)
			} else {
				return nil, fmt.Errorf("Unexpected char: %c", json[*pos])
			}
		}
	}

	return builder, nil
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
		if json[end] == balanced_c {
			end--
			balanced = true
			break
		}
		end--
	}

	return balanced
}

func parseObject(json string, pos *int, end int) (map[string]any, error) {
	if json[*pos] != '{' {
		return nil, fmt.Errorf("Unexpected array beggining pos: %d", *pos)
	}

	builder := make(map[string]any)

	for *pos < end {
		*pos++

		switch json[*pos] {
		case '{':
			if !isBalancedBrackets(json, *pos, end) {
				return nil, fmt.Errorf("Parse body error disbalanced pos: %d", *pos)
			}
		case '}':
			return builder, nil
		case '"':
			err := readKeyValue(&builder, json, pos, end)
			if err != nil {
				return nil, err
			}
		case ',':
		default:
			return nil, fmt.Errorf("Unexpected char %c pos %d", json[*pos], *pos)
		}
	}

	return builder, nil
}

func readStringLeteral(json string, pos *int, end int) (string, error) {
	if json[*pos] != '"' {
		return "", fmt.Errorf("Unexpected literal beggining pos: %d", *pos)
	}

	var key bytes.Buffer
	closed := false

	for *pos < end && !closed {
		*pos++

		switch json[*pos] {
		case '"':
			closed = true
		default:
			if !closed {
				key.WriteByte(json[*pos])
			}
		}
	}

	return key.String(), nil
}

func isNumber(char byte) bool {
	return char > 47 && char < 58
}

func parseNumber(json string, pos *int, end int) (int, error) {
	var number bytes.Buffer
	closed := false

	for *pos < end && !closed {
		next := json[*pos]

		switch {
		case isNumber(next):
			number.WriteByte(next)
			*pos++
		case next == ',' || next == ']':
			*pos--
			closed = true
		default:
			return 0, fmt.Errorf("Unexpected char while parsing number pos %d %c", *pos, next)
		}
	}

	return strconv.Atoi(number.String())
}

func readKeyValue(builder *map[string]any, json string, pos *int, end int) error {
	key, err := readStringLeteral(json, pos, end)
	if err != nil {
		return err
	}
	*pos++

	if json[*pos] != ':' {
		return fmt.Errorf("Invalid key value pair %d", *pos)
	}
	*pos++

	var value any

	switch json[*pos] {
	case '{':
		obj, err := parseObject(json, pos, end)
		if err != nil {
			return err
		}
		value = obj
	case '[':
		arr, err := parseArray(json, pos, end)
		if err != nil {
			return err
		}
		value = arr
	case '"':
		str, err := readStringLeteral(json, pos, end)
		if err != nil {
			return err
		}
		value = str
	default:
		return fmt.Errorf("Unexpected char on pos", *pos)
	}

	(*builder)[key] = value
	return nil
}
