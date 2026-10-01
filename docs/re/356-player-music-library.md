# 356 — 玩家自產曲庫的音色庫來源與授權

日期：2026-10-01。工作項目：[Issue #69](https://github.com/wicanr2/logh3/issues/69)。
本輪只建立來源與授權前置證據，不宣稱原生轉檔器或三平台封包已完成。
重生查詢入口：`bash tools/audio.sh <命令>`；映像檔、版本及雜湊均記於
[`player-music-library.json`](../internals/data/player-music-library.json)。

## 已證實的音色庫

既有 `logh3-audio:2` 實際安裝 `fluid-soundfont-gm 3.1-5.3`。
`/usr/share/sounds/sf2/FluidR3_GM.sf2` 為 148,398,306 bytes（約 142 MiB），
SHA-256 `74594e8f4250680adf590507a306655a299935343583256f3b722c48a1bc1cb0`。
這是目前 GM 轉檔使用的來源，不是遊戲的原版素材。

隨附 `copyright` SHA-256 為
`f2f5e6358fc7c1fe62aafc6c8229c4c5898e5afb0d4ce1322b15362a8cbd600e`，
完整檔案保存於 [`fluidr3-gm-copyright.txt`](../internals/data/fluidr3-gm-copyright.txt)。
音色庫本體採 MIT：可隨包散布，但必須保留著作權與授權聲明；貢獻者清單一併保留。
Debian 包裝另列 GPL-2-or-later，不將它與音色庫本體的 MIT 混稱。
外部交叉核對：[同版本上游 README](https://sources.debian.org/src/fluid-soundfont/3.1-5.3/README/)、
[COPYING](https://sources.debian.org/src/fluid-soundfont/3.1-5.3/COPYING/)。
不採用其他網站對衍生音色庫的 CC BY 標示來覆蓋這份實物的授權。

查詢時發現 image 只保留 `copyright`；dpkg 檔案清單仍列出 README，實際檔案已刪減。
猜測壓縮檔名也不存在。這兩次失敗是工具文件覆蓋問題，並非授權不明或產品缺陷；
改以實際保留的完整授權與同版本 Debian 上游文件交叉核對，沒有新建重複 image。

## FM 基準

重新計算 `assets/original/music/ogg/SOURCES.tsv` 的 11 首 `_F`：
原始 `.MDS`、VGM、OGG 三項均與記錄相符。TSV SHA-256 為
`9bf596e9ba29a6f996d88271e2a8a72827110a37277d1671045be19701044c5f`。
逐首雜湊、既有 dosgolem 版本及增益列於上述 JSON；沒有冒稱本輪重新跑過原版。
原生管線仍須由玩家自備原版產生同一基準，不能將原版曲庫內嵌公開封包。

## 固定來源取得

Debian `fluid-soundfont_3.1-5.3.dsc` 提供 `fluid-soundfont_3.1.orig.tar.gz` 的
SHA-256 `2621acaa1c78e4abdb24bdd163230cc577e61276936d6aa6e3180582142f0343`，
實際下載 134,835,922 bytes 後核對相同。archive 裡的 SF2 與既有 image 相同；
README／COPYING 的原始成員雜湊也記進上述 JSON。

建置準備入口：`LOGH3_NET=1 tools/py.sh tools/scripts/soundfont_prepare.py --download`。
已下載後可離線重跑 `tools/py.sh tools/scripts/soundfont_prepare.py`，
只提取已鎖定的 SF2、README 與 COPYING，逐一驗雜湊；錯誤快取明確拒絕。
空快取下載、無網路離線提取與錯誤快取拒絕三項均已實跑通過；
下載檢查的重複大檔於核對後清除，只保留上述來源快取。
輸入 archive 與暫存輸出在 `workplace/music69-source/`，不進 Git，
也不把這個暫存目錄當正式交付根目錄。這是建置工具，玩家免 Docker 轉檔器仍待實作。

## 已採用決定與未完成範圍

2026-09-19 的「GM 音色庫隨發行包、玩家不安裝 fluidsynth」決定仍有效。
2026-10-01 使用者選定固定版本下載與 SHA-256 核對，排除 Git LFS；發行包內含音色庫，
玩家不另行下載。Git 保存來源、雜湊與完整授權，建置可用網路或已驗證快取。

原生轉檔器、FM 重生、合成器／編碼器的跨平台封裝、實際聆聽及循環接點驗收
仍未完成。現有 Debian fluidsynth 連結多項系統函式庫，不能只複製一個 executable
便宣稱玩家免安裝；亦不把 OGG 可解碼視為人耳驗收。
