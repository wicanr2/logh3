# 126：修護艦每回合補同艦隊的缺額（`1AAB:1286` 第二階段）

狀態：DRAFT（2026-10-09）。尚未動 production path，等使用者決定要不要照原版補上。
類別：原版行為，remake 尚未實作。

## 證據

- 差異測試：`decomp/src/f_1AAB_1286.c`（658 案、指令覆蓋 112/112、13 個突變全被抓到，向量檔 `74e076b598cbfa04`），
  主工作階段獨立重跑得到相同雜湊。完整語意見檔頭與 [`docs/internals/67`](../internals/67-decomp-verified-details.md)。
- 舊文件 [`docs/re/346`](../re/346-region-resupply.md) §四、[`docs/internals/64`](../internals/64-ship-pool.md)、[`docs/spec/063`](063-region-resupply.md)
  只寫了第一階段（重設）。

## 原版行為

`1AAB:1286(艦隊)` 對艦隊裡每一個艦種 6（修護艦）的編制組 `S`，依序做：

1. 重設：`S[+0x10]（word）← 0`、`S[+0x0E]（word）← S[+0x08]`，也就是存量（dword）等於現有艘數。remake 已有（`rules.ResetRepairShips`）。
2. 補缺額：對同艦隊的每個編制組 `D`，依編制順序：
   - `S` 的存量（`+0x0E` dword）為 0 就結束；
   - `D` 的艦種是 6 就跳過；
   - `D[+0x0C] <= D[+0x08]`（無號 word）就跳過；
   - `need ← D[+0x0C] − D[+0x08]`（16 位元）；`give ← min(S 存量, need)`；
   - `D[+0x08] += give`、`S 存量 −= give`（32 位元）。

只動 `D` 的現有艘數：修理上限、士氣、飛彈與艦載機存量都不變。呼叫者是 `1AAB:1235`，`1AAB:0FAF` 補充流程最前面，
掃的是這一格**所有陣營**的艦隊（不限擁有者，`docs/re/346` §三），無主星域整段不跑。

## 與 remake 的差距

`rules.ResetRepairShips` 只做第 1 步。`rules.ResupplyRegion` 之後的流程（星域庫存補到滿編）不碰這一步的結果，
所以在自家（或敵方艦隊所在的）星域，原版每回合會由修護艦把戰損補回修理上限，remake 不會。

差異測試轉接器：`internal/rules/decomp_repair_reset_test.go`（環境變數 `LOGH3_DECOMP_REPAIR_RESET`），
對良構案例（編制組數 ≤ 12、池索引互異且小於 352）逐欄位比對。2026-10-09 實測：658 案中良構 272 案，與原版完全相符 160 案，
原版多做第 2 階段而不符 112 案，其他不符 0 案。也就是第 1 步全部相符，差距完全來自第 2 步。

## 預計改動（待決定）

| 落點 | 修改 |
|---|---|
| `internal/rules/regionresupply.go` 的 `ResetRepairShips` | 第 1 步之後加第 2 步，算式如上；只動 `Ships` 與修護艦的 `Missiles` |
| 測試 | 上列轉接器改成嚴格比對；加單元測試（存量不足、存量足夠、有多組缺額、修護艦之間不互補） |

## 驗收

- 轉接器對良構案例零不符。
- 既有 `TestResupply*` 在沒有缺額的情境結果不變。
- 玩家可見的影響：自家星域的艦隊，修護艦每回合把戰損補回修理上限（受修護艦存量限制），存量等於修護艦的現有艘數。
  `RepairShipsRepairable` 偏好（規格 125）不影響這一步。
