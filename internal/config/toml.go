package config

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type tomlStatement struct {
	text string
	line int
}

// ParseTOML parses a TOML v1.0.0 document into a nested map.
func ParseTOML(data []byte) (map[string]any, error) {
	root := make(map[string]any)
	current := root
	statements, err := splitStatements(string(data))
	if err != nil {
		return nil, err
	}
	for _, stmt := range statements {
		line := strings.TrimSpace(stmt.text)
		if line == "" {
			continue
		}
		clean := stripComments(line)
		clean = strings.TrimSpace(clean)
		if clean == "" {
			continue
		}
		if strings.HasPrefix(clean, "[[") {
			path, err := parseTablePath(clean, 2)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
			}
			arr, err := getOrCreateArrayTable(root, path)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
			}
			entry := make(map[string]any)
			arr = append(arr, entry)
			if err := setPath(root, path, arr); err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
			}
			current = entry
			continue
		}
		if strings.HasPrefix(clean, "[") {
			path, err := parseTablePath(clean, 1)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
			}
			current, err = getOrCreateTable(root, path)
			if err != nil {
				return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
			}
			continue
		}
		key, valueRaw, err := splitKeyValue(clean)
		if err != nil {
			return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
		}
		value, err := parseValue(valueRaw)
		if err != nil {
			return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
		}
		if err := setKey(current, key, value); err != nil {
			return nil, fmt.Errorf("parse toml: line %d: %w", stmt.line, err)
		}
	}
	return root, nil
}

func splitStatements(data string) ([]tomlStatement, error) {
	scanner := bufio.NewScanner(strings.NewReader(data))
	var (
		statements []tomlStatement
		buf        strings.Builder
		startLine  int
	)
	state := statementState{}
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if buf.Len() == 0 {
			startLine = lineNo
		}
		buf.WriteString(line)
		buf.WriteString("\n")
		state.scan(line)
		if state.complete() {
			statements = append(statements, tomlStatement{text: buf.String(), line: startLine})
			buf.Reset()
			startLine = 0
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse toml: %w", err)
	}
	if buf.Len() > 0 {
		statements = append(statements, tomlStatement{text: buf.String(), line: startLine})
	}
	return statements, nil
}

type statementState struct {
	inBasic        bool
	inLiteral      bool
	inMultiBasic   bool
	inMultiLiteral bool
	bracketDepth   int
	braceDepth     int
	lastWasEscape  bool
}

func (s *statementState) scan(line string) {
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if s.inMultiBasic {
			if r == '\\' {
				s.lastWasEscape = !s.lastWasEscape
				continue
			}
			if r == '"' && !s.lastWasEscape && hasTriple(runes, i, '"') {
				s.inMultiBasic = false
				i += 2
			}
			s.lastWasEscape = false
			continue
		}
		if s.inMultiLiteral {
			if r == '\'' && hasTriple(runes, i, '\'') {
				s.inMultiLiteral = false
				i += 2
			}
			continue
		}
		if s.inBasic {
			if r == '\\' {
				s.lastWasEscape = !s.lastWasEscape
				continue
			}
			if r == '"' && !s.lastWasEscape {
				s.inBasic = false
			}
			s.lastWasEscape = false
			continue
		}
		if s.inLiteral {
			if r == '\'' {
				s.inLiteral = false
			}
			continue
		}
		s.lastWasEscape = false
		switch r {
		case '"':
			if hasTriple(runes, i, '"') {
				s.inMultiBasic = true
				i += 2
				continue
			}
			s.inBasic = true
		case '\'':
			if hasTriple(runes, i, '\'') {
				s.inMultiLiteral = true
				i += 2
				continue
			}
			s.inLiteral = true
		case '[':
			s.bracketDepth++
		case ']':
			if s.bracketDepth > 0 {
				s.bracketDepth--
			}
		case '{':
			s.braceDepth++
		case '}':
			if s.braceDepth > 0 {
				s.braceDepth--
			}
		}
	}
}

func (s statementState) complete() bool {
	return !s.inBasic && !s.inLiteral && !s.inMultiBasic && !s.inMultiLiteral && s.bracketDepth == 0 && s.braceDepth == 0
}

func hasTriple(runes []rune, idx int, target rune) bool {
	return idx+2 < len(runes) && runes[idx] == target && runes[idx+1] == target && runes[idx+2] == target
}

func stripComments(line string) string {
	var b strings.Builder
	runes := []rune(line)
	inBasic := false
	inLiteral := false
	inMultiBasic := false
	inMultiLiteral := false
	var lastWasEscape bool
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inMultiBasic {
			if r == '\\' {
				lastWasEscape = !lastWasEscape
				b.WriteRune(r)
				continue
			}
			if r == '"' && !lastWasEscape && hasTriple(runes, i, '"') {
				inMultiBasic = false
				b.WriteString("\"\"\"")
				i += 2
				continue
			}
			lastWasEscape = false
			b.WriteRune(r)
			continue
		}
		if inMultiLiteral {
			if r == '\'' && hasTriple(runes, i, '\'') {
				inMultiLiteral = false
				b.WriteString("'''")
				i += 2
				continue
			}
			b.WriteRune(r)
			continue
		}
		if inBasic {
			if r == '\\' {
				lastWasEscape = !lastWasEscape
				b.WriteRune(r)
				continue
			}
			if r == '"' && !lastWasEscape {
				inBasic = false
			}
			lastWasEscape = false
			b.WriteRune(r)
			continue
		}
		if inLiteral {
			if r == '\'' {
				inLiteral = false
			}
			b.WriteRune(r)
			continue
		}
		if r == '#' {
			break
		}
		switch r {
		case '"':
			if hasTriple(runes, i, '"') {
				inMultiBasic = true
				b.WriteString("\"\"\"")
				i += 2
				continue
			}
			inBasic = true
		case '\'':
			if hasTriple(runes, i, '\'') {
				inMultiLiteral = true
				b.WriteString("'''")
				i += 2
				continue
			}
			inLiteral = true
		}
		b.WriteRune(r)
	}
	return b.String()
}

func parseTablePath(line string, brackets int) ([]string, error) {
	end := strings.LastIndex(line, "]")
	if end == -1 {
		return nil, fmt.Errorf("table missing closing bracket")
	}
	content := strings.TrimSpace(line[brackets : end-(brackets-1)])
	if content == "" {
		return nil, fmt.Errorf("empty table name")
	}
	return parseKeyPath(content)
}

func splitKeyValue(line string) (string, string, error) {
	idx := splitOnEquals(line)
	if idx == -1 {
		return "", "", fmt.Errorf("missing '='")
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	if key == "" {
		return "", "", fmt.Errorf("empty key")
	}
	if value == "" {
		return "", "", fmt.Errorf("empty value")
	}
	return key, value, nil
}

func splitOnEquals(line string) int {
	runes := []rune(line)
	inBasic := false
	inLiteral := false
	inMultiBasic := false
	inMultiLiteral := false
	depthBracket := 0
	depthBrace := 0
	var lastWasEscape bool
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inMultiBasic {
			if r == '\\' {
				lastWasEscape = !lastWasEscape
				continue
			}
			if r == '"' && !lastWasEscape && hasTriple(runes, i, '"') {
				inMultiBasic = false
				i += 2
			}
			lastWasEscape = false
			continue
		}
		if inMultiLiteral {
			if r == '\'' && hasTriple(runes, i, '\'') {
				inMultiLiteral = false
				i += 2
			}
			continue
		}
		if inBasic {
			if r == '\\' {
				lastWasEscape = !lastWasEscape
				continue
			}
			if r == '"' && !lastWasEscape {
				inBasic = false
			}
			lastWasEscape = false
			continue
		}
		if inLiteral {
			if r == '\'' {
				inLiteral = false
			}
			continue
		}
		switch r {
		case '"':
			if hasTriple(runes, i, '"') {
				inMultiBasic = true
				i += 2
				continue
			}
			inBasic = true
		case '\'':
			if hasTriple(runes, i, '\'') {
				inMultiLiteral = true
				i += 2
				continue
			}
			inLiteral = true
		case '[':
			depthBracket++
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
		case '{':
			depthBrace++
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
		case '=':
			if depthBracket == 0 && depthBrace == 0 {
				return i
			}
		}
	}
	return -1
}

func parseKeyPath(raw string) ([]string, error) {
	segments := splitOnDots(raw)
	if len(segments) == 0 {
		return nil, fmt.Errorf("empty key")
	}
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		trimmed := strings.TrimSpace(segment)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid key")
		}
		if strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "'") {
			parsed, err := parseStringValue(trimmed)
			if err != nil {
				return nil, err
			}
			parts = append(parts, parsed)
			continue
		}
		parts = append(parts, trimmed)
	}
	return parts, nil
}

func splitOnDots(raw string) []string {
	var segments []string
	var buf strings.Builder
	runes := []rune(raw)
	inBasic := false
	inLiteral := false
	var lastWasEscape bool
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inBasic {
			if r == '\\' {
				lastWasEscape = !lastWasEscape
				buf.WriteRune(r)
				continue
			}
			if r == '"' && !lastWasEscape {
				inBasic = false
			}
			lastWasEscape = false
			buf.WriteRune(r)
			continue
		}
		if inLiteral {
			if r == '\'' {
				inLiteral = false
			}
			buf.WriteRune(r)
			continue
		}
		switch r {
		case '"':
			inBasic = true
		case '\'':
			inLiteral = true
		case '.':
			segments = append(segments, buf.String())
			buf.Reset()
			continue
		}
		buf.WriteRune(r)
	}
	if buf.Len() > 0 {
		segments = append(segments, buf.String())
	}
	return segments
}

func parseValue(raw string) (any, error) {
	parser := &valueParser{data: []rune(strings.TrimSpace(raw))}
	value, err := parser.parseValue()
	if err != nil {
		return nil, err
	}
	parser.skipSpace()
	if parser.more() {
		return nil, fmt.Errorf("unexpected trailing data")
	}
	return value, nil
}

type valueParser struct {
	data []rune
	pos  int
}

func (p *valueParser) more() bool {
	return p.pos < len(p.data)
}

func (p *valueParser) peek() rune {
	if p.pos >= len(p.data) {
		return 0
	}
	return p.data[p.pos]
}

func (p *valueParser) next() rune {
	if p.pos >= len(p.data) {
		return 0
	}
	r := p.data[p.pos]
	p.pos++
	return r
}

func (p *valueParser) skipSpace() {
	for p.more() {
		r := p.peek()
		if r == '#' {
			p.skipComment()
			continue
		}
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			p.pos++
			continue
		}
		return
	}
}

func (p *valueParser) skipComment() {
	for p.more() {
		r := p.next()
		if r == '\n' {
			return
		}
	}
}

func (p *valueParser) parseValue() (any, error) {
	p.skipSpace()
	if !p.more() {
		return nil, fmt.Errorf("empty value")
	}
	switch p.peek() {
	case '"', '\'':
		return p.parseString()
	case '[':
		return p.parseArray()
	case '{':
		return p.parseInlineTable()
	}
	return p.parseLiteral()
}

func (p *valueParser) parseString() (any, error) {
	if hasTriple(p.data, p.pos, p.peek()) {
		return p.parseMultiLineString()
	}
	start := p.pos
	quote := p.next()
	var b strings.Builder
	for p.more() {
		r := p.next()
		if r == quote {
			return b.String(), nil
		}
		if quote == '"' && r == '\\' {
			if !p.more() {
				return nil, fmt.Errorf("unterminated string")
			}
			esc := p.next()
			switch esc {
			case 'b':
				b.WriteByte('\b')
			case 't':
				b.WriteByte('\t')
			case 'n':
				b.WriteByte('\n')
			case 'f':
				b.WriteByte('\f')
			case 'r':
				b.WriteByte('\r')
			case '\\', '"':
				b.WriteRune(esc)
			case 'u', 'U':
				length := 4
				if esc == 'U' {
					length = 8
				}
				code, err := p.readUnicode(length)
				if err != nil {
					return nil, err
				}
				b.WriteRune(code)
			default:
				return nil, fmt.Errorf("invalid escape")
			}
			continue
		}
		b.WriteRune(r)
	}
	_ = start
	return nil, fmt.Errorf("unterminated string")
}

func (p *valueParser) parseMultiLineString() (any, error) {
	quote := p.next()
	p.next()
	p.next()
	var b strings.Builder
	for p.more() {
		r := p.next()
		if r == quote && hasTriple(p.data, p.pos-1, quote) {
			p.pos += 2
			return b.String(), nil
		}
		if quote == '"' && r == '\\' {
			if !p.more() {
				return nil, fmt.Errorf("unterminated string")
			}
			esc := p.next()
			switch esc {
			case 'b':
				b.WriteByte('\b')
			case 't':
				b.WriteByte('\t')
			case 'n':
				b.WriteByte('\n')
			case 'f':
				b.WriteByte('\f')
			case 'r':
				b.WriteByte('\r')
			case '\\', '"':
				b.WriteRune(esc)
			case '\n':
				continue
			case 'u', 'U':
				length := 4
				if esc == 'U' {
					length = 8
				}
				code, err := p.readUnicode(length)
				if err != nil {
					return nil, err
				}
				b.WriteRune(code)
			default:
				return nil, fmt.Errorf("invalid escape")
			}
			continue
		}
		b.WriteRune(r)
	}
	return nil, fmt.Errorf("unterminated string")
}

func (p *valueParser) readUnicode(length int) (rune, error) {
	if p.pos+length > len(p.data) {
		return 0, fmt.Errorf("invalid unicode escape")
	}
	var hex strings.Builder
	for i := 0; i < length; i++ {
		hex.WriteRune(p.data[p.pos+i])
	}
	p.pos += length
	value, err := strconv.ParseInt(hex.String(), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid unicode escape")
	}
	return rune(value), nil
}

func (p *valueParser) parseArray() ([]any, error) {
	if p.next() != '[' {
		return nil, fmt.Errorf("array must start with [")
	}
	var items []any
	for {
		p.skipSpace()
		if !p.more() {
			return nil, fmt.Errorf("unterminated array")
		}
		if p.peek() == ']' {
			p.next()
			return items, nil
		}
		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		p.skipSpace()
		if !p.more() {
			return nil, fmt.Errorf("unterminated array")
		}
		r := p.peek()
		if r == ',' {
			p.next()
			continue
		}
		if r == ']' {
			p.next()
			return items, nil
		}
		return nil, fmt.Errorf("invalid array separator")
	}
}

func (p *valueParser) parseInlineTable() (map[string]any, error) {
	if p.next() != '{' {
		return nil, fmt.Errorf("inline table must start with {")
	}
	result := make(map[string]any)
	for {
		p.skipSpace()
		if !p.more() {
			return nil, fmt.Errorf("unterminated inline table")
		}
		if p.peek() == '}' {
			p.next()
			return result, nil
		}
		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.next() != '=' {
			return nil, fmt.Errorf("inline table missing '='")
		}
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if err := setInlineKey(result, key, value); err != nil {
			return nil, err
		}
		p.skipSpace()
		if !p.more() {
			return nil, fmt.Errorf("unterminated inline table")
		}
		r := p.peek()
		if r == ',' {
			p.next()
			continue
		}
		if r == '}' {
			p.next()
			return result, nil
		}
		return nil, fmt.Errorf("invalid inline table separator")
	}
}

func (p *valueParser) parseKey() (string, error) {
	p.skipSpace()
	if !p.more() {
		return "", fmt.Errorf("empty key")
	}
	if p.peek() == '"' || p.peek() == '\'' {
		value, err := p.parseString()
		if err != nil {
			return "", err
		}
		parsed, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("invalid key")
		}
		return parsed, nil
	}
	start := p.pos
	for p.more() {
		r := p.peek()
		if r == '=' || r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ',' || r == '}' {
			break
		}
		p.pos++
	}
	key := strings.TrimSpace(string(p.data[start:p.pos]))
	if key == "" {
		return "", fmt.Errorf("empty key")
	}
	return key, nil
}

func (p *valueParser) parseLiteral() (any, error) {
	start := p.pos
	for p.more() {
		r := p.peek()
		if r == '#' || r == ',' || r == ']' || r == '}' || r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			break
		}
		p.pos++
	}
	raw := strings.TrimSpace(string(p.data[start:p.pos]))
	if raw == "" {
		return nil, fmt.Errorf("empty value")
	}
	lower := strings.ToLower(raw)
	if lower == "true" || lower == "false" {
		return lower == "true", nil
	}
	if lower == "inf" || lower == "+inf" || lower == "-inf" || lower == "nan" || lower == "+nan" || lower == "-nan" {
		value, err := strconv.ParseFloat(strings.ReplaceAll(lower, "_", ""), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float")
		}
		return value, nil
	}
	if parsed, ok := parseDateTime(raw); ok {
		return parsed, nil
	}
	if strings.ContainsAny(raw, ".eE") {
		return parseFloat(raw)
	}
	return parseInteger(raw)
}

func parseDateTime(raw string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
		"15:04:05.999999999",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseFloat(raw string) (float64, error) {
	clean := strings.ReplaceAll(raw, "_", "")
	value, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float")
	}
	return value, nil
}

func parseInteger(raw string) (int64, error) {
	clean := strings.ReplaceAll(raw, "_", "")
	base := 10
	value := clean
	sign := ""
	if strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		sign = value[:1]
		value = value[1:]
	}
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		base = 16
		value = value[2:]
	} else if strings.HasPrefix(value, "0o") || strings.HasPrefix(value, "0O") {
		base = 8
		value = value[2:]
	} else if strings.HasPrefix(value, "0b") || strings.HasPrefix(value, "0B") {
		base = 2
		value = value[2:]
	}
	parsed, err := strconv.ParseInt(sign+value, base, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer")
	}
	return parsed, nil
}

func parseStringValue(raw string) (string, error) {
	parser := &valueParser{data: []rune(raw)}
	value, err := parser.parseValue()
	if err != nil {
		return "", err
	}
	parser.skipSpace()
	if parser.more() {
		return "", fmt.Errorf("invalid string")
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid string")
	}
	return str, nil
}

func setInlineKey(root map[string]any, key string, value any) error {
	parts := splitOnDots(key)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) == 1 {
		if _, exists := root[parts[0]]; exists {
			return fmt.Errorf("duplicate key %s", parts[0])
		}
		root[parts[0]] = value
		return nil
	}
	nested, err := getOrCreateTable(root, parts[:len(parts)-1])
	if err != nil {
		return err
	}
	last := parts[len(parts)-1]
	if _, exists := nested[last]; exists {
		return fmt.Errorf("duplicate key %s", last)
	}
	nested[last] = value
	return nil
}

func getOrCreateTable(root map[string]any, path []string) (map[string]any, error) {
	current := root
	for _, part := range path {
		nextRaw, ok := current[part]
		if !ok {
			next := make(map[string]any)
			current[part] = next
			current = next
			continue
		}
		next, ok := nextRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("table conflict at %s", part)
		}
		current = next
	}
	return current, nil
}

func getOrCreateArrayTable(root map[string]any, path []string) ([]any, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty array table path")
	}
	parentPath := path[:len(path)-1]
	key := path[len(path)-1]
	parent, err := getOrCreateTable(root, parentPath)
	if err != nil {
		return nil, err
	}
	if existing, ok := parent[key]; ok {
		arr, ok := existing.([]any)
		if !ok {
			return nil, fmt.Errorf("array table conflict at %s", key)
		}
		return arr, nil
	}
	arr := []any{}
	parent[key] = arr
	return arr, nil
}

func setKey(current map[string]any, key string, value any) error {
	parts, err := parseKeyPath(key)
	if err != nil {
		return err
	}
	if len(parts) == 1 {
		if _, exists := current[parts[0]]; exists {
			return fmt.Errorf("duplicate key %s", parts[0])
		}
		current[parts[0]] = value
		return nil
	}
	nested, err := getOrCreateTable(current, parts[:len(parts)-1])
	if err != nil {
		return err
	}
	last := parts[len(parts)-1]
	if _, exists := nested[last]; exists {
		return fmt.Errorf("duplicate key %s", last)
	}
	nested[last] = value
	return nil
}

func setPath(root map[string]any, path []string, value any) error {
	if len(path) == 0 {
		return fmt.Errorf("empty path")
	}
	current := root
	for _, part := range path[:len(path)-1] {
		nextRaw, ok := current[part]
		if !ok {
			next := make(map[string]any)
			current[part] = next
			current = next
			continue
		}
		next, ok := nextRaw.(map[string]any)
		if !ok {
			return fmt.Errorf("path conflict at %s", part)
		}
		current = next
	}
	current[path[len(path)-1]] = value
	return nil
}
