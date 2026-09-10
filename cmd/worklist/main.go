// worklist 是未完成項與完成度的核實工具（`~/.claude/rulebook/61`）。
//
// **權威是 `docs/worklist.json`，markdown 由 `render` 產生。**
// markdown 的 `- [ ]` 清單會長出過期斷言——東西做好了，而沒有人回頭改
// 那一條。症狀不是報錯，是清單上留著一句自信的「還沒接」，然後有人照它
// 去重做一遍、或拿它當「還剩多少」的依據去回報進度。
//
// 每一條掛一個 `verify`，跑起來回答**「這一條的斷言還成立嗎」**：
//
//	status ＝ open  斷言是「還沒做」——verify 說做好了就是過期斷言
//	status ＝ done  斷言是「做好了」——verify 說沒做好就是回歸
//
// 兩種方向由 `signalMeansDone` 統一，報表只列**不一致**的那幾條。
//
//	tools/go.sh run ./cmd/worklist verify          # 核實全部
//	tools/go.sh run ./cmd/worklist verify 18q      # 只核實某幾條
//	tools/go.sh run ./cmd/worklist render          # 產生 markdown 到 stdout
//	tools/go.sh run ./cmd/worklist check           # 只檢查 schema 與交叉引用
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// dataPath 是權威檔。工具跑在 repo 根目錄（`tools/go.sh` 把它掛成 /src）。
const dataPath = "docs/worklist.json"

// renderTarget 是那一節的家。`render -w` 只動標記之間，其餘一個字不改。
const renderTarget = "CONTEXT.md"

// schemaID 跟著結構走：欄位語意改了就升版，舊檔會被 check 擋下來。
const schemaID = "logh3-worklist/1"

// Status 是一條斷言的方向。
const (
	StatusOpen = "open" // 斷言：這件事還沒做
	StatusDone = "done" // 斷言：這件事做好了
)

// Doc 是整份 worklist。
type Doc struct {
	Schema string            `json:"schema"`
	Note   string            `json:"note"`
	Layers map[string]string `json:"layers"`
	Items  []Item            `json:"items"`
}

// Item 是一條斷言。
//
// `Body` 是逐行的敘述，render 出來就是 markdown 那一段——**原文照搬**，
// 因為那些摘要是跨多份 `docs/re/` 的綜合，丟了就要重寫。
type Item struct {
	ID     string `json:"id"`
	Legacy string `json:"legacy_id,omitempty"` // CONTEXT.md 原本的編號，docs/re 有交叉引用
	Status string `json:"status"`
	Layer  string `json:"layer"`
	Title  string `json:"title"`

	Body     []string `json:"body,omitempty"`
	Evidence []string `json:"evidence,omitempty"`
	Accept   string   `json:"acceptance,omitempty"` // 怎樣算做完
	Blocked  string   `json:"blocked_by,omitempty"` // 卡在什麼

	Verify Verify `json:"verify"`
}

// Verify 是那個會自己開口的訊號。欄位按 Kind 取用，其餘留空。
type Verify struct {
	Kind string `json:"kind"`

	// present／absent：在 Paths 底下找 Pattern。
	Paths   []string `json:"paths,omitempty"`
	Pattern string   `json:"pattern,omitempty"`

	// go_test：某個套件裡有沒有這支測試函式。
	Pkg string `json:"pkg,omitempty"`
	Run string `json:"run,omitempty"`

	// spec_status：某份規格有沒有標到某個狀態。
	Spec  string `json:"spec,omitempty"`
	State string `json:"state,omitempty"`

	// receipt：某批收據的檔案與 SHA-256。
	Batch string            `json:"batch,omitempty"`
	Files map[string]string `json:"files,omitempty"`

	// json_len：某份 JSON 的欄位長度是不是還沒超過 Max。
	Path  string `json:"path,omitempty"`
	Field string `json:"field,omitempty"`
	Max   int    `json:"max,omitempty"`

	Note string `json:"note,omitempty"`
}

// Result 是核實的結論。
type Result int

const (
	Holds   Result = iota // 斷言成立，沒事
	Broken                // 斷言不成立 ← 要人看的那些
	Unknown               // 拿不到訊號：manual，或收據不在這台機器上
)

func (r Result) String() string {
	switch r {
	case Holds:
		return "一致"
	case Broken:
		return "⚠ 斷言不成立"
	default:
		return "？ 無法核實"
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "用法: worklist {verify [id...]|render [-w]|check}")
		os.Exit(2)
	}
	doc, err := Load(dataPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch args[0] {
	case "verify":
		os.Exit(runVerify(doc, args[1:]))
	case "render":
		out, err := Render(doc)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if len(args) > 1 && args[1] == "-w" {
			if err := WriteBack(renderTarget, out); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("寫回 %s\n", renderTarget)
			return
		}
		fmt.Print(out)
	case "check":
		if errs := Check(doc); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, " ", e)
			}
			fmt.Fprintf(os.Stderr, "%d 個問題\n", len(errs))
			os.Exit(1)
		}
		fmt.Printf("%d 條，schema 與交叉引用都對得上\n", len(doc.Items))
	default:
		fmt.Fprintf(os.Stderr, "不認得的子命令 %q\n", args[0])
		os.Exit(2)
	}
}

// Load 讀權威檔。
func Load(path string) (*Doc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀不到 %s：%w", path, err)
	}
	var d Doc
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields() // 打錯欄位名要當場知道，不要安靜地被忽略
	if err := dec.Decode(&d); err != nil {
		return nil, fmt.Errorf("%s 解不開：%w", path, err)
	}
	return &d, nil
}

// runVerify 核實指定的（或全部的）條目，回 exit code。
func runVerify(doc *Doc, only []string) int {
	want := map[string]bool{}
	for _, id := range only {
		want[id] = true
	}
	var broken, unknown int
	var lines []string
	for _, it := range doc.Items {
		if len(want) > 0 && !want[it.ID] && !want[it.Legacy] {
			continue
		}
		res, why := VerifyItem(&it, ".")
		switch res {
		case Broken:
			broken++
		case Unknown:
			unknown++
		}
		lines = append(lines, fmt.Sprintf("%-32s %-6s %-14s %s",
			it.ID, it.Status, res, why))
	}
	sort.Strings(lines)
	for _, l := range lines {
		fmt.Println(l)
	}
	fmt.Printf("\n%d 條核實：%d 條斷言不成立、%d 條無法核實\n",
		len(lines), broken, unknown)
	if broken > 0 {
		fmt.Println("\n⚠ 「斷言不成立」的意思是**清單與現況對不上**：")
		fmt.Println("   status=open 卻驗到做好了 → 過期斷言，把它改成 done；")
		fmt.Println("   status=done 卻驗不到     → 回歸，或者那個訊號被改名了。")
		return 1
	}
	return 0
}

// Check 驗 schema 與交叉引用：id 不重複、status／layer 認得、evidence 指到
// 的檔案真的在、verify 的必填欄位有填。
func Check(doc *Doc) []string {
	var errs []string
	if doc.Schema != schemaID {
		errs = append(errs, fmt.Sprintf("schema 是 %q，這支工具吃的是 %q",
			doc.Schema, schemaID))
	}
	seen := map[string]bool{}
	for _, it := range doc.Items {
		if it.ID == "" {
			errs = append(errs, "有一條沒有 id")
			continue
		}
		if seen[it.ID] {
			errs = append(errs, fmt.Sprintf("%s：id 重複", it.ID))
		}
		seen[it.ID] = true
		if it.Status != StatusOpen && it.Status != StatusDone {
			errs = append(errs, fmt.Sprintf("%s：status ＝ %q，只能是 open 或 done",
				it.ID, it.Status))
		}
		if it.Layer != "" && doc.Layers != nil {
			if _, ok := doc.Layers[it.Layer]; !ok {
				errs = append(errs, fmt.Sprintf("%s：layer %q 不在 layers 裡",
					it.ID, it.Layer))
			}
		}
		if it.Title == "" {
			errs = append(errs, fmt.Sprintf("%s：沒有 title", it.ID))
		}
		for _, e := range it.Evidence {
			// 允許帶行號或錨點的寫法，取 # 之前那一段。
			p := e
			if i := strings.IndexByte(p, '#'); i >= 0 {
				p = p[:i]
			}
			if _, err := os.Stat(filepath.Clean(p)); err != nil {
				errs = append(errs, fmt.Sprintf("%s：evidence %q 找不到", it.ID, e))
			}
		}
		errs = append(errs, checkVerify(&it)...)
	}
	return errs
}
