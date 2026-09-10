package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ImportMatrix 是**一次性**的搬家：把 `VERIFICATION-MATRIX.md` 的每一列
// 變成一條 Item。跑過一次之後矩陣就由 `render-matrix` 產生，這支不再用。
//
// 自動綁 verify 的順序：備註裡有 `TestXxx` 就綁那支測試；有收據批次名就
// 綁收據；都沒有就 `manual`，note 寫明「完成度由這一列自己的證據欄撐著」。
func ImportMatrix(doc *Doc, out string) error {
	m, rows, err := ParseMatrix(matrixTarget)
	if err != nil {
		return err
	}
	doc.Matrix = m

	have := map[string]bool{}
	for _, it := range doc.Items {
		have[it.ID] = true
	}
	var added int
	for si, sec := range m.Sections {
		for ri, cells := range rows[si] {
			if len(cells) != len(sec.Columns) {
				return fmt.Errorf("%s 第 %d 列有 %d 欄，表頭是 %d 欄：%q",
					sec.Title, ri+1, len(cells), len(sec.Columns), cells[0])
			}
			cm := map[string]string{}
			for i, col := range sec.Columns {
				cm[col] = cells[i]
			}
			id := matrixID(sec, cells, ri)
			for have[id] {
				id += "'"
			}
			have[id] = true
			it := Item{
				ID: id, Status: matrixStatus(cells), Layer: "matrix",
				Title: plainTitle(cells[titleCol(sec)]),
				Section: sec.ID, Cells: cm,
				Verify: matrixVerify(strings.Join(cells, " ")),
			}
			doc.Items = append(doc.Items, it)
			added++
		}
	}
	if doc.Layers["matrix"] == "" {
		doc.Layers["matrix"] = "驗證矩陣的一列（完成度）"
	}
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("搬進 %d 列，現在 %d 條\n", added, len(doc.Items))
	return nil
}

// titleCol 回哪一欄拿來當標題：有 `#` 欄的用第二欄，否則用第一欄。
func titleCol(s Section) int {
	if len(s.Columns) > 1 && s.Columns[0] == "#" {
		return 1
	}
	return 0
}

// matrixID 優先用矩陣自己的識別碼（`4y`、`5q`——`docs/` 裡有「矩陣 4m」
// 這種引用），沒有就用章節序號。
func matrixID(s Section, cells []string, ri int) string {
	if len(s.Columns) > 0 && s.Columns[0] == "#" && cells[0] != "" {
		return "m" + cells[0]
	}
	return fmt.Sprintf("%s-%02d", s.ID, ri+1)
}

var mdNoise = regexp.MustCompile(`\*\*|\[|\]\([^)]*\)|` + "`")

// plainTitle 把 markdown 的粗體、連結與 code span 去掉當標題。
func plainTitle(s string) string {
	// 不要在括號處切——標題裡的「移動（艦隊）」會被切成「戰場的「索敵」「移動」。
	return strings.TrimSpace(mdNoise.ReplaceAllString(s, ""))
}

// matrixStatus：等級或狀態欄是 `—` 的就是還沒做。
func matrixStatus(cells []string) string {
	for _, c := range cells {
		if strings.TrimSpace(c) == "—" {
			return StatusOpen
		}
	}
	if strings.Contains(strings.Join(cells, " "), "**未做**") {
		return StatusOpen
	}
	return StatusDone
}

var (
	reTestName = regexp.MustCompile("`(Test[A-Za-z0-9_]+)`")
	reReceipt  = regexp.MustCompile("收據 `([A-Za-z0-9_.-]+)`")
)

func matrixVerify(row string) Verify {
	if mm := reTestName.FindStringSubmatch(row); mm != nil {
		// 測試釘著就是最強的回歸偵測；套件掃整個 internal 與 cmd。
		return Verify{Kind: "go_test", Pkg: "internal", Run: mm[1]}
	}
	if mm := reReceipt.FindStringSubmatch(row); mm != nil {
		return Verify{Kind: "manual",
			Note: fmt.Sprintf("完成度由收據 `%s` 撐著，還沒把雜湊釘進來；"+
				"要自動回歸偵測就改成 receipt 並補 files 的 SHA-256", mm[1])}
	}
	return Verify{Kind: "manual",
		Note: "這一列的完成度由它自己的證據欄撐著，沒有可綁的自動訊號"}
}
