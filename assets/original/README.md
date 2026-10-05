# 原版素材（保存用）

從《銀河英雄傳說III SP》DOS 版（亨和企業．微波軟體，1994）原版檔案解出來的素材，
進版控是為了保存遊戲資產、給日後重新繪圖當參考。規則見 `AGENTS.md`「Docker-only 與
原版素材邊界」：只放匯出腳本產生的可檢視格式，不手改；**remake 的發行包不含這個目錄**；
repo 轉公開前要先處理這個目錄（它在 git 歷史裡）。

重做：自備原版放 `org_game/extracted/`，`tools/py.sh tools/scripts/<命令>` 多半寫到
`workplace/assets/`，再複製過來（`bmg_export.py` 直接寫這裡）；然後 `tools/py.sh tools/scripts/assets_manifest.py`
重產這份說明與 `MANIFEST.tsv`（逐檔 SHA-256）。音樂那兩支是 shell 包裝器，要 dosgolem。

五類補完素材使用正式 Go 解碼器：`bash tools/go.sh run tools/scripts/remaining_assets_export.go`
`-orig /orig -out workplace/assets-remaining-<新批次> -manifest workplace/assets-remaining-<新批次>.json`。
輸出拒絕覆寫，逐張驗 PNG 重讀尺寸、色號與 palette；驗證後將各分類目錄複製到本目錄。
來源及輸出雜湊見 [保存素材收據](../../docs/internals/data/remaining-assets.json)。

| 目錄 | 檔數 | 原版檔 | 產生命令 | 內容 | 格式文件 |
|---|---:|---|---|---|---|
| `face/` | 83 | `FACE.AMG` | `amg_export.py` | 人物頭像 83 張 80×100（第 82 張是空白格） | [`09-face-amg.md`](../../docs/internals/09-face-amg.md) |
| `sprite/` | 75 | `SPRITE.SPR` | `spr_export.py 16 SPRITE.SPR` | 圖塊 75 張 16×16（艦隊組成格盤） | [`10-sprite-spr.md`](../../docs/internals/10-sprite-spr.md) |
| `frame/` | 33 | `FRAME.BLK` | `spr_export.py 20 FRAME.BLK` | 視窗外框圖塊 33 張 16×20 | [`47-window-frame-tiles.md`](../../docs/internals/47-window-frame-tiles.md) |
| `block_3/` | 128 | `BLOCK_3.BLK` | `spr_export.py 20 BLOCK_3.BLK` | 戰場圖塊 128 張 16×20 | [`35-block3-blk.md`](../../docs/internals/35-block3-blk.md) |
| `block_0/` | 128 | `BLOCK_0.BLK` | `remaining_assets_export.go` | 圖塊 128 張16×20 | [`35-block3-blk.md`](../../docs/internals/35-block3-blk.md) |
| `block_1/` | 128 | `BLOCK_1.BLK` | `remaining_assets_export.go` | 圖塊 128 張16×20 | [`35-block3-blk.md`](../../docs/internals/35-block3-blk.md) |
| `block_2/` | 128 | `BLOCK_2.BLK` | `remaining_assets_export.go` | 圖塊 128 張16×20 | [`35-block3-blk.md`](../../docs/internals/35-block3-blk.md) |
| `ani_pnl/` | 3 | `ANI_PNL0–2.A98` | `remaining_assets_export.go` | 面板三張640×400 | [`37-ani-pnl-a98.md`](../../docs/internals/37-ani-pnl-a98.md) |
| `map/` | 32 | `MAP00–31.CMP ＋ BLOCK_3.BLK` | `remaining_assets_export.go` | 戰場地圖32張2048×2040，驗圖塊索引範圍 | [`36-map-cmp.md`](../../docs/internals/36-map-cmp.md) |
| `intro/` | 374 | `IBM16.350 ＋ IBMFLS16.350 ＋ BU16IBM.RGB` | `remaining_assets_export.go` | 來源374格640×227；含玩家播放略過的349格 | [`19-ibm16-350.md`](../../docs/internals/19-ibm16-350.md) |
| `title/` | 1 | `SPTIT.350 ＋ SPTIT.RGB` | `remaining_assets_export.go` | 標題一張640×350；色號與固定原版雙跑相同 | [`111-title-image.md`](../../docs/spec/111-title-image.md) |
| `3dblock/` | 1 | `3DBLOCK.SPR` | `spr_sheet.py 3DBLOCK.SPR 48 64 4 4 row` | 等距方塊 48×64（放大 4 倍的單張） | [`21-3dblock-spr.md`](../../docs/internals/21-3dblock-spr.md) |
| `comflag/` | 6 | `COM_FLAG.SPR` | `comflag_export.py` | 指揮旗標籤五張 ＋ 總表 | [`01-assets-inventory.md`](../../docs/internals/01-assets-inventory.md) |
| `strmap/` | 1 | `STRMAP.BIT` | `strmap_export.py` | 戰略星圖底圖 | [`12-strmap-bit.md`](../../docs/internals/12-strmap-bit.md) |
| `vo5/` | 50 | `END*.VO5` | `vo5_export.py` | 結局圖 50 張 | [`18-vo5.md`](../../docs/internals/18-vo5.md) |
| `comani/` | 1438 | `COMANI.BMG` | `bmg_export.py` | 交戰演出動畫 254 段 1,438 格，每張 256×102 是畫完那一格之後的視窗；XOR 開頭的段疊在同單位 D0 第 0 格上 | [`20-comani-bmg.md`](../../docs/internals/20-comani-bmg.md) |
| `font/` | 2 | `GRAPH.IMG` | `font_extract.py；font_head.py` | 倚天中文字型總表與半形字頭 | [`06-font-graph-img.md`](../../docs/internals/06-font-graph-img.md) |
| `music/mid/` | 22 | `*.MDS` | `mds_export.py` | 22 首樂曲轉標準 MIDI | [`17-mds.md`](../../docs/internals/17-mds.md) |
| `music/vgm/` | 22 | `*_F.MDS` | `tools/music_fm_capture.sh` | ＦＭ版 11 首在 dosgolem 裡跑出來的 OPL 暫存器記錄（.meta 是擷取參數） | [`17-mds.md`](../../docs/internals/17-mds.md) |
| `music/ogg/` | 23 | `*.MDS` | `tools/music_to_ogg.sh` | `_F` 由 VGM 合成（OPL 模擬）、`_M` 由 MIDI 以 fluidsynth ＋ FluidR3_GM 合成；`SOURCES.tsv` 是 `_F` 的來源雜湊 | [`17-mds.md`](../../docs/internals/17-mds.md) |

上述PNG是解碼素材保存；不代表每張圖都已接入正式玩家路徑。
