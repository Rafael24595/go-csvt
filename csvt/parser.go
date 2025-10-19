package csvt

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func parseTable(table string) (*nexus, error) {
	root := false

	items := []group{}

	fragments := strings.Split(table, "\n")

	if strings.Contains(fragments[0], string(TBL_HEAD_ROOT)+string(TBL_HEAD_ROOT)) {
		root = true
	}

	re := regexp.MustCompile(`/\*\*\s|///\s`)
	name := re.ReplaceAllString(fragments[0], "")

	heads := parseHeaders(fragments[1])

	for _, v := range fragments[2:] {
		if len(v) == 0 {
			continue
		}
		row, err := parseRow(v, heads)
		if err != nil {
			return nil, err
		}
		items = append(items, *row)
	}

	result := newNexus(name, root, items)

	return &result, nil
}

func parseHeaders(row string) []string {
	re := regexp.MustCompile(`[A-Za-z0-9]+->\s`)
	row = re.ReplaceAllString(row, "")
	if row == "" {
		return []string{}
	}
	return strings.Split(row, string(HEA_SEPARATOR))
}

func parseRow(row string, header []string) (*group, error) {
	re := regexp.MustCompile(`\d+-> `)
	row = re.ReplaceAllString(row, "")

	instance := categoryOf(row, len(header) != 0)

	var group interface{}
	var err error
	switch instance {
	case MAP:
		group, err = parseMap(row)
	case ARR:
		group, err = parseArray(row)
	case STR:
		group, err = parseStructure(row)
	case OBJ:
		group, err = parseObject(row)
	default:
		err = fmt.Errorf("row type not recognized: \n%s", row)
	}

	if err != nil {
		return nil, err
	}

	result := newGroup(instance, header, group)
	return &result, nil
}

func categoryOf(row string, header bool) category {
	switch rune(row[len(row)-1]) {
	case ARR_CLOSING:
		return ARR
	case MAP_CLOSING:
		return MAP
	case STR_CLOSING:
		return STR
	}

	if !header {
		return OBJ
	}

	return STR
}

func parseMap(row string) (map[string]node, error) {
	mapp := map[string]node{}

	if rune(row[len(row)-1]) != MAP_CLOSING {
		return nil, errors.New("invalid map closing character")
	}

	row = row[:len(row)-1]

	buffer := row
	for len(buffer) > 0 {
		var index int
		if buffer[0] == '"' {
			index = strings.Index(buffer[1:], "\"") + 2
		} else {
			index = strings.Index(buffer, string(MAP_LINKER))
		}

		if index == -1 {
			return nil, errors.New("undefined value")
		}

		key := buffer[:index]
		buffer = buffer[index+1:]

		node, err := parseObject(key)
		if err != nil {
			return nil, err
		}
		key = node.key()

		if buffer[0] == '"' {
			index = strings.Index(buffer[1:], "\"") + 1
			if index < len(buffer)-1 && buffer[index+1] == byte(MAP_SEPARATOR) {
				index = index + 1
			} else {
				index = -1
			}
		} else {
			index = strings.Index(buffer, string(MAP_SEPARATOR))
		}

		var content string
		if index != -1 {
			if len(buffer) >= index && rune(buffer[index]) != MAP_SEPARATOR {
				return nil, errors.New("invalid map entry")
			}
			content = buffer[:index]
			buffer = buffer[index+1:]
		} else {
			content = buffer
			buffer = ""
		}

		node, err = parseObject(content)
		if err != nil {
			return nil, err
		}
		mapp[key] = node
	}

	return mapp, nil
}

func parseArray(row string) ([]node, error) {
	return parseList(row, ARR_SEPARATOR, ARR_CLOSING)
}

func parseStructure(row string) ([]node, error) {
	return parseList(row, STR_SEPARATOR, STR_CLOSING)
}

func parseList(row string, separator, closing rune) ([]node, error) {
	lst := []node{}

	if rune(row[len(row)-1]) != closing {
		return nil, errors.New("invalid list closing character")
	}

	row = row[:len(row)-1]

	buffer := row
	for len(buffer) > 0 {
		var index int
		if buffer[0] == '"' {
			index = strings.Index(buffer[1:], "\"") + 2
			if len(buffer) == index {
				index = -1
			}
		} else {
			index = strings.Index(buffer, string(separator))
		}

		var content string
		if index != -1 {
			if len(buffer) >= index && rune(buffer[index]) != separator {
				return nil, errors.New("invalid list separator character")
			}
			content = buffer[:index]
			buffer = buffer[index+1:]
		} else {
			content = buffer
			buffer = ""
		}

		node, err := parseObject(content)
		if err != nil {
			return nil, err
		}
		lst = append(lst, node)
	}

	return lst, nil
}

func parseObject(obj string) (node, error) {
	if len(obj) == 0 {
		return fromEmpty(), nil
	}
	if v, i, ok, err := isPointer(obj); ok {
		if err != nil {
			return node{}, nil
		}
		return fromPointer(v, i), nil
	}
	if v, ok := isString(obj); ok {
		return fromNonPointer(v), nil
	}
	lower := strings.ToLower(obj)
	if lower == "false" {
		return fromNonPointer(false), nil
	}
	if lower == "true" {
		return fromNonPointer(true), nil
	}
	if strings.Contains(obj, ".") {
		if v, err := strconv.ParseFloat(obj, 64); err == nil {
			return fromNonPointer(v), nil
		}
	}
	if v, err := strconv.Atoi(obj); err == nil {
		return fromNonPointer(v), nil
	}

	return node{}, fmt.Errorf("type not recognized: \n%s", obj)
}

func isPointer(obj string) (string, int, bool, error) {
	if obj[0] != byte(PTR_HEADER) {
		return obj, 0, false, nil
	}

	fragments := strings.Split(obj[1:], string(PTR_SEPARATOR))

	key := fragments[0]
	index := 0

	index, err := strconv.Atoi(fragments[1])
	if err != nil {
		err := fmt.Errorf("index \"%s\" type not recognized: %s", fragments[1], err.Error())
		return "", 0, false, err
	}

	return key, index, true, nil
}

func isString(obj string) (string, bool) {
	len := len(obj)
	if obj[0] == '"' && obj[len-1] == '"' {
		fixed := strings.ReplaceAll(obj[1:len-1], "\\'", "\"")
		replacer := strings.NewReplacer(
			"\\\\", "\\",
			"\\\"", "\"",
			"\\n", "\n",
			"\\r", "\r",
			"\\t", "\t",
			"\\b", "\b",
			"\\f", "\f")
		fixed = replacer.Replace(fixed)
		return fixed, true
	}
	return obj, false
}
