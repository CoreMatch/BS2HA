package bsdump

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Value struct {
	Valid bool
	Text  string
}

type Row map[string]Value

func (r Row) Get(name string) (string, bool) {
	value, ok := r[name]
	if !ok || !value.Valid {
		return "", false
	}
	return value.Text, true
}

type Dump struct {
	Tables     map[string][]Row
	ColumnDefs map[string][]string
}

func ParseFile(path string) (*Dump, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(data))
}

func Parse(sql string) (*Dump, error) {
	dump := &Dump{
		Tables:     map[string][]Row{},
		ColumnDefs: map[string][]string{},
	}

	statements := splitStatements(sql)
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		trimmed = stripLeadingComments(trimmed)
		trimmed = strings.TrimSuffix(trimmed, ";")
		if trimmed == "" {
			continue
		}

		if table, cols, ok := parseCreateTable(trimmed); ok {
			dump.ColumnDefs[table] = cols
			continue
		}

		table, rows, err := parseInsert(trimmed, dump.ColumnDefs)
		if err == nil && table != "" {
			dump.Tables[table] = append(dump.Tables[table], rows...)
		}
	}

	return dump, nil
}

func splitStatements(sql string) []string {
	var out []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	inBacktick := false
	escaped := false

	for _, r := range sql {
		current.WriteRune(r)

		if escaped {
			escaped = false
			continue
		}

		switch r {
		case '\\':
			if inSingle || inDouble {
				escaped = true
			}
		case '\'':
			if !inDouble && !inBacktick {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle && !inBacktick {
				inDouble = !inDouble
			}
		case '`':
			if !inSingle && !inDouble {
				inBacktick = !inBacktick
			}
		case ';':
			if !inSingle && !inDouble && !inBacktick {
				out = append(out, current.String())
				current.Reset()
			}
		}
	}

	if strings.TrimSpace(current.String()) != "" {
		out = append(out, current.String())
	}
	return out
}

func stripLeadingComments(stmt string) string {
	for {
		trimmed := strings.TrimSpace(stmt)
		switch {
		case strings.HasPrefix(trimmed, "--"):
			if idx := strings.Index(trimmed, "\n"); idx >= 0 {
				stmt = trimmed[idx+1:]
				continue
			}
			return ""
		case strings.HasPrefix(trimmed, "/*"):
			end := strings.Index(trimmed, "*/")
			if end >= 0 {
				stmt = trimmed[end+2:]
				continue
			}
			return ""
		default:
			return trimmed
		}
	}
}

var createTablePattern = regexp.MustCompile("(?is)^CREATE\\s+TABLE\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?`?([a-zA-Z0-9_]+)`?\\s*\\((.*)\\)\\s*(?:ENGINE|DEFAULT|CHARSET|COLLATE|COMMENT|ROW_FORMAT|$).*")
var columnLinePattern = regexp.MustCompile("(?m)^\\s*`([^`]+)`\\s+")
var insertPattern = regexp.MustCompile("(?is)^INSERT\\s+INTO\\s+`?([a-zA-Z0-9_]+)`?\\s*(?:\\((.*?)\\))?\\s*VALUES\\s*(.*)$")

func parseCreateTable(stmt string) (string, []string, bool) {
	matches := createTablePattern.FindStringSubmatch(stmt)
	if len(matches) != 3 {
		return "", nil, false
	}

	table := matches[1]
	body := matches[2]
	columnMatches := columnLinePattern.FindAllStringSubmatch(body, -1)
	if len(columnMatches) == 0 {
		return "", nil, false
	}

	cols := make([]string, 0, len(columnMatches))
	for _, match := range columnMatches {
		cols = append(cols, match[1])
	}
	return table, cols, true
}

func parseInsert(stmt string, schemas map[string][]string) (string, []Row, error) {
	matches := insertPattern.FindStringSubmatch(strings.TrimSpace(stmt))
	if len(matches) != 4 {
		return "", nil, fmt.Errorf("not an insert")
	}

	table := matches[1]
	columnList := strings.TrimSpace(matches[2])
	rawValues := strings.TrimSpace(matches[3])

	var columns []string
	if columnList != "" {
		columns = parseColumnList(columnList)
	} else {
		columns = schemas[table]
	}
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("missing column definition for table %s", table)
	}

	tuples, err := parseTuples(rawValues)
	if err != nil {
		return "", nil, err
	}

	rows := make([]Row, 0, len(tuples))
	for _, tuple := range tuples {
		if len(tuple) != len(columns) {
			return "", nil, fmt.Errorf("column count mismatch for table %s", table)
		}
		row := Row{}
		for i, col := range columns {
			row[col] = tuple[i]
		}
		rows = append(rows, row)
	}

	return table, rows, nil
}

func parseColumnList(raw string) []string {
	parts := strings.Split(raw, ",")
	cols := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "`")
		if part != "" {
			cols = append(cols, part)
		}
	}
	return cols
}

func parseTuples(raw string) ([][]Value, error) {
	var tuples [][]Value
	i := 0
	for i < len(raw) {
		for i < len(raw) && (raw[i] == ' ' || raw[i] == '\n' || raw[i] == '\r' || raw[i] == '\t' || raw[i] == ',') {
			i++
		}
		if i >= len(raw) {
			break
		}
		if raw[i] != '(' {
			return nil, fmt.Errorf("expected tuple start")
		}
		i++

		values := []Value{}
		for {
			for i < len(raw) && (raw[i] == ' ' || raw[i] == '\n' || raw[i] == '\r' || raw[i] == '\t') {
				i++
			}
			value, next, err := parseValue(raw, i)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
			i = next

			for i < len(raw) && (raw[i] == ' ' || raw[i] == '\n' || raw[i] == '\r' || raw[i] == '\t') {
				i++
			}
			if i >= len(raw) {
				return nil, fmt.Errorf("unterminated tuple")
			}
			if raw[i] == ',' {
				i++
				continue
			}
			if raw[i] == ')' {
				i++
				break
			}
			return nil, fmt.Errorf("unexpected tuple separator %q", raw[i])
		}

		tuples = append(tuples, values)
	}

	return tuples, nil
}

func parseValue(raw string, start int) (Value, int, error) {
	if start >= len(raw) {
		return Value{}, start, fmt.Errorf("unexpected end of input")
	}

	if raw[start] == '\'' {
		var out strings.Builder
		i := start + 1
		for i < len(raw) {
			switch raw[i] {
			case '\\':
				if i+1 >= len(raw) {
					return Value{}, i, fmt.Errorf("bad escape sequence")
				}
				i++
				switch raw[i] {
				case '0':
					out.WriteByte(0)
				case 'n':
					out.WriteByte('\n')
				case 'r':
					out.WriteByte('\r')
				case 't':
					out.WriteByte('\t')
				default:
					out.WriteByte(raw[i])
				}
				i++
			case '\'':
				if i+1 < len(raw) && raw[i+1] == '\'' {
					out.WriteByte('\'')
					i += 2
					continue
				}
				return Value{Valid: true, Text: out.String()}, i + 1, nil
			default:
				out.WriteByte(raw[i])
				i++
			}
		}
		return Value{}, i, fmt.Errorf("unterminated string literal")
	}

	i := start
	for i < len(raw) && raw[i] != ',' && raw[i] != ')' {
		i++
	}
	token := strings.TrimSpace(raw[start:i])
	if strings.EqualFold(token, "NULL") {
		return Value{}, i, nil
	}
	return Value{Valid: true, Text: token}, i, nil
}
