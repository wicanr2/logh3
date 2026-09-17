# 原版素材（保存用）

從《銀河英雄傳說III SP》DOS 版（亨和企業．微波軟體，1994）原版檔案解出來的素材，
進版控是為了保存遊戲資產、給日後重新繪圖當參考。規則見 `AGENTS.md`「Docker-only 與
原版素材邊界」：只放匯出腳本產生的可檢視格式，不手改；**remake 的發行包不含這個目錄**；
repo 轉公開前要先處理這個目錄（它在 git 歷史裡）。

重做：自備原版放 `org_game/extracted/`，`tools/py.sh tools/scripts/<命令>` 多半寫到
`workplace/assets/`，再複製過來（`bmg_export.py` 直接寫這裡）；然後 `tools/py.sh tools/scripts/assets_manifest.py`
重產這份說明與 `MANIFEST.tsv`（逐檔 SHA-256）。音樂那兩支是 shell 包裝器，要 dosgolem。

| 目錄 | 檔數 | 原版檔 | 產生命令 | 內容 | 格式文件 |
|---|---:|---|---|---|---|
| `face/` | 83 | `FACE.AMG` | `amg_export.py` | 人物頭像 83 張 80×100（第 82 張是空白格） | [`09-face-amg.md`](../../docs/internals/09-face-amg.md) |
| `sprite/` | 75 | `SPRITE.SPR` | `spr_export.py 16 SPRITE.SPR` | 圖塊 75 張 16×16（艦隊組成格盤） | [`10-sprite-spr.md`](../../docs/internals/10-sprite-spr.md) |
| `frame/` | 33 | `FRAME.BLK` | `spr_export.py 20 FRAME.BLK` | 視窗外框圖塊 33 張 16×20 | [`47-window-frame-tiles.md`](../../docs/internals/47-window-frame-tiles.md) |
| `block_3/` | 128 | `BLOCK_3.BLK` | `spr_export.py 20 BLOCK_3.BLK` | 戰場圖塊 128 張 16×20 | [`35-block3-blk.md`](../../docs/internals/35-block3-blk.md) |
| `3dblock/` | 1 | `3DBLOCK.SPR` | `spr_sheet.py 3DBLOCK.SPR 48 64 4 4 row` | 等距方塊 48×64（放大 4 倍的單張） | [`21-3dblock-spr.md`](../../docs/internals/21-3dblock-spr.md) |
| `comflag/` | 6 | `COM_FLAG.SPR` | `comflag_export.py` | 指揮旗標籤五張 ＋ 總表 | [`01-assets-inventory.md`](../../docs/internals/01-assets-inventory.md) |
| `strmap/` | 1 | `STRMAP.BIT` | `strmap_export.py` | 戰略星圖底圖 | [`12-strmap-bit.md`](../../docs/internals/12-strmap-bit.md) |
| `vo5/` | 50 | `END*.VO5` | `vo5_export.py` | 結局圖 50 張 | [`18-vo5.md`](../../docs/internals/18-vo5.md) |
| `comani/` | 1438 | `COMANI.BMG` | `bmg_export.py` | 交戰演出動畫 254 段 1,438 格，每張 256×102 是畫完那一格之後的視窗；XOR 開頭的段疊在同單位 D0 第 0 格上 | [`20-comani-bmg.md`](../../docs/internals/20-comani-bmg.md) |
| `font/` | 2 | `GRAPH.IMG` | `font_extract.py；font_head.py` | 倚天中文字型總表與半形字頭 | [`06-font-graph-img.md`](../../docs/internals/06-font-graph-img.md) |
| `music/mid/` | 22 | `*.MDS` | `mds_export.py` | 22 首樂曲轉標準 MIDI | [`17-mds.md`](../../docs/internals/17-mds.md) |
| `music/vgm/` | 22 | `*_F.MDS` | `tools/music_fm_capture.sh` | ＦＭ版 11 首在 dosgolem 裡跑出來的 OPL 暫存器記錄（.meta 是擷取參數） | [`17-mds.md`](../../docs/internals/17-mds.md) |
| `music/ogg/` | 23 | `*.MDS` | `tools/music_to_ogg.sh` | `_F` 由 VGM 合成（OPL 模擬）、`_M` 由 MIDI 以 fluidsynth ＋ FluidR3_GM 合成；`SOURCES.tsv` 是 `_F` 的來源雜湊 | [`17-mds.md`](../../docs/internals/17-mds.md) |

還沒匯出成圖的（解碼在 `internal/asset` 或只有格式文件）：`ANI_PNL0`–`2.A98`、
`BLOCK_0`–`2.BLK`、`MAP*.CMP`、`IBM16.350`（開頭動畫）、`SPTIT.350`。
