# 444 — GIN3PS.EXE 解包映像、重定位表與函式台帳

日期：2026-10-09。起點是 matching decompilation 需要一份乾淨的解包映像、重定位表與函式清單。
`docs/re/336` §9.3 記錄過一份舊解包映像（`c3fa47da504d3b94…`），但當時的解包工具 `workplace/r7glue/chain.py` 與輸出已不在磁碟上。
這份紀錄用 dosgolem 的執行期傾印重做，不依賴解包殼的實作。

## 結論

| 項目 | 結論 | 等級 |
|---|---|---|
| 解包後映像 371,594 bytes，最終映像 SHA-256 開頭 `c3fa47da504d3b94` | 與 `docs/re/336` 記錄的舊映像相同 | 已證實 |
| 映像解壓完成、重定位尚未開始的時刻 | 第 58,313,900 步 | 已證實 |
| 最終入口（映像已重定位、啟動碼尚未執行） | 第 58,339,129 步，`CS:IP=0349:0079`，`SS:SP=63EB:0080`，`DS=ES=0339` | 已證實 |
| 重定位數 | 5,991 筆，與 `docs/internals/02` 記錄的 5,991 項相同 | 已證實 |
| 未重定位映像加上這 5,991 筆等於最終映像 | 逐位元組相同 | 已證實 |
| 映像分為程式碼、零填充緩衝區、DGROUP 初始資料三段 | 見下表 | 已證實（邊界由傾印與 IDA 範圍量得） |

## 輸入

| 項目 | 值 |
|---|---|
| `GIN3.COM` SHA-256 | `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| `GIN3PS.EXE` SHA-256 | `1efa4bd95e73415efba440af63c6a480665d60a67d8876b5f230f5c391aa2381` |
| dosgolem HEAD | `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨 |
| probe 旗標 | `-exe /orig/GIN3.COM -root /orig -empty-env -tick 20000 -clicks 40000000:300:170 -click-hold 3000000` |

## 方法

1. 冷開機跑 `GIN3.COM`，第 4000 萬步點掉 `OPEN.EXE` 的開場動畫，GIN3PS 開始六層解包。
2. `0349:0079` 在第 18,445,091 步（`OPEN.EXE` 啟動附近，`DS=075A`）與第 44,219,204 步起（外層 GLUE 殼的迴圈，之後每約 30 步一次）都會被命中，
   不能當作解包完成的訊號。改在多個步數對 `0x3509` 取樣 8 bytes：第 5500 萬步前是殼的程式碼，第 5600 萬步變成
   `FA B8 6A 51`（`mov ax,516Ah`，**尚未重定位**的資料段），第 5850 萬步變成 `FA B8 B3 54`（`mov ax,54B3h`，已重定位）。
3. 解壓與重定位是兩個先後階段。對最終入口傾印與各取樣點做差異：解壓期間「加上 `0x0349` 的 word」數遞增、其他差異遞減；
   第 58,313,900 步其他差異為 0 且正好 5,991 筆，之後重定位開始，筆數遞減。
4. 重定位 word 的低位 byte 一定改變（`0x0349` 的低位是 `0x49`），所以掃描每個不同的 byte 即可。
5. 以 `decomp/tools/build_mz.py` 組出標準 MZ 檔並驗證。

重做入口：`decomp/tools/gin3ps_image.sh <新批次名>`，約 5 秒，輸出在 `workplace/decomp/image/<批次>/`。

## 輸出雜湊

| 檔 | SHA-256 |
|---|---|
| 未重定位映像（371,594 bytes） | `4d68ba4786d57129580cb8d1e83904d8c36c19d0ed21edda20dbe39f5b4f8eab` |
| 最終映像（371,594 bytes，載入於 `0349`） | `c3fa47da504d3b94551e0e41362942cbbca9256661809cdcf4941b2f367e51d9` |
| `GIN3PS_UNPACKED.EXE`（395,594 bytes，24,000 bytes 檔頭） | `6a787dcf4046c787496553c18190068e696f3d58568717859e0eb7ba035efab0` |

MZ 檔的入口與堆疊（`CS:IP`＝`0000:0079`、`SS:SP`＝`60A2:0080`）取自最終入口的暫存器減去載入段 `0349`。`min alloc` 與 `max alloc` 不影響反組譯，沿用 0 與 `0xFFFF`。

## 映像分段

位移是映像內位移。IDA 線性位址等於位移加 `0x10000`，dosgolem 線性位址等於位移加 `0x3490`，`docs/re` 使用的 `seg:off` 與 dosgolem 線性位址是同一個空間。

| 位移範圍 | 內容 | 大小 |
|---|---|---:|
| `0x00000`–`0x39D11` | 程式碼，IDA 辨識出 1,440 個函式共 218,606 bytes，其餘 18,211 bytes 是函式之間尚未分類的位元組 | 236,817 |
| `0x39D11`–`0x516A0` | 幾乎全是 0 的緩衝區（`0x3A000` 起每個 4 KB 區塊零佔 99% 以上；`0x48000` 附近有一塊 71% 為零的非零島） | 96,655 |
| `0x516A0`–`0x5AB8A` | DGROUP 初始資料（`DS=54B3`，相對段 `0x516A`）；其後的 BSS 不在映像裡 | 38,890 |

零填充區的存在代表整檔 byte-identical 還需要重現這些緩衝區的配置。`0x48000` 附近的非零島尚未分類。

## IDA

使用現存的 `ida-pro-9.4-idapython:py312-v1`（映像 ID `4ac62de83339`）。專案預設的 `locked-v1` 已不在本機，這次沒有重建，`tools/ida.sh` 以 `LOGH3_IDA_IMAGE` 覆寫。

- `idat -A -B /work/GIN3PS_UNPACKED.EXE`：基底 `1000h`，載入長度 `5AB8Ah`（371,594），入口 `1000:79`，約 5 秒完成。
- IDAPython 可用。`decomp/tools/ida_export_functions.py` 匯出 1,440 個函式：far 1,305、near 135、thunk 1、含 tail chunk 3。
- 函式大小：最小 1，中位 106，第 90 百分位 316，最大 2,615 bytes。
- 座標換算用已知函式驗證：`203A:0424` 對應 `sub_2D334`（173 bytes、7 個呼叫者，位元組開頭 `55 56 57 83 EC 16 89 E5`）；`3CC9:0009` 對應 `sub_49809`（23 bytes、150 個呼叫者）；`3CC9:0020` 對應 `sub_49820`（80 bytes、72 個呼叫者）。
- IDA 資料庫載入時用基底段 `0x1000` 套用重定位，所以資料庫內的 bytes 與 dosgolem 的 `0349` 版本在重定位 word 上不同。台帳的 `sha256_16` 取自**未重定位**映像，與載入段無關。

## 台帳

`decomp/ledger/functions.tsv` 由 `decomp/tools/ledger.py build` 產生，欄位見該腳本的說明。`decomp/ledger/status.tsv` 手動維護每個函式的等級，
`ledger.py summary` 印出覆蓋率。目前：1,440 個函式中 `diff-tested` 1 個（173 bytes，0.1%），`byte-matching` 0 個。
台帳的分母是 IDA 的函式範圍，不是整個映像。

## 限制與未決

- 函式分類（遊戲碼、Datalight runtime、手寫組語、資料）尚未做。`3CC9`（long 輔助函式、亂數）與開頭的啟動碼確定是 runtime。
- IDA 的函式邊界是自動分析結果，可能有切錯或漏掉的函式，`docs/re` 已有數次進入點錯位的紀錄（`CLAUDE.md` §3a-1）。
- byte-matching 的前提是同版編譯器，見 [`442`](442-compiler-is-datalight-c.md)。
