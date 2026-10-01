# 354 — AI 距離暫存與行星地圖座標

2026-10-01；[#85](https://github.com/wicanr2/logh3/issues/85)。
入口：[第十三輪](../goal/2026-10-01-runtime-fields-round-13.md)；
規格：[066](../spec/066-post-turn-runtime-fields.md)。
已審查的[欄位／位址語意索引](../internals/data/post-turn-runtime-fields.json)
由 IDA 匯出工具自動合併；未列入索引的指令保留 unknown 警示。
驗收入口：[對拍 027](../playtest/027-post-turn-runtime-fields.md)；
完整輸入與輸出雜湊：[收據摘要](../internals/data/post-turn-runtime-receipts.json)。

## 原始輸入與工具

`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
dosgolem-logh3 HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨；
原版收據入口 `tools/oracle.sh`，Go image `golang:1.24-bookworm`。
IDA Pro 9.4（`locked-v1` image ID
`6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`），
查詢 `tools/scripts/runtime_fields_ida.py` 僅操作新切段資料庫，沒有改名。

記憶體採 dosgolem 線性位址；程式另保留 `CS:IP`。
IDA 的一次性 `.COM` 基底 `0x10200`，段內位移須另加，不當成原版線性位址。
切段來自研究 353 的 `command.bin`，SHA-256
`9193bb29a878a1d7fbe4ca86f6a6f6f1b4fdebe16b334aa530152ba6ab4ea212`。
`re85-2E54.com` SHA-256 `77d4f56617ac34ead76e49bef8f591e1dc695421384b2f08edc9685191df5b8f`；
`re85-2AF5.com` SHA-256 `84e13897f022b395ed03850dafca84ff39e3ce9d9030e85d6cd3a0cd26d49b57`。
`re85-113A.com` SHA-256 `6be21e744ef286297eee366a6c129ca0b7fbd1a6f968e20390cb6732f8bb14d3`。

## 星域 `+0x78`（proven）

既有研究 [177](177-ai-phase-2-and-5.md) 與規格 043 已指出它是 AI 階段二距離。
本輪重新核對 IDA 的 `2E54:0F09`：從 `ds:A79C` 開始，每筆加 `0x7E`，
只對 `+0x12 == 2` 的星域查 `2B99:028F(i, 首都, 陣營)`。
`2E54:0F5B` 原始 bytes `26 89 47 78` 寫入 word，**之後**才比 `0x800`
及費沙閘門；不可只寫通過過濾的候選。非無主記錄保留原值。

排序的呼叫鏈為 `2E54:0EC2` → `0F09` → `3C6A:000D`，比較器遠指標
由 `0EDD–0EE4` 載入 `2E54:1001`；排序後 `0EFC` 呼叫 `2E54:1108`。
比較器讀 `+0x78` 升冪；相同再讀 `+0x20` 防衛力升冪、`+0x1A` 經濟力降冪。
原始位移與 bytes 存於一次性資料庫 JSON；自訂函式範圍僅為導覽，
實際邊界以序言、各條 `retf` 及原版命中審查。

## 行星 `+0x46／+0x48`（proven）

`2AF5:0036` 依星域檔頭定位 82 bytes 行星記錄；旗標非零才呼叫 `00CB`。
`00CB` 收最多五個圖塊 6，不足補圖塊 22；偏移 X 加 8、Y 加 12，排序後逐筆寫入。
寫入指令是 `2AF5:0170`／`0184`，bytes `26 89 47 46`／`26 89 44 48`，
後續指標加 `0x52`。`0198–01F6` 完整比較器先比 Y，再比 X，兩者皆升冪。

**勘誤**：研究 320、行星資料參考及 `PlanetSites` 舊註解把計算常式寫成
`2AF5:01CB`；`01CB` 落在比較器內，不是計算入口。本輪寫入監看與
`0036 → 00CB` 原始 call bytes 足以否定該定位，不否定已驗的座標算法。

## 動態監看（proven，指定正常回合）

`runtime85-watch`／`-r2` 從 `fcmd-base.state`，不注入；固定 tick 20000，
按「實行」421M／424M，執行到 433600000，完整 flags 存於收據。
兩輪畫面／PNG／色盤逐位元組相同，六筆監看及步數一致。

- `#427762748` 星域 3 寫 `0x0C6D`，監看 IP `2E54:0F5F`。
- `#430530866` 行星 0 的 X 寫 312，監看 IP `2AF5:0174`。
- `#430530873` 行星 0 的 Y 寫 272，監看 IP `2AF5:0188`。

監看 IP 是執行後位置，不能把它當成寫入指令起點。IDA 前一條指令與 bytes
才能給出上述 `0F5B／0170／0184`。目前證據支持按真正地圖載入時機回寫，
不能在開局或每幀把所有行星提前填好。

## 工具限制與後續

首次 IDA 查詢未輸出；加日誌定位後修正版本 API 與資料陣列阻擋建指令，
最後輸出 schema／輸入雜湊／非空指令均核對。這些是工具問題，沒有修改原版或執行器。
目標為完整表及玩家垂直鏈驗收；不宣稱所有 AI 階段或所有局面一致。

## 重生 IDA 匯出

`tools/py.sh tools/scripts/seg_to_com.py` 從固定雜湊的
`workplace/oracle/supply84-stationed/command.bin` 取 CS `2E54／2AF5／113A`，
參數依序為輸入、線性基底 `0`、CS、長度 `2000`、
`workplace/ida/re85-<CS>.com`。IDA 保留固定 bytes，索引複本放
`workplace/ida/re85-semantics.json`，來源為受版控的 `post-turn-runtime-fields.json`。

```sh
timeout 3m tools/ida.sh raw idat -A -L/work/re85-2E54-reviewed.log '-S/idascripts/runtime_fields_ida.py /work/re85-2E54-reviewed.json 2E54 --index=/work/re85-semantics.json 0EC2:47 0F09:8A 0F93:6E 1001:107' /work/re85-2E54.com
timeout 3m tools/ida.sh raw idat -A -L/work/re85-2AF5-reviewed.log '-S/idascripts/runtime_fields_ida.py /work/re85-2AF5-reviewed.json 2AF5 --index=/work/re85-semantics.json 0036:80 00CB:CD 0198:5F' /work/re85-2AF5.com
timeout 3m tools/ida.sh raw idat -A -L/work/re85-113A-reviewed.log '-S/idascripts/runtime_fields_ida.py /work/re85-113A-reviewed.json 113A --index=/work/re85-semantics.json 0350:128' /work/re85-113A.com
```

匯出自動合併語意索引且驗輸入雜湊／指令 bytes，不從散文批次提升分級。
CS:IP、IDA 位址、執行期線性位址與原始名稱／運算元並列，unknown 保留警示。
切段資料庫沒有跨段完整映射，直接 xref 為零不代表沒有 caller；本輪以原始
far call bytes、完整比較器指標與實際命中堆疊補證。
輸出雜湊保存於收據摘要；舊文件回填護欄由 `tools/scripts/runtime_fields_summary.py`
逐份驗證勘誤指向，缺入口即失敗。

## 驗收與新增限制

兩組原版重跑的階段記憶體與畫面完全重現；remake 指定完整回合的星域 4032／
行星 6642 bytes 全同，無排除區。正常 GUI、待命／情報抽驗及標準全部套件通過。
詳見對拍 027。保存後取進度另揭露現況艦隊／艦艇未寫入底稿，#86 開放追蹤；
本輪只完成兩張表完整寫出、新增 AI 距離及全部行星讀回，未宣稱全局存讀閉環。
