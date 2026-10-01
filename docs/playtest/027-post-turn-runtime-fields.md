# 正常回合 AI 距離與行星座標全表對拍

日期：2026-10-01；[#85](https://github.com/wicanr2/logh3/issues/85)。
比較等級：指定 S1 帝國首回合、全部星域與行星表的 `same-state`；正式正常 GUI 同狀態。
證據：[研究 354](../re/354-post-turn-runtime-fields.md)、[規格 066](../spec/066-post-turn-runtime-fields.md)，
機器摘要：[完整收據與雜湊](../internals/data/post-turn-runtime-receipts.json)。

## 固定來源與重生

原版 bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`；
`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
dosgolem-logh3 HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨；
Go image `golang:1.24-bookworm`、GUI `logh3-ebiten:2`、分析 `python:3.12-slim`。
執行器未改，原版唯讀。兩側執行前固定 seed `0xD3DC167E`；原版初始記憶體實讀
線性 `0x5E00C`（DS:94DC），remake 使用 `strategyRandSeed`，不永久鎖住正式遊戲。

```sh
bash tools/runtime_fields_receipts.sh runtime85-normal gui
bash tools/runtime_fields_receipts.sh runtime85-normal-r2 gui
bash tools/runtime_fields_receipts.sh runtime85-stationed stationed
bash tools/runtime_fields_receipts.sh runtime85-stationed-r2 stationed
LOGH3_GUI_PREFIX=runtime85-normal bash tools/gui_check.sh
tools/py.sh tools/scripts/plan_arrow_stats.py runtime85-normal
tools/gui.sh test ./internal/... ./cmd/...
tools/py.sh tools/scripts/runtime_fields_summary.py
```

所有原版 flags、檢查點雜湊、階段步數及記憶體雜湊由包裝器保存；remake GUI metadata
記錄 HEAD、未提交狀態、來源檔雜湊、正式命令與完整 `remake-script.txt`。
`normal` 沿原版編成檢查點按實行，不注入；remake 從正常主選單／防拷／月報／編成到實行。
`stationed` 沿相同路徑，另保留研究 353 的六項明示注入，只屬低士氣補給診斷分支。
正常原版五階段、診斷十二階段的 1 MB 全記憶體及畫面／色盤各重跑兩次，相同。

## 結果與界線

兩組回合結束皆逐位元組比較全部星域 4032／行星 6642 bytes，**零差異，沒有遮罩或排除欄位**。
舊的星域 18／行星 15 bytes 差異已解成 AI 距離與載圖座標；原始舊差異仍保留於研究 353 索引。
正式 GUI 原生 640×480、無裁切／縮放／遮罩，完整色號及 RGB 零差異：

| 區域 | 像素 | 差異 | 最大 RGB 差 |
|---|---:|---:|---:|
| 全畫面 | 307200 | 0 | 0 |
| 地圖 `(0,40,320,360)` | 115200 | 0 | 0 |
| 命令面板 `(320,40,320,260)` | 83200 | 0 | 0 |
| 艦隊清單 `(320,300,320,100)` | 32000 | 0 | 0 |

原版／remake PNG、差分 PNG、逐區統計均由腳本重生。另抽驗 `fcmd-standby-pick`、
`fcmd-rinfo`，通過既有小圖最後一列的允許差異契約；不把這兩張升格為無排除區一致。
戰術與戰略打行星兩條產品路徑由全部套件回歸抽驗，不另宣稱新戰場畫面同狀態。

全表寫入 STELLAR／PLANET 存檔與當前表一致；正常取進度保存全部行星與 32 格新增 AI 距離。
**現況艦隊與艦艇池尚未同步保存**，由 [#86](https://github.com/wicanr2/logh3/issues/86) 追蹤；
因此不宣稱完整局面存讀閉環。載入本來會重數艦隊、重算未安定度，不能把衍生欄位當永久原值。

距離非步數、所有無主格先寫再過濾、兩項排序次鍵與費沙 238／318 門檻有獨立規則測試；
合成邊界輸入只證明實作遵守已審查的原版控制流，不替代正常回合收據。
完全平手的 qsort 次序、十回合後 AI 全局與所有局面未裁決，沒有全遊戲一致性聲明。
月末稅收測試仍屬相鄰局面：本輪只有星域 25 擁有者不同，照原有最多三格額度跳過；
其餘逐格稅收斷言不變，不用這支測試宣稱十回合 AI 對拍。
