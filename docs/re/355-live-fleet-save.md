# 355 — 存檔缺少現況艦隊與艦艇池

工作項目：[Issue #86](https://github.com/wicanr2/logh3/issues/86)。
規格入口：[067](../spec/067-live-fleet-save.md)。證據審查及指定範圍驗收已通過，
收據入口：[028](../playtest/028-live-fleet-save.md)。

## 輸入與工具

原始 `GIN3.COM` SHA-256：
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
固定原版記憶體來自[研究 354](354-post-turn-runtime-fields.md)的 `runtime85-normal/command.bin`；
完整來源 metadata／命令與雜湊在[對拍 027](../playtest/027-post-turn-runtime-fields.md)。
本輪記憶體 SHA-256 `abf7512aef90dd50e5b809026c12008dc9bf2eda775ea30e95f399df72ef7e52`。
dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，Go 1.24；
IDA Pro 9.4 `locked-v1`，image ID
`6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780`。

一次性 COM 切段有 256-byte 前綴；IDA 指令基底 `0x10200`。
原版執行期 `CS=297B`，線性定位 `0x297B0+IP`，不是 IDA 位址。
`re86-297B.com` SHA-256 `45adb5a8ae308d85bc0b06c88ea1c3fca6d4a5dd07ed82637c7f0d1a367fb309`；
`re86-speeds2.com` SHA-256 `9f28a7b740ca064f760dffe6b6fcf941dbd9fd4f3a4ed9d2e37ca771b57bd120`。
由 `tools/ida.sh raw idat -A` 與 `tools/scripts/runtime_fields_ida.py` 建獨立資料庫；
查詢 `3EB:420`、`E8B:2B0`。原始名稱／運算元與 bytes 保留，未審查項均警示 unknown。

## 已確認的產品缺口

`Game.SaveSlot` 新遊戲只建立全 FF 兩表；存檔未序列化 `g.fleets`。
因此原有四支艦隊取進度後變成零支，星域 `+28/+29` 重數後有四個位元組不同。
這是 remake 缺口，不是原版 oracle 未知。

第一次保存又把 `g.slot` 從 nil 改成底稿，`fleetShipTotals` 隨即切換到過期原始表。
只加序列化而不檢查保存後的回合，可能留下 AI 讀舊艘數的分支不對稱。

## 原版資料流（原始定位保留）

`297B:03EB` 使用編制索引取池記錄，`0445`（`26 88 45 03`）寫組序號、
`0452`（`26 88 45 05`）寫隊形槽、`045F`（`26 88 45 02`）寫星域。
索敵隊由艦隊 `+38` 單獨取址，`048A`（`26 C6 44 03 FF`）寫序號 FF、
`0491`（`26 C6 44 06 64`）寫士氣 100、`0498`（`26 C6 44 05 14`）寫隊形 20。
等級：已證實；出處為上述固定雜湊 IDA 查詢及研究 344 的配置／編組讀寫鏈。

`297B:0E8B` 對每組查艦種表，基礎回轉／巡航／戰鬥速度分別是
`kind[08]*fleet[43]/100`、`kind[06]*fleet[42]/80`、`kind[07]*fleet[42]/60`。
再計 `p=(270-2*fleet[46])*mobility[groups]/100`，各速度逐項扣 `speed*p/100`，
寫入艦艇 `+1F/+21/+22`，艦隊 `+4D/+4E/+4F` 取每組最小。
索敵隊尾端直接抄艦種 7 的 `+08/+06/+07`，不吃艦隊能力／機動力折扣。
等級：已證實。以原版執行期艦隊 0／23／28／31 的 15 組編制與 4 組索敵隊
獨立驗算三項速度，19 組全部零差異。只證明指定資料流與樣本，不宣稱存讀閉環已完成。

仍未知的欄位沿用[艦艇池參考](../internals/64-ship-pool.md)，不能以新初始化猜補。

## 已完成的保存修正與證據索引

`Slot.SyncFleets` 在副本保留未知欄位、既有索引，再同步現況；新組以 `Pool=-1` 明示未配置，
需要原始表時按第一空槽配置旗艦／索敵隊／其餘組。這是 remake 物化原始格式的時點，
不冒稱與原版每次點格盤的完整內部池序列相同。保存、戰術部署與結算前接入同步；
AI 艘數統一從活的編制讀取，保存不再讓來源切換到過期底稿。

原版 `297B:0428` 是 word 讀池索引（`26 8B 04`），池有 352 筆；空槽哨兵是 `FFFF`。
remake 三個解析器以前只看低位元組 `FF`，會誤跳過合法 `00FF`（255 號）槽。
已改成兩 byte 都 FF 才是空；原始定位／bytes 保留，等級：已證實。
部署與戰損結算沿用現有組數 `+0F`，縮編的後方舊槽不能再當成現役分隊。

[`live-fleet-save.json`](../internals/data/live-fleet-save.json) 保存 2 個函式與 15 個指令的
經審查語意，含原始定位、bytes、分級、來源及工具位址基準；IDA 匯出自動合併這份索引，
未審查指令仍警示 unknown。重生切段查詢 `3EB:104 E8B:253`，完整函式尾端分別為 `04EE retf`、
`10DD retf`。較早較大查詢範圍只作導覽，不當作正式函式邊界。
原版三份未操作整表、19 組速度與原版非零位置／命令／編制樣本通過；正常 GUI 存讀兩表全 bytes
一致，正常續回合與全部產品套件通過。限制與命令見對拍 028。
