# SPEC 009 — 存檔寫出

狀態：**CONFORMED**（已實作；讀一格再寫出去，六個檔與原版逐位元組相同）
證據：[`docs/re/41`](../re/41-save-files-are-the-runtime-tables.md)、
[`docs/internals/30`](../internals/30-save-format.md)
規格相依：[`005`](005-save-slot-loading.md)（讀存檔）

## 要做什麼

存檔就是執行期那幾張表原樣寫出去，所以「存檔」沒有格式轉換——把手上的
表直接寫成六個檔，再更新 `SAVEDATA.IDX` 的目錄。

## 寫哪些

| 檔案 | 內容 |
|---|---|
| `STELLAR.DT?` | 星域表 32 × 126 |
| `PLANET.DT?` | 行星表 81 × 82（**去掉** `PLADAT.DAT` 的 64 bytes 檔頭）|
| `ADMIRAL.DT?` | 提督表 82 × 28 |
| `FLEET.DT?` | 艦隊表 32 × 122 |
| `SHIP.DT?` | 艦艇池 352 × 78 |
| `WORK.DT?` | 全域狀態 230 bytes |

`WORK` **只改解出來的三欄**（回合數 `+0x04`、軍事費 `+0x3E`、
總稅收 `+0x56`），其餘 227 bytes 照抄。那些還沒解，動它等於把不懂的
東西寫壞。`SAVEDATA.IDX` 同理——只改回合數與名稱，開頭那 12 bytes、
序號與 `FF FF` 原樣留著。

## 兩條硬規則

- **不寫進原版目錄**。輸出目錄由呼叫端給，`org_game/` 一律唯讀
  （`AGENTS.md` 的素材邊界）。
- **只有從存檔進來的局面存得了**。新遊戲沒有 `WORK` 的底稿可以照抄
  ——`WORK.DTX` 不存在，而那 227 bytes 還沒解，湊不出來。

## 驗收

- `internal/rules.TestSlotSaveRoundTrip`：讀第 9、第 10 格再寫出去，
  六個檔**逐位元組**與原版相同。
- `internal/rules.TestWriteSaveIndex`：把目錄寫回原值之後與原版一個 byte
  都不差；改回合數與名稱不會動到序號與時間戳。
- `internal/game.TestSaveSlotRoundTrip`：走完「取進度」那條玩家路徑
  （`docs/spec/007`）之後直接存到另一個目錄，六個檔與原版相同——
  只有 `STELLAR` 的未安定度那 32 個 byte 允許不同，因為載入時重算過
  （`docs/spec/006`）。

## 不在這份規格裡

- **「存進度」畫面的版面**：量到的只有 7 行、位置與「取進度」不同
  （`docs/spec/008`），所以點哪裡存到第幾格還沒接起來。
- 存檔時間戳怎麼寫（`SAVEDATA.IDX` 開頭那六個 byte，`docs/spec/008`
  的解法只有一組樣本）。
- 新遊戲的存檔。
