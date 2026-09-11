# SPEC 032 — 星域態勢重算（`+0x13` 與 `+0x24`）

狀態：**CONFORMED**（純函式通過 32 格同狀態驗證，已接進 `advanceTurn`）
證據：[`docs/re/243`](../re/243-who-fills-the-distance-table.md)、
[`244`](../re/244-region-posture-recompute.md)、
[`docs/internals/22`](../internals/22-stellar-dt.md)、
[`docs/internals/26`](../internals/26-star-region-graph.md)
收據：[`data/region-posture-s3ai6.json`](../internals/data/region-posture-s3ai6.json)
（32 格的 `+0x24` 與 `+0x13`，`tools/scripts/region_posture.py` 從 `s3ai6.bin` 倒出）
相關：[`docs/spec/029`](029-strategic-ai-turn.md)（AI 回合的順序）、
[`docs/spec/006`](006-unrest-at-load.md)（未安定度也吃距離）

## 要做什麼

原版每個回合結束會做這一串（`2703:08DE`，`docs/re/244`）：

```
逐格重算擁有者、易主的走 056A:0855
2B99:0003               ; 重建四張路徑表（距離 ×2、中點 ×2）
for 星域 i in 0..31:
    2703:0933(i)        ; 態勢旗標 + 到首都的成本   ← 這份規格
    2703:12EB(i)        ; 艦艇與資源彙總
    2703:0DA1(i)        ; 未安定度
判勝負
```

**順序不能換**：路徑表是擁有者的函數，所以要先易主、再重建、才重算態勢。

remake 目前讀 `+0x13`（`regionContested`／`regionBlocked`）但**沒有任何
地方寫它**——那個位元組凍在載入時的值，所以「這一格有戰鬥」永遠是檔案裡
的舊資料，艦隊怎麼移動都不會變。

## 算什麼

```
+0x13 &= 0F2h                          ; 只清 bit0 | bit2 | bit3
bit2 = 這一格帝國艦隊數 > 0 且同盟艦隊數 > 0
bit3 = 至少一個鄰居的擁有者跟自己不同
if 擁有者 != 2:                         ; 無主的整段跳過
    +0x24 = 距離矩陣[擁有者][這一格][該陣營首都]
    bit0  = (+0x24 >= 0400h)
```

三個細節照原版：

- **數艦隊時跳過 `+0x00` bit0 有設的那些**（`297B:14F8`）。
- **無主星域不寫 `+0x24`**，留著上一次的值。
- `and 0F2h` 只清三個位元，**bit1 與 bit4–bit7 由別處維護**，不要一起歸零。

bit0 的門檻是 `0x400`（`k >= 1`）。**別跟 `2B99:0392` 的 `0x800` 搞混**
——那個是「從首都走得到嗎」用的（`docs/re/185`）。

## remake

`rules.StarGraph.Postures(owner, fleets) [32]RegionPosture`，配
`RegionPosture.Bits()` 與 `ApplyPosture(舊值, p)`。距離就地用
`BuildTables(owner, 陣營)` 重建，不吃外面傳進來的舊表——那正是這份規格
存在的理由。

同狀態驗證（`TestPosturesMatchOracle`）：拿 `s3ai6.bin` 那一刻的擁有者與
18 支在用艦隊重算，32 格的 `+0x24`、bit0、bit2、bit3 **逐格與原版相同**。

## 接進回合流程

`Game.recomputeRegions()` 在 `advanceTurn` 的回合結束處呼叫它，**在依行星
重算星域擁有者之後**，把 32 格的 `+0x13`／`+0x24` 寫回 `g.res.Stellar`
（整條鏈見 [`docs/spec/033`](033-region-ownership-chain.md)）。

順序照量到的（收據 `flagpass`，`docs/re/245`）：`056A:0469`（bit1／bit2／
bit4，`rules.Occupations`；進入點見 `docs/re/295` 末節）先跑，再跑這一支，
最後才是回合計數器進位。

`advanceTurn` 遇到戰鬥會提早 `return`，那條路徑跳過這一段——見 `033` 末節。
