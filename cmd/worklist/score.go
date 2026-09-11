package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// 自評分數。**唯一的輸入是 `docs/worklist.json`**，所以它不可能與驗證矩陣
// 講不同的話——兩份都是同一批資料的切面。
//
// ⚠ **不要手寫一個總分放進 README。** 手寫的數字沒有辦法跟著資料變：
// 條目一改它就過期，而過期的完成度**看起來完全正常**，只有在別人照它
// 排工作時才會發現。要引用就引用這支印出來的。
//
// ⚠ **章節之間不能直接比。** 每一列的份量差很多——`s2`（畫面）有 68 列、
// `s4`（介面覆蓋）只有 2 列，一列「主選單」與一列「戰鬥畫面」在表上
// 一樣重。所以下面**先列各章節**，總計只當「條目通過的比例」讀，
// 不是「遊戲完成度」。

// sectionScore 是一個矩陣章節的完成度。
type sectionScore struct {
	ID    string
	Title string
	Done  int
	Total int
}

func (s sectionScore) pct() int {
	if s.Total == 0 {
		return 0
	}
	return s.Done * 100 / s.Total
}

// layerCount 是非矩陣層剩下幾條——那才是「還差多少」的答案。
type layerCount struct {
	Layer string
	Desc  string
	Open  int
}

// Score 從 doc 算出自評，回傳可直接印的報表。
func Score(doc *Doc) string {
	titles := map[string]string{}
	if doc.Matrix != nil {
		for _, s := range doc.Matrix.Sections {
			titles[s.ID] = s.Title
		}
	}

	secs := map[string]*sectionScore{}
	var matrixDone, matrixTotal int
	openByLayer := map[string]int{}
	var openMatrix []Item

	for _, it := range doc.Items {
		if it.Layer == "matrix" {
			s := secs[it.Section]
			if s == nil {
				s = &sectionScore{ID: it.Section, Title: titles[it.Section]}
				secs[it.Section] = s
			}
			s.Total++
			matrixTotal++
			if it.Status == "done" {
				s.Done++
				matrixDone++
			} else {
				openMatrix = append(openMatrix, it)
			}
			continue
		}
		if it.Status != "done" {
			openByLayer[it.Layer]++
		}
	}

	var ids []string
	for id := range secs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	fmt.Fprintf(&b, "自評（算自 docs/worklist.json，%d 條）\n\n", len(doc.Items))
	fmt.Fprintf(&b, "%-4s %-14s %6s %s\n", "章節", "", "完成", "")
	for _, id := range ids {
		s := secs[id]
		fmt.Fprintf(&b, "%-4s %-14s %3d%%  %d/%d %s\n",
			s.ID, s.Title, s.pct(), s.Done, s.Total, bar(s.pct()))
	}
	fmt.Fprintf(&b, "\n矩陣合計 %d/%d 條通過（%d%%）——**只是條目比例，不是遊戲完成度**：\n",
		matrixDone, matrixTotal, matrixDone*100/max1(matrixTotal))
	fmt.Fprintf(&b, "每一列的份量差很多（%s）。\n", weightNote(ids, secs))

	fmt.Fprint(&b, receiptHealth())
	fmt.Fprintf(&b, "\n還差多少（矩陣以外的層，只數還沒做完的）\n")
	var layers []layerCount
	for l, n := range openByLayer {
		layers = append(layers, layerCount{l, doc.Layers[l], n})
	}
	sort.Slice(layers, func(i, j int) bool { return layers[i].Open > layers[j].Open })
	for _, l := range layers {
		fmt.Fprintf(&b, "  %-10s %2d 條  %s\n", l.Layer, l.Open, l.Desc)
	}
	if len(openMatrix) > 0 {
		fmt.Fprintf(&b, "\n矩陣裡還沒通過的 %d 列\n", len(openMatrix))
		for _, it := range openMatrix {
			fmt.Fprintf(&b, "  %-4s %s\n", it.Section, it.Title)
		}
	}
	return b.String()
}

func weightNote(ids []string, secs map[string]*sectionScore) string {
	var most, least *sectionScore
	for _, id := range ids {
		s := secs[id]
		if most == nil || s.Total > most.Total {
			most = s
		}
		if least == nil || s.Total < least.Total {
			least = s
		}
	}
	if most == nil || least == nil {
		return "章節列數不一"
	}
	return fmt.Sprintf("%s 有 %d 列、%s 只有 %d 列", most.ID, most.Total, least.ID, least.Total)
}

func bar(pct int) string {
	n := pct / 10
	return strings.Repeat("█", n) + strings.Repeat("·", 10-n)
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// receiptHealth 是「撐著上面那些通過的收據，現在還有幾份重現得出來」。
//
// ⚠ **這一行不能省。** 矩陣的每一個「通過」背後都是一份收據，而收據綁
// 執行器版本（AGENTS.md）——換了執行器之後重現不出來的那些，它們撐著的
// 結論就只剩文件上的斷言。少了這一行，上面的百分比會顯得比實際可靠。
//
// 資料是 `tools/check_receipts.sh` 產的，落在 gitignore 的 workplace，
// 所以**沒有它是正常的**（別人 clone 之後本來就沒有收據）——那時說不知道，
// 不要假裝健康。
func receiptHealth() string {
	const path = "workplace/receipt-check.json"
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("\n對拍證據：不知道（沒有 %s，跑 tools/check_receipts.sh 產生）\n", path)
	}
	var rc struct {
		GolemHead string `json:"golem_head"`
		Generated string `json:"generated"`
		Results   []struct {
			ScreenDiff int `json:"screen_diff"`
			PalDiff    int `json:"pal_diff"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &rc); err != nil || len(rc.Results) == 0 {
		return fmt.Sprintf("\n對拍證據：%s 讀不動\n", path)
	}
	ok := 0
	for _, r := range rc.Results {
		if r.ScreenDiff == 0 && r.PalDiff == 0 {
			ok++
		}
	}
	return fmt.Sprintf("\n對拍證據：%d/%d 份收據在執行器 %s 上逐位元組重現（%d%%，量於 %s）\n"+
		"  對不上的 %d 份已逐份分形狀（docs/re/276）：一部分只要重錄，另一部分是\n"+
		"  輸入沒變但時序變了、同樣的步數落在遊戲的另一個時刻，那些要重挑步數。\n"+
		"  在處理完之前，上面的「通過」有這 %d 份的收據還沒重新驗過。\n",
		ok, len(rc.Results), rc.GolemHead, ok*100/len(rc.Results), rc.Generated[:10],
		len(rc.Results)-ok, len(rc.Results)-ok)
}
