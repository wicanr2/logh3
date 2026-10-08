# decomp：原版函式的 C 參考實作

把 `GIN3PS.EXE` 的函式用 C 重寫成參考實作，逐函式對原版做差異測試。通過的 C 版可以在 Docker 內原生編譯，
當作 remake 規則層的 oracle，不必每次都起 dosgolem。驗收契約見 [`docs/spec/124-decomp-reference-oracle.md`](../docs/spec/124-decomp-reference-oracle.md)，
專案契約在 `AGENTS.md` §3c。

## 目標與現況

目標是完成 `GIN3PS.EXE` 的 matching decompilation：每個函式的 C 原始碼，用原版編譯器編出與原版相同的 bytes。

進度數字用 `decomp/tools/ledger.sh summary` 取得，不手寫。

| 項目 | 現況 |
|---|---|
| 函式台帳 | 1,440 個函式、129 個模組（`ledger/functions.tsv`），機械分類見 `ledger.py` 檔頭 |
| `byte-matching` | 0 個 |
| `diff-tested` | 見 `ledger.sh summary`。每個函式的收據在 [`docs/re/443`](../docs/re/443-fleet-attack-differential-test.md)、[`445`](../docs/re/445-differential-test-framework.md) |
| 編譯器 | **缺 Datalight C 3.x**。`GIN3PS.EXE` 的 runtime 是 Datalight C 3.10 Large（[`docs/re/442`](../docs/re/442-compiler-is-datalight-c.md)），已取得的 Zortech C++ 2.06／3.1 代碼產生器風格不同，編不出相同 bytes |

沒有同版編譯器，`byte-matching` 無法開始。目前只能做 `diff-tested`（功能等價），取得編譯器後可把已完成的函式升級，等級不倒退。

## 目錄

| 路徑 | 內容 |
|---|---|
| `src/f_<SEG>_<OFF>.c` | C 參考實作。一個函式一個檔，用 `GIN3_REGISTER` 自我登記。 |
| `include/` | `gin3_mem.h`（1 MB 記憶體模型）、`gin3_reg.h`（登記）、`gin3_rt.h`（runtime 數學語意） |
| `specs/<SEG_OFF>.json` | 差異測試描述檔；`<SEG_OFF>.mutants.json` 是該函式的突變清單 |
| `harness/` | 在 dosgolem module 內執行的 Go 探針。獨立 `go.mod`，主 module 不會編譯它 |
| `test/` | 讀向量檔、用 C 版重算並比對的程式 |
| `tools/` | 執行入口與映像、IDA、台帳工具 |
| `ledger/` | `functions.tsv`、`callgraph.tsv`、`insns.tsv` 由工具產生；`status.d/<SEG_OFF>.tsv` 一個函式一個檔記錄等級 |
| `WORKFLOW.md` | 逐函式的作業流程與驗收門檻 |

## 取得映像、IDA 與台帳

```sh
# 1. 解包映像、重定位表與標準 MZ 檔（約 5 秒；輸出在 workplace/decomp/image/<批次>/）
decomp/tools/gin3ps_image.sh <新批次名>

# 2. IDA 分析（專案預設的 locked-v1 已不在本機時用 py312-v1）
cp workplace/decomp/image/<批次>/GIN3PS_UNPACKED.EXE workplace/ida/
LOGH3_IDA_IMAGE=ida-pro-9.4-idapython:py312-v1 tools/ida.sh raw idat -A -B /work/GIN3PS_UNPACKED.EXE
cp workplace/ida/GIN3PS_UNPACKED.EXE.i64 workplace/ida/GIN3PS_UNPACKED.EXE.i64.pristine     # 匯出前先備份
cp decomp/tools/ida_export_functions.py workplace/ida/
LOGH3_IDA_IMAGE=ida-pro-9.4-idapython:py312-v1 tools/ida.sh raw idat -A -S/work/ida_export_functions.py /work/GIN3PS_UNPACKED.EXE.i64

# 3. 台帳（Docker 內執行）
decomp/tools/ledger.sh build <映像批次名>
decomp/tools/ledger.sh summary
```

## 日常指令

```sh
decomp/tools/ledger.sh show 203A:0424       # 函式特徵、分類、呼叫與被呼叫
decomp/tools/ledger.sh order                # 由下而上的反編譯順序
decomp/tools/asm.sh 203A:0424               # IDA 反組譯，呼叫目標附 seg:off 與分類
decomp/tools/diff_run.sh 203A_0424 <批次>   # 單一函式：產生原版向量、比對 C 版、跑突變對照
decomp/tools/diff_all.sh <批次>             # 所有描述檔
decomp/tools/check_diff.sh <vectors.txt>    # 只重跑 C 版比對（不需要原版與執行器）
decomp/tools/ledger.sh mark 203A:0424 diff-tested decomp/src/f_203A_0424.c decomp/specs/203A_0424.json
```

`diff_run.sh` 需要 `dosgolem-logh3` 工作樹乾淨且 HEAD 符合腳本內固定值。向量檔與映像在 `workplace/decomp/`（gitignore），
檔頭記錄執行器 HEAD、保存狀態與函式位元組的雜湊。收據綁執行器版本，HEAD 改變後要重跑。

## 邊界

- 這裡是重寫的 C、測試程式與函式台帳（位址、大小、雜湊），不含原版執行檔，不含 Datalight 的 runtime 原始碼或由它改寫的內容。
- 原版執行檔、解包映像、保存狀態、IDA 資料庫、向量檔都留在本機，不進 Git。
- 本 repo 為 private。轉公開前，`decomp/` 與 `assets/original/` 一併請使用者裁決。
