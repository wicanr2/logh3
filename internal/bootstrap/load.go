// Package bootstrap 從玩家自備的原版檔案與規格 JSON 組出一份遊戲資源。
//
// 兩個入口共用：`cmd/play`（無頭、把每一步倒成 PNG）與 `cmd/gin3`
// （Ebiten 視窗）。**remake 不內嵌原版素材**，路徑一律由呼叫端給。
package bootstrap

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/font"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/screen"
	"github.com/wicanr2/logh3/internal/ui"
)

// Load 讀原版檔案與規格，組出遊戲資源。
func Load(orig, data string) (*game.Resources, error) {
	fnt, err := font.Load(filepath.Join(orig, "GRAPH.IMG"))
	if err != nil {
		return nil, err
	}
	faceData, err := os.ReadFile(filepath.Join(orig, "FACE.AMG"))
	if err != nil {
		return nil, err
	}
	faces, err := asset.OpenAMG(faceData)
	if err != nil {
		return nil, err
	}
	admiral, err := os.ReadFile(filepath.Join(orig, "ADMIRAL.DTX"))
	if err != nil {
		return nil, err
	}
	menu, err := screen.LoadMenuSpec(filepath.Join(data, "screen-main-menu.json"))
	if err != nil {
		return nil, err
	}
	cp, err := screen.LoadCopyProtSpec(filepath.Join(data, "screen-copy-protection.json"))
	if err != nil {
		return nil, err
	}
	mg, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-main-game.json"))
	if err != nil {
		return nil, err
	}
	so, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-ship-order.json"))
	if err != nil {
		return nil, err
	}
	bp, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-build-plan.json"))
	if err != nil {
		return nil, err
	}
	bg, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-budget-plan.json"))
	if err != nil {
		return nil, err
	}
	al, err := screen.LoadMainGameSpec(filepath.Join(data, "screen-admiral-list.json"))
	if err != nil {
		return nil, err
	}
	cur, err := ui.LoadCursor(filepath.Join(data, "cursor.json"))
	if err != nil {
		return nil, err
	}
	bit, err := os.ReadFile(filepath.Join(orig, "STRMAP.BIT"))
	if err != nil {
		return nil, err
	}
	rgn, err := os.ReadFile(filepath.Join(orig, "STRMAP.RGN"))
	if err != nil {
		return nil, err
	}
	colours, err := regionColours(filepath.Join(data, "strmap-regions.json"))
	if err != nil {
		return nil, err
	}
	star, err := asset.BuildStarMap(bit, rgn, colours, 16, 20, 20)
	if err != nil {
		return nil, err
	}
	regions, err := asset.OpenRGN(rgn, 16, 20)
	if err != nil {
		return nil, err
	}
	screens, maps, err := game.LoadCommandScreens(
		data, filepath.Join(data, "strmap-regions.json"), bit, rgn)
	if err != nil {
		return nil, err
	}
	// 左下角的星域摘要：版面 ＋ 執行期星域記錄 ＋ 行星表（規格 `docs/spec/001`）。
	panel, err := screen.LoadRegionPanel(filepath.Join(data, "region-panel.json"))
	if err != nil {
		return nil, err
	}
	capital, err := screen.LoadCapitalPanel(filepath.Join(data, "capital-panel.json"))
	if err != nil {
		return nil, err
	}
	budget, err := screen.LoadBudgetPanel(filepath.Join(data, "budget-panel.json"))
	if err != nil {
		return nil, err
	}
	report, err := screen.LoadReportPanel(filepath.Join(data, "report-panel.json"))
	if err != nil {
		return nil, err
	}
	bioPanel, err := screen.LoadBioPanel(filepath.Join(data, "admiral-bio-panel.json"))
	if err != nil {
		return nil, err
	}
	bios, err := loadBiographies(filepath.Join(data, "keireki.json"))
	if err != nil {
		return nil, err
	}
	flagships, err := screen.LoadFlagships(filepath.Join(data, "flagships.json"))
	if err != nil {
		return nil, err
	}
	admPanel, err := screen.LoadAdmiralPanel(filepath.Join(data, "admiral-detail-panel.json"))
	if err != nil {
		return nil, err
	}
	fleetPanel, err := screen.LoadFleetPanel(filepath.Join(data, "fleet-org-panel.json"))
	if err != nil {
		return nil, err
	}
	sprData, err := os.ReadFile(filepath.Join(orig, "SPRITE.SPR"))
	if err != nil {
		return nil, err
	}
	sprites, err := asset.DecodeTiles(sprData, 16, 16)
	if err != nil {
		return nil, err
	}
	composition, err := screen.LoadCompositionPanel(filepath.Join(data, "fleet-composition-panel.json"))
	if err != nil {
		return nil, err
	}
	buildPanel, err := screen.LoadReportPanel(filepath.Join(data, "build-plan-panel.json"))
	if err != nil {
		return nil, err
	}
	delivery, err := screen.LoadDeliveryPanel(filepath.Join(data, "delivery-panel.json"))
	if err != nil {
		return nil, err
	}
	production, err := screen.LoadProductionPanel(filepath.Join(data, "production-panel.json"))
	if err != nil {
		return nil, err
	}
	stellarFile, err := os.ReadFile(filepath.Join(orig, "STELLAR.DTA"))
	if err != nil {
		return nil, err
	}
	stellar, err := game.LoadRegionRecords(stellarFile,
		filepath.Join(data, "month-rollover-796-01.json"),
		filepath.Join(data, "region-ships-796-01.json"))
	if err != nil {
		return nil, err
	}
	pladat, err := os.ReadFile(filepath.Join(orig, "PLADAT.DAT"))
	if err != nil {
		return nil, err
	}
	planets, err := rules.LoadPlanets(pladat)
	if err != nil {
		return nil, err
	}
	growth, err := rules.LoadGrowthTable(filepath.Join(data, "growth-table.json"))
	if err != nil {
		return nil, err
	}
	work, err := rules.LoadWorkTemplate(filepath.Join(data, "work-796-01.json"))
	if err != nil {
		return nil, err
	}
	slotList, err := screen.LoadSlotList(filepath.Join(data, "save-slot-list.json"))
	if err != nil {
		return nil, err
	}
	var saveIndex []rules.SaveIndexEntry
	if b, err := os.ReadFile(filepath.Join(orig, "SAVEDATA.IDX")); err == nil {
		if saveIndex, err = rules.ParseSaveIndex(b); err != nil {
			return nil, err
		}
	}
	graph, err := rules.LoadStarGraph(filepath.Join(data, "star-graph.json"))
	if err != nil {
		return nil, err
	}
	names, err := loadAdmiralNames(filepath.Join(data, "keireki.json"))
	if err != nil {
		return nil, err
	}
	starOrder, err := loadStarListOrder(filepath.Join(data, "star-list-order.json"))
	if err != nil {
		return nil, err
	}
	return &game.Resources{Font: fnt, OrigDir: orig, Graph: graph, AdmiralNames: names, StarListOrder: starOrder,
		UnrestParams: rules.UnrestParams{
			Ruler: [2]int{16, 60}, Blend: [2]int{1, 2}, Rate: [2]int{2, 2},
			Table: []int{170, 150, 130, 90, 60},
			Base:  []int{50, 75, 100, 150, 200},
		}, Faces: faces, Cursor: cur, Menu: menu,
		CopyProt: cp, MainGame: mg, StarMap: star, Admiral: admiral,
		StarBit: bit, StarRGN: rgn, RegionColours: colours,
		ShipOrder: so, BuildPlan: bp, BudgetPlan: bg, AdmiralList: al,
		Screens: screens, StarMaps: maps, Regions: regions,
		RegionPanel: panel, CapitalPanel: capital, BudgetPanel: budget, ProductionPanel: production,
		DeliveryPanel: delivery, ReportPanel: report, BuildPlanPanel: buildPanel, AdmiralPanel: admPanel, FleetPanel: fleetPanel, CompositionPanel: composition, Sprites: sprites, Flagships: flagships, BioPanel: bioPanel, Biographies: bios,
		Stellar: stellar, Planets: planets, Growth: growth,
		SlotList: slotList, SaveIndex: saveIndex, WorkTemplate: work,
		StartMilitary: 50000, ShipPct: 25}, nil
}

func regionColours(path string) ([]uint8, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Regions []struct {
			Colour uint8 `json:"colour_796_01"`
		} `json:"regions"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make([]uint8, len(doc.Regions))
	for i, r := range doc.Regions {
		out[i] = r.Colour
	}
	return out, nil
}


// loadAdmiralNames 讀 82 名提督姓名的 Big5 字碼。
func loadAdmiralNames(path string) ([][]uint16, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Records []struct {
			Index     int      `json:"index"`
			NameCodes []uint16 `json:"name_codes"`
		} `json:"records"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make([][]uint16, len(doc.Records))
	for _, r := range doc.Records {
		if r.Index >= 0 && r.Index < len(out) {
			out[r.Index] = r.NameCodes
		}
	}
	return out, nil
}

// loadStarListOrder 讀「星系資料」那張清單的列順序，兩個陣營各一份。
func loadStarListOrder(path string) ([][]int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Orders [][]int `json:"orders"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	return doc.Orders, nil
}


// loadBiographies 讀 82 篇人物傳的字碼（`docs/internals/15`）。
func loadBiographies(path string) ([][]uint16, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Records []struct {
			Index int      `json:"index"`
			Codes []uint16 `json:"text_codes"`
		} `json:"records"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	out := make([][]uint16, 0, len(doc.Records))
	for _, r := range doc.Records {
		for len(out) <= r.Index {
			out = append(out, nil)
		}
		out[r.Index] = r.Codes
	}
	return out, nil
}
