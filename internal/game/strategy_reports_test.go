package game

import "testing"

// TestStrategyReportFamiliesRegistered 是整個戰略報告族群的單一產品接線閘門。
// 各 family 的內容與 lifecycle 仍由同套件的專用測試逐條驗證；這裡防止其中
// 一個狀態或共用 panel 在 bootstrap／狀態名稱層被漏接。
func TestStrategyReportFamiliesRegistered(t *testing.T) {
	r := load(t)
	if r.FlipReport == nil || r.PhezzanReport == nil || r.VictoryReport == nil ||
		r.FrameTiles == nil || r.Faces == nil {
		t.Fatal("戰略報告所需 panel／FRAME／FACE 沒有全部載入")
	}
	want := map[State]string{
		StateRegionFlipReport:  "StateRegionFlipReport",
		StateRebellionReport:   "StateRebellionReport",
		StateOccupationReport:  "StateOccupationReport",
		StateAdmiralFateReport: "StateAdmiralFateReport",
		StatePhezzanReport:     "StatePhezzanReport",
		StateVictoryReport:     "StateVictoryReport",
	}
	for state, name := range want {
		if got := state.String(); got != name {
			t.Errorf("%s 的穩定狀態名=%q", name, got)
		}
	}
}
