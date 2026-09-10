package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// 矩陣（`VERIFICATION-MATRIX.md`）是完成度的唯一數字來源，九個章節、
// 五種欄位組合。它跟 worklist 是同一批資料的兩個切面：
//
//	worklist ＝ status 是 open 的那些
//	matrix   ＝ 全部，按章節分組
//
// 所以「改一件事只改一處」才成立——以前改一條規則要同時改 CONTEXT 的
// worklist、矩陣那一列與 README，三處各自會過期。

// Section 是矩陣的一個章節。`Sep` 存分隔線原文，因為對齊標記（`|---:|`）
// 是版面的一部分，round-trip 要還原得回去。
type Section struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Intro   []string `json:"intro,omitempty"`
	Columns []string `json:"columns"`
	Sep     string   `json:"sep"`

	// Outro 是表格**後面**的文字（矩陣有幾節在表後掛了註腳）。
	// 少了它，搬家會安靜地吃掉那幾行。
	Outro []string `json:"outro,omitempty"`
}

// Matrix 是矩陣的骨架：前言 ＋ 章節定義。列本身掛在 `Doc.Items` 上。
type Matrix struct {
	Head     []string  `json:"head"`
	Sections []Section `json:"sections"`
}

// sentinel 讓拆欄時避開跳脫的 `\|`——備註欄裡有人寫絕對值 `\|角差\|`，
// 那是字面字元不是分隔符。
const sentinel = "\x00PIPE\x00"

var sepCell = regexp.MustCompile(`^:?-{2,}:?$`)

// splitRow 把一列 markdown 表格拆成欄。
func splitRow(line string) []string {
	body := strings.TrimSpace(strings.ReplaceAll(line, `\|`, sentinel))
	body = strings.TrimPrefix(body, "|")
	body = strings.TrimSuffix(body, "|")
	cells := strings.Split(body, "|")
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = strings.ReplaceAll(strings.TrimSpace(c), sentinel, `\|`)
	}
	return out
}

func isSepRow(cells []string) bool {
	for _, c := range cells {
		if c != "" && !sepCell.MatchString(c) {
			return false
		}
	}
	return true
}

// ParseMatrix 讀 markdown，回骨架與逐列的欄位內容。
//
// 回傳的 rows 與 sections 同索引：`rows[i]` 是 `sections[i]` 的列。
func ParseMatrix(path string) (*Matrix, [][][]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	m := &Matrix{}
	var rows [][][]string
	var pending []string
	cur := -1
	for _, line := range strings.Split(string(b), "\n") {
		switch {
		case strings.HasPrefix(line, "## "):
			m.Sections = append(m.Sections, Section{Title: strings.TrimSpace(line[3:])})
			rows = append(rows, nil)
			cur++
			pending = nil
		case cur < 0:
			m.Head = append(m.Head, line)
		case strings.HasPrefix(line, "|"):
			cells := splitRow(line)
			s := &m.Sections[cur]
			if s.Columns == nil {
				s.Columns = cells
				s.Intro = append(s.Intro, pending...) // 原樣，含空行
				pending = nil
				continue
			}
			if s.Sep == "" && isSepRow(cells) {
				s.Sep = line
				continue
			}
			rows[cur] = append(rows[cur], cells)
		default:
			pending = append(pending, line)
		}
		// 表格已經開始之後再出現的非表格行，是這一節的註腳。
		if cur >= 0 && m.Sections[cur].Sep != "" && !strings.HasPrefix(line, "|") &&
			!strings.HasPrefix(line, "## ") {
			m.Sections[cur].Outro = append(m.Sections[cur].Outro, line)
			pending = nil
		}
	}
	// 章節 id：用序號加標題，穩定又不必猜英文名。
	for i := range m.Sections {
		m.Sections[i].ID = fmt.Sprintf("s%d", i+1)
	}
	return m, rows, nil
}

// RenderMatrix 把骨架與條目畫回 markdown。
func RenderMatrix(doc *Doc) (string, error) {
	if doc.Matrix == nil {
		return "", fmt.Errorf("這份 worklist.json 沒有 matrix 骨架")
	}
	bySection := map[string][]*Item{}
	for i := range doc.Items {
		if s := doc.Items[i].Section; s != "" {
			bySection[s] = append(bySection[s], &doc.Items[i])
		}
	}
	var b strings.Builder
	for _, l := range doc.Matrix.Head {
		b.WriteString(l + "\n")
	}
	for _, s := range doc.Matrix.Sections {
		b.WriteString("## " + s.Title + "\n")
		for _, l := range s.Intro {
			b.WriteString(l + "\n")
		}
		b.WriteString("| " + strings.Join(s.Columns, " | ") + " |\n")
		b.WriteString(s.Sep + "\n")
		for _, it := range bySection[s.ID] {
			cells := make([]string, len(s.Columns))
			for i, col := range s.Columns {
				cells[i] = it.Cells[col]
			}
			b.WriteString("| " + strings.Join(cells, " | ") + " |\n")
		}
		for _, l := range s.Outro {
			b.WriteString(l + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n", nil
}
