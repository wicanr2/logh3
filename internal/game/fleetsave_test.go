package game

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/logh3/internal/rules"
)

func fleetSaveEqual(t *testing.T, got, want []rules.Fleet) {
	t.Helper()
	copyFleets := func(in []rules.Fleet) []rules.Fleet {
		out := append([]rules.Fleet(nil), in...)
		for i := range out {
			out[i].Strength = 0
		} // 解析時依權重重建的快取，不是存檔欄位。
		return out
	}
	if !reflect.DeepEqual(copyFleets(got), copyFleets(want)) {
		t.Fatalf("保存／載入改變現況艦隊\n取得：%+v\n原有：%+v", got, want)
	}
}

func TestSaveLiveFleetStateLoop(t *testing.T) {
	control, g := reachFleetDecided(t), reachFleetDecided(t)
	if control.rand94DC().Seed() != strategyRandSeed || g.rand94DC().Seed() != strategyRandSeed {
		t.Fatal("執行前未固定種子")
	}
	executeTurn(t, control)
	executeTurn(t, g)
	if len(g.fleets) != 4 {
		t.Fatalf("未到達雙方建隊局面：%d", len(g.fleets))
	}
	for _, f := range g.fleets {
		if f.Faction == 1 && len(f.Composition) == 0 {
			t.Fatal("AI 編制為空")
		}
	}
	before := bytes.Clone(g.res.Stellar)
	planets := bytes.Clone(g.res.Planets.Data)
	dir := t.TempDir()
	g.res.SaveDir = dir
	g.Click(544, 125)
	g.Click(560, 165)
	wantState(t, g, StateSaveGame, "正常存進度")
	g.Click(300, g.res.SlotList.Save.TextTop)
	if err := g.SaveErr(); err != nil {
		t.Fatal(err)
	}
	for _, name := range rules.SlotFileNames(0) {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	want := g.slot.Fleets(g.res.Weights)
	fleetSaveEqual(t, want, g.fleets)
	// 保存後仍由活的編制讀 AI 艘數；故意改底稿不能影響計分輸入。
	f := g.fleets[1]
	p := f.Composition[0].Pool*rules.ShipRecord + 8
	old := binary.LittleEndian.Uint16(g.slot.Ship[p:])
	binary.LittleEndian.PutUint16(g.slot.Ship[p:], 0)
	ships, _ := g.fleetShipTotals(f)
	if ships == 0 {
		t.Fatal("AI 改讀過期存檔底稿")
	}
	binary.LittleEndian.PutUint16(g.slot.Ship[p:], old)
	g.RightClick(300, 122)
	wantState(t, g, StateCommandPhase, "正常返回")
	g.Click(544, 125)
	g.Click(560, 145)
	wantState(t, g, StateLoadGame, "正常取進度")
	g.Click(300, 122)
	wantState(t, g, StateLoadFaction, "存檔列後選陣營")
	g.Click(318, 208)
	wantState(t, g, StateCommandPhase, "取進度完成")
	fleetSaveEqual(t, g.fleets, want)
	for r := 0; r < rules.RegionCount; r++ {
		o := r*rules.RegionRecord + 0x28
		supplyTimingEqual(t, "載入後雙方艦隊計數", g.res.Stellar[o:o+2], before[o:o+2])
	}
	supplyTimingEqual(t, "載入後行星全表", g.res.Planets.Data, planets)
	// 控制組不經存讀。兩側下一回合執行前設定同一受控種子；不要求存檔新增 RNG 欄位。
	control.rand94DC().SetSeed(strategyRandSeed)
	g.rand94DC().SetSeed(strategyRandSeed)
	executeTurn(t, control)
	executeTurn(t, g)
	wantState(t, g, control.State(), "取進度後繼續回合")
	// 控制組尚未物化池索引；先同步出一份原版格式，才能比較身份欄位。
	if err := control.SaveSlot(t.TempDir(), 0); err != nil {
		t.Fatal(err)
	}
	if err := g.SaveSlot(t.TempDir(), 0); err != nil {
		t.Fatal(err)
	}
	// 控制組直到第二回合後才物化原始表，新艦隊排序會改變初配池索引。
	// 首次存讀的池身份已在上面完整比較；此處比較不經存讀的玩法結果。
	logical := func(in []rules.Fleet) []rules.Fleet {
		out := append([]rules.Fleet(nil), in...)
		for i := range out {
			out[i].Scout = -1
			out[i].Composition = append([]rules.FleetGroup(nil), out[i].Composition...)
			for j := range out[i].Composition {
				out[i].Composition[j].Pool = -1
			}
		}
		return out
	}
	fleetSaveEqual(t, logical(g.fleets), logical(control.fleets))
	supplyTimingEqual(t, "後續回合行星", g.res.Planets.Data, control.res.Planets.Data)
	supplyTimingEqual(t, "後續回合星域", g.res.Stellar, control.res.Stellar)
}
