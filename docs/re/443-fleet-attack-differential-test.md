# 443 — 戰略艦隊攻擊力 `203A:0424` 的差異測試

日期：2026-10-08。第一支進入 `decomp/` 的函式，驗證流程見 [`spec/124`](../spec/124-decomp-reference-oracle.md)。

## 輸入

| 項目 | 值 |
|---|---|
| 原版 `GIN3.COM` SHA-256 | `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| 保存狀態 `dosgolem-logh3/workplace/fcmd/fcmd-base.state` SHA-256 | `c55de581a6393e0e39e304dd0f191e9d1cfa8457b86b93f2a9e3368b5751ba26` |
| `203A:0424` 的 173 bytes SHA-256 | `a307dbe5319b50d57ca2dd57ae6638468a5402de5f5cc3de8a7fb5c73a2a23cf` |
| dosgolem HEAD | `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，工作樹乾淨，分支 `logh3-oracle` |
| 該狀態下的艦艇池 | 段 `668B`，位移 `0002`（`ds:CFB6`／`ds:CFB4`） |

## 方法

1. 載入保存狀態，驗證 `203A:0424` 的開頭 8 bytes 是 `55 56 57 83 EC 16 89 E5` 且最後一個 byte 是 `CB`。
2. 每一案把艦隊記錄（122 bytes）寫進 `ds:C074 + 31 × 122`，把被引用的艦艇記錄（78 bytes）寫進艦艇池。
   記錄裡函式不讀的位元組填隨機值。
3. 照 `oracle.Call` 的作法呼叫：推入 far 指標與假返回位址，設 `DS=54B3`，清 IF，逐道執行到回到假返回位址。
4. 呼叫前後比對整個 1 MB，只允許堆疊下方 `0x100` bytes 改變。
5. 向量檔交給 `decomp/test/check_fleet_attack.c` 重算比對，再交給 remake 的 Go 測試。

6 個固定案加 5000 個隨機案。隨機案刻意加重邊界：槽數 0／1／11／12、艘數 0／1／65535／32768、
能力 0／1／255，所以溢位區案數遠高於實際遊戲會遇到的比例。

## 結果

| 項目 | 結果 |
|---|---|
| 原版對 C 版 | 5006 案，不符 0 |
| 原版呼叫是否動到堆疊以外的記憶體 | 0 案。函式是純函式 |
| 執行步數 | 359 到 697 道指令 |
| 重複產生 | 三次產生的向量檔逐位元組相同，SHA-256 `36d47bd922e2a050c974028c2526bf1ff465524e4ae74cf2625e46cd4dc3ec70` |
| 負向對照 1：商不截成 16 位元 | 206 案中不符 18 |
| 負向對照 2：艘數合計不回繞 | 206 案中不符 58 |
| 負向對照 3：`+250` 改 `+251` | 206 案中不符 206 |
| remake `strategicAttack` 對原版 | 溢位區以外 3297 案零不符；溢位區 1709 案中不符 1624 |

固定案的原版輸出與手算相符：

| 案 | 輸入 | 原版 AX |
|---|---|---|
| 1 | 一組艦種 0、1500 艘，能力 100 | `0x015E`（350） |
| 2 | 兩組同一筆艦種 1、各 40000 艘，能力 255 | `0x0A94`（2708，艘數合計 80000 回繞成 14464） |
| 3 | 一組艦種 1、65535 艘，能力 255 | `0x0A5C`（2652，不截斷會是 `0x2C7E`） |

## 溢位區

remake 的 `strategicAttack` 與 `FleetAttack`（`planetassault.go`）都用 Go 的 `int`，原版在兩處縮成 16 位元。
兩者只在下列條件之一成立時可能不同，其餘輸入一律相同：

```
艦種 0 或 1 的艘數合計 >= 65536
艘數合計 × 能力 ÷ 100 >= 65536
```

這是 remake 與原版的已知差異，類別 remake-owned。是否有遊戲內路徑能把艘數累到這個量，目前沒有證據。
`+0x40` 的觀察範圍是 78 到 120，此範圍內第二個條件要艘數合計超過 54,613 才成立。尚未查到艦艇記錄 `+0x08` 的上限。
重新開啟條件：找到產生大艘數的路徑，或確認艦艇記錄上限。

## 限制

- 輸入域：槽數 0 到 12、艦艇池索引小於 352。域外原版會讀到鄰近記憶體，C 版沒有涵蓋。
- 只驗常式實作，不驗呼叫時機（`oracle.Call` 的邊界）。
- 回傳值只取 AX。DX 是呼叫後的暫存器殘值，原版不使用。
- `FleetAttack`（`planetassault.go`）是戰場版，輸入結構不同，這一輪沒有對它跑向量。

## 重跑

```sh
decomp/tools/fleet_attack_run.sh <新批次名> 5000
LOGH3_DECOMP_FLEET_ATTACK_VECTORS=/src/workplace/decomp/fleet_attack/<批次名>/vectors.txt \
  tools/go.sh test ./internal/rules -run TestStrategicAttackAgainstOriginalVectors -v -count=1
```

## 後續（2026-10-09）

上面的探針與比對程式是這支函式專用的版本。通用框架（描述檔、讀取集導向的突變、覆蓋率回饋、全部副作用比對、突變對照）見
[`445`](445-differential-test-framework.md)，專用版本的檔案（`fleet_attack_probe.go`、`check_fleet_attack.c`、`fleet_attack_run.sh`、
`check_vectors.sh`、平坦記憶體版 C 檔）已從工作樹刪除，內容保存在 git 歷史（commit `41f9b91` 與 `c6aa066`）。
新框架下 `203A:0424` 的結果：556 案、指令覆蓋 64/64、5 個突變全被抓到。remake 的重放測試改用新向量：
`LOGH3_DECOMP_FLEET_ATTACK=<向量檔> tools/go.sh test ./internal/rules -run TestStrategicAttackAgainstOriginalVectors`，
結果相同：溢位區以外一律相符，溢位區內全部不同。
