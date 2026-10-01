# 040 — 原版自動補給排序的受控驗證

日期：2026-10-02；工作項[#79](https://github.com/wicanr2/logh3/issues/79)。
身份、工具、位址基準、完整命令與限制：[研究368](../re/368-battle-preparation-sort.md)。
固定欄位：[battle-preparation-sort.json](../internals/data/battle-preparation-sort.json)。

比較等級：**原版常式診斷**；本輪沒有remake正式畫面比較，不套same-state畫面通過標籤。
兩側固定種子／初始狀態相同的診斷模型，只驗指定候選排序，不能替代正常玩家入口。

- 原版75案雙跑完整收據SHA-256：`370cfe8b8c7ca2efc1d23e8db6f4dab646be29cda33b2b3123b737a3fc49cb99`。
- 72個排序案例逐索引符合可丟棄模型；另外三案確認行星資源給到的艦隊及來源指標。
- 固定種子`0x12345678`在75案全都維持原bytes；不是重擲到結果一致。
- 受版控欄位SHA-256：`9df9ec35eb56af3e9445cff7be6359b4107803a2528fcfe2d922ccd0a7c8292f`。
- 來源／輸出與工具SHA均保存，原始完整注入與資料庫只留本機。

穩定排序反例：混合七筆得到`[4,0,1,2,3,5,6]`，原版為`[4,0,5,6,1,2,3]`。
不能只以「來源艦隊先拿」的測試放行另一種排序。

`bash tools/gui.sh test -count=1 ./internal/... ./cmd/...`標準18個套件通過或無測試；
本機`product-tests.log` SHA-256
`ae1e90875876038921f4d7e84394f1904cd0c112a6219b603d3fb4d2bac064aa`。
工具Python語法通過、四份舊文件回填入口由欄位匯出程式自動檢查。
產品測試只證明既有remake回歸，本輪沒有新增正式畫面或存讀驗收。
輸出檔與目錄UID/GID1000:1000，161975個工作樹項目未發現root-owned或同名`.md`目錄。
本輪容器皆已退出刪除；同工具image的另一專案容器經掛載核對後保留，沒有全域清理。

## IDA分級匯出核對

工具為IDA Pro 9.4，沿用`ida-pro-9.4-idapython:locked-v1`，UID/GID1000:1000。
新的切段副本透過`tools/ida.sh raw idat -A -L... -S...`，查詢如下：

| 段 | 查詢（位移:長度，十六進位） | 匯出SHA-256 |
|---|---|---|
| 3C6A | `000D:111 011E:3A` | `3e26d59caed90e4ea3d8fb32b538a2ac940c7accc90829d7a38a1a14e250d26e` |
| 1930 | `019D:117 0510:DC` | `74ade3d4095a4f5e0d017ff1e5ddbb58bba59fca58255fd3ae724de131c4eb66` |
| 1AAB | `006D:123 04BD:C6` | `01d382c0badb83f01ff048151b0822b0d888dcb2fb8794f6e31a6ffc96ecc905` |
| 1261 | `000B:5C` | `0f9b4db2eb8cbcd604be67dc3ef71ffa9a91b4a145fda3257e2afd67570e2217` |

各自傳`--index=/work/re79-sort-reviewed-index.json`；匯出程式為
`/idascripts/battle_rules_ida.py`，輸入為`/work/re79-sort-reviewed-{段}.com`，
輸出`/work/re79-sort-reviewed-{段}.json`。原始CS:IP、IDA位址與dosgolem線性位址逐列分開。
七個函式、十三項審查語意皆合併；其餘逐項顯示unknown警示，不把自動名稱當證據。
核對JSON在本機`workplace/oracle/battle79-sort/ida-check.json`。

重驗欄位命令見研究368；新版工具不能靜默覆寫舊收據。正式第一階段仍DRAFT，
首次來源狀態、人類結果顯示、士氣報告續接、正式回合與存讀必須接續；#79保持開啟。

## 2026-10-02／研究369續證

來源局部診斷與六組第一階段結果／士氣收據見
[驗證041](041-battle-preparation-lifecycle.md)、[研究369](../re/369-battle-preparation-lifecycle.md)。
本篇75案仍只驗原排序；換戰場／重新開始／讀檔重設與正式接線仍待完成。
