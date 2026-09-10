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

// Render 把權威檔畫成 markdown。
//
// **未完成的排在前面**：worklist 要回答的是「還剩什麼」，不是「做過什麼」。
// 已完成的留著是為了「別重做一遍」與交叉引用（`docs/re/` 裡有「CONTEXT
// 第 21 條」這種寫法，所以原編號要印出來）。
func Render(doc *Doc) (string, error) {
	var open, done []Item
	for _, it := range doc.Items {
		if it.Status == StatusOpen {
			open = append(open, it)
		} else {
			done = append(done, it)
		}
	}

	var b strings.Builder
	b.WriteString(markBegin + "\n\n")
	fmt.Fprintf(&b, "權威是 [`docs/worklist.json`](docs/worklist.json)：**%d 條還沒做、%d 條做完了**。\n",
		len(open), len(done))
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

	b.WriteString("### 做完的\n\n")
	b.WriteString("留著是為了不要重做一遍，也為了 `docs/re/` 裡「CONTEXT 第 N 條」那種交叉引用。\n\n")
	for i := range done {
		renderItem(&b, doc, &done[i])
	}
	b.WriteString(markEnd + "\n")
	return b.String(), nil
}

func renderItem(b *strings.Builder, doc *Doc, it *Item) {
	fmt.Fprintf(b, "#### %s — %s\n\n", it.ID, it.Title)

	var tags []string
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
	out := s[:i] + body + s[j+len(markEnd):]
	return os.WriteFile(path, []byte(out), 0o644)
}
