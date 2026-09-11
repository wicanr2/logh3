package main

import (
	"fmt"
	"os"
	"strings"
)

// 這兩個標記把 markdown 裡「由工具產生」的那一段夾起來。標記之外的內容
// `render -w` 一個字都不動。
const (
	markBegin = "<!-- worklist:begin 由 `tools/go.sh run ./cmd/worklist render` 產生，不要手改 -->"
	markEnd   = "<!-- worklist:end -->"
)

// issueURLBase 是 issue 連結的前綴。repo 改名時改這一行。
const issueURLBase = "https://github.com/wicanr2/logh3/issues"

// Render 把權威檔畫成 markdown。
//
// **未完成的排在前面**：worklist 要回答的是「還剩什麼」，不是「做過什麼」。
// 已完成的留著是為了「別重做一遍」與交叉引用（`docs/re/` 裡有「CONTEXT
// 第 21 條」這種寫法，所以原編號要印出來）。
func Render(doc *Doc) (string, error) {
	var open, done []Item
	for _, it := range doc.Items {
		switch {
		case it.Ref != "":
			// 由別的條目代表，不重複列。
		case it.Status == StatusOpen:
			open = append(open, it)
		case it.Section != "":
			// 已完成的矩陣列只出現在矩陣——它的等級、未核准差異與收據
			// 都在那張表上，在 worklist 再抄一份就是第二份會過期的東西。
		default:
			done = append(done, it)
		}
	}

	var b strings.Builder
	b.WriteString(markBegin + "\n\n")
	fmt.Fprintf(&b, "權威是 [GitHub issue（label `worklist`）](%s?q=is%%3Aopen+label%%3Aworklist)：**還剩 %d 條**。\n",
		issueURLBase, len(open))
	b.WriteString("搬移（2026-09-11）之前做完的在 [`docs/worklist.json`](docs/worklist.json)；\n")
	b.WriteString("這一節由 JSON ＋ issue 快照合併產生，產生前先跑 `tools/issues_snapshot.sh`。\n")
	b.WriteString("每一條掛一個 `verify`，`tools/go.sh run ./cmd/worklist verify` 會回答\n")
	b.WriteString("「這一條的斷言還成立嗎」——清單與現況對不上會當場紅掉。\n\n")
	if doc.Note != "" {
		fmt.Fprintf(&b, "> %s\n\n", doc.Note)
	}

	b.WriteString("### 還沒做的\n\n")
	if len(open) == 0 {
		b.WriteString("（空）\n\n")
	}
	for i := range open {
		renderItem(&b, doc, &open[i])
	}

	if len(done) > 0 {
		b.WriteString("### 做完的（不在矩陣裡的那幾條）\n\n")
		fmt.Fprintf(&b, "%d 條。完成度的全貌在 [`VERIFICATION-MATRIX.md`](VERIFICATION-MATRIX.md)"+
			"——已完成的矩陣列不在這裡重抄一份。更早的逆向歷程（舊 worklist 那 65 條的原文）在\n"+
			"[`docs/solved-history.md`](docs/solved-history.md)，那是快照、不再維護。\n\n", len(done))
	}
	for i := range done {
		renderItem(&b, doc, &done[i])
	}
	b.WriteString(markEnd + "\n")
	return b.String(), nil
}

func renderItem(b *strings.Builder, doc *Doc, it *Item) {
	fmt.Fprintf(b, "#### %s — %s\n\n", it.ID, it.Title)

	var tags []string
	if it.Issue > 0 {
		tags = append(tags, fmt.Sprintf("[#%d](%s/%d)", it.Issue, issueURLBase, it.Issue))
	}
	if it.Legacy != "" {
		tags = append(tags, "原編號 "+it.Legacy)
	}
	if it.Layer != "" {
		if desc, ok := doc.Layers[it.Layer]; ok {
			tags = append(tags, fmt.Sprintf("%s（%s）", it.Layer, desc))
		} else {
			tags = append(tags, it.Layer)
		}
	}
	if len(tags) > 0 {
		fmt.Fprintf(b, "*%s*\n\n", strings.Join(tags, " · "))
	}
	for _, line := range it.Body {
		b.WriteString(line + "\n")
	}
	if len(it.Body) > 0 {
		b.WriteString("\n")
	}
	if it.Blocked != "" {
		fmt.Fprintf(b, "**卡在**：%s\n\n", it.Blocked)
	}
	if it.Accept != "" {
		fmt.Fprintf(b, "**做完的判準**：%s\n\n", it.Accept)
	}
	if len(it.Evidence) > 0 {
		links := make([]string, 0, len(it.Evidence))
		for _, e := range it.Evidence {
			links = append(links, fmt.Sprintf("[`%s`](%s)", e, e))
		}
		fmt.Fprintf(b, "**證據**：%s\n\n", strings.Join(links, "、"))
	}
	fmt.Fprintf(b, "**核實**：%s\n\n", describeVerify(&it.Verify))
}

// describeVerify 用一行講清楚這一條靠什麼訊號開口。
func describeVerify(v *Verify) string {
	switch v.Kind {
	case "manual":
		return "要人判——" + v.Note
	case "present":
		return fmt.Sprintf("`%s` 裡還找得到 `%s`（自承還在 ＝ 還沒做）",
			strings.Join(v.Paths, "`、`"), v.Pattern)
	case "absent":
		return fmt.Sprintf("`%s` 裡出現 `%s` 就是做好了",
			strings.Join(v.Paths, "`、`"), v.Pattern)
	case "go_test":
		return fmt.Sprintf("`%s` 釘在 `%s`", v.Run, v.Pkg)
	case "spec_status":
		return fmt.Sprintf("`%s` 標到 `%s`", v.Spec, v.State)
	case "receipt":
		return fmt.Sprintf("收據 `%s` 的 %d 個檔案雜湊相符", v.Batch, len(v.Files))
	case "json_len":
		return fmt.Sprintf("`%s` 的 `%s` 超過 %d 項", v.Path, v.Field, v.Max)
	}
	return "（不認得的 verify）"
}

// WriteBack 把 render 出來的內容塞回 markdown 的標記之間。
// 找不到標記就把它接在檔尾——第一次接管那一節時會用到。
func WriteBack(path, body string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(b)
	i := strings.Index(s, markBegin)
	j := strings.Index(s, markEnd)
	if i < 0 || j < 0 || j < i {
		return fmt.Errorf("%s 裡找不到 worklist:begin／end 標記——"+
			"第一次接管請自己在該節放好標記，避免蓋掉手寫內容", path)
	}
	// ⚠ **要冪等**：`Render` 的結尾已經帶一個換行（`markEnd + "\n"`），
	// 舊檔 `markEnd` 後面也還留著上一次寫進去的那一個。兩個都留的話
	// 每跑一次就多一個空行——而那一行在 diff 裡幾乎看不見，於是十幾次
	// 之後 `## 已解的逆向歷程` 前面就堆了一整片空白（2026-09-11 量到）。
	rest := s[j+len(markEnd):]
	if strings.HasSuffix(body, "\n") && strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}
	out := s[:i] + body + rest
	return os.WriteFile(path, []byte(out), 0o644)
}
