package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// write 在暫存樹裡放一個檔案。
func write(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestPresentBothWays 是 `present` 的正反對照。
//
// **只驗「它說仍未完成」證明不了機制有在看**——空的檔案、寫錯的路徑、
// 永遠為真的 pattern 都會印出一樣的好消息。所以先確認訊號在時斷言成立，
// **再把訊號拿掉、確認它真的開口**。
func TestPresentBothWays(t *testing.T) {
	root := t.TempDir()
	it := &Item{ID: "x", Status: StatusOpen, Title: "t", Verify: Verify{
		Kind: "present", Paths: []string{"internal/game"}, Pattern: "remake 還沒有播放層",
	}}

	// 訊號在：自承註解還在 ＝ 還沒做 ＝ open 這個斷言成立。
	write(t, root, "internal/game/audio.go", "package game\n// remake 還沒有播放層\n")
	if got, why := VerifyItem(it, root); got != Holds {
		t.Errorf("訊號在的時候回 %v（%s），預期一致", got, why)
	}

	// 訊號沒了：東西做好了而條目沒改 → 必須開口。
	write(t, root, "internal/game/audio.go", "package game\n// 播放層在這裡\n")
	if got, why := VerifyItem(it, root); got != Broken {
		t.Errorf("訊號拿掉之後回 %v（%s），預期斷言不成立", got, why)
	}
}

// TestAbsentBothWays 是 `absent` 的正反對照：東西還沒出現 ＝ 還沒做。
func TestAbsentBothWays(t *testing.T) {
	root := t.TempDir()
	it := &Item{ID: "x", Status: StatusOpen, Title: "t", Verify: Verify{
		Kind: "absent", Paths: []string{"internal/rules"}, Pattern: `func PlayMusic\(`,
	}}
	write(t, root, "internal/rules/other.go", "package rules\nfunc Nothing() {}\n")
	if got, why := VerifyItem(it, root); got != Holds {
		t.Errorf("東西還沒出現時回 %v（%s），預期一致", got, why)
	}
	write(t, root, "internal/rules/music.go", "package rules\nfunc PlayMusic() {}\n")
	if got, why := VerifyItem(it, root); got != Broken {
		t.Errorf("東西出現之後回 %v（%s），預期斷言不成立", got, why)
	}
}

// TestVerifyIgnoresTestFiles 釘住那條 [HARD]：**verify 只看產品程式碼**。
//
// 測試本來就會提到還沒接上的東西（為了釘住將來的行為）。把 `*_test.go`
// 算進來，`absent` 會因為測試裡有一行呼叫就判成「已經做了」，真缺口就這樣
// 被蓋掉。
func TestVerifyIgnoresTestFiles(t *testing.T) {
	root := t.TempDir()
	it := &Item{ID: "x", Status: StatusOpen, Title: "t", Verify: Verify{
		Kind: "absent", Paths: []string{"internal/rules"}, Pattern: `func PlayMusic\(`,
	}}
	// 只有測試檔提到它——產品程式碼還沒有，所以斷言仍然成立。
	write(t, root, "internal/rules/music_test.go",
		"package rules\nfunc TestX(t *testing.T) { PlayMusic() }\nfunc PlayMusic() {}\n")
	if got, why := VerifyItem(it, root); got != Holds {
		t.Errorf("只有測試檔提到時回 %v（%s），預期一致（測試檔不算數）", got, why)
	}
	// 產品程式碼出現了才算。
	write(t, root, "internal/rules/music.go", "package rules\nfunc PlayMusic() {}\n")
	if got, why := VerifyItem(it, root); got != Broken {
		t.Errorf("產品程式碼出現之後回 %v（%s），預期斷言不成立", got, why)
	}
}

// TestGoTestSignal：`go_test` 問的是另一個問題（有沒有測試釘著），
// 所以它**只**看得到測試檔。
func TestGoTestSignal(t *testing.T) {
	root := t.TempDir()
	it := &Item{ID: "x", Status: StatusDone, Title: "t", Verify: Verify{
		Kind: "go_test", Pkg: "internal/rules", Run: "TestStaffPick",
	}}
	write(t, root, "internal/rules/staff_test.go",
		"package rules\nfunc TestStaffPick(t *testing.T) {}\n")
	if got, why := VerifyItem(it, root); got != Holds {
		t.Errorf("測試在的時候回 %v（%s），預期一致", got, why)
	}
	// 測試被刪掉或改名 → done 這個斷言就沒有東西撐著了。
	write(t, root, "internal/rules/staff_test.go", "package rules\n")
	if got, why := VerifyItem(it, root); got != Broken {
		t.Errorf("測試沒了之後回 %v（%s），預期斷言不成立", got, why)
	}
}

// TestManualAlwaysUnknown 釘住那條 [HARD]：**沉默不等於通過**。
func TestManualAlwaysUnknown(t *testing.T) {
	root := t.TempDir()
	for _, st := range []string{StatusOpen, StatusDone} {
		it := &Item{ID: "x", Status: st, Title: "t", Verify: Verify{
			Kind: "manual", Note: "接上時綁 rules.PlayMusic",
		}}
		got, why := VerifyItem(it, root)
		if got != Unknown {
			t.Errorf("status=%s 的 manual 回 %v（%s），預期無法核實", st, got, why)
		}
		if !strings.Contains(why, "要人判") {
			t.Errorf("manual 的說明是 %q，應該標明要人判", why)
		}
	}
}

// TestReceiptThreeWays：收據是本機證據，三種結果都要分得開。
func TestReceiptThreeWays(t *testing.T) {
	root := t.TempDir()
	body := []byte("oracle bytes")
	sum := sha256.Sum256(body)
	good := hex.EncodeToString(sum[:])
	it := &Item{ID: "x", Status: StatusDone, Title: "t", Verify: Verify{
		Kind: "receipt", Batch: "staffpick", Files: map[string]string{"staffpre.bin": good},
	}}

	// 換一台機器：`workplace/` 在 .gitignore 裡，收據不存在。
	// 那既不是通過也不是壞了。
	if got, why := VerifyItem(it, root); got != Unknown {
		t.Errorf("沒有收據時回 %v（%s），預期無法核實", got, why)
	}
	write(t, root, "workplace/oracle/staffpick/staffpre.bin", string(body))
	if got, why := VerifyItem(it, root); got != Holds {
		t.Errorf("雜湊相符時回 %v（%s），預期一致", got, why)
	}
	write(t, root, "workplace/oracle/staffpick/staffpre.bin", "被改過了")
	if got, why := VerifyItem(it, root); got != Broken {
		t.Errorf("雜湊對不上時回 %v（%s），預期斷言不成立", got, why)
	}
}

// TestSpecStatusBothWays：規格標到 CONFORMED 才算數。
func TestSpecStatusBothWays(t *testing.T) {
	root := t.TempDir()
	it := &Item{ID: "x", Status: StatusDone, Title: "t", Verify: Verify{
		Kind: "spec_status", Spec: "docs/spec/029.md", State: "CONFORMED",
	}}
	write(t, root, "docs/spec/029.md", "# 規格\n狀態：**READY**\n")
	if got, why := VerifyItem(it, root); got != Broken {
		t.Errorf("規格還停在 READY 時回 %v（%s），預期斷言不成立", got, why)
	}
	write(t, root, "docs/spec/029.md", "# 規格\n狀態：**CONFORMED**\n")
	if got, why := VerifyItem(it, root); got != Holds {
		t.Errorf("規格標到 CONFORMED 時回 %v（%s），預期一致", got, why)
	}
}

// TestJSONLenBothWays：進度型的清單綁**項數**，不綁註解。
func TestJSONLenBothWays(t *testing.T) {
	root := t.TempDir()
	it := &Item{ID: "x", Status: StatusOpen, Title: "t", Verify: Verify{
		Kind: "json_len", Path: "docs/internals/data/x.json", Field: "frames", Max: 10,
	}}
	write(t, root, "docs/internals/data/x.json", `{"frames":[1,2,3]}`)
	if got, why := VerifyItem(it, root); got != Holds {
		t.Errorf("還沒超過門檻時回 %v（%s），預期一致", got, why)
	}
	write(t, root, "docs/internals/data/x.json",
		`{"frames":[1,2,3,4,5,6,7,8,9,10,11]}`)
	if got, why := VerifyItem(it, root); got != Broken {
		t.Errorf("超過門檻之後回 %v（%s），預期斷言不成立", got, why)
	}
}

// TestBadPathIsUnknownNotPass 釘住一件容易安靜溜過去的事：**寫錯路徑的
// verify 不能長得像通過**。
func TestBadPathIsUnknownNotPass(t *testing.T) {
	root := t.TempDir()
	it := &Item{ID: "x", Status: StatusOpen, Title: "t", Verify: Verify{
		Kind: "absent", Paths: []string{"internal/typo"}, Pattern: "whatever",
	}}
	got, why := VerifyItem(it, root)
	if got != Unknown {
		t.Errorf("路徑不存在時回 %v（%s），預期無法核實", got, why)
	}
	if !strings.Contains(why, "不存在") {
		t.Errorf("說明是 %q，應該講清楚路徑不存在", why)
	}
}

// TestCheckCatchesBadItems：schema 檢查要抓得到寫壞的條目。
func TestCheckCatchesBadItems(t *testing.T) {
	doc := &Doc{
		Schema: schemaID,
		Layers: map[string]string{"rule": "遊戲規則"},
		Items: []Item{
			{ID: "dup", Status: StatusOpen, Layer: "rule", Title: "a",
				Verify: Verify{Kind: "manual", Note: "n"}},
			{ID: "dup", Status: "maybe", Layer: "nope", Title: "",
				Verify: Verify{Kind: "present", Paths: []string{"x"}}},
			{ID: "noverify", Status: StatusDone, Layer: "rule", Title: "c"},
		},
	}
	errs := Check(doc)
	joined := strings.Join(errs, "\n")
	for _, want := range []string{"id 重複", "status", "layer", "沒有 title",
		"少了 pattern", "沒有 verify"} {
		if !strings.Contains(joined, want) {
			t.Errorf("Check 沒抓到 %q，實際抓到：\n%s", want, joined)
		}
	}
}

// TestRenderAndWriteBack：render 出來的東西要能塞回標記之間，
// **標記之外一個字都不能動**。
func TestRenderAndWriteBack(t *testing.T) {
	doc := &Doc{
		Schema: schemaID,
		Layers: map[string]string{"rule": "遊戲規則"},
		Items: []Item{
			{ID: "open-one", Status: StatusOpen, Layer: "rule", Title: "還沒做的",
				Body: []string{"一行敘述。"}, Accept: "對拍 0 差異",
				Verify: Verify{Kind: "manual", Note: "n"}},
			{ID: "done-one", Legacy: "18q", Status: StatusDone, Layer: "rule",
				Title: "做完的", Verify: Verify{Kind: "manual", Note: "n"}},
		},
	}
	out, err := Render(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"還剩 1 條", "### 還沒做的",
		"#### open-one", "原編號 18q", "不要手改"} {
		if !strings.Contains(out, want) {
			t.Errorf("render 少了 %q", want)
		}
	}
	// **未完成的要排在已完成的前面**——worklist 回答的是「還剩什麼」。
	if strings.Index(out, "#### open-one") > strings.Index(out, "#### done-one") {
		t.Error("未完成項應該排在已完成項前面")
	}

	root := t.TempDir()
	p := filepath.Join(root, "CONTEXT.md")
	if err := os.WriteFile(p, []byte("前言\n"+markBegin+"\n舊的\n"+markEnd+"\n後記\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteBack(p, out); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !strings.HasPrefix(s, "前言\n") || !strings.HasSuffix(s, "後記\n") {
		t.Errorf("標記以外的內容被動到了：\n%s", s)
	}
	if strings.Contains(s, "舊的") {
		t.Error("標記之間的舊內容沒有被換掉")
	}
}

// TestWriteBackRefusesWithoutMarkers：沒有標記就拒絕寫，不要蓋掉手寫內容。
func TestWriteBackRefusesWithoutMarkers(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "CONTEXT.md")
	if err := os.WriteFile(p, []byte("整份都是手寫的\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteBack(p, "新內容"); err == nil {
		t.Fatal("沒有標記時應該拒絕寫入")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "整份都是手寫的\n" {
		t.Errorf("拒絕之後檔案還是被動到了：%q", b)
	}
}

// TestMatrixRoundTrip 是搬家的安全網：**把真的那份矩陣 parse 完再 render
// 回去，內容必須一模一樣**。掉一列或吃掉一個註腳，在 174 列裡不會有人發現。
//
// 比的是**語意**不是字元：原檔的欄位間距是手打的、不一致（全形括號後面
// 常常沒有空格），render 會統一成 `| a | b |`。那是一次性的正規化，
// 之後矩陣都由工具產生。真正不能變的是欄位內容、列數與註腳。
func TestMatrixRoundTrip(t *testing.T) {
	t.Chdir("../..")
	m1, rows1, err := ParseMatrix(matrixTarget)
	if err != nil {
		t.Skipf("讀不到 %s：%v", matrixTarget, err)
	}
	doc := &Doc{Matrix: m1}
	for si, sec := range m1.Sections {
		for ri, cells := range rows1[si] {
			if len(cells) != len(sec.Columns) {
				t.Fatalf("%s 第 %d 列有 %d 欄，表頭 %d 欄",
					sec.Title, ri+1, len(cells), len(sec.Columns))
			}
			cm := map[string]string{}
			for i, col := range sec.Columns {
				cm[col] = cells[i]
			}
			doc.Items = append(doc.Items, Item{
				ID: fmt.Sprintf("%s-%d", sec.ID, ri), Section: sec.ID, Cells: cm})
		}
	}
	out, err := RenderMatrix(doc)
	if err != nil {
		t.Fatal(err)
	}

	// render 出來的再 parse 一次，兩次的結果要相同。
	tmp := filepath.Join(t.TempDir(), "m.md")
	if err := os.WriteFile(tmp, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	m2, rows2, err := ParseMatrix(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m1.Head, m2.Head) {
		t.Errorf("前言變了：\n%v\n%v", m1.Head, m2.Head)
	}
	if len(m1.Sections) != len(m2.Sections) {
		t.Fatalf("章節數 %d → %d", len(m1.Sections), len(m2.Sections))
	}
	for i := range m1.Sections {
		a, b := m1.Sections[i], m2.Sections[i]
		if a.Title != b.Title || !reflect.DeepEqual(a.Columns, b.Columns) ||
			!reflect.DeepEqual(a.Intro, b.Intro) || !reflect.DeepEqual(a.Outro, b.Outro) {
			t.Errorf("章節 %q 變了：\n%+v\n%+v", a.Title, a, b)
		}
		if !reflect.DeepEqual(rows1[i], rows2[i]) {
			if len(rows1[i]) != len(rows2[i]) {
				t.Fatalf("章節 %q 的列數 %d → %d", a.Title, len(rows1[i]), len(rows2[i]))
			}
			for r := range rows1[i] {
				if !reflect.DeepEqual(rows1[i][r], rows2[i][r]) {
					t.Fatalf("章節 %q 第 %d 列變了：\n%q\n%q",
						a.Title, r+1, rows1[i][r], rows2[i][r])
				}
			}
		}
	}
}

// 自評的數字要**算自資料**，而且矩陣以外的層要分開數。
//
// ⚠ 反面不會報錯：一個手寫死的百分比照樣印得出來，只是條目一改就過期，
// 而過期的完成度看起來完全正常——只有在別人照它排工作時才會發現。
func TestScoreCountsFromData(t *testing.T) {
	doc := &Doc{
		Layers: map[string]string{"matrix": "矩陣", "re": "逆向", "feature": "功能"},
		Matrix: &Matrix{Sections: []Section{{ID: "s1", Title: "顯示層"}, {ID: "s2", Title: "畫面"}}},
		Items: []Item{
			{ID: "a", Layer: "matrix", Section: "s1", Status: StatusDone, Title: "甲"},
			{ID: "b", Layer: "matrix", Section: "s2", Status: StatusDone, Title: "乙"},
			{ID: "c", Layer: "matrix", Section: "s2", Status: StatusOpen, Title: "丙"},
			{ID: "d", Layer: "re", Status: StatusOpen, Title: "丁"},
			{ID: "e", Layer: "feature", Status: StatusOpen, Title: "戊"},
			{ID: "f", Layer: "feature", Status: StatusDone, Title: "己"},
		},
	}
	// **比對前把連續空白收成一個**：欄寬是靠 `%-14s` 排的，而中文字的
	// 寬度與 rune 數不一致，把 padding 寫進預期值等於把測試綁在排版上。
	out := Score(doc)
	flat := strings.Join(strings.Fields(out), " ")
	for _, want := range []string{
		"s1 顯示層 100% 1/1",   // 全過
		"s2 畫面 50% 1/2",     // 一過一沒過
		"矩陣合計 2/3 條通過（66%）", // 只數矩陣層
		"re 1 條",            // 矩陣以外的層分開數
		"feature 1 條",       // 只數 open，不含那條 done
		"丙",                 // 沒過的那一列要點名
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("報表少了 %q，實際：\n%s", want, out)
		}
	}
	// 已經做完的不該出現在「還差多少」裡。
	if strings.Contains(out, "己") {
		t.Errorf("做完的條目不該出現在還差多少：\n%s", out)
	}
}

// WriteBack 跑兩次要與跑一次相同。
//
// ⚠ 不冪等的症狀是**每跑一次多一個空行**，在 diff 裡幾乎看不見，而
// `render -w` 是每一輪收尾都會跑的——十幾次之後 CONTEXT.md 的某一節
// 前面就堆了一整片空白，沒有任何東西會開口。
func TestWriteBackIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/ctx.md"
	orig := "# 標題\n\n" + markBegin + "\n舊內容\n" + markEnd + "\n\n## 下一節\n"
	if err := os.WriteFile(path, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	body := markBegin + "\n\n新內容\n\n" + markEnd + "\n"
	if err := WriteBack(path, body); err != nil {
		t.Fatal(err)
	}
	once, _ := os.ReadFile(path)
	if err := WriteBack(path, body); err != nil {
		t.Fatal(err)
	}
	twice, _ := os.ReadFile(path)
	if string(once) != string(twice) {
		t.Errorf("跑兩次不同：\n一次 %q\n兩次 %q", once, twice)
	}
	if !strings.Contains(string(once), markEnd+"\n\n## 下一節") {
		t.Errorf("標記後面原本的空行應該保留：%q", once)
	}
}
