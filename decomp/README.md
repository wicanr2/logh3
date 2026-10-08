# decomp：原版函式的 C 參考實作

把 `GIN3PS.EXE` 的函式用 C 重寫成參考實作，逐函式對原版做差異測試。通過的 C 版可以在 Docker 內原生編譯，
當作 remake 規則層的 oracle，不必每次都起 dosgolem。驗收契約見 [`docs/spec/124-decomp-reference-oracle.md`](../docs/spec/124-decomp-reference-oracle.md)，
專案契約在 `AGENTS.md` §3c。

## 目標與現況

目標是完成 `GIN3PS.EXE` 的 matching decompilation：每個函式的 C 原始碼，用原版編譯器編出與原版相同的 bytes。

| 項目 | 現況 |
|---|---|
| 函式台帳 | 1,440 個函式、218,606 bytes（`ledger/functions.tsv`） |
| `byte-matching` | 0 個 |
| `diff-tested` | 1 個：`203A:0424`，173 bytes（[`docs/re/443`](../docs/re/443-fleet-attack-differential-test.md)） |
| 編譯器 | **缺 Datalight C 3.x**。`GIN3PS.EXE` 的 runtime 是 Datalight C 3.10 Large（[`docs/re/442`](../docs/re/442-compiler-is-datalight-c.md)），已取得的 Zortech C++ 2.06／3.1 代碼產生器風格不同，編不出相同 bytes |

沒有同版編譯器，`byte-matching` 無法開始。目前只能做 `diff-tested`（功能等價），取得編譯器後可把已完成的函式升級，等級不倒退。
進度數字用 `ledger.py summary` 取得，不手寫。

## 目錄

| 路徑 | 內容 |
|---|---|
| `src/` | C 重寫版。檔名帶原版位址 |
| `include/gin3_model.h` | 資料版面常數（記錄大小、欄位位移） |
| `harness/` | 在 dosgolem module 內執行的 Go 探針，直接呼叫原版函式並產生向量。獨立 `go.mod`，主 module 不會編譯它 |
| `test/` | 讀向量檔、用 C 版重算並逐筆比對的程式 |
| `tools/` | 執行入口與映像、IDA、台帳工具 |
| `ledger/` | 函式台帳。`functions.tsv` 由工具產生，`status.tsv` 手動維護 |

## 執行

```sh
# 1. 取得解包映像、重定位表與標準 MZ 檔（約 5 秒；輸出在 workplace/decomp/image/<批次>/）
decomp/tools/gin3ps_image.sh <新批次名>

# 2. IDA 分析（用現存映像；專案預設的 locked-v1 已不在本機時用 py312-v1）
cp workplace/decomp/image/<批次>/GIN3PS_UNPACKED.EXE workplace/ida/
LOGH3_IDA_IMAGE=ida-pro-9.4-idapython:py312-v1 tools/ida.sh raw idat -A -B /work/GIN3PS_UNPACKED.EXE
cp decomp/tools/ida_export_functions.py workplace/ida/
LOGH3_IDA_IMAGE=ida-pro-9.4-idapython:py312-v1 tools/ida.sh raw idat -A -S/work/ida_export_functions.py /work/GIN3PS_UNPACKED.EXE.i64

# 3. 台帳（python3 -I，Docker 內）
python3 -I decomp/tools/ledger.py build workplace/ida/functions.json workplace/decomp/image/<批次>/relocs.json \
    workplace/decomp/image/<批次>/gin3ps-unrelocated.bin decomp/ledger
python3 -I decomp/tools/ledger.py summary decomp/ledger

# 4. 單一函式的差異測試：產生原版向量，再用 C 版比對（需要 dosgolem-logh3 乾淨且 HEAD 符合腳本內固定值）
decomp/tools/fleet_attack_run.sh <新批次名> [案數]

# 只重跑 C 版比對（不需要原版與執行器）
decomp/tools/check_vectors.sh workplace/decomp/fleet_attack/<批次名>/vectors.txt

# remake 的 Go 實作對同一份向量
LOGH3_DECOMP_FLEET_ATTACK_VECTORS=/src/workplace/decomp/fleet_attack/<批次名>/vectors.txt \
  tools/go.sh test ./internal/rules -run TestStrategicAttackAgainstOriginalVectors -v -count=1
```

向量檔與映像在 `workplace/decomp/`（gitignore），檔頭記錄執行器 HEAD、保存狀態與函式位元組的雜湊。收據綁執行器版本，
HEAD 改變後要重跑。

## 邊界

- 這裡是重寫的 C、測試程式與函式台帳（位址、大小、雜湊），不含原版執行檔，不含 Datalight 的 runtime 原始碼或由它改寫的內容。
- 原版執行檔、解包映像、保存狀態、IDA 資料庫、向量檔都留在本機，不進 Git。
- 本 repo 為 private。轉公開前，`decomp/` 與 `assets/original/` 一併請使用者裁決。
