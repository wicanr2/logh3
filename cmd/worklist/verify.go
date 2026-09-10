package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// signal 是訊號說的話。**它只描述現況，不管條目自己宣稱什麼**——
// 兩者對不對得上由 VerifyItem 判。
type signal int

const (
	sigDone signal = iota // 訊號說：這件事做好了
	sigOpen               // 訊號說：這件事還沒做
	sigNone               // 拿不到訊號
)

// 掃描時只看這幾種副檔名——`workplace/` 裡有 1 MB 的傾印，別去讀它們。
var scanExt = map[string]bool{
	".go": true, ".py": true, ".sh": true, ".md": true, ".json": true,
}

// VerifyItem 回「這一條的斷言還成立嗎」。
//
//	status ＝ open，訊號說做好了 → 斷言不成立（過期斷言）
//	status ＝ done，訊號說沒做好 → 斷言不成立（回歸，或訊號被改名了）
func VerifyItem(it *Item, root string) (Result, string) {
	sig, why := readSignal(&it.Verify, root)
	if sig == sigNone {
		// [HARD] 沉默不等於通過：拿不到訊號一律標出來要人判。
		return Unknown, why
	}
	holds := (sig == sigDone) == (it.Status == StatusDone)
	if holds {
		return Holds, why
	}
	return Broken, why
}

func readSignal(v *Verify, root string) (signal, string) {
	switch v.Kind {
	case "manual":
		note := v.Note
		if note == "" {
			note = "沒有可綁的訊號"
		}
		return sigNone, "要人判：" + note

	case "present":
		// 綁的是程式碼裡的**自承註解**（「remake 還沒有…」）：還在 ＝ 還沒做。
		where, err := grep(root, v.Paths, v.Pattern, false)
		if err != nil {
			return sigNone, err.Error()
		}
		if where != "" {
			return sigOpen, "自承還在 " + where
		}
		return sigDone, "找不到 " + v.Pattern

	case "absent":
		// 綁的是**接上時一定會出現**的東西（型別名、函式名）：出現了 ＝ 做好了。
		where, err := grep(root, v.Paths, v.Pattern, false)
		if err != nil {
			return sigNone, err.Error()
		}
		if where != "" {
			return sigDone, "已經出現在 " + where
		}
		return sigOpen, "還沒出現：" + v.Pattern

	case "go_test":
		// 只問「有沒有測試釘著」，不跑它——測試紅不紅是 `go test ./...` 的職責，
		// 這裡便宜到每次都跑得起。
		where, err := grep(root, []string{v.Pkg}, `func `+regexp.QuoteMeta(v.Run)+`\(`, true)
		if err != nil {
			return sigNone, err.Error()
		}
		if where != "" {
			return sigDone, v.Run + " 釘在 " + where
		}
		return sigOpen, "沒有這支測試：" + v.Run

	case "spec_status":
		b, err := os.ReadFile(filepath.Join(root, v.Spec))
		if err != nil {
			return sigNone, fmt.Sprintf("讀不到規格 %s", v.Spec)
		}
		if strings.Contains(string(b), v.State) {
			return sigDone, fmt.Sprintf("%s 標著 %s", v.Spec, v.State)
		}
		return sigOpen, fmt.Sprintf("%s 沒有標 %s", v.Spec, v.State)

	case "receipt":
		return receiptSignal(v, root)

	case "json_len":
		b, err := os.ReadFile(filepath.Join(root, v.Path))
		if err != nil {
			return sigNone, fmt.Sprintf("讀不到 %s", v.Path)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(b, &m); err != nil {
			return sigNone, fmt.Sprintf("%s 解不開：%v", v.Path, err)
		}
		raw, ok := m[v.Field]
		if !ok {
			return sigNone, fmt.Sprintf("%s 沒有欄位 %s", v.Path, v.Field)
		}
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return sigNone, fmt.Sprintf("%s 的 %s 不是陣列", v.Path, v.Field)
		}
		msg := fmt.Sprintf("%s 的 %s 有 %d 項（門檻 %d）", v.Path, v.Field, len(arr), v.Max)
		if len(arr) > v.Max {
			return sigDone, msg
		}
		return sigOpen, msg
	}
	return sigNone, "不認得的 verify.kind：" + v.Kind
}

// receiptSignal 比對收據檔的 SHA-256。
//
// **收據是本機證據**（`workplace/` 在 .gitignore 裡），換一台機器就不存在
// ——那時回 sigNone 標成「無法核實」，不是「通過」也不是「壞了」。
func receiptSignal(v *Verify, root string) (signal, string) {
	dir := filepath.Join(root, "workplace", "oracle", v.Batch)
	if _, err := os.Stat(dir); err != nil {
		return sigNone, "這台機器沒有收據 " + v.Batch
	}
	var bad []string
	for name, want := range v.Files {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return sigNone, fmt.Sprintf("收據 %s 缺 %s", v.Batch, name)
		}
		sum := sha256.Sum256(b)
		if got := hex.EncodeToString(sum[:]); got != want {
			bad = append(bad, fmt.Sprintf("%s（%s… 應為 %s…）", name, got[:12], want[:12]))
		}
	}
	if len(bad) > 0 {
		return sigOpen, "收據雜湊對不上：" + strings.Join(bad, "、")
	}
	return sigDone, fmt.Sprintf("收據 %s 的 %d 個檔案雜湊相符", v.Batch, len(v.Files))
}

// grep 在 paths 底下找 pattern，回第一個命中的檔案（相對路徑）。
//
// **[HARD] `includeTests` 預設是 false，因為條目問的是「產品程式碼做了這件事
// 沒有」。** 測試本來就會提到還沒接上的東西——為了釘住將來的行為，或為了測
// 那個資料結構本身。把測試檔算進來，`absent` 會因為測試裡有一行呼叫就判成
// 「已經做了」，真缺口就這樣被蓋掉。
//
// 只有 `go_test` 傳 true：它問的是另一個問題（有沒有測試釘著），
// 而答案只可能在測試檔裡。
func grep(root string, paths []string, pattern string, includeTests bool) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("verify 少了 pattern")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("pattern 編不起來：%w", err)
	}
	for _, target := range paths {
		full := filepath.Join(root, target)
		info, err := os.Stat(full)
		if err != nil {
			return "", fmt.Errorf("verify 指到的 %s 不存在", target)
		}
		if !info.IsDir() {
			if !includeTests && isTestFile(full) {
				continue
			}
			hit, err := fileHas(full, re)
			if err != nil {
				return "", err
			}
			if hit {
				return target, nil
			}
			continue
		}
		var found string
		err = filepath.WalkDir(full, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || found != "" {
				return err
			}
			if !scanExt[filepath.Ext(p)] {
				return nil
			}
			if !includeTests && isTestFile(p) {
				return nil
			}
			hit, err := fileHas(p, re)
			if err != nil {
				return err
			}
			if hit {
				rel, _ := filepath.Rel(root, p)
				found = rel
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		if found != "" {
			return found, nil
		}
	}
	return "", nil
}

// isTestFile 認 Go 的 `*_test.go` 與 Python 的 `test_*.py`。
func isTestFile(path string) bool {
	name := filepath.Base(path)
	return strings.Contains(name, "_test.") || strings.HasPrefix(name, "test_")
}

func fileHas(path string, re *regexp.Regexp) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return re.Match(b), nil
}

// checkVerify 驗一條 verify 的必填欄位。**寫錯路徑的 verify 會永遠回同一個
// 答案**，那比沒有 verify 更糟——它給人「有在看」的假象。
func checkVerify(it *Item) []string {
	v := &it.Verify
	need := func(cond bool, msg string) []string {
		if cond {
			return nil
		}
		return []string{fmt.Sprintf("%s：%s", it.ID, msg)}
	}
	switch v.Kind {
	case "manual":
		return need(v.Note != "", "manual 一定要寫 note：將來接上時該綁什麼")
	case "present", "absent":
		var errs []string
		errs = append(errs, need(len(v.Paths) > 0, v.Kind+" 少了 paths")...)
		errs = append(errs, need(v.Pattern != "", v.Kind+" 少了 pattern")...)
		if v.Pattern != "" {
			if _, err := regexp.Compile(v.Pattern); err != nil {
				errs = append(errs, fmt.Sprintf("%s：pattern 編不起來：%v", it.ID, err))
			}
		}
		return errs
	case "go_test":
		var errs []string
		errs = append(errs, need(v.Pkg != "", "go_test 少了 pkg")...)
		errs = append(errs, need(v.Run != "", "go_test 少了 run")...)
		return errs
	case "spec_status":
		var errs []string
		errs = append(errs, need(v.Spec != "", "spec_status 少了 spec")...)
		errs = append(errs, need(v.State != "", "spec_status 少了 state")...)
		return errs
	case "receipt":
		var errs []string
		errs = append(errs, need(v.Batch != "", "receipt 少了 batch")...)
		errs = append(errs, need(len(v.Files) > 0, "receipt 少了 files")...)
		return errs
	case "json_len":
		var errs []string
		errs = append(errs, need(v.Path != "", "json_len 少了 path")...)
		errs = append(errs, need(v.Field != "", "json_len 少了 field")...)
		return errs
	case "":
		return []string{it.ID + "：沒有 verify"}
	}
	return []string{fmt.Sprintf("%s：不認得的 verify.kind %q", it.ID, v.Kind)}
}
