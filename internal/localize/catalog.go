// Package localize translates display records without changing game data.
package localize

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

//go:embed ui.tsv font.json font-12.png font-36.png big5.json
var embedded embed.FS

var Locales = []string{"zh-TW", "zh-CN", "ja", "en", "ko"}

type Catalog struct {
	Locale string
	values map[string]string
	terms  map[string]string
	order  []string
}

func Valid(locale string) bool {
	for _, code := range Locales {
		if code == locale {
			return true
		}
	}
	return false
}

// Load uses the editable catalog when supplied, otherwise the same embedded TSV.
func Load(path, locale string) (*Catalog, error) {
	if !Valid(locale) {
		return nil, fmt.Errorf("invalid locale: %s", locale)
	}
	b, err := os.ReadFile(path)
	if path == "" || os.IsNotExist(err) {
		b, err = embedded.ReadFile("ui.tsv")
	}
	if err != nil {
		return nil, err
	}
	return Parse(string(b), locale)
}

func Parse(text, locale string) (*Catalog, error) {
	if !Valid(locale) {
		return nil, fmt.Errorf("invalid locale: %s", locale)
	}
	c := &Catalog{Locale: locale, values: map[string]string{}, terms: map[string]string{}}
	column := 0
	for i, code := range Locales {
		if code == locale {
			column = i + 1
		}
	}
	s := bufio.NewScanner(strings.NewReader(text))
	s.Buffer(make([]byte, 4096), 1024*1024)
	line := 0
	for s.Scan() {
		line++
		row := strings.Split(strings.TrimSuffix(s.Text(), "\r"), "\t")
		if len(row) != 6 {
			return nil, fmt.Errorf("catalog line %d: expected 6 columns", line)
		}
		if line == 1 {
			if strings.Join(row, "\t") != "key\tzh-TW\tzh-CN\tja\ten\tko" {
				return nil, fmt.Errorf("invalid catalog header")
			}
			continue
		}
		for _, field := range row {
			if strings.TrimSpace(field) == "" {
				return nil, fmt.Errorf("catalog line %d: empty translation", line)
			}
		}
		if _, exists := c.values[row[0]]; exists {
			return nil, fmt.Errorf("duplicate catalog key: %s", row[0])
		}
		c.values[row[0]] = row[column]
		// Legacy Big5 records are matched only after rendering. They are never
		// substituted into game identifiers, parsers, or save buffers.
		if strings.HasPrefix(row[0], "ui.") {
			c.terms[row[1]] = row[column]
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if line < 2 {
		return nil, fmt.Errorf("empty catalog")
	}
	for source := range c.terms {
		c.order = append(c.order, source)
	}
	sort.Slice(c.order, func(i, j int) bool {
		if len(c.order[i]) == len(c.order[j]) {
			return c.order[i] < c.order[j]
		}
		return len(c.order[i]) > len(c.order[j])
	})
	return c, nil
}

func (c *Catalog) Value(key string) string {
	if value, ok := c.values[key]; ok {
		return value
	}
	return key
}

// Translate matches full labels first, then longest source terms. Numeric and
// ASCII fields remain literal. Unknown characters remain visible for diagnosis.
func (c *Catalog) Translate(source string) string {
	if c == nil || c.Locale == "zh-TW" {
		return source
	}
	// Some original menu labels deliberately leave full cells between Han
	// characters. Those cells are layout, not part of the translation key.
	input := []rune(strings.TrimSpace(source))
	var joined []rune
	for i, ch := range input {
		if unicode.IsSpace(ch) && i > 0 && i+1 < len(input) && unicode.Is(unicode.Han, input[i-1]) && unicode.Is(unicode.Han, input[i+1]) {
			continue
		}
		joined = append(joined, ch)
	}
	source = string(joined)
	if v, ok := c.terms[source]; ok {
		return v
	}
	var out strings.Builder
	runes := []rune(source)
	for len(runes) > 0 {
		left := string(runes)
		found := false
		for _, term := range c.order {
			if strings.HasPrefix(left, term) {
				if c.Locale == "en" && out.Len() > 0 {
					last := out.String()[out.Len()-1]
					if last >= 'a' && last <= 'z' || last >= 'A' && last <= 'Z' {
						out.WriteByte(' ')
					}
				}
				out.WriteString(c.terms[term])
				runes = runes[len([]rune(term)):]
				found = true
				break
			}
		}
		if !found {
			out.WriteRune(runes[0])
			runes = runes[1:]
		}
	}
	result := out.String()
	if c.Locale == "en" {
		result = strings.NewReplacer("，", ", ", "。", ".", "：", ": ", "﹕", ": ", "？", "?", "！", "!", "／", "/", "％", "%", "　", " ").Replace(result)
		result = strings.Map(func(ch rune) rune {
			if ch >= 0xFF01 && ch <= 0xFF5E {
				return ch - 0xFEE0
			}
			return ch
		}, result)
	}
	return result
}
