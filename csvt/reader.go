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
	buffer = strings.ReplaceAll(buffer, "\r\n", "\n")

	fragments := strings.Split(buffer, "\n\n")
	for _, v := range fragments {
		table := strings.TrimSpace(v)
		if table == "" {
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
