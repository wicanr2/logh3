# 交戰演出同狀態對拍（S3 第五回合，同盟艦隊打帝國行星雷奧米爾）

日期：2026-09-16

比較等級：`same-state`（整張 640×480，扣掉事先列出的四塊排除區：行星圖 (8,56) 257×102、
艦隊圖 (376,56) 256×102、量表 (297,62) 48×68、戰場 (6,207) 627×226——理由在規格 056 已知差異）。
原版全部是 dosgolem-logh3 `b5485eb`、S3「巴米利恩會戰」帝國，沒有注入（`eng-skip`
那兩批例外：`-poke lin:5DA66@1139500000=00` 把動畫效果關掉）；序列 `tools/engagement_receipts.sh`
（主選單第 3 列「取進度」×2 → 防拷 → S3 (300,282) → 帝國 (318,168) → 系統指令／實行 ×6
→ 每個戰術回合 (240,200) 兩下）。remake 這一側是**直接進入**的診斷比對：戰場模擬走不到
與原版同一個第五回合，所以把那一場的數（3300/3300、22000/22000、擊毀 290）直接排進演出
（`internal/game/engagement_test.go`）；正常路徑的收據等戰術 AI 與亂數對齊之後。

`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
十二份 `pal.bin` SHA-256 都是 `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`。
座標是滑鼠座標（畫面 y ＝ 滑鼠 y ＋ 40）。

| 批次 | 多的那一步 | 畫面 SHA-256 | 差異 |
|---|---|---|---|
| `eng-entry` | 冷啟動 → 達希利戰場第 1 回合（#900M） | `8c5cdcd52597a0920d0298626deb36919bae369e53a39e4cf11dbd0e72f9f88d` | 不比（戰場） |
| `eng-t5` | (240,200) ×8 → #1139M | `aab76fde02de646533feb5b6c0833de4710390deb71cf3de2159f5a793e266c8` | 不比 |
| `eng-open` | 實行 #1140M → #1141.5M 演出開場 | `88cede654131e3d029fd350ff857d8704c32dbe396d6d3f3185ee7f9ae24a7d7` | 0（排除四塊） |
| `eng-text` | #1144M 兩張圖與方塊 | `7de0cbfa522127cef57d37d55d3ec6b9b5f40977b2451ddbd9d145e46a1aabb3` | 不比（圖） |
| `eng-mid` | #1150M | `19c32065a6068aff595267325c873d5c8b0cccf009756ee67e508d74038e6b8d` | 不比 |
| `eng-mid2` | #1156M | `9093adebfb91af37a03a075089a132cbdab73d8df30055aed802310bfd046534` | 不比 |
| `eng-end` | #1169M 擊毀 290、21710/22000 | `c00c7c40ad9c3f51daac5acfe6b320155601e7b83f952548f3c0b14692dae738` | 0（排除四塊） |
| `eng-after` | #1172M 回戰場 | `4e855a5306f3769008fa09a60e0d48961c42d718ce421238e2b43d78e842651f` | 不比 |
| `eng-skip` | 動畫效果關：#1172M | `1e9a68f0a514234bd45efc638c38f0b3f1326efd1716a2b6216c47467b8bb080` | 不比 |
| `eng-skip-mid` | 動畫效果關：#1150M 沒有方塊 | `875e1e189cd3bbe84168045399c108dc30ffab96d34d5929b339141c00433bb2` | 不比 |
| `eng-click` | #1146M 左鍵 → #1152M 已是結果幀 | `2740b8980cdc07527bdfd57209c5a245aa10e33f99264c1981437c682b419de9` | ＝ `eng-end` |
| `eng-rclick` | #1146M 右鍵 → 同上 | `2740b8980cdc07527bdfd57209c5a245aa10e33f99264c1981437c682b419de9` | ＝ `eng-end` |

測試：`TestEngagementMatchesOracle`（`eng-open`、`eng-end`）、`TestEngagementQueuedFromAssault`；
戰鬥曲 `TestEveryTuneIsReachable`。
