# 011 — 艦隊短程平移與抵達朝向同狀態收據

日期：2026-09-14
比較等級：`same-state`（只裁決槽 30 的 `+0x48`、`+0x4A`、`+0x4C`、
`+0x56`、`+0x5C`、`+0x5E`、`+0x64`、`+0x66`；不宣稱整張戰場畫面一致）

## 輸入與執行器

- 原版：`GIN3.COM`，SHA-256
  `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
- dosgolem 工作副本：`e0b6a03c5338ea4aa5676e81c5a19825d847ab51`，乾淨工作樹。
- 基底 checkpoint：`logh3-s3short-receipt-ck.gz`，SHA-256
  `bd7ab6affb48a3f83c75b2fdc08c0777a6983443de7185cda85c22dfda2f11b3`。
  它由正常玩家路徑的 `s3short-receipt` 在第 1,573,100,000 道指令保存；該次
  `report.txt` SHA-256 是
  `f4412a5215c9a03aa3e2fbb767717c5a9c4eaaedaf5ecfcddbfb4fe9c90d6f90`，
  內含完整冷啟動按鍵、滑鼠、注入與步數。
- 重播收據：`workplace/oracle/s3short-slot30-east-seq/`。`meta.txt` SHA-256
  `b4d5e614e3e44cdf497ee5cd8c16b489f0f72d488fad27cff0e5e83ad78b08d0`；
  `flags.txt` SHA-256
  `0929a9a7dbabdf2cf4c2a340da5e820da5c66f1aa7f63f507058f4a62893e439`；
  `report.txt` SHA-256
  `a4a983e38f07b401de5f15e50c9cda3badacb052cf1ca9b1f6fd3b48cc4e5156`。

原版記憶體傾印的位址空間是 dosgolem 線性位址。艦隊表為
`ds:C074`（本次 `ds=54B3`，線性 `0x60BA4`），記錄大小 122 bytes；槽 30
記錄起點是 `0x619F0`。重播只在第 1,573,100,100 道指令把目的地
`+0x64/+0x66`（線性 `0x61A54`）改成 `(1181,1163)`：

```text
-load-state /src/workplace/logh3-s3short-receipt-ck.gz
-steps 1574410000
-poke lin:61A54@1573100100=9d 04 8b 04
-dump-mem-at 1573200000:lin:60000:65536:/src/workplace/logh3-s3short-slot30-east-seq-00.bin;1573750000:lin:60000:65536:/src/workplace/logh3-s3short-slot30-east-seq-01.bin;1574100000:lin:60000:65536:/src/workplace/logh3-s3short-slot30-east-seq-02.bin;1574400000:lin:60000:65536:/src/workplace/logh3-s3short-slot30-east-seq-03.bin
-regs-at 1C03:0361 -regs-from 1573100000 -regs-max 30
```

`report.txt` 回讀為 `改記憶體 lin:61A54（61A54）4 bytes`；四份傾印均為
65,536 bytes。沒有改朝向、額度、位置、行動或速度。

## 原版結果

四份傾印的 SHA-256 與槽 30 欄位如下。`+0x4A/+0x4C` 是 byte；不要把相鄰
的 `+0x4B/+0x4D` 合成 word 解讀。

| 狀態 | SHA-256 | `+0x48` | 朝向／目標 | `+0x56` 額度 | 位置 → 目的地 |
|---|---|---:|---:|---:|---|
| `00` | `db5c4a3e6c4672dff73ae15aa7c631584f81ec1e861bd305ab61c1caeb9b4b8e` | `10` | 178／170 | `0x00046C50` | (1169,1163) → (1181,1163) |
| `01` | `7e04ddadf26bfa07825dfd2c5577f91ef1fca330eec30f641d2f1f45249575ae` | `10` | 178／170 | `0x00C80000` | (1177,1163) → (1181,1163) |
| `02` | `f5418e97dd48cfe4e2284be9c646ae24a5b6dc0a13936b2ecd0ad7a2b343f585` | `10` | 178／170 | `0x00C40000` | (1181,1163) → (1181,1163) |
| `03` | `8c831abfc8bc732a2e2180562cb3458f6e733c5926d2bc9731eceb1f2f9567a2` | `10` | 170／170 | `0x00C40000` | (1181,1163) → (1181,1163) |

`00` 是回合補額度前；之後原版把額度補成 208.0。距離 12 大於一步 8：
第一步直指目的地平移 8，朝向仍為 178、額度剩 200；第二步平移 4 到位，
朝向仍為 178、額度剩 196；下一次實行才把朝向 178 → 170。角差 8 不大於
回轉速度 10，所以轉到位不扣額度。這直接區分 `1C03:0361` 與一般先轉身的
`1C03:0185`。

## remake 對照

`internal/rules.TestShortFleetStrafeMatchesControlledReceipt` 使用同一初始狀態，
逐次核對 `(1177,1163,178,200.0)`、`(1181,1163,178,196.0)`、
`(1181,1163,170,196.0)`。舊收據 `s3fltseq` 的距離 3.16 抵達樣本另由
`TestShortFleetArrivalMatchesS3FleetSequence` 核對最終額度 `0x00CCD678`。

兩項規則測試與艦隊命令正常玩家路徑測試均在專案 Docker 工具鏈通過。
此收據只證實短程平移、額度與抵達轉向；分隊動畫、整張 framebuffer、音訊與
實機手感不在裁決範圍。
