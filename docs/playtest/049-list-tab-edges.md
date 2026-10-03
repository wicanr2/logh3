# 049：一覽表分頁邊界的完整畫面對拍

2026-10-03，第59輪，#59。十張完整640×480全圖與活動區均零差，沒有像素排除。
完成範圍是[090](../spec/090-list-tab-edge-display.md)的四格顯示與既有旗標／士氣分支。
情報更新生命週期未驗，#59保持OPEN。

完整凍結收據：[list-tab-edges.json](../internals/data/list-tab-edges.json)，SHA-256
`9a0e7a220fa8b1a91217c3f06b5c8daa0887d36af0c31951e15b70398693c2b8`。
JSON包含每個原版run的完整flags、metadata、RAM／圖／色盤／狀態雜湊，原始指令與來源雜湊，
Game逐檔產物、11份來源指紋及六份測試log。凍結後不得重生來遷就新版本。

## 比較條件

- 原版GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`，只讀。
- 原版bundle為`org_game/银河英雄传说3SP.zip`，SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- dosgolem固定`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`／乾淨，tick20000，mode12h。
- remake父提交`2d3cac5071a3519135b718bf8a48d0a5e8cf9b9f`，執行時工作樹dirty；來源逐檔SHA固定。
- 帝國父狀態為正常新遊戲建隊後420M，S1／帝國；同盟父狀態為正常新遊戲關月初報告後170M，S1／同盟。
- 原版WORK+04為0，九份正式保存WORK+04亦為0，日期796/1/1；原版線性5DA48、正式存檔偏移4分開核對。
- 本輪沒有亂數規則比較，不調種子找結果；原版94DC四bytes及各初始snapshot雜湊保存於JSON。
- 九個案例為明示等價局面注入：原版lin位址與步數完整寫flags，Game修改相同星域／艦隊欄位後走正常Click開清單。
- 第十張同盟清單由主選單→防拷138→兩次右鍵關報告→一覽表→星系資料，沒有局面或RAM注入。
- 星系畫面游標460,125，艦隊士氣460,145，作戰584,70；原版虛擬滑鼠Y+40。
- 不縮放、不裁切，色號及原圖的有效16色色盤均比較。GUI視窗640×480／1280×960，邏輯圖皆640×480。
- 所有案例原版兩跑RAM1048576 bytes、畫面307200 bytes、PNG與色盤相同；兩組同盟盲探針另外記錄，不算正式比較。

## 完整結果

| 案例 | 原版screen.bin SHA-256 | 全圖差異 |
|---|---|---|
| 帝國非零四格 | 15030cbdb4a4a0fb36b8ca478dee48e65e5a4426a2a8b076bdb62c3c14f73d1b | 0 |
| 帝國FF兩格 | 89b8fe811c7b864e567ef4672ad6be3084cd9a3521a98a2b9a9aa6100e3120a8 | 0 |
| 同盟非零四格 | 357d39fa360c8556397cb6b3388ee360445c42b30685cb3d72fa0cb72e0d53af | 0 |
| 同盟FF兩格 | 64ac0924c3985cf88709e392a39ec6f301902670fa6919d9534e2b96a88665cd | 0 |
| 星系戰鬥冒號 | 0413a8f541f8ab0966c8529ea6e3530ce17cb5ac21ec4822dc9eaf17b3c68c94 | 0 |
| 士氣49 | 27c30c1b2c467aa5fc013dc420a49457c1e64fcd74e4db8e31e0b84dcc0fd11a | 0 |
| 士氣50 | 7ec2786405184f6e0e1b09653e9879a894f9b39f1d8753902f0c451818b68ae4 | 0 |
| 作戰bit2+bit4 | 9aea2a31eaa023f4775372ce172893d4a8a5dd7e113c50a1bee2a260aa38d4e3 | 0 |
| 作戰bit4 | 9262128cb8f647338ad5f3946490aed483fd6c5570b5a3162d9cf648345fc3b2 | 0 |
| 正常同盟清單 | ebf406cea30ff5c9f8bc8f470386d39b76b7e2e83d81e59df83dc927923c5326 | 0 |

每張全圖307200 pixels相同，面板320,40,320,380與活動列336,80,272,320皆零差。
最大通道差0，差異bounding box為null。差分圖在`workplace/round59-list-tab-differences/`，逐圖SHA於JSON。
Game雙跑所有輸出檔逐byte相同，四個數值案例的正式STELLAR保存／重讀保留+2A/+2B。
不要求另外兩個導出實數在LoadSlot後保留注入值，原版既有重數契約仍適用。

## 重播與測試

原版使用`bash tools/list_tab_edge_oracle.sh ally-base`建立同盟父狀態。
帝國父狀態由`tools/left_hold_oracle.sh base`產生，沿用固定SHA，不覆寫第58輪輸出。
對上表案例名稱加`-a`／`-b`重播；工具的合法名稱與命令例子見腳本，完整實際旗標在JSON。
新的版本驗收使用新目錄及新收據，不覆寫本輪凍結檔。

正式測試：`bash tools/gui.sh test ./internal/game -run '^TestListTab(EdgesMatchOracle|NormalPlayer)$' -count=1 -v`。
雙跑各11個PASS事件，零FAIL／SKIP。輸出目錄由LOGH3_LIST_TAB_EDGE_OUT指定，生成器只讀。
標準命令：`bash tools/gui.sh test ./cmd/... ./internal/... -count=1`，18產品套件通過。
真實視窗：`bash tools/gui.sh xrun bash tools/list_tab_edge_gui.sh`，一倍／兩倍均307200 pixels相同，已目視。
僅用intro=false跳過開頭動畫，不使用direct-entry、script或debug狀態旗標。
GUI是正常0:0清單抽樣，不把它外推為九個注入局面的真實玩法。

三份IDA共679條原指令／18項proven附加語意；限定證據見[研究397](../re/397-list-tab-edges.md)。
輸出UID/GID1000:1000，所有一次性容器由--rm收尾；逐輪擁有權與清理核對見WORKLOG。
