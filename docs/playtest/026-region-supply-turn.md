# 正常回合的星域／行星庫存鏈對拍

日期：2026-10-01；[#84](https://github.com/wicanr2/logh3/issues/84)。
比較等級：指定庫存及畫面的 `same-state`。S1 帝國 796 年 1 月 1 日 → 4 日。
範圍：全部 32 格兩陣營七種待命艦艇、飛彈／艦載機，以及全部 81 顆行星對應庫存。
完整執行期表格仍有差異，另由 [#85](https://github.com/wicanr2/logh3/issues/85) 追查。

## 來源與重生

原版 `GIN3.COM` SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
專用 dosgolem-logh3 HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨；
Go image `golang:1.24-bookworm`，正式 GUI image `logh3-ebiten:2`，分析
`python:3.12-slim`。原版輸入唯讀，未改執行器。

```sh
tools/region_supply_receipts.sh
tools/region_supply_receipts.sh supply84-stationed-r2
tools/region_supply_receipts.sh supply84-gui gui
tools/region_supply_receipts.sh supply84-gui-r2 gui
LOGH3_GUI_PREFIX=supply84-gui bash tools/gui_check.sh
tools/py.sh tools/scripts/plan_arrow_stats.py supply84-gui
tools/py.sh tools/scripts/region_supply_summary.py
tools/gui.sh test ./internal/... ./cmd/...
```

原版從 `fcmd-base.state` 完整檢查點還原。每次的父檢查點雜湊、完整旗標、
階段步數與輸出雜湊保存在 [結構化收據](../internals/data/region-supply-turn-timing.json)。
`initial.bin` 在執行前實讀 `DS:94DC`＝`0xD3DC167E`；remake 使用正式
`strategyRandSeed` 同值，在執行前固定。局部庫存公式沒有亂數呼叫。
此次比較同一初始局面與等價輸入的結果，不把同 seed 當成整段骰序一致。

`stationed` 局面沿正常主選單／防拷／月初報告／艦隊編成抵達；只為窄幅補給
診斷注入六處：旗標、艦隊與旗艦士氣、現有艘數、飛彈與艦載機。旗艦設
100 艘／兩種存量 0，避免混入使用者已選定的負缺口夾零差異。
所有注入線性位址、bytes 及時點在 `flags.txt`，規格及命中見 [研究 353](../re/353-region-supply-turn-timing.md)。
`gui` 完全沒有狀態注入；`cmd/gin3` 正式入口與 renderer 以等價輸入從主選單
走到編成與實行，完整事件序列及程式雜湊在 `remake-script.txt`／`remake-meta.json`。
沒有直接進畫面；不宣稱另做人工手感或真實作業系統事件驗收。

## 結果與限制

兩輪原版重播的 12 個診斷階段及 2 個未注入階段，各 1 MB 記憶體、畫面／色盤
全部相同。分配後的行星全 6642 bytes 相同；補給後星域 0 整筆及行星整表
相同；無艦隊星域 1 也執行彙總與分配；下一座位只回充星域，32 格全
4032 bytes 相同，行星全表不動。

正常完整回合庫存欄位共星域 2304、行星 2916 bytes 全同；沒有略掉陣營或艦種。
正常存檔保留兩份完整表，正常「取進度」讀回行星一致。跨月交付、原有
回合／戰鬥／AI 及全產品套件通過；既有交付測試改為先完成九回合再比較
第十回合交付前後，避免把開局未分配的行星分布當成收據的交付前值。

未注入 GUI 原生 640×480，沒有裁切、縮放、遮罩或排除矩形：

| 區域 | 比較像素 | 不同像素 | 最大 RGB 差 |
|---|---:|---:|---:|
| 全畫面 | 307200 | 0 | 0 |
| 地圖 `(0,40,320,360)` | 115200 | 0 | 0 |
| 命令面板 `(320,40,320,260)` | 83200 | 0 | 0 |
| 艦隊清單 `(320,300,320,100)` | 32000 | 0 | 0 |

原版／remake PNG、色號資料、差分 PNG、逐區統計與雜湊齊全；資料見結構化收據。
另抽驗 `fcmd-exec` 正常移動與 `vs5` 對戰交棒完整畫面相同。

**完整正常回合表未達一致**：九格星域 `+0x78`／`+0x79` 共 18 bytes，首都四顆
行星 `+0x46`–`+0x49` 共 15 bytes。所有原值、remake 值與定位保留在結構化
收據，等級仍為 `unknown`，由 #85 追查；沒有把未知欄位分類成允許差異。
此次完成聲明限於已驗的庫存鏈與畫面，不宣稱所有回合、超編算術或整體原版一致。

## 2026-10-01 後續勘誤

本輪留下的星域 18／行星 15 bytes 差異已由[研究 354](../re/354-post-turn-runtime-fields.md)
證實為 AI 距離暫存及載圖座標，正式接線後兩組回合全部星域 4032／行星 6642 bytes
零差異。此處保留發現時的差異與未知分級，不回寫成當時已證實。新增存讀檔驗證
另發現現況艦隊／艦艇未同步保存，由 #86 追蹤；不宣稱完整局面保存完成。
