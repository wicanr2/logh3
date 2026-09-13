# 第 8 格存檔載入當下的未安定度收據

日期：2026-09-14

比較等級：`same-state`（載入完成後 32 筆星域記錄的 `+0x7D`）

原版執行器：dosgolem-logh3 `e0b6a03`（工作樹乾淨）

原版：`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`

## 方法

批次 `workplace/oracle/load8unrest/` 走正常玩家路徑：主選單「取進度」→
第 8 格「亞姆力札會戰」，在 `#260,000,000` 載入完成後倒出執行期星域表
線性 `0x5F2CC–0x6028C`（32×126 bytes）。完整輸入、執行器版本與畫面雜湊
保存在 `flags.txt`／`meta.txt`。

第 8 格 `WORK.DT8` 的直接輸入為：

- 回合數 `+0x04`：93；
- `Blend +0xC0`：2／2；
- `Rate +0xE4`：2／2；
- 內政專任參謀：提督 32／60（專任參謀表 `+0x37` 的第三組）。

這正是 issue #21 與舊固定參數不同的局面：舊值會拿統治者 16／60、
`Blend` 1／2；remake 現在從剛載入的 `Slot.WorkFields` 與 `WORK` 專任參謀
欄位組成 `UnrestParams`，才呼叫 `Slot.RecomputeUnrest`。

## 結果

`TestLoadSlot8UnrestMatchesReceipt` 由產品的 `loadSlotFromDisk(8)` 正常載入
六份 `DT8` 表，32 格 `+0x7D` 逐格與原版收據相同：

```
19 21 8 21 10 7 29 36 100 34 25 36 25 35 100 43
38 36 34 32 19 11 38 43 36 40 36 34 39 34 32 10
```

收據 `regions.bin` 為 4,032 bytes，SHA-256：
`134b242dd6ca9f63155c5e18e1e173b8a3597b3eb91ce8107aef9992ccce832d`。

本項只裁決載入當下的未安定度重算，不把其餘星域欄位或後續玩家操作納入
完成聲明。
