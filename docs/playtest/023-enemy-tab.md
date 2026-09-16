# 右下角「敵艦隊」分頁有列的同狀態對拍（S3 帝國第一回合，改記憶體造的局面）

日期：2026-09-17

比較等級：`same-state`（整張 640×480，沒有排除區）。原版全部是 dosgolem-logh3 `b5485eb`、
S3「巴米利恩會戰」帝國第一回合的戰略地圖（`docs/re/324` 同一條序列停在 #300M，檢查點
`enm-map`），之後每批在 #300M 改記憶體（規格 014：要局面就改記憶體，不靠亂數）：
同盟第２艦隊（槽 17）`+0x06 ← 0`（瓦爾哈拉）、它四組艦艇的 `+0x17` 寫成要的索敵等級；
`enm-mix` 再加士氣 `+0x3C ← 30` 與槽 16（六組全 2）。位址與算法在
[`docs/re/325`](../re/325-enemy-tab-rows-and-window.md) §一，序列 `tools/enemytab_receipts.sh`。
remake 這一側是**直接改狀態**的診斷比對（`internal/game/enemytab_test.go`：S3 走
`loadSlotFromDisk(10)`，改 `g.fleets` 的 Region 與各組 `Detail`）；正常玩家路徑
（打過仗回到戰略地圖）的收據等戰術層與原版對齊到同一回合。

`GIN3.COM` SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
十四份 `pal.bin` SHA-256 都是 `28b933fe81e79a380123b932302ba2dd8983bd3ac85a8f922e9e3b6ac1bd02e8`。
座標是滑鼠座標（畫面 y ＝ 滑鼠 y ＋ 40）；每一下都先移過去、10M 步之後才按
（`docs/re/318` §六）。

| 批次 | 改記憶體／多的那一步 | 畫面 SHA-256 | 差異 |
|---|---|---|---|
| `enm-map` | 冷啟動 → S3 帝國第一回合 #300M | `69a5fbf714d616c48a0ba08e10c98fa2bf55d25d083a2e4f24040feb6b1d9bcc` | 不比（載具） |
| `enm-l3` | 槽 17 進瓦爾哈拉、四組 3；敵艦隊 (490,250) | `da3b7f595282f13f5d42b727ecebc4607ff9bd12fffb1fa8d139120088338015` | 0 |
| `enm-l3-row` | 列 0 (450,268) → 視窗 | `e5e914ded8bdef1fb397c0ce627f2f0409cf223f998211ee5c3c4d9ba28ba2c6` | 0 |
| `enm-l3-ok` | 結束 (488,245) | `960819273f3aae93aa9e66446f515c6fe3e64771bea7ab8509dc6a092be9b579` | 0 |
| `enm-l3-rclick` | 右鍵 (300,200) | `18ca8760677226a3e9b304d3bef260d8675043c7acc43608f7501172417aef18` | 0 |
| `enm-l1` | 四組 1；敵艦隊 | `a3ab95b16a7489cb814abec00620835832207fb4a8dc096e824e5a5830dd73c2` | 0 |
| `enm-l1-row` | 列 0 → 「情報資料來源不足」 | `00fc1f3faab123252c8f9d1783c3c713c76f7432c82e60b75f0a9456e7669e5d` | 0 |
| `enm-l1-row-miss` | 框內別處 (300,200) 左鍵 | `fc8d5f2f3b8a036d05b9aaf988b3f2b9f4833b5b0ad5503662e921b2b97b039e` | 0 |
| `enm-l1-row-rclk` | 右鍵 (300,200) | `fc8d5f2f3b8a036d05b9aaf988b3f2b9f4833b5b0ad5503662e921b2b97b039e` | 0（＝ `-miss`） |
| `enm-l1-row-ok` | 確認 (432,230) | `99bd840c22740f573ff5e9fcd6f2e019a66651b573c3093cc3e3cc29ae40fdaf` | 0 |
| `enm-mix` | 槽 17 四組 3／2／1／0、士氣 30；槽 16 六組 2 | `2d23fe412a4cc5b3b1e32d5016ebb691538cc6c86f09d0ebe2219b140c3bd71a` | 0 |
| `enm-mix-row0` | 列 0（槽 16）→ 訊息框 | `4f1e655cea57b128ecc2fe9aa73482f705420997aaccf6fee7e92149ef7f585a` | 0 |
| `enm-mix-row1` | 列 1（槽 17）→ 視窗「ＢＣ？」 | `8eceffdc111b5bd06aca09d89799afd7677add03bc41603b8194b1cab15cc10a` | 0 |
| `enm-l0` | 四組 0；敵艦隊 | `6937d5b989f7b83dcc25d2c8ebcedb7f668668dead3351d03d010abe22ce8024` | 0 |

每批另倒字元平面（`plane.bin`）、槽 17 那一筆（`fleet17.bin`）與五個 peek
（艦艇池指標 `lin:61AE4`、`ds:9588`、`ds:8F1F`、`ds:95B4/95B5`、槽 17 `+0x70`），
筆數與等級都與畫面一致。

差異分類：全部 0。做的時候修掉兩處 remake 的既有偏差：列按下去的底色蓋到了鈕左邊
兩個像素的亮邊（`fleet-list-panel.json` 的 `row.x`／`w` 改 338／270）；清單的視窗或
訊息框開著時左下角仍跟著游標重算（`refreshHoverRegion`）。
