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
`pal.bin` SHA-256：`eng-entry`／`eng-t5`／`eng-skip` 是 `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`，
其餘九份（進演出之後）是 `765462f42f56e5e5d135948bf27f15c171abe8824290b1ae7a73fd445e8f73f9`（各批 `meta.txt`）。
座標是滑鼠座標（畫面 y ＝ 滑鼠 y ＋ 40）。

| 批次 | 多的那一步 | 畫面 SHA-256 | 差異 |
|---|---|---|---|
| `eng-entry` | 冷啟動 → 達希利戰場第 1 回合（#900M） | `8c5cdcd52597a0920d0298626deb36919bae369e53a39e4cf11dbd0e72f9f88d` | 不比（戰場） |
| `eng-t5` | (240,200) ×8 → #1139M | `aab76fde02de646533feb5b6c0833de4710390deb71cf3de2159f5a793e266c8` | 不比 |
| `eng-open` | #1141.5M 演出開場（tick 7，還沒有圖） | `88cede654131e3d029fd350ff857d8704c32dbe396d6d3f3185ee7f9ae24a7d7` | 0（排除量表與戰場；`TestEngagementMatchesOracle`） |
| `eng-text` | #1144M 兩張靜圖（tick 20） | `7de0cbfa522127cef57d37d55d3ec6b9b5f40977b2451ddbd9d145e46a1aabb3` | 同上 |
| `eng-mid` | #1150M（tick 50，PPAB 第 4 格之後） | `19c32065a6068aff595267325c873d5c8b0cccf009756ee67e508d74038e6b8d` | 同上；`TestBMGEngagementMatchesOracle` 兩個視窗 0 |
| `eng-mid2` | #1156M（tick 80，PPAB 第 8 格之後） | `9093adebfb91af37a03a075089a132cbdab73d8df30055aed802310bfd046534` | 同上 |
| `eng-end` | #1169M 擊毀 290、21710/22000（tick 145；戰場區截在 `321F:0897` 的暫態） | `c00c7c40ad9c3f51daac5acfe6b320155601e7b83f952548f3c0b14692dae738` | 同上 |
| `eng-after` | #1172M **仍在演出**（停頓中，tick 160；`docs/re/331` §八） | `4e855a5306f3769008fa09a60e0d48961c42d718ce421238e2b43d78e842651f` | 同上 |
| `eng-skip` | 動畫效果關：#1172M | `1e9a68f0a514234bd45efc638c38f0b3f1326efd1716a2b6216c47467b8bb080` | 不比 |
| `eng-skip-mid` | 動畫效果關：#1150M 兩個視窗只剩底圖（戰場方塊照畫） | `875e1e189cd3bbe84168045399c108dc30ffab96d34d5929b339141c00433bb2` | 0（`TestEngagementAnimationOffMatchesOracle`） |
| `eng-click` | #1146M 左鍵 → #1152M | `2740b8980cdc07527bdfd57209c5a245aa10e33f99264c1981437c682b419de9` | ＝ 沒按鍵的 #1151.15M（PPAB 第 5 格之後；`docs/re/328` §六），≠ `eng-end` |
| `eng-rclick` | #1146M 右鍵 → 同上 | `2740b8980cdc07527bdfd57209c5a245aa10e33f99264c1981437c682b419de9` | 同上 |
| `eng-morale-low` | （第 5 段，從 `eng-t5`）#1170M 把第１３艦隊 `+0x00` 寫 05、`+0x3C` 寫 40 → 第五回合結算最後一步 #1197M「士氣值低落」臨時視窗（`3826:0405`） | `918b92643f65e17c56bbc24f37fdc73f9e54b6a82a08cc9dc1e0f81cdcea379e` | 視窗內 0（其餘戰場到不了同一局面，規格 059） |

| `eng-known` | （第 6 段）同 `eng-text` 的 #1144M，但七筆同盟艦艇 `+0x17` 寫 3：右框印「第16艦隊　費雪」而不是「第？艦隊」（#74）| `614bc7da162865a49918273e32e2688addc6c4d320e894fe3d52e00bdcc7cd3d` | 0（`TestEngagementTwoDigitMatchesOracle`，排除區同 `eng-text`）|
| `eng-annihilate` | （第 7 段，從 `eng-t5`）#1187.90M 把帝國第１３艦隊第 0 組艘數（`lin:6811A`）寫 0 → #1188.40M「帝國軍第13艦隊舒坦梅茲／已經全滅。」臨時視窗，**沒有頭像** | `803ec3880500ace39075fc5e704b62a13ee4fd359e73102a71e0bb5245083bee` | 視窗內 0（`TestBattleAnnihilationReportWindowMatchesOracle`）|
| `eng-annihilate-mem` | 同一條，mem-in #1187.91M／mem-out #1234.0M（報告等滿 90 tick 才跑 `185C:00F0`）| `6b753819f22269e5a6630472d43f5fa97f85b640b68be1f1c274e627834c3a6d` | 逐欄對，結果在 `docs/re/339` §九 |

測試：`TestEngagementMatchesOracle`（六批，步數換 tick 的公式在測試開頭註解）、`TestEngagementClickQueuesUntilPause`、
`TestEngagementAnimationOffMatchesOracle`、`TestEngagementQueuedFromAssault`；
戰鬥曲 `TestEveryTuneIsReachable`。
