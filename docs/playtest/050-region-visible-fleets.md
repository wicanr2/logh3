# 050：可見艦隊數與正常命令入口

2026-10-03，第60輪，#59。契約見[091](../spec/091-region-visible-fleets.md)，
完整證據見[region-visible-fleets.json](../internals/data/region-visible-fleets.json)。
上一輪十張邊界收據保持原雜湊；讀檔契約勘誤見[研究398](../re/398-region-visible-fleets.md)。

| 案例 | 方法 | 結果與限制 |
|---|---|---|
| normal | 正常建隊父狀態→實行→一覽表 | 原版雙跑七檔相同；Game雙跑畫面及七個保存檔相同；全圖零差 |
| late1 | 043B入口後注入艦艇Detail1 | +2A=1，全部64格與原版相同；受控規則 |
| late255 | 注入DetailFF及艦隊陣營1 | +2B=1，64格相同；實數在注入前已重算，不驗全圖 |
| clear | 一回合前注入Detail1 | 到達配置先清零，命令入口再清過期兩格；原版雙跑 |
| peace | 0AF2入口注入Detail1／Action2 | 原版兩跑清理前後RAM相同；正式helper兩張整表相同 |
| 保存／載入 | 正常Game建隊後明示索敵／陣營變更 | 保存不改快照；LoadSlot依保存Detail重算；兩個陣營均驗 |
| GUI 1×／2× | 真實X11鍵鼠由主選單建隊、實行及開清單 | StateStarList，視窗640×480／1280×960，邏輯圖皆640×480，零差 |

正常screen.bin SHA-256：257b2e6ffdd365d35055888838cb14ebbccb4f461eca6421ab53f139b7fee265。
全圖307200 pixels相同，面板320,40,320,380與活動列336,80,272,320均零差；
最大通道差0，bbox=null，沒有排除區。差分圖在workplace/round60-visible-differences/normal.png，雜湊於JSON。
色盤從PNG讀有效屬性色號，不以DAC pal.bin取代完整顯示色盤。

dosgolem固定f7523f9／clean，tick20000，mode12h，不裁切或縮放。
remake父提交22af102，測試時dirty，JSON保存16份來源SHA。
兩側正常初始seed D3DC167E，不重擲或挑結果；導出規則不消費亂數。
原版虛擬滑鼠Y加40對應產品座標；游標460,125，回合1，日期796/1/4。
GUI走主選單→138→月初報告→編成→實行→一覽表；沒有script或direct-entry，intro=false僅略過開頭動畫。
晚注入只作規則隔離，不外推正常全圖。雙人換邊只驗資料接線。

## 重播

- 原版：`bash tools/region_visible_fleets_oracle.sh <normal|late1|late255|clear|peace>-<a|b>`。
- 重點：`bash tools/gui.sh test ./internal/rules ./internal/game -run 'Test(RegionVisible|ListTabEdges|ListTabNormal)' -count=1 -v`。
- 正常圖／保存：LOGH3_REGION_VISIBLE_OUT指定新的絕對目錄，執行TestRegionVisibleFleetNormalCommand。
- 真實GUI：`bash tools/gui.sh xrun bash tools/region_visible_fleets_gui.sh`。
- 凍結：`bash tools/gui.sh xrun python3 tools/scripts/region_visible_fleets_receipt.py`，已有輸出即拒絕覆寫。
- 標準：`bash tools/gui.sh test ./cmd/... ./internal/... -count=1`。

固定父狀態沿第58輪重播，不覆寫凍結輸出。完整旗標、metadata與檔案SHA保存於JSON。
初次-poke分隔符及相對輸出路徑失敗屬驗證腳本問題，已修正後用同容器重跑。
一次性容器UID/GID1000:1000、資源有界、網路關閉；原版RAM與圖片留本機，Git保存方法與雜湊。

凍結收據SHA-256：07b5c05e76c3435d84822308bc17667157c91e062f6d3dc31dded80a84c7b3c0。
18產品套件通過，focused零FAIL／SKIP；原版與Game雙跑及正常GUI均通過。
