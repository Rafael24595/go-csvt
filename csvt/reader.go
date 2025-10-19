package csvt

import (
	"strings"
)

type reader struct {
	structures map[string][]string
}

func newReader() *reader {
	return &reader{
		structures: make(map[string][]string),
	}
}

func (r *reader) read(data []byte) (*table, error) {
	tables := map[string]nexus{}

	buffer := string(data)
	for len(buffer) > 0 {
		var table string
		table, buffer = r.readNext(buffer)
		if len(strings.TrimSpace(table)) == 0 {
			continue
		}

		nexus, err := parseTable(table)
		if err != nil {
			return nil, err
		}

		tables[nexus.key] = *nexus
	}
	tbl := newTable(tables)
	return &tbl, nil
}

func (d *reader) readNext(csv string) (string, string) {
	initial := 0
	headCount := 0

	end := 0
	tailCount := 0

	index := 0
	for i, v := range csv {
		index = i
		if v == TBL_HEAD_BASE {
			if headCount == 0 {
				initial = i
			}
			headCount++
		} else if v == TBL_HEAD_ROOT && headCount > 0 {
			headCount++
		} else if v == '\n' {
			if tailCount == 1 {
				end = i
				break
			}
			tailCount++
		} else {
			if headCount < 3 {
				headCount = 0
			}
			if tailCount < 2 {
				tailCount = 0
			}
		}
	}

	if index == len(csv)-1 {
		end = index
	}

	if end == 0 {
		return "", ""
	}

	return csv[initial:end], csv[end:]
}
