package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"os"
)

type testcase struct {
	name string
	in   string
	out  any
}

func main() {
	if len(os.Args) == 2 {
		path := os.Args[1]
		bytes, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		got, err := Parse(string(bytes))
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", got)
	}

	runTest()
}

func runTest() {
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
		testcase{"an array with objects", "[{\"name\":\"ivan\"}]", []any{map[string]any{"name": "ivan"}}},
		testcase{"an array with empty object", "[{}]", []any{map[string]any{}}},
		testcase{"an array with an array with empty object", "[[{}]]", []any{[]any{map[string]any{}}}},
		testcase{"an array with objects and string literal", "[{\"name\":\"ivan\"},\"hello\"]", []any{map[string]any{"name": "ivan"}, "hello"}},
		testcase{"object with array field", "{\"names\":[\"Vasya\",\"Ivan\"]}", map[string]any{"names": []any{"Vasya", "Ivan"}}},
		testcase{"array of integers", "[1,2,3]", []any{1, 2, 3}},
		testcase{"object with array integers field", "{\"ages\":[21,23]}", map[string]any{"ages": []any{21, 23}}},
		testcase{"object with boolean value", "{\"isValid\":true}", map[string]any{"isValid": true}},
		testcase{"object with boolean value", "{\"isValid\":false}", map[string]any{"isValid": false}},
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
	if json[*pos] != '[' {
		return nil, fmt.Errorf("Unexpected array beggining pos: %d", *pos)
	}

	builder := make([]any, 0)

	for *pos < end {
		*pos++

		switch json[*pos] {
		case ',':
		case ']':
			return builder, nil
		default:
			value, err := parseValue(json, pos, end)
			if err != nil {
				return nil, err
			}
			builder = append(builder, value)
		}
	}

	return builder, nil
}



func parseObject(json string, pos *int, end int) (map[string]any, error) {
	if json[*pos] != '{' {
		return nil, fmt.Errorf("Unexpected array beggining pos: %d", *pos)
	}

	builder := make(map[string]any)

	for *pos < end {
		*pos++

		switch json[*pos] {
		case '}':
			return builder, nil
		case '"':
			key, value, err := parseKeyValue(json, pos, end)
			if err != nil {
				return nil, err
			}
			builder[key] = value
		case ',':
		default:
			return nil, fmt.Errorf("Unexpected char %c pos %d", json[*pos], *pos)
		}
	}

	return builder, nil
}

func parseString(json string, pos *int, end int) (string, error) {
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

func parseKeyValue(json string, pos *int, end int) (string, any, error) {
	key, err := parseString(json, pos, end)
	if err != nil {
		return "", nil, err
	}
	*pos++

	if json[*pos] != ':' {
		return "", nil, fmt.Errorf("Invalid key value pair %d", *pos)
	}
	*pos++

	value, err := parseValue(json, pos, end)
	if err != nil {
		return "", nil, err
	}

	return key, value, nil
}

func parseBoolean(json string, pos *int) (bool, error) {
	switch json[*pos] {
	case 't':
		str_true := json[*pos:(*pos + 4)]
		if str_true == "true" {
			*pos = *pos + 3
			return true, nil
		} else {
			return false, fmt.Errorf("Failed parse boolean true value %v", str_true)
		}
	case 'f':
		str_false := json[*pos:(*pos + 5)]
		if str_false == "false" {
			*pos = *pos + 4
			return false, nil
		} else {
			return false, fmt.Errorf("Failed parse boolean false value %v", str_false)
		}
	default:
		return false, fmt.Errorf("Failed parse boolean value pos %d", *pos)
	}
}

func parseValue(json string, pos *int, end int) (any, error) {
	switch json[*pos] {
	case '{':
		return parseObject(json, pos, end)
	case '[':
		return parseArray(json, pos, end)
	case '"':
		return parseString(json, pos, end)
	case 't':
		return parseBoolean(json, pos)
	case 'f':
		return parseBoolean(json, pos)
	default:
		if isNumber(json[*pos]) {
			return parseNumber(json, pos, end)
		}
		return nil, fmt.Errorf("Unexpected char on pos", *pos)
	}
}
