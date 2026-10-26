# 第168輪：提交、推送與工作區結案清理

使用者要求commit、push後結案，清理workspace只留下必要內容。

## 目前狀態表

| 範圍 | 狀態 | 契約 |
|---|---|---|
| 目前來源 | 已提交推送 | `ea26f166`已入main，保存其餘來源工具、歷史收據與規格；正式產品程式及藝術層保持 |
| 私人備忘 | 保留本機 | 原文字檔含私人完整版下載連結，內容及路徑保持，明確忽略避免公開 |
| 可重建副本 | 已清理 | 兩批57項，另清目前macOS的一個Go快取，合計58項；逐檔CRC／SHA、實際權限及AppImage連結已記錄 |
| 系統快取 | 已清理 | Wine非Windows系統部分5970878 bytes、36個檔案讀回通過；registry、使用者資料及連結保留備份。Go編譯快取可重建 |
| 必要保留 | 保留 | 原始資料、Git歷史、來源快照、oracle、decomp、IDA、獨有音畫、存檔與收據；目前166完整驗證輸入及三平台封包／影片 |
| 結案 | 驗收通過 | 淨釋放22265954304 bytes，約20.74GiB；兩批check、ZIP／AppImage還原抽驗、v1.6.3完整交付與55份影片來源核對通過 |

只在本專案內清理。正式ZIP／AppImage及影片不刪；舊版解壓目錄只有逐檔證實可從相鄰ZIP還原才可刪。最新v.1.6.3整版目錄保持，避免破壞可直接啟動及獨立驗收入口。其他專案、Docker映像及外部工作樹不動。

第二批亦使用舊工作目錄中既有ZIP作還原來源。權限不同但內容相同的副本，還原時沿候選原有權限；安裝目錄含額外資料時只刪已核對的素材子目錄，保留額外日誌、存檔及來源。尚未證實可完整還原的歷史產物保留。

已從保留ZIP還原第一批ID0 macOS副本，逐檔SHA與權限通過；第二批ID16 AppDir從正式AppImage還原，檔案、權限及`.DirIcon`連結通過。抽驗副本均再次移除。清理後`poster_frame_receipt.py --check`通過，原始bundle SHA與166影片55份來源完全相同。這是本次交付結案，#48仍由GitHub追蹤，沒有改稱整個remake完成。

`tools/gui.sh`會先建立空的`workplace/gocache`，檢查器只接受空目錄；任何非空快取仍會失敗。第一批計畫的產生器SHA對應`ea26f166`，後續只修這項檢查器環境誤判，不改刪除清單。Wine非系統資料保存在`workplace/cleanup168/wine-state.tar.gz`，所有還原對照及完整檔案指紋保存在同目錄的兩份PLAN。

需重跑舊驗證而缺少安裝副本時，在既有Docker入口執行對應還原命令。不要重建或覆寫歷史收據。

```sh
# 第一批：改ID可還原其他已核對副本；Go與Wine系統由既有映像重建。
python3 tools/scripts/workspace_closeout.py restore --id 0
# 第二批：此例為已驗AppDir，會同時核對連結。
python3 tools/scripts/workspace_closeout_extra.py restore --id 16
```

以上Python命令在容器內執行；主機使用下列既有入口包裝，再接Python命令：

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh tools/gui.sh xrun \
python3 tools/scripts/workspace_closeout.py check
```

入口：[目前脈絡](../../CONTEXT.md)、[總收據](../internals/data/workspace-closeout168.json)、[第一批收據](../internals/data/workspace-cleanup168.json)、[第二批收據](../internals/data/workspace-cleanup168-extra.json)、[既有清理148](2026-10-24-workspace-cleanup-round-148.md)、[166交付](2026-10-26-poster-full-promo-round-166.md)、[第一批工具](../../tools/scripts/workspace_closeout.py)、[第二批工具](../../tools/scripts/workspace_closeout_extra.py)。
