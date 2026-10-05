package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/logh3/internal/asset"
	"github.com/wicanr2/logh3/internal/game"
	"github.com/wicanr2/logh3/internal/screen"
)

// LoadIntro 把開頭動畫接上一份已經組好的資源（`docs/spec/020`）。
//
// **`Load` 不會自己接。** 開頭動畫排在主選單前面，接上去整條輸入序列
// 就往後位移一格畫面——對拍收據與 README 的截圖腳本都是照沒有動畫的
// 路徑寫的。要動畫的入口自己呼叫這一支。
//
// 三個檔缺任何一個就當作沒有動畫（回 nil，不是錯）：玩家自備的原版
// 目錄不保證完整，少一個檔不該讓遊戲開不起來。
func LoadIntro(res *game.Resources, orig string) error {
	data, err := os.ReadFile(filepath.Join(orig, "IBM16.350"))
	if err != nil {
		return nil
	}
	lengths, err := os.ReadFile(filepath.Join(orig, "IBMFLS16.350"))
	if err != nil {
		return nil
	}
	rgb, err := os.ReadFile(filepath.Join(orig, "BU16IBM.RGB"))
	if err != nil {
		return nil
	}
	dec, err := asset.NewIntroPlaybackDecoder(data, lengths)
	if err != nil {
		return err
	}
	pal, err := screen.IntroPalette(rgb)
	if err != nil {
		return err
	}
	res.Intro, res.IntroPal = dec, pal
	if tailRGB, err := os.ReadFile(filepath.Join(orig, "TITSPT.RGB")); err == nil {
		tailPal, err := screen.IntroPalette(tailRGB)
		if err != nil {
			return fmt.Errorf("TITSPT.RGB: %w", err)
		}
		res.IntroTailPal = &tailPal
	} else if !os.IsNotExist(err) {
		return err
	}
	res.IntroRestart = func() (asset.IntroPlayer, error) {
		return asset.NewIntroPlaybackDecoder(data, lengths)
	}
	// 112：標題缺檔仍沿020進主選單；完整但損壞不能當作缺檔。
	titleData, err := os.ReadFile(filepath.Join(orig, "SPTIT.350"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	titleRGB, err := os.ReadFile(filepath.Join(orig, "SPTIT.RGB"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	title, err := asset.DecodeTitle350(titleData)
	if err != nil {
		return err
	}
	titlePal, err := screen.IntroPalette(titleRGB)
	if err != nil {
		return fmt.Errorf("SPTIT.RGB: %w", err)
	}
	res.Title, res.TitlePal = title, titlePal
	return nil
}
