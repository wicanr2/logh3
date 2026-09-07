package font

import "testing"

// TestBig5IndexMatchesVerifiedOffsets：索引公式是整個文字系統的地基，
// 錯一格就是整片字全錯位，而畫面上看起來只是「字怪怪的」。
// 這裡釘的九個位移是 dosgolem 實測遊戲 seek 過去的值
// （`docs/internals/data/font-graph-img.json` 的 verified_seeks）。
func TestBig5IndexMatchesVerifiedOffsets(t *testing.T) {
	cases := []struct {
		name string
		code uint16
		off  int
	}{
		{"帝", 0xABD2, 63360}, {"國", 0xB0EA, 87630}, {"同", 0xA650, 36930},
		{"盟", 0xB7F9, 121050}, {"對", 0xB9EF, 130170}, {"戰", 0xBED4, 152910},
		{"取", 0xA8FA, 50430}, {"進", 0xB669, 113040}, {"度", 0xABD7, 63510},
	}
	for _, c := range cases {
		if got := offsetOf(c.code); got != c.off {
			t.Errorf("%s（%04X）位移 = %d，預期 %d", c.name, c.code, got, c.off)
		}
	}
}

// TestGlyphRejectsInvalidCode：不合法字碼要回 nil，不能回空白。
// 回空白的話缺字會長得跟空格一模一樣，翻譯漏字永遠查不出來。
func TestGlyphRejectsInvalidCode(t *testing.T) {
	f := &Font{data: make([]byte, 500000)}
	for _, code := range []uint16{0x0041, 0xA13F, 0xA17F, 0xFA40} {
		if g := f.Glyph(code); g != nil {
			t.Errorf("字碼 %04X 應該回 nil，得到 %v", code, g)
		}
	}
	if g := f.Glyph(0xA650); g == nil {
		t.Error("合法字碼 A650 回了 nil")
	}
}

// TestGlyphLevel2 釘住次常用字（Big5 0xC940 起）的位移。
//
// 0xC6A1–0xC8FE 那一段沒有存進字型檔，所以次常用字要扣掉 408 個碼；
// 不扣的話畫出來是別的字，而且**畫得出來**——最難發現的那種錯。
func TestGlyphLevel2(t *testing.T) {
	// 璐（0xED69，遊戲設定的第一首 BGM「波琍璐」）的點陣在 359850
	// ——量法是把收據 settings1e 上那個字的 16×15 點陣打包成 30 bytes，
	// 拿去 GRAPH.IMG 裡搜。
	want := 359850
	if got := offsetOf(0xED69); got != want {
		t.Errorf("0xED69 的位移 %d，預期 %d", got, want)
	}
	// 常用字不受影響：待（0xA5DD）。
	if got := offsetOf(0xA5DD); got != 35430 {
		t.Errorf("0xA5DD 的位移 %d，預期 35430", got)
	}
}
