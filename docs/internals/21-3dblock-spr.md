# `3DBLOCK.SPR` — 等距方塊

素材：`assets/original/3dblock/sheet-3dblock-48x64.png`（進版控；
以 `tools/py.sh tools/scripts/spr_sheet.py 3DBLOCK.SPR 48 64 4 4 row` 重建；清單與雜湊見 [`assets/original/README.md`](../../assets/original/README.md)）
證據等級：**強證據**（版面與 `SPRITE.SPR` 相同，圖形完整對齊）。

## 版面

1,536 bytes ＝ **48 × 64**，4 位元色號，未壓縮，沒有檔頭。
與 `SPRITE.SPR` 同一套：**每一列依序放 plane 0–3 各 6 bytes**
（`docs/internals/10`），一列 24 bytes、共 64 列。

## 內容

四張48×16圖塊用於交戰下方的分隊投影。321F:06FC依分隊陣營取
檔案+0或+0x300，再依分隊序號／情報條件加+0x180。
不是領地／建設等級；舊用途假說由[研究406](../re/406-fleet-scene.md)的實際消費端推翻。
載入記憶體與原檔1,536 bytes全部相同，SHA-256
ab77e32f4714064a14477a381eec22c373bc567b5b06134ffeab70261e17851a。

proven：321F:07B6–07D5以word(剩餘艘數+499)／500作堆疊數，
傳寬度6 bytes、高16列給2376:07FA；後者每畫一塊把Y減4。
相同來源及投影前後畫面保存在[本輪JSON](data/fleet-scene.json)。
第70輪由[研究407](../re/407-fleet-blocks.md)、[095](../spec/095-fleet-blocks.md)與
[原版收據](data/fleet-blocks.json)閉合透明遮罩、裁切、網格、連線及箭頭。
正式解析器／完整合成器40層＋20裁切零差；Game未接，完整092仍DRAFT。
圖塊零色號透明，每列四個plane分組；背景x8..631／y208..431，圖塊裁切x8..615／y208..429。
第一條亮標沿用圖塊最後留下的平面遮罩8，畫完才恢復15，不是XOR或四平面覆寫。
2376:0B43只有CB返回指令，不搬移背景buffer。裁切零高度的直接注入未返回，玩家可達性未知。

## 怎麼定出尺寸的

先看位元組分佈：只有 22 種值、其中 0x00 與 0xFF 佔 82%，一眼就是
遮罩類資料。再逐一試「一列幾個 byte」，用「每一列的 1 位元是不是
恰好一段連續」當判準——6 bytes 時 **256 列全部**都是一段連續，
其餘寬度都有破口。6 bytes × 4 個 plane ＝ 一列 24 bytes，
1536 ÷ 24 ＝ 64 列。

## `COM_FLAG.SPR`（confirmed）

1,440 bytes ＝ **64 × 45，每列依序放四個plane**。
與3DBLOCK.SPR都是整列plane分組，每個plane的列寬不同：

```
一個影像列 ＝ 32 bytes
    +0   plane 0 的 8 bytes（64 像素）
    +8   plane 1
    +16  plane 2
    +24  plane 3
色號 = p0 | p1<<1 | p2<<2 | p3<<3
```

內容是**五個英文標籤**，一個 9 列高（45 ＝ 5 × 9）：

| 列 | 字樣 |
|---|---|
| 0–8 | `BEEM` |
| 9–17 | `MISSILE` |
| 18–26 | `FIGHTER` |
| 27–35 | `ATTACK` |
| 36–44 | `DEFENCE` |

字是色 4 的填色加色 11 的描邊，背景 0。看名稱與內容，這是**戰鬥畫面的
指令字樣**（`COM` ＝ command）。

第69輪proven：DS:F134載入資料與原檔逐byte相同；321F:0A08實際畫
ATTACK／武器字樣及DEFENCE，寬8 bytes、高9列，保留BEEM拼字。
asset.DecodeFleetSceneSprites依精確長度解析四塊／五標籤，DecodeFleetSceneSprite供單張解析。
既有asset.DecodeTiles在寬度大於16時按word交錯，不能直接解這兩份整列plane素材；
可檢視匯出仍用tools/scripts/spr_sheet.py的row模式，入口見本頁首段。

> `BEEM` 是**原版自帶的錯字**（應為 `BEAM`）。不要「順手修好」——
> 這種東西是保存的對象。

先前判成未解，是因為卡在「16 像素寬、只用到 plane 0 與 1」這個讀法上：
每 8 個位元組的後兩個幾乎一定是 0，看起來像 plane 2／3 沒用到，
實際上那是**第 3 個 plane 那一列的右半邊是空的**——同一組數字，
換一個切法就從雜訊變成字。

## `FRAME.BLK`／`BLOCK_*.BLK` 未解（已排除的讀法）

`BLOCK_0..3.BLK` 各 20,480 bytes、`FRAME.BLK` 5,280 bytes。

已經量到的事實：

- 開頭 `ff ff ff ff 00 00 00 00` **正好重複 40 組 ＝ 320 bytes** 就停，
  所以 320 這個週期是真的（`BLOCK_*` 三份都一樣）。
- 自相關最強的位移：`BLOCK_*` 是 8、160、320；`FRAME` 是 8、16。
- `BLOCK_0` 與 `BLOCK_1` 只有 3,583 / 20,480 個位元組不同（17%），
  是同一張圖的變體。
- 位元組直方圖用掉 **254 個相異值**，只有 `00`（10,987）與 `FF`（4,400）
  特別多。一張四 plane 的 UI 圖不會長這樣——**比較像壓縮過或 8bpp**。

已經試過而且是雜訊的讀法（不必重試）：

| 讀法 | 結果 |
|---|---|
| 四個 plane 以 8 像素為單位交錯（＝ `COM_FLAG` 那種）| 雜訊 |
| 一列 ＝ 四個 plane 各 40 bytes（320 寬 × 128 列）| 雜訊 |
| 一列 ＝ 四個 plane 各 80 bytes（640 寬 × 64 列）| 雜訊 |
| 32×32 圖塊，塊內每列四個 plane | 雜訊 |
| 64×64 圖塊（10 塊），塊內每列四個 plane | 雜訊 |

下一步該問的是「這份資料有沒有壓縮」，不是「切法對不對」——
254 個相異值那一項是判準。
