# 130：全局提督能力編輯

契約：[149](../spec/149-admiral-editor.md)。本輪唯一狀態表：[142](../goal/2026-10-19-admiral-editor-round-142.md)。這是remake新增功能驗證，不宣稱原版編輯器或新增特殊能力具有DOS parity。

## 使用者決定

- 核准三分頁介面，沿現行兩陣營配色與立體框。
- 楊威利與萊因哈特的攻擊加成先答50%，隨即更正為30%，以30%為準；移動加成維持50%。
- 編輯結果全局共用，新遊戲也套用。沿既有偏好檔保存稀疏覆寫，不新增六表欄位。
- 後續指定坎普與波布蘭所在艦隊戰鬥機攻擊預設+30%，確認擔任參謀也生效。這項能力獨立可編輯。

## 功能與邊界

設定頁增加「編輯提督」入口。十項基本能力、三項內部能力、換艦率及移動／攻擊增幅均可編輯。草稿與Game分開，取消不寫檔；套用前驗證能力、派生資料及偏好寫入，失敗保留原局。全局只覆寫編輯過的能力，其餘欄位保留劇本、存檔與升級結果。

三項內部能力依[原版用途證據](../internals/08-admiral-dt.md#沒有名字那三格在做什麼)顯示，原版欄名仍未知。畫面使用用途說明，不把標籤當作原名。營運與內部除數不允許0，能力上限200。

全軍加成由存活且在場的帶隊旗艦提供，取同陣營最高值。攻擊只改戰術艦隊交火係數，保持三武器庫存、原傷害流程與動畫時鐘。既有名將移動總開關、轉移旗艦總開關與原版模式閘門保持；行星反擊、要塞砲及戰略攻城不放大。

新增第四項特殊能力「本艦隊戰鬥機攻擊」，提督或參謀均可提供，取兩者較高值，僅強化模式武器2套用。坎普索引10、波布蘭74，預設30%；其他0%。與全軍攻擊倍率順序乘算，各自取整。這是使用者指定擴充，沒有新增原版隱藏效果的證據聲明。

## 驗證方法

正式GUI由[鍵鼠重播器](../../tools/scripts/admiral_editor_gui.py)從片頭走到S1、設定、編輯頁，沒有腳本入口、狀態或seed注入。首次啟動驗帝國基本／內部／特殊、數值輸入、取消提示、套用及原版／HD；第二次正常啟動驗同盟楊威利預設30%，再查帝國編輯值已跨重啟及S1讀檔保留。新增坎普、波布蘭各30%預設，分別改為65%／75%並保存，坎普65%跨第二次正常啟動與S1讀檔核對；波布蘭跨重啟序列另由偏好檔測試驗證。ELF身份從執行中程序取SHA-256，兩次都與獨立凍結二進位一致。

新版批次為`workplace/admiral-editor-gui-r6/`。十一張完整960×720截圖、輸入／狀態及偏好保存於`RESULT.json`。另驗武器能力1被拒絕，原局與偏好不變；下限由實際16艦種／256角度公式推導為3。最終來源與圖像身份由[唯一索引](../images/admiral-editor.json)保存；[核對器](../../tools/scripts/admiral_editor_receipt.py)可重核。前版r5-r2收據及標準日誌分別留在該舊目錄的`exported-receipt.json`與`standard-tests.log`，不混用來源身份。

```sh
tools/gui.sh build -o workplace/admiral-editor-gin3-r6 ./cmd/gin3
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=tools/with-xvfb.sh tools/gui.sh xrun python3 tools/scripts/admiral_editor_gui.py --binary /src/workplace/admiral-editor-gin3-r6 --out /src/workplace/admiral-editor-gui-r6
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=tools/with-xvfb.sh tools/gui.sh xrun python3 tools/scripts/admiral_editor_receipt.py --check
```

重跑GUI需改用新的獨立輸出目錄，不能蓋掉已綁定收據。最後一項核對器針對本批凍結身份，不把其他批次自動升格。

## 草稿對照與限制

實際畫面：[帝國](../images/admiral-editor-empire.png)、[同盟](../images/admiral-editor-alliance.png)、[內部能力](../images/admiral-editor-internal.png)、[坎普](../images/admiral-editor-kemp.png)、[波布蘭](../images/admiral-editor-poplan.png)。三武器實際戰損、RNG／庫存、全局新局／讀檔、艦隊派生與六表保存沿標準測試驗證。戰鬥機加成另驗兩角色、任意編輯提供者、同艦隊取最大值、原版模式停用、正式Game傷害命令帶入，以及只有武器2增加實際戰損。光線砲及飛彈整份命中結果保持不變。

[核准草稿原圖](../goal/2026-10-19-admiral-editor-round-142.md#核准草稿原圖)保留原樣。正式頁面沿相同外框、頭像、分頁與底部操作配置，補齊14字格標題、16字格姓名及10字格說明的大小與靠左排列。依使用者追加需求，特殊能力改四列、列距40，基本與內部能力保持。對照為`layout-only`，草稿與正式S1的等級、能力、日期、背景、攻擊待定／30%、第四列及游標狀態不同，不宣稱整張逐像素一致。

本輪不重打包、不建立Release、不修改#48。Windows／macOS實機及原版DOS新增行為不在本輪驗收範圍。

r6最後標準測試19套件15 PASS／4無測試，日誌為`workplace/admiral-editor-tests.log`，其SHA-256保存於來源索引。這次沒有以正常GUI新增實戰觸發收據；戰鬥機實際傷害與兩角色觸發由固定局面規則測試及正式Game傷害命令測試證明，正常GUI驗證編輯、保存及讀回。

後續[131正式交付](131-admiral-editor-delivery.md)延伸同一GUI工具，加入正式包、Wine及錄影選項。此頁與r6索引保留原凍結工具身份，現行封包及工具核對以131的來源索引為準，不以後續工具雜湊套回舊收據。
