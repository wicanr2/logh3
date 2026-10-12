package game

import (
	"image"

	"github.com/wicanr2/logh3/internal/gfx"
	"github.com/wicanr2/logh3/internal/rules"
	"github.com/wicanr2/logh3/internal/ui"
)

// 規格134：矩形同時驅動繪製與命中；這裡是640×480畫布座標。
var enemyVisibilityButton = image.Rect(176, 360, 304, 380)

// EnemyVisibilityButtonBounds 回正式控制的畫布矩形，供版面驗證共用。
func EnemyVisibilityButtonBounds() image.Rectangle { return enemyVisibilityButton }

func enemyVisibilityButtonContains(x, y int) bool {
	return image.Pt(x, y).In(enemyVisibilityButton)
}

// ShowEnemies 回執行期顯示偏好；不屬於原版索敵狀態。
func (g *Game) ShowEnemies() bool { return g.showEnemies }

// SetShowEnemies 只改顯示偏好及游標摘要，不回寫原始六表或規則資料。
func (g *Game) SetShowEnemies(on bool) {
	if g.showEnemies == on {
		return
	}
	g.showEnemies = on
	g.battleHoverUnit, g.battleHoverOn = -1, false
	g.battleHoverMinimapLine = false
	g.refreshBattleHover()
}

// CanToggleEnemyVisibility 保留啟動畫面的限制；玩家入口在設定選單。
func (g *Game) CanToggleEnemyVisibility() bool {
	switch g.state {
	case StateIntro, StateTitle, StateMainMenu, StateCopyProtection:
		return false
	}
	return true
}

// ToggleEnemyVisibility 供遊戲內設定鈕使用（規格139）。
func (g *Game) ToggleEnemyVisibility() {
	if g.CanToggleEnemyVisibility() {
		g.SetShowEnemies(!g.showEnemies)
	}
}

// battleDisplayUnits 只給renderer副本。AI、移動、交戰與存檔仍讀battleUnits。
func (g *Game) battleDisplayUnits() ([]rules.BattleUnit, map[int]int) {
	if !g.showEnemies {
		return g.battleUnits, g.battleSpotting.directions
	}
	units := append([]rules.BattleUnit(nil), g.battleUnits...)
	for i := range units {
		u := &units[i]
		if u.Out || u.Faction == g.faction {
			continue
		}
		u.Detail = rules.SpottingMax
		// 清除副本中的舊位置快取，強制取當前座標及陣營色。
		u.MarkerFlags = 0
		rules.UpdateBattleMarker(u, true, true)
	}
	directions := make(map[int]int, len(g.battleSpotting.directions)+len(g.battleFleets))
	for i, d := range g.battleSpotting.directions {
		directions[i] = d
	}
	for _, f := range g.battleFleets {
		for _, u := range units {
			if u.Fleet == f.Index && u.Squadron == 0 && !u.Out && u.Faction != g.faction {
				directions[f.Index] = rules.DirectionIndex(f.Facing)
				break
			}
		}
	}
	return units, directions
}

func (g *Game) displayBattleSprite(u rules.BattleUnit) int {
	if g.showEnemies && u.Faction != g.faction && !u.Out {
		return rules.BattleCachedSprite(u, true, rules.BlankSprite)
	}
	return g.cachedBattleSprite(u)
}

// 「顯示敵軍」Big5，沿原字格及高解析文字命令。
var enemyVisibilityLabel = []uint16{0xC5E3, 0xA5DC, 0xBCC4, 0xAD78}

func (g *Game) drawEnemyLabel(fb *gfx.Framebuffer, r image.Rectangle, suffix string, ink uint8) {
	if g.res.Font == nil {
		return
	}
	x, y := r.Min.X+4, r.Min.Y+2
	for _, code := range enemyVisibilityLabel {
		ui.DrawGlyphClipped(fb, g.res.Font, x, y, code, ink, 3, -1, false, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
		x += 16
	}
	for _, code := range []byte(suffix) {
		ui.DrawHalfGlyphClipped(fb, g.res.Font, x, y, code, ink, 3, -1, false, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
		x += 8
	}
}

func (g *Game) drawEnemyVisibilitySetting(fb *gfx.Framebuffer) {
	r := enemyVisibilityButton
	fill, label := uint8(settingsPlainFill), " OFF"
	if g.showEnemies {
		fill, label = settingsPickedFill, " ON"
	}
	(ui.Frame{Role: "ui/button", Layers: []ui.Layer{{Top: 7, Right: 7, Bottom: 3, Left: 3}}, Fill: fill}).Draw(fb, r.Min.X, r.Min.Y, r.Dx(), r.Dy())
	g.drawEnemyLabel(fb, r.Inset(1), label, 7)
}

func (g *Game) drawEnemyVisibilityNotice(fb *gfx.Framebuffer) {
	if !g.showEnemies {
		return
	}
	r := image.Rect(8, 420, 168, 440)
	fb.SurfaceRect(r.Min.X, r.Min.Y, r.Dx(), r.Dy(), 9)
	g.drawEnemyLabel(fb, r, " ON", 7)
}
