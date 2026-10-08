# 逐函式作業流程

一個函式從「台帳上的一列」到「`diff-tested`」的步驟與門檻。框架說明見 [`README.md`](README.md)，
契約見 [`docs/spec/124`](../docs/spec/124-decomp-reference-oracle.md) 與 `AGENTS.md` §3c。

## 0. 環境與邊界

- 分析、建置、測試一律在 Docker 內。主機只做檔案編輯、`git` 與呼叫 `decomp/tools/*.sh`（這些腳本內部走 Docker）。
- 可以寫的位置：`decomp/src/`、`decomp/specs/`、`decomp/ledger/status.d/`、`workplace/decomp/`。
- 不要改 `decomp/harness/`、`decomp/test/`、`decomp/include/`、`decomp/tools/`、`decomp/ledger/*.tsv`。需要改這些就在回報裡寫明，由主工作階段處理。
- 不要 commit、push、開 issue、改 `AGENTS.md` 或 `CLAUDE.md`、動 `~/.claude`、`~/.cache`、其他 repo。
- 不要執行 `docker prune`、`docker rmi`、`docker rm` 別人的容器。腳本用 `--rm`，容器不會殘留。
- 不要在主機安裝或執行 python／pip。需要小工具就放在 Docker 內跑。
- 原版執行檔、保存狀態、傾印都只讀。

## 1. 選函式

```sh
decomp/tools/ledger.sh order            # 由下而上；game-pure 與 game-logic
decomp/tools/ledger.sh show 203A:0424   # 特徵、分類、呼叫與被呼叫
decomp/tools/asm.sh 203A:0424           # IDA 反組譯，呼叫目標附 seg:off 與分類
```

- 先做深度 0（沒有呼叫其他函式）的葉函式，再做只呼叫「已經 diff-tested」函式的函式。
- `runtime`、`video-dos`、`hw-dependent`、`unknown` 這幾類目前不做。
- 同一個模組（`seg_off` 的 `SEG`）的函式盡量一起做，因為它們共用資料結構。

## 2. 讀函式

- 先 `grep -rn "SEG:OFF" docs/` 看專案是否已經有記載（`docs/re/`、`docs/internals/`、`docs/spec/`），已證實的語意直接採用並引用。
- 反組譯裡的 `sub_XXXXX` 是 IDA 自動名稱，只是座標。台帳的 `seg_off` 才是 `docs/re` 的記法。
- 不要猜欄位的語意。不確定的寫「未知」，並在檔頭標推論等級（已證實／強推論／假說／未知）。

## 3. 呼叫慣例（Datalight C 3.10 Large）

| 項目 | 規則 |
|---|---|
| 序言 | 典型是 `push bp; push si; push di; sub sp,N; mov bp,sp`：區域變數在 `[bp+0 .. bp+N-1]`，向上是 saved bp／si／di（6 bytes），再向上是返回位址 |
| 參數位置 | far 函式：第一個參數在 `[bp+N+10]`，near 函式：`[bp+N+8]`。往上依序是後面的參數。序言不同（沒有 push si／di、或沒有 `sub sp`）時自己推算 |
| 參數寬度 | 每個參數佔 1 個字（byte 也推成 1 個字）；far 指標佔 2 個字（位移在低位址、段在高位址）；long 佔 2 個字（低字在低位址） |
| 回傳 | 16 位元回 `AX`；32 位元與 far 指標多半回 `DX:AX`，但也有 `AX:BX`（配置器 `3982:0001`、`3ABF:0000` 是 AX＝段、BX＝位移）。**一律看 asm 確認**。long 運算輔助函式（`3CC9`）用 `AX:BX`、`CX:DX` 暫存器傳參，輸出也可能是旗標（`3CC9:00B9`） |
| 資料段 | `DS` 固定是 `0x54B3`。全域變數用 `DS(位移)` 取線性位址 |
| 位移運算 | 段內位址是 16 位元，超過 `0xFFFF` 回繞。C 端用 `uint16_t` 做位移運算，再交給 `LIN(seg, off)` |
| 有號與無號 | 看是 `cbw`／`movsx` 還是 `xor ah,ah`；看是 `jl`／`jg` 還是 `jb`／`ja`。C 端的型別要完全對應 |
| 長整數 | 乘法、除法、比較走 `3CC9` 的輔助函式，C 端用 `gin3_lmul`、`gin3_uldivmod`、`gin3_ldivmod`。除法的商是否只取低 16 位元要看呼叫端怎麼用 `BX` |
| 呼叫其他函式 | C 端呼叫對應的 `f_SEG_OFF`，該函式必須已經 `diff-tested`。在自己的 `.c` 檔宣告 `extern` 原型，不要去改共用標頭 |

## 4. 寫 C

`decomp/src/f_<SEG>_<OFF>.c`，範本是 `f_203A_0424.c`。

```c
#include "gin3_reg.h"

/* 檔頭註解：位址、用途、參數、回傳、副作用、輸入域、證據來源、推論等級 */
uint16_t f_203A_0424(Mem *m, uint16_t fleet_off, uint16_t fleet_seg) { ... }

static uint64_t run(Mem *m, const uint16_t *a) { return f_203A_0424(m, a[0], a[1]); }
GIN3_REGISTER("203A_0424", 2, run)
```

- `args` 陣列依描述檔的參數順序展開，farptr 佔兩格（位移、段）。
- 函式名稱固定用 `f_<SEG>_<OFF>`。描述性名稱只放在變數與註解，並標推論等級。
- 讀寫記憶體一律用 `rd8`／`rd16`／`rd32`／`wr8`／`wr16`／`wr32`。
- 檔頭註解要寫輸入域，例如「槽數超過 12 時原版會讀到鄰近記憶體，本版不涵蓋」。

## 5. 寫描述檔與突變清單

`decomp/specs/<SEG_OFF>.json`，範本是 `203A_0424.json`（堆疊參數）與 `3CC9_0009.json`（暫存器參數）。

- `size` 與 `code_sha256_16` 從台帳取：`ledger.sh show` 的 `size`，`functions.tsv` 的 `sha256_final_16` 欄。
- `ret`：`ax`、`dxax`、`none`，或 `pack:` 加暫存器清單（由高位到低位串成一個值，最多 4 個）。清單項目可用 `ax bx cx dx si di bp`、`al ah`、`flags`（OF SF ZF CF 四個旗標），以及單一旗標 `cf zf sf of`。例：`pack:ax,bx`、`pack:al`、`pack:flags`。
- 參數 `kind`：`word`、`byte`、`dword`、`farptr`（`table`、`ds`、`scratch`、`null`）、`nearptr`、`reg`。`reg` 的 `reg` 欄位可以是 `ax cx dx bx si di bp` 或段暫存器 `es`、`ds`（`ss`、`cs` 不開放，堆疊與程式段維持基底狀態）。`kind: byte` 只會推入高位元組為 0 的字；要驗證「原版只讀低位元組」就改用 `word`，`choices` 放 0..255 並加 `also_random`。
  `gen` 可以給 `min`／`max`、`choices`（加 `also_random` 混入均勻隨機）。
- 指向記錄的指標用 `tables` 定義基底、stride、筆數，基底來自 `docs/internals`。
- `base_state` 預設是 `fcmd-base.state`（戰略層）。函式需要戰場狀態時換成戰場狀態並填它的 SHA-256。
- 突變清單 `<SEG_OFF>.mutants.json` 至少 3 個，針對語意關鍵處：邊界、符號、16 位元截斷、寫入。
  `find` 字串在 C 檔裡必須剛好出現一次，否則腳本中止。

## 6. 跑與門檻

```sh
decomp/tools/diff_run.sh <SEG_OFF> <批次名>
```

全部成立才算通過：

1. 回傳值不符 0、寫入集合不符 0、格式錯誤 0。
2. 指令覆蓋 ≥ 95%。未覆蓋的指令要在 `decomp/specs/<SEG_OFF>.notes.md` 逐一說明為什麼到不了（例如錯誤路徑需要的狀態不在基底狀態裡）。
3. 突變存活 0。
4. 丟棄的執行不能占多數。大量被丟棄代表指標範圍或參數描述有問題。

不通過時，先判斷是 C 版寫錯、描述檔不對、還是框架限制。不要為了通過而放寬門檻、刪突變、或把輸入域縮到只剩好測的部分。

## 7. 記錄

```sh
decomp/tools/ledger.sh mark <SEG:OFF> diff-tested decomp/src/f_<SEG>_<OFF>.c decomp/specs/<SEG>_<OFF>.json
```

## 8. 回報

每個函式一列：`seg:off`、案數、覆蓋、突變存活、是否通過；有發現舊文件錯誤或需要框架支援的，另列。
沒做完的函式寫明卡在哪裡與已嘗試的方法。

## 9. 進階描述檔功能

### `ds_seg`：呼叫時的 DS

預設 `54B3`（DGROUP）。函式的呼叫者先把 DS 設成別的資料段時（例如 `2376:097D` 的呼叫者先載入 seg129 的 `3D22`），
用 `"ds_seg": "3D22"`。`tables` 的 `ds:` 基底與 `farptr` 的 `gen.ds` 都跟著這個值。C 版要把 `ds` 當參數，不能用 `DS()` 巨集。

### `seeds`：手工邊界案例

隨機突變很難剛好打到「某個欄位等於特定值」的邊界，例如 `Y + H == 0xA8`。對這種邊界手工給輸入：

```json
"seeds": [
  {"args": [], "pokes": [{"addr": "ds:786", "hex": "7002 8700"}, {"addr": "1234:0010", "hex": "ff"}]}
]
```

`args` 省略就隨機產生；`pokes` 的位址可寫 `ds:OFF`、`SEG:OFF`（hex）或 `lin:ADDR`。種子一律執行並記錄，也會進語料庫讓後續突變從它出發。
種子無法執行（丟棄）會直接中止。

### `stubs`：硬體葉函式替身與呼叫軌跡

函式的呼叫閉包碰到硬體（int、I/O、視訊記憶體）時，把那些葉函式當替身，其餘函式照常用 C 實作並先通過 diff-tested。

```json
"stubs": [
  {"entry": "3ABF:0000", "conv": "far", "nargs": 1, "ret": "pack:ax,bx",
   "gen": {"choices": [0, 1, 4660, 65535], "also_random": true}}
]
```

- 原版執行到替身入口時，探針不執行它，而是讀出 `nargs` 個堆疊參數（加上 `argregs` 指定的暫存器），記錄一筆呼叫，設定回傳值後直接返回。
- `ret`：`ax`、`dxax`、`none`，或 `pack:` 加最多 2 個暫存器。`gen` 決定回傳值，記錄後重放，不是每次不同。
- 參數數量要看替身函式的 asm（它用 `[bp+N]` 存取到哪裡），或看呼叫端呼叫後清除了多少 `add sp,N`。
- C 版以 `gin3_stub(m, "SEG_OFF", nargs, args)` 取代該呼叫，回傳值是記錄的 `AX<<16|BX`（依 `pack:` 順序）。
- 檢查程式比對呼叫的**順序、數量與參數**。替身沒有記憶體副作用，所以只適合不透過指標回傳資料的函式。
- 替身內部真正的行為（例如配置器失敗就停機）不在驗證範圍，要在 C 版檔頭註明。

### `step_budget`、`repeat`、`note`

- `"step_budget": 400000`：每案的步數上限，預設 200000。輸出很大的函式（例如 `1588:06C0` 解壓 13,056 格約需 158,400 步，每個單格封包再多約 10 步）才需要調高。
  調高會讓每案更慢，只在原版本來就要這麼多步時使用。
- `pokes` 的 `"repeat": N`：把 `hex` 重複 N 次再寫入，用來填大段資料，不必逐格寫 100 KB 的描述檔。
- `"note"`：說明文字，探針不使用。
- 暫存緩衝區（`scratch`）固定在線性 `0x9D000` 到 `0xA0000`，共 12,288 bytes，讀到 `0xA0000` 以上不會被發現。需要更大的輸出緩衝時，描述檔先限制輸入域並在 C 檔頭寫明。
