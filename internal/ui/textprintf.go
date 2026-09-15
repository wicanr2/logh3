package ui

import "strconv"

// 格式化輸出（原版 `3ACA:0079`，由 `3617:003A` 帶著視窗一起呼叫）。
//
// 原版的 printf 把結果一個位元組一個位元組交給輸出回呼，所以**格式化
// 出來的位元組與字串本身走同一條路**——控制碼、全形前後半、自動折行
// 都在 PutByte 那一層處理，這裡只負責把 `%…` 換成位元組。
//
// 只做原版格式字串用得到的幾種：`%s`、`%c`、`%d`／`%u`／`%ld`／`%lu`，
// 加上 `-`（靠左）、`0`（補零）與寬度。`l` 只影響取值的寬度，輸出一樣。
// 沒見過的轉換原樣輸出，**不猜**。

// Printf 把格式化的結果寫進緩衝區。args 依序對應格式字串裡的轉換：
// `%s` 吃 []byte 或 string，數字吃 int／int64／uint16 之類。
func (b *TextBuffer) Printf(format []byte, args ...any) {
	formatTo(b.PutByte, format, args...)
}

// Sprintf 是同一支 printf，但把結果收成位元組串而不是寫進視窗
// ——原版 `3ACA:001B` 就是這樣先排進暫存區再交給訊息框的
// （確認框「要移動至…星系嗎？」，`docs/re/312` §四）。
func Sprintf(format []byte, args ...any) []byte {
	var out []byte
	formatTo(func(c byte) { out = append(out, c) }, format, args...)
	return out
}

func formatTo(put func(byte), format []byte, args ...any) {
	ai := 0
	next := func() any {
		if ai >= len(args) {
			return nil
		}
		v := args[ai]
		ai++
		return v
	}
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' {
			put(c)
			continue
		}
		if i+1 < len(format) && format[i+1] == '%' {
			put('%')
			i++
			continue
		}
		j := i + 1
		left, zero := false, false
		for j < len(format) && (format[j] == '-' || format[j] == '0') {
			if format[j] == '-' {
				left = true
			} else {
				zero = true
			}
			j++
		}
		width := 0
		for j < len(format) && format[j] >= '0' && format[j] <= '9' {
			width = width*10 + int(format[j]-'0')
			j++
		}
		for j < len(format) && (format[j] == 'l' || format[j] == 'h') {
			j++
		}
		if j >= len(format) {
			put(c)
			continue
		}
		var out []byte
		switch format[j] {
		case 's':
			switch v := next().(type) {
			case []byte:
				out = v
			case string:
				out = []byte(v)
			}
		case 'c':
			if v, ok := toInt(next()); ok {
				out = []byte{byte(v)}
			}
		case 'd', 'u', 'i':
			if v, ok := toInt(next()); ok {
				out = []byte(strconv.FormatInt(v, 10))
			}
		default:
			// 沒見過的轉換：原樣輸出那一段，等有證據再補。
			for _, c := range format[i : j+1] {
				put(c)
			}
			i = j
			continue
		}
		pad := width - len(out)
		if pad > 0 && !left {
			fill := byte(' ')
			if zero {
				fill = '0'
			}
			for k := 0; k < pad; k++ {
				put(fill)
			}
		}
		for _, c := range out {
			put(c)
		}
		if pad > 0 && left {
			for k := 0; k < pad; k++ {
				put(' ')
			}
		}
		i = j
	}
}

func toInt(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case int32:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint8:
		return int64(n), true
	}
	return 0, false
}
