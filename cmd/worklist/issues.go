package main

// GitHub issue 是**未完成項的權威**（使用者定案 2026-09-11）。
//
// 分工：
//
//	GitHub issue（open）   還剩什麼——新增、修改、關閉都在那裡做
//	GitHub issue（closed） 做完了——關掉 ＝ status done
//	docs/worklist.json      搬移之前就做完的那一批（歷史 ＋ 回歸偵測）
//
// **`verify` 沒有跟著消失。** 每張 issue 的本文尾巴有一個 JSON 區塊，
// 放 verify 與矩陣欄位；`tools/issues_snapshot.sh` 在主機用 `gh` 把
// issue 抓成 `workplace/issues.json`，這支工具在容器裡讀它、合併、核實。
// 容器沒有網路，所以抓取一定在主機、核實一定在容器，中間靠那份快照。
//
// ⚠ **沒有快照時要說出來。** 快照不在的時候，未完成項整批不在清單上，
// 而 `verify` 對剩下的歷史照樣回「0 條斷言不成立」——那句話在這種情況下
// 是假的。所以 `Load` 找不到快照會把這件事記在 `doc.IssuesNote`，
// 每個子命令都印出來。
//
// ⚠ **本文的格式是資料，不是排版。** 「### 判準」「### 卡在」「### 證據」
// 三個標題與尾巴那個 JSON 區塊是解析的錨點；改了標題文字，那一段就會被
// 當成敘述的一部分讀回來。在 GitHub 上編輯時只改內容，不改標題。

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// issuesPath 是主機抓下來的快照（gitignore）。
const issuesPath = "workplace/issues.json"

// stubTitlePrefix 是佔位條目的標題開頭。**佔位不抄真正的標題**——抄了就是
// 第二份會過期的內容；沒有快照時看到「見 GitHub #N」反而是對的提示。
const stubTitlePrefix = "（見 GitHub #"

// Stub 回一條的佔位：只留 id、層、章節、issue 編號。
func Stub(it *Item, number int) Item {
	return Item{
		ID: it.ID, Legacy: it.Legacy, Status: it.Status, Layer: it.Layer, Section: it.Section,
		Title: fmt.Sprintf("%s%d）", stubTitlePrefix, number),
		Issue: number,
		Verify: Verify{Kind: "manual",
			Note: fmt.Sprintf("權威在 GitHub issue #%d；先跑 tools/issues_snapshot.sh 抓快照才核實得到", number)},
	}
}

const (
	issueMarkerPrefix = "<!-- worklist:id="
	issueMarkerSuffix = " -->"
	headAccept        = "### 判準"
	headBlocked       = "### 卡在"
	headEvidence      = "### 證據"
	machineOpen       = "<details><summary>機器核對（verify 與矩陣欄位；改這一段就是改核對方式）</summary>"
	machineClose      = "</details>"
	labelAll          = "worklist"
	labelLayerPrefix  = "layer:"
	labelSecPrefix    = "section:"
)

// IssueRecord 是 `gh issue list --json number,title,body,state,labels` 的一筆。
type IssueRecord struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"` // OPEN／CLOSED
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// machine 是本文尾巴那個 JSON 區塊：人不太會去改、改了就是改語意的欄位。
type machine struct {
	ID      string            `json:"id"`
	Legacy  string            `json:"legacy_id,omitempty"`
	Layer   string            `json:"layer"`
	Section string            `json:"section,omitempty"`
	Cells   map[string]string `json:"cells,omitempty"`
	Ref     string            `json:"ref,omitempty"`
	Verify  Verify            `json:"verify"`
}

// IssueLabels 回這一條要掛的 label。
func IssueLabels(it *Item) []string {
	ls := []string{labelAll, labelLayerPrefix + it.Layer}
	if it.Section != "" {
		ls = append(ls, labelSecPrefix+it.Section)
	}
	return ls
}

// IssueBody 把一條畫成 issue 本文。ParseIssue 要能原樣讀回來
// （`TestIssueRoundTrip` 對現有每一條驗）。
func IssueBody(it *Item) (string, error) {
	var b strings.Builder
	b.WriteString(issueMarkerPrefix + it.ID + issueMarkerSuffix + "\n\n")
	if len(it.Body) > 0 {
		b.WriteString(strings.Join(it.Body, "\n"))
		b.WriteString("\n\n")
	}
	if it.Accept != "" {
		b.WriteString(headAccept + "\n\n" + it.Accept + "\n\n")
	}
	if it.Blocked != "" {
		b.WriteString(headBlocked + "\n\n" + it.Blocked + "\n\n")
	}
	if len(it.Evidence) > 0 {
		b.WriteString(headEvidence + "\n\n")
		for _, e := range it.Evidence {
			b.WriteString("- `" + e + "`\n")
		}
		b.WriteString("\n")
	}
	m := machine{ID: it.ID, Legacy: it.Legacy, Layer: it.Layer, Section: it.Section,
		Cells: it.Cells, Ref: it.Ref, Verify: it.Verify}
	js, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	b.WriteString(machineOpen + "\n\n```json\n" + string(js) + "\n```\n\n" + machineClose + "\n")
	return b.String(), nil
}

// ParseIssue 把一張 issue 讀回條目。
func ParseIssue(rec IssueRecord) (Item, error) {
	body := strings.ReplaceAll(rec.Body, "\r\n", "\n") // GitHub 網頁編輯會存成 CRLF
	var it Item

	// 機器區塊。
	i := strings.Index(body, machineOpen)
	if i < 0 {
		return it, fmt.Errorf("#%d 找不到機器核對區塊（本文尾巴那個 <details>）", rec.Number)
	}
	mpart := body[i+len(machineOpen):]
	a := strings.Index(mpart, "```json\n")
	z := strings.LastIndex(mpart, "\n```")
	if a < 0 || z < 0 || z < a {
		return it, fmt.Errorf("#%d 的機器核對區塊裡找不到 ```json 圍欄", rec.Number)
	}
	var m machine
	dec := json.NewDecoder(strings.NewReader(mpart[a+len("```json\n") : z]))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return it, fmt.Errorf("#%d 的機器核對區塊解不開：%w", rec.Number, err)
	}
	human := body[:i]

	// 識別碼：本文開頭的標記與 JSON 裡的 id 要一致——兩處不同代表有人
	// 複製了別張 issue 的本文，那時候哪一個都不能信。
	if !strings.HasPrefix(human, issueMarkerPrefix) {
		return it, fmt.Errorf("#%d 本文不是以 %sID%s 開頭", rec.Number, issueMarkerPrefix, issueMarkerSuffix)
	}
	end := strings.Index(human, issueMarkerSuffix)
	if id := human[len(issueMarkerPrefix):end]; id != m.ID {
		return it, fmt.Errorf("#%d 開頭標記的 id %q 與機器區塊的 %q 不同", rec.Number, id, m.ID)
	}
	human = human[end+len(issueMarkerSuffix):]

	// 人看的四段：敘述在最前面，後面三個標題各一段。
	sections := map[string]string{}
	cur, curText := "", []string{}
	flush := func() {
		sections[cur] = strings.Trim(strings.Join(curText, "\n"), "\n")
		curText = nil
	}
	for _, line := range strings.Split(human, "\n") {
		switch line {
		case headAccept, headBlocked, headEvidence:
			flush()
			cur = line
			continue
		}
		curText = append(curText, line)
	}
	flush()

	it.ID, it.Legacy, it.Layer = m.ID, m.Legacy, m.Layer
	it.Section, it.Cells, it.Ref, it.Verify = m.Section, m.Cells, m.Ref, m.Verify
	it.Title = rec.Title
	if s := sections[""]; s != "" {
		it.Body = strings.Split(s, "\n")
	}
	it.Accept = sections[headAccept]
	it.Blocked = sections[headBlocked]
	if s := sections[headEvidence]; s != "" {
		for _, line := range strings.Split(s, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			e := strings.TrimSuffix(strings.TrimPrefix(line, "- `"), "`")
			if e == line {
				return it, fmt.Errorf("#%d 的證據那一段有一行不是 \"- `路徑`\"：%q", rec.Number, line)
			}
			it.Evidence = append(it.Evidence, e)
		}
	}
	if strings.EqualFold(rec.State, "CLOSED") {
		it.Status = StatusDone
	} else {
		it.Status = StatusOpen
	}
	return it, nil
}

// MergeIssues 把快照裡的 issue 併進清單。**同一個 id 以 issue 為準**
// ——它是權威；JSON 裡那一筆只是搬移當下的副本。
func MergeIssues(doc *Doc, recs []IssueRecord) error {
	idx := map[string]int{}
	for i, it := range doc.Items {
		idx[it.ID] = i
	}
	var errs []string
	for _, rec := range recs {
		if !hasLabel(rec, labelAll) {
			continue // 不是 worklist 的 issue（一般的 bug 回報之類）
		}
		it, err := ParseIssue(rec)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		it.Issue = rec.Number
		if i, ok := idx[it.ID]; ok {
			doc.Items[i] = it // 原地替換：佔位的位置就是這一條在矩陣裡的順序
		} else {
			idx[it.ID] = len(doc.Items)
			doc.Items = append(doc.Items, it)
		}
	}
	if len(errs) > 0 {
		sort.Strings(errs)
		return fmt.Errorf("%d 張 issue 讀不回來：\n  %s", len(errs), strings.Join(errs, "\n  "))
	}
	return nil
}

func hasLabel(rec IssueRecord, name string) bool {
	for _, l := range rec.Labels {
		if l.Name == name {
			return true
		}
	}
	return false
}

// loadIssues 讀快照；沒有快照回 (nil, false)。
func loadIssues(path string) ([]IssueRecord, bool, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var recs []IssueRecord
	if err := json.Unmarshal(b, &recs); err != nil {
		return nil, true, fmt.Errorf("%s 解不開：%w", path, err)
	}
	return recs, true, nil
}

// exportIssue 是 `export-issues` 的一筆：搬移腳本照它開 issue。
type exportIssue struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Labels []string `json:"labels"`
}

// ExportOpen 把 JSON 裡**還沒做完**的條目畫成 issue 負載。
// 只在搬移時用一次；之後新增條目直接在 GitHub 開（照 issue 範本）。
func ExportOpen(doc *Doc) ([]exportIssue, error) {
	var out []exportIssue
	for i := range doc.Items {
		it := &doc.Items[i]
		if it.Status != StatusOpen || it.Issue > 0 {
			continue // 做完的不搬；已經是佔位的代表 issue 早就開了
		}
		body, err := IssueBody(it)
		if err != nil {
			return nil, err
		}
		out = append(out, exportIssue{ID: it.ID, Title: it.Title, Body: body, Labels: IssueLabels(it)})
	}
	return out, nil
}

// StubMigrated 把 JSON 裡**已經開成 issue 的條目**換成佔位，寫回 path。
//
// 只看 JSON 本身與快照裡的 id → 編號，**不做合併**——合併後的清單是
// issue 的內容，寫回去就等於把 GitHub 上的東西抄一份進 repo，又回到
// 兩份各自過期的老問題。
func StubMigrated(path string, recs []IssueRecord) (int, error) {
	doc, err := Load(path)
	if err != nil {
		return 0, err
	}
	num := map[string]int{}
	for _, rec := range recs {
		if !hasLabel(rec, labelAll) {
			continue
		}
		it, err := ParseIssue(rec)
		if err != nil {
			return 0, err
		}
		num[it.ID] = rec.Number
	}
	n := 0
	for i := range doc.Items {
		it := &doc.Items[i]
		if it.Issue > 0 {
			continue
		}
		if no, ok := num[it.ID]; ok {
			*it = Stub(it, no)
			n++
		}
	}
	if n == 0 {
		return 0, nil
	}
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false) // 內文有 `<`、`>`、`&`，預設會被寫成 \u003c
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return 0, err
	}
	return n, os.WriteFile(path, []byte(b.String()), 0o644)
}
