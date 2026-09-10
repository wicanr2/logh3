// Package audio 是配樂的播放層：OGG Vorbis，循環播放。
//
// # 為什麼是 OGG 而不是直接播 `.MDS`
//
// 原版的樂曲是 `.MDS`（`docs/internals/17`）——固定 4 bytes 一個事件的
// 序列，交給 FM 音源或 MPU-401。要在 remake 裡「照原樣播」等於要重做
// 一整個 OPL 合成器，而**那不是這個專案要解的問題**：畫面對拍要的是
// 逐位元組相同，配樂要的是聽起來是同一首。所以走「離線轉成 OGG、
// 執行期只做解碼與循環」這條路，合成的部分留在轉檔工具裡。
//
// # 為什麼曲子不進版控
//
// OGG 是從原版素材轉出來的，**它就是原版素材**（AGENTS.md 的硬規則）。
// 所以：曲子放在 gitignore 的目錄，玩家自備原版之後自己轉；
// 檔案不在的時候**不播、而且要說出來**——安靜地當作沒事會讓
// 「還沒做完」看起來像做完了。
package audio

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SampleRate 是輸出的取樣率。轉檔工具照這個值產 OGG，播放端就不必重取樣
// ——重取樣要在每一個 buffer 上做，而它會把 CPU 花在聽不出差別的地方。
// **轉檔與播放共用這一個常數**，對不上時播放端要出聲抗議，不要默默重取樣。
const SampleRate = 44100


// Library 是配樂目錄的檢視：哪一首找得到、哪幾首缺。
//
// **它不解碼、也不需要音效裝置**，所以「曲目對應」這件事可以單獨測。
// 把它跟 Player 綁在一起的話，沒有音效裝置的環境（CI、容器）就一行都測不了。
type Library struct {
	dir    string
	tracks map[string]string // 曲名 → 檔案路徑
	want   []string          // 期望有的曲名（原版有幾首就幾首）
}

// Tracks 是原版的 22 首曲子，名字取自 `.MDS` 的檔名去掉副檔名
// （`docs/internals/data/mds/index.json`）。
//
// `_F` 與 `_M` 是同一首的兩個版本（FM 音源 / MIDI），原版用 `MUSIC.INI`
// 選一組；兩邊都列進來，缺哪一組由 Missing 說。
var Tracks = []string{
	"BGM_0F", "BGM_0M", "BGM_1F", "BGM_1M", "BGM_2F", "BGM_2M",
	"BGM_3F", "BGM_3M", "BGM_4F", "BGM_4M", "BGM_5F", "BGM_5M",
	"BGM_6F", "BGM_6M", "BGM_7F", "BGM_7M",
	"END_1F", "END_1M", "END_2F", "END_2M",
	"OPEN_F", "OPEN_M",
}

// NewLibrary 掃一次 dir，記下找得到哪幾首。dir 不存在不是錯誤——
// 那只是「玩家還沒轉檔」，Missing 會把整份清單交出來。
func NewLibrary(dir string) *Library {
	l := &Library{dir: dir, tracks: map[string]string{}, want: Tracks}
	for _, name := range l.want {
		p := filepath.Join(dir, name+".ogg")
		if st, err := os.Stat(p); err == nil && !st.IsDir() && st.Size() > 0 {
			l.tracks[name] = p
		}
	}
	return l
}

// Path 回這一首的檔案路徑。第二個回傳值是「找不找得到」——
// **呼叫端要看它**，不要拿空字串去開檔然後把錯誤吞掉。
func (l *Library) Path(track string) (string, bool) {
	p, ok := l.tracks[strings.ToUpper(track)]
	return p, ok
}

// Missing 是期望有、但目錄裡沒有的曲名（排序過，好比對）。
//
// ⚠ **這個清單要被看見。** 配樂缺檔的症狀是「安靜」，而安靜與
// 「播了但沒聲音」「根本沒接播放層」長得一模一樣。
func (l *Library) Missing() []string {
	var out []string
	for _, name := range l.want {
		if _, ok := l.tracks[name]; !ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// Have 是找得到的首數。
func (l *Library) Have() int { return len(l.tracks) }

// Dir 是這份檢視掃的目錄（報訊息時要講清楚是哪裡沒有）。
func (l *Library) Dir() string { return l.dir }
