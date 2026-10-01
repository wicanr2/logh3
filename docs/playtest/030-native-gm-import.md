# 030 — 原生 GM 匯入階段驗收

日期：2026-10-01。工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
規格：[069](../spec/069-native-gm-import.md)，研究：[358](../re/358-native-gm-import.md)。
完整來源、產物、工具及程式雜湊見
[`native-gm-import.json`](../internals/data/native-gm-import.json)。
JSON SHA-256：`088127df62730d0d588ba6f3f32c930434985d5e4b74256132c749e9fa42209d`。

## 已驗證結果

- 22 首原版 MDS 經原生 `asset.MDSMIDI` 後，與保存 MIDI 逐位元組相同。
  M1 簽章、完整事件、軌界、大小及空音符軌的錯誤拒絕通過。
- Linux 原生 `music-import` 真實轉完整十一首 GM，輸出在含中文／空白的路徑，
  每首 OGG 均 44100 Hz 立體聲、可完整解碼、時長為正且非靜音。
  BGM_0M 201.50 秒、END_2M 284.56 秒、OPEN_M 139.12 秒。
- 三首由舊保存 MIDI 與相同 FluidSynth／FFmpeg 命令重生；兩側經 FFmpeg 解碼，
  `cmp` 比完整 f32le PCM 全部相同。這是既有 GM 管線移植一致，不是原版 MPU 音色對拍。
- 正式 `musicPlayer.decode` 逐首完整讀出十一首產物，正常玩家路徑選曲測試通過。
  沒有音效裝置的容器只驗正式解碼／選曲，不冒稱已實際聽到或實機播放完成。
- 未明示 GM、誤選 FM、寫入原版目錄、缺工具、錯誤 SF2 都拒絕；同長度但錯 SHA-256
  另由 `TestMusicImportPreflight` 驗證。前置拒絕沒有建立假成功輸出。
- Windows amd64、macOS amd64／arm64 交叉建置通過，檔頭架構與 SHA-256 已記錄；
  不將交叉建置當成合成器依賴封裝或實機驗收。

完整標準測試 `tools/gui.sh test ./internal/... ./cmd/...` 通過，game 約 72.6 秒。
之後新增同長度錯音色庫與正式串流驗收，相關目標測試乾淨重跑通過。
Go 1.24.13；image、實際工具版本、完整每首子程序參數均保存於上述 JSON。

## 重生入口

```sh
tools/go.sh test ./internal/asset -run TestMDSMIDI -count=1 -v
tools/go.sh build -o workplace/music69-native/music-import ./cmd/music-import
bash tools/audio.sh /work/music69-native/music-import -gm-only \
  -source /orig -output '/work/music69-native/曲庫 測試' \
  -tools /usr/bin -soundfont /usr/share/sounds/sf2/FluidR3_GM.sf2
bash tools/audio.sh sh /src/tools/scripts/music_native_check.sh
tools/gui.sh test ./cmd/gin3 \
  -run 'TestNativeGMImportPlayerDecode|TestPlayerPathMusicSwitches' -count=1 -v
tools/gui.sh xrun sh -c '
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o workplace/music69-native/music-import-windows-amd64.exe ./cmd/music-import &&
  CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o workplace/music69-native/music-import-darwin-amd64 ./cmd/music-import &&
  CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o workplace/music69-native/music-import-darwin-arm64 ./cmd/music-import'
```

測試使用 Docker 是主機隔離契約；原生 `music-import` 本身不呼叫 Docker、Python、shell，
玩家最終包須提供 tools/ 與 soundfont/，不依賴玩家安裝合成器。
收據合併入口 `tools/scripts/music_native_receipt.py` 的必填版本／image／HEAD 參數
須取本輪實讀值，不把本文件舊版身份套到新工具。

## 限制與停止線

保持既有 FA BPM、raw 時間及不展開反覆的近似；沒有新增原版 driver 逆向。
部分 GM 解碼峰值 1.0，沿用舊增益，未聲稱無 clipping 或聆聽通過。
正式 runtime 的存檔、音源選擇與曲目對應未改。
FM 原版 OPL 擷取、三平台合成器／編碼器與依賴封裝、實際聆聽和 loop 接點仍未完成；
#69 保持開啟，不將本階段當完整玩家曲庫管線。
