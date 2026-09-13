package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
)

// TestFleetOrgMatchesOracle 提督詳細資料按「出擊」開出艦隊編成
// （收據 `fleet1e`／`fleet2e`，`docs/spec/013`）。
//
// 兩張收據點的是不同提督：萊因哈特的副官是修特萊、吉爾菲艾斯的是錦茲
// （記錄 `+0x07`）。艦隊能力十項逐項取大，副官贏的那幾格印成色 5
// ——兩張的墨色分布不同，是這條規則最直接的證據（`docs/re/66`）。
func TestFleetOrgMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		y     int
		batch string
	}{
		{85, "fleet1e"},  // 第 0 列 萊因哈特，副官修特萊
		{105, "fleet2e"}, // 第 1 列 吉爾菲艾斯，副官錦茲
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachCommandPhase(t)
			g.Click(352, 125) // 戰略命令
			g.Click(400, 125) // 出擊命令
			g.Click(400, c.y)
			if g.State() != StateAdmiralDetail {
				t.Fatalf("點了 y=%d 之後在 %v", c.y, g.State())
			}
			g.Click(488, 330) // 「出擊」
			if g.State() != StateFleetOrg {
				t.Fatalf("點了出擊之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// reachFleetOrg 走玩家路徑走到萊因哈特那張艦隊編成。
func reachFleetOrg(t *testing.T) *Game {
	t.Helper()
	g := reachCommandPhase(t)
	g.Click(352, 125) // 戰略命令
	g.Click(400, 125) // 出擊命令
	g.Click(400, 85)  // 第 0 列 萊因哈特
	g.Click(488, 330) // 「出擊」
	if g.State() != StateFleetOrg {
		t.Fatalf("走到艦隊編成之前就跑到 %v", g.State())
	}
	return g
}

// TestFleetOrgHotspotsAreNarrowerThanTheBoxes 點在畫出來的框內、熱區外
// 一個像素，原版一個位元組都不動（收據 `fnum2`／`fadj2` 與 `fleet1e`
// 雜湊相同，`docs/re/68`）。
//
// 這兩個是**反例**：正例只說得出「這裡點得到」，反例才定得住邊界。
func TestFleetOrgHotspotsAreNarrowerThanTheBoxes(t *testing.T) {
	for _, c := range []struct {
		x, y  int
		batch string
		what  string
	}{
		{400, 350, "fnum2", "艦隊編號框的右邊緣"},
		{448, 350, "fadj2", "副官框的右邊緣"},
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(c.x, c.y)
			if g.State() != StateFleetOrg {
				t.Fatalf("點 %s (%d,%d) 之後跑到 %v", c.what, c.x, c.y, g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetNumberListMatchesOracle 「艦隊編號」開出的十六列清單
// （收據 `fnum`／`fnum3`，`docs/re/67`）。滑鼠停哪一列，那一列整條反白。
func TestFleetNumberListMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		x     int
		batch string
	}{
		{368, "fnum"},  // 游標壓在「艦」上
		{338, "fnum3"}, // 游標壓在「第」上
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(c.x, 350)
			if g.State() != StateFleetNumber {
				t.Fatalf("點了艦隊編號之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetAdjutantMenuMatchesOracle 「副官」開出的「變更／卸任」兩列
// 小選單（收據 `fadj`／`fadj3`）。
func TestFleetAdjutantMenuMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		x     int
		batch string
	}{
		{432, "fadj"},
		{417, "fadj3"},
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(c.x, 350)
			if g.State() != StateFleetAdjutant {
				t.Fatalf("點了副官之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetCompositionMatchesOracle 艦隊編成按「確認」開出的**艦隊組成**
// （收據 `fok`，`docs/re/67`、`docs/re/69`）。
//
// 七列的「待命部隊」是首都星域的艦艇艘數 ÷ 一單位幾艘，「數量」印的就是
// 一單位幾艘；格盤是 9 × 9 的 `SPRITE.SPR` 圖塊。
func TestFleetCompositionMatchesOracle(t *testing.T) {
	g := reachFleetOrg(t)
	g.Click(480, 350) // 「確認」
	if g.State() != StateFleetComposition {
		t.Fatalf("點了確認之後在 %v", g.State())
	}
	matches(t, g, "fok")
}

// TestFleetCompositionMultipleGroupsMatchesOracle 用收據 `fmx1` 的記憶體傾印
// 重建三組編制，再逐像素比較編隊八格與 9×9 格盤。這份局面把艦隊 0 改成 3／8 組：
// 艦種 B／C／D 分別在隊形槽 20／10／30；原版壓實工作表必須與艦隊記錄
// 逐組相同，畫面上的字母、空格、部隊數與 9×9 格盤因此全部一起受驗。
func TestFleetCompositionMultipleGroupsMatchesOracle(t *testing.T) {
	g := reachFleetOrg(t)
	g.Click(480, 350) // 「確認」→ 艦隊組成

	fleet, err := os.ReadFile(filepath.Join(root, "workplace/fmx1_fleet.bin"))
	if err != nil {
		t.Skip("沒有 fmx1 艦隊傾印，跳過（tools/oracle.sh fmx1 可重建）")
	}
	work, err := os.ReadFile(filepath.Join(root, "workplace/fmx1_work.bin"))
	if err != nil {
		t.Skip("沒有 fmx1 工作表傾印，跳過（tools/oracle.sh fmx1 可重建）")
	}
	if len(fleet) != rules.FleetRecord || len(work) != 0xE6 {
		t.Fatalf("fmx1 傾印尺寸：fleet=%d work=%d，應為 %d／230",
			len(fleet), len(work), rules.FleetRecord)
	}
	wantGroups, wantSlots := int(fleet[0x0F]), int(fleet[0x0E])
	if wantGroups != 3 || wantSlots != 8 || work[0xD1] != byte(wantGroups) ||
		work[0xD2] != byte(wantSlots-wantGroups) {
		t.Fatalf("fmx1 組數不符：艦隊 %d/%d，工作表 %d+%d",
			wantGroups, wantSlots, work[0xD1], work[0xD2])
	}

	groups := make([]screen.CompositionGroup, wantGroups)
	for i := range groups {
		fleetGroup := fleet[0x10+i*4 : 0x14+i*4]
		packedGroup := work[0xA4+i*4 : 0xA8+i*4] // ds:981E - ds:977A
		if string(fleetGroup) != string(packedGroup) {
			t.Fatalf("第 %d 組：艦隊 % X，壓實工作表 % X", i, fleetGroup, packedGroup)
		}
		groups[i] = screen.CompositionGroup{
			Kind: int(packedGroup[2]), Slot: int(packedGroup[3]),
			// 收據刻意讓三組都指艦艇池第 0 筆；該筆現有量等於滿編量，
			// 所以三個字母都走滿編色 7（docs/re/293）。
			Ships: 1, Full: 1,
		}
	}
	for i, want := range []screen.CompositionGroup{
		{Kind: 0, Slot: 20},
		{Kind: 2, Slot: 10},
		{Kind: 4, Slot: 30},
	} {
		if groups[i].Kind != want.Kind || groups[i].Slot != want.Slot {
			t.Fatalf("第 %d 組是艦種 %d／槽 %d，應為 %d／%d",
				i, groups[i].Kind, groups[i].Slot, want.Kind, want.Slot)
		}
	}

	region := rules.CapitalRegion(g.faction)
	o := region * rules.RegionRecord
	ships := rules.RegionShips(g.res.Stellar[o:o+rules.RegionRecord], g.faction)
	v := screen.CompositionView{
		Ships: ships, PerUnit: rules.ShipsPerUnit[:], Mobility: rules.NewFleetMorale,
		Groups: groups, Slots: wantSlots,
	}
	// 動態內容必須在 overlay 畫完、游標畫上去之前落筆；若在 g.Frame
	// 完成後才覆畫，舊的「□」與游標會留殘影，測到的是錯誤的圖層順序。
	spec := g.commandSpec(g.state)
	g.setListRows(spec)
	g.liveCursor(spec)
	g.setHoverRow(spec)
	g.applyStaffFaces(spec)
	fb, _ := screen.MainGameLayered(spec, g.res.Font, g.starMapFor(spec),
		g.res.Faces, g.res.Cursor, g.drawBoardPanels, func(name string, fb *gfx.Framebuffer) {
			if name != "overlay" {
				return
			}
			g.drawFleetOrg(fb)
			g.res.CompositionPanel.Draw(fb, g.res.Font, v)
			g.drawCompositionGrid(fb, v)
			g.drawShipOrder(fb)
		})
	g.drawBudget(fb)
	g.drawProduction(fb)
	g.drawDelivery(fb)
	g.drawSlotList(fb)
	g.drawReport(fb)
	g.drawAdmiralDetail(fb)
	g.drawAdmiralBio(fb)
	g.drawSpecStaff(fb, spec)

	want, err := os.ReadFile(filepath.Join(root, "workplace/oracle/fmx1/screen.bin"))
	if err != nil {
		t.Skip("沒有 fmx1 畫面收據，跳過（tools/oracle.sh fmx1 可重建）")
	}
	// 這一條驗的是規則控制的每一格。游標壓在「部隊數」上，3／8 與
	// 原本 1／8 的游標交疊順序另有 88 像素差異，不屬於編隊字母規則；
	// 組數本身已由上面的艦隊記錄與 ds:984B／984C 雙重核對。
	boxes := [][4]int{
		{136, 168, 280, 312}, // 9×9 格盤
		{160, 342, 288, 358}, // 編隊八格
	}
	bad, total := 0, 0
	for _, box := range boxes {
		for y := box[1]; y < box[3]; y++ {
			for x := box[0]; x < box[2]; x++ {
				i := y*640 + x
				total++
				if want[i] != fb.Pix[i] {
					bad++
				}
			}
		}
	}
	if bad != 0 {
		if os.Getenv("LOGH3_DUMP_REMAKE") != "" {
			out := filepath.Join(root, "workplace/remake-fmx1.bin")
			if err := os.WriteFile(out, fb.Pix, 0o644); err != nil {
				t.Logf("倒 remake 畫面失敗：%v", err)
			} else {
				t.Logf("remake 畫面寫到 %s", out)
			}
		}
		t.Fatalf("fmx1 編隊八格與格盤：與原版不同的像素 %d / %d", bad, total)
	}
}

// TestFleetCompositionRowSelect 點七個艦種列的哪一列，選取條就跳到哪一列
// （收據 `row3`：點第 2 列巡航艦）。
func TestFleetCompositionRowSelect(t *testing.T) {
	g := reachFleetOrg(t)
	g.Click(480, 350) // 「確認」
	g.Click(360, 210) // 第 2 列 巡航艦
	if g.State() != StateFleetComposition {
		t.Fatalf("點了艦種列之後在 %v", g.State())
	}
	matches(t, g, "row3")
}

// TestFormationListMatchesOracle 艦隊組成按「陣形」開出的八列清單：
// 圓錐／紡錘／半月／Ｕ字／Ｍ字／Ｔ字／分散／密集（`docs/re/70`）。
//
// 三張收據點的是同一顆按鈕的不同位置——游標一次蓋得住兩列，八列要三份
// 收據才讀得齊，而三張的反白條都在同一列（滑鼠都落在那一格裡）。
func TestFormationListMatchesOracle(t *testing.T) {
	for _, c := range []struct {
		x, y  int
		batch string
	}{
		{510, 310, "form1"},
		{497, 300, "form2"},
		{497, 318, "form3"},
	} {
		t.Run(c.batch, func(t *testing.T) {
			g := reachFleetOrg(t)
			g.Click(480, 350) // 「確認」→ 艦隊組成
			g.Click(c.x, c.y) // 「陣形」
			if g.State() != StateFormation {
				t.Fatalf("點了陣形之後在 %v", g.State())
			}
			matches(t, g, c.batch)
		})
	}
}

// TestFleetDecideMatchesOracle 艦隊組成按「決定」：艦隊建成，回戰略地圖，
// 右下角清單的「艦隊」分頁多出一列（收據 `dec1`，`docs/re/70`）。
//
// 左下角那塊**不會被擦掉**——艦隊那幾張畫的是自己的首都，回到地圖之後
// 原版就讓它留在那裡。
func TestFleetDecideMatchesOracle(t *testing.T) {
	g := reachFleetOrg(t)
	g.Click(480, 350) // 「確認」→ 艦隊組成
	g.Click(510, 350) // 「決定」
	if g.State() != StateCommandPhase {
		t.Fatalf("點了決定之後在 %v", g.State())
	}
	matches(t, g, "dec1")
}
