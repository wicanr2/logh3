# 031 — 完整原生曲庫匯入驗收

日期：2026-10-01。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
規格：[071](../spec/071-native-fm-import.md)，研究：[359](../re/359-native-fm-import.md)。
狀態：完整原生匯入與正式解碼驗收通過；不外推平台依賴封裝或人耳驗收。
完整來源、每曲 flags、工具、輸出雜湊及重跑結果見
[`native-fm-import.json`](../internals/data/native-fm-import.json)。
JSON SHA-256：`3283ef4f5305fef7c36a772f7c6b1480b8c00aa1e1b2e6e823c787a817ab2fa1`。

## 已驗結果

- 預設完整原生入口產生十一首 GM、十一首 FM。全部 FM 的 VGM／OGG 全檔與
  `SOURCES.tsv` 一致；沒有裁切、遮罩、重新挑種子或複製保存曲庫作為輸出。
- BGM_0F、END_2F、OPEN_F 三種觸發路徑第二次擷取／合成，VGM 與 OGG 均逐位元組一致。
  BGM_0F 127.13 秒、END_2F 136.18 秒、OPEN_F 101.27 秒。
- 正式 `musicPlayer.decode` 完整讀出二十二首 PCM，約 13.0 秒；正常玩家路徑
  選曲切換通過。44100 Hz、立體聲、正時長及非靜音通過，未宣稱實際聽到。
- 六種 CLI 前置拒絕通過，均未建立輸出；錯原版／probe 不會執行替代 DOS 程式。
  070 的真實封裝還原與壞 CRC／截斷／頁序／串接拒絕測試通過。
- 標準 `tools/gui.sh test ./internal/... ./cmd/...` 全綠，game 約 48.1 秒。
  新增完整曲庫驗收測試於全批結束後另乾淨重跑通過。
- Windows amd64、macOS amd64／arm64 匯入器交叉建置及檔頭核對通過，
  僅 Linux 原生入口在本輪實際執行。原版 bundle／已鎖定來源雜湊保持不變。
- 輸出 UID/GID 1000:1000；原版暫存副本清除，沒有 root-owned 或誤建 `.md` 目錄。
  按兩個工作樹掛載核對無本專案執行中／已停止容器，其他專案未動。

## 重生入口

本專案所有驗證工作透過 Docker 包裝器隔離；`music-import` 本身不呼叫
Docker、shell、Python 或玩家 PATH。原版只讀，執行器使用專用工作副本。

```sh
# 在 /home/anr2/cht/dosgolem-logh3 內，先確認 clean HEAD 符合規格 071。
tools/go.sh build -o workplace/music69-probe ./cmd/probe
# 回到 /home/anr2/cht/logh3；在容器內複製探針及建立診斷工具連結。
LOGH3_EXTRA_RO=/home/anr2/cht/dosgolem-logh3 tools/py.sh -c '
from pathlib import Path
import os, shutil
p=Path("/work/workplace/music69-fm/tools")
p.mkdir(parents=True, exist_ok=True)
assert p.stat().st_uid == os.getuid()
src=Path("/ref/workplace/music69-probe")
assert src.is_file()
shutil.copyfile(src, p/"dosgolem-probe")
(p/"dosgolem-probe").chmod(0o755)
for name, target in [("ffmpeg", "/usr/bin/ffmpeg"), ("fluidsynth", "/usr/bin/fluidsynth"), ("opl2wav", "/usr/local/bin/opl2wav")]:
    dest=p/name
    if dest.is_symlink(): dest.unlink()
    assert not dest.exists()
    dest.symlink_to(target)'
tools/go.sh build -o workplace/music69-fm/music-import ./cmd/music-import
LOGH3_TIMEOUT=90m bash tools/audio.sh /work/music69-fm/music-import \
  -source /orig -output /work/music69-fm/full -tools /work/music69-fm/tools \
  -soundfont /usr/share/sounds/sf2/FluidR3_GM.sf2
LOGH3_TIMEOUT=20m bash tools/audio.sh /work/music69-fm/music-import \
  -fm-only -tracks BGM_0F,END_2F,OPEN_F \
  -source /orig -output /work/music69-fm/repeat -tools /work/music69-fm/tools \
  -soundfont /usr/share/sounds/sf2/FluidR3_GM.sf2
tools/gui.sh test ./cmd/gin3 \
  -run 'TestNativeFullMusicImportPlayerDecode|TestPlayerPathMusicSwitches' -count=1 -v \
  > workplace/music69-fm/player-test.log 2>&1
bash tools/audio.sh sh /src/tools/scripts/music_fm_check.sh
LOGH3_EXTRA_RO=/home/anr2/cht/logh3 tools/py.sh \
  /work/tools/scripts/music_fm_receipt.py
```

完整旗標、程式與原始來源雜湊由當輪收據保留，不能重用前輪身份。
合併前須保存當輪 `full-runmeta.json`，包含 Git HEAD／工作樹狀態、匯入器雜湊、
四份正式來源雜湊、實讀 image 身份、probe HEAD 與完整命令；不可從歷史文件猜填。
`tools/scripts/music_fm_receipt.py` 只接受完成的全批及三路徑重跑，
且核對當輪正式播放器解碼測試日誌；沒有本機原版時測試略過，不製造替代音樂。
`tools/scripts/music_fm_check.sh` 驗互斥旗標、錯曲名、缺工具、來源內輸出、
錯執行器身份與缺原版檔案的前置拒絕；錯誤案例不執行替代 DOS 程式。

## 限制

研究批次的工具連結指向既有 image，只用於本機驗證；尚不是玩家平台工具封包。
原版 cue／OPL 暫存器串與完整 VGM／OGG 雜湊可驗，合成硬體時序保持
hardware-spec approximation。GM 沿用規格 069 近似。
三平台第三方依賴封裝、實際聆聽與 loop 接點仍待完成，#69 保持開啟。
