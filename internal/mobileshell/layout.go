package mobileshell

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/logh3/internal/gfx"
)

// 行動版版面（規格161 M1）。座標分三種：
//
//   - 原版座標：遊戲的 640×480 畫面。
//   - 邏輯座標：高固定 480，寬依螢幕比例（20:9 約 1066）。版面與按鈕都用它。
//   - 畫布座標：實際像素。畫布＝邏輯 × u，u＝畫布高／480。
//
// 每個區塊從原版畫面裁出、只縮放一次（先整數倍最近鄰、再線性補剩下的倍率），
// 原版點陣字在非整數倍下仍然銳利。
//
// M1 只在「純戰略地圖」時重排（PlainStrategic：地圖上沒有任何視窗疊著，裁出來的
// 地圖是乾淨的）。其他狀態（報告、清單、計劃、下拉選單、戰場）照原版整張置中，
// 原版視窗本來就在畫面中央附近。繪製分層與裁切個別視窗在 M2。

// 原版畫面上的區塊（以 screen-command-phase 實際畫面的色號掃描定界，第176輪）。
var (
	srcFull  = image.Rect(0, 0, gfx.Width, gfx.Height)
	srcMap   = image.Rect(0, 40, 320, 340)    // 銀河地圖 320×300
	srcStrip = image.Rect(320, 40, 640, 112)  // 日期、軍事費、總稅收
	srcCard  = image.Rect(320, 146, 640, 420) // 星域摘要＋四分頁＋清單
)

const (
	sidebarW = 88  // 左側欄寬（邏輯）
	slotGap  = 8   // 地圖與卡片槽、卡片之間的間距
	minSlot  = 280 // 卡片槽至少要這麼寬（卡片至少 0.875 倍），不夠就退回整張置中；16:9 會退回
)

// region 是一塊從原版畫面裁出、畫到邏輯座標 (x,y)、放大 scale 倍的區塊。
type region struct {
	src   image.Rectangle
	x, y  float64
	scale float64
}

// screenLayout 是這一格的版面。
type screenLayout struct {
	lw        int // 邏輯寬
	strategic bool
	regions   []region
}

// computeLayout 依邏輯寬與是否為純戰略地圖排版。
func computeLayout(lw int, plainStrategic bool) screenLayout {
	full := screenLayout{lw: lw, regions: []region{{src: srcFull, x: float64(gameOrigin(lw)), scale: 1}}}
	if !plainStrategic {
		return full
	}
	mapScale := float64(gfx.Height) / float64(srcMap.Dy()) // 1.6：地圖滿高
	mapX := float64(sidebarW)
	slotX := mapX + float64(srcMap.Dx())*mapScale + slotGap
	slotW := float64(lw) - slotX - slotGap
	if slotW < minSlot {
		return full
	}
	h := float64(srcStrip.Dy() + srcCard.Dy())
	s := slotW / float64(srcCard.Dx())
	if hs := (float64(gfx.Height) - slotGap) / h; hs < s {
		s = hs
	}
	top := (float64(gfx.Height) - h*s - slotGap) / 2
	x := slotX + (slotW-float64(srcCard.Dx())*s)/2
	return screenLayout{lw: lw, strategic: true, regions: []region{
		{src: srcMap, x: mapX, y: 0, scale: mapScale},
		{src: srcStrip, x: x, y: top, scale: s},
		{src: srcCard, x: x, y: top + float64(srcStrip.Dy())*s + slotGap, scale: s},
	}}
}

// toGame 把邏輯座標換成原版座標；不在任何區塊上回 ok=false。
func (l screenLayout) toGame(x, y float64) (int, int, bool) {
	for _, r := range l.regions {
		w, h := float64(r.src.Dx())*r.scale, float64(r.src.Dy())*r.scale
		if x >= r.x && x < r.x+w && y >= r.y && y < r.y+h {
			return r.src.Min.X + int((x-r.x)/r.scale), r.src.Min.Y + int((y-r.y)/r.scale), true
		}
	}
	return 0, 0, false
}

// gameOrigin 回整張原版畫面置中時的左緣（邏輯座標）。
func gameOrigin(logicalW int) int { return (logicalW - gfx.Width) / 2 }

// sharp 用「先整數倍最近鄰、再線性」把 src 的一塊畫到 dst（畫布座標）。
// 中間圖依尺寸快取。
type sharp struct{ tmp map[image.Point]*ebiten.Image }

func (sh *sharp) draw(dst, src *ebiten.Image, sr image.Rectangle, x, y, scale float64) {
	k := int(scale)
	if k < 1 {
		k = 1
	}
	size := image.Pt(sr.Dx()*k, sr.Dy()*k)
	if sh.tmp == nil {
		sh.tmp = map[image.Point]*ebiten.Image{}
	}
	tmp := sh.tmp[size]
	if tmp == nil {
		tmp = ebiten.NewImage(size.X, size.Y)
		sh.tmp[size] = tmp
	}
	tmp.Clear()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(k), float64(k))
	tmp.DrawImage(src.SubImage(sr).(*ebiten.Image), op)
	op = &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(scale/float64(k), scale/float64(k))
	op.GeoM.Translate(x, y)
	dst.DrawImage(tmp, op)
}
