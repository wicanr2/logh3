# 048：左鍵保持與同結果全圖

日期：2026-10-03。工作項：[#60](https://github.com/wicanr2/logh3/issues/60)。
契約：[089](../spec/089-left-button-repeat.md)，研究：[396](../re/396-left-button-repeat.md)。
原版補證：[left-button-repeat.json](../internals/data/left-button-repeat.json)，
SHA-256為4873d0449777bd1ae6be2e558c82493b582c32f38f96d8d1e20d96e4715dfc5a。
正式成果：[left-button-repeat-game.json](../internals/data/left-button-repeat-game.json)，
SHA-256為8020036870c8441d36371c182703abde7b693487ac7879fbac5451cadbc3b0ae。

## 核准差異與來源

使用者選定清單按下立即一列，350ms後每100ms一列。
顯示層按幀更新，時間邊界量化至下一個更新幀。
原版清單沒有固定延遲，速度隨重畫成本變動；不再要求相同保持指令數得到相同列數。
比較等級為same-state，僅限明示等價輸入後的靜態全圖，不宣稱保持耗時一致。
徵稅／預算四箭頭每次一級或一點，約137.3ms沿PIT2048、十IRQ一倒數及八倒數契約。
此時間是hardware-spec approximation，原版首間隔與重畫相位不作逐週期聲明。

原版GIN3.COM為770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918，
bundle為de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
dosgolem固定f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean；本輪未改執行器。
父提交2c44b828097d9680e689419089809e6cdde36efc，建置時工作樹未提交，精確來源指紋在成果JSON。
工具為既有logh3-ebiten:2、Go 1.24及IDA Pro 9.4 locked-v1；原始位址空間分欄保留。

## 結果

每張保存未縮放原圖、remake圖、差分圖與全畫面／活動區統計；沒有像素排除區。
原版與Game各自雙跑，原版完整RAM／色號／PNG／色盤相同，Game每份輸出逐byte相同。

| 畫面 | 原版保持 | Game等價保持 | 結果 | 全圖差異 |
|---|---|---|---|---|
| 右下提督清單 | 300000指令 | 650ms | 位移5 | 0／307200 |
| 星系一覽表 | 300000指令 | 350ms | 位移2 | 0／307200 |
| 徵稅檔位 | 3000000指令 | 約137.3ms | 2→4，WORK+C0保存4 | 0／307200 |
| 建造費百分比 | 3000000指令 | 約137.3ms | 25→27，WORK+7E保存27 | 0／307200 |
| 月初提督清單 | 300000指令 | 350ms | 位移2 | 0／307200 |
| 正常同盟首戰資料 | 第57輪固定偏移26 | 向下保持300幀 | 40列、上界26 | 0／307200 |
| 正常同盟首戰資料 | 第57輪固定偏移0 | 向上保持300幀 | 回到0 | 0／307200 |

首戰使用正式bootstrap／Game，主選單→138→S3同盟→實行→808 tick首戰→資料。
沒有RAM或局面注入，保持後六份原版存檔表與INDEX寫出；每個偏移的九檔雙跑相同。
其他五張使用實際原版素材的測試資源，沿正常新遊戲及命令輸入；不擴張成所有啟動流程驗收。
稅率／百分比的存讀與上下界由TestLeftHoldEconomicTimingBoundsAndMove核對。
月初報告未結算時保存仍被拒絕，不用保持功能繞過既有閘門。

正常GUI使用真實X11鍵盤與滑鼠，從主選單走到同盟首戰。
唯讀比對右側「輸入」標籤612像素後再開資料，按住下箭頭五秒後放開。
一倍640×480、兩倍1280×960均取邏輯640×480，游標440,390，完整色號與正常Game及原版零差。
兩份PNG已目視；不宣稱其他GUI畫面或所有輸入模式均已驗。

## 根因與修正

- 顯示層補LeftHold／LeftRelease，只重複捕獲的箭頭動作。放開、另一按下、右鍵、分頁或模態切換取消捕獲。
- 右下清單與副官候選短按改為一列，舊五列／十六列收據以等價保持輸入核對，原始收據不改。
- 待命提督與參謀候選先辨識箭頭，再做列選擇，避免把箭頭當最後一列。
- 月初清單圖塊寬8、待命清單內框寬14；DS:1BF2證實共同熱區寬16。
  命中由相同右緣導出，游標改用正式座標。舊量測混入兩列框線的25個游標像素由產生器清理。
  月初全圖249→25→0，沒有遮罩、裁切或重寫原版收據。
- 首次兩倍GUI的固定等待過短，尚未開資料窗。改以畫面就緒觀測，乾淨重跑同流程兩種倍率通過。

月初保持後的右鍵探針同時移出熱區，兩跑畫面仍留清單。
未隔離移動與事件接受，不能證明原版右鍵不關窗；返回契約為unknown，沿研究323的事件時序限制。
兩份名稱含close的探索收據只保存實際輸出，完整旗標留在成果JSON。
月初測試的狀態斷言核既有Game行為；失敗訊息中的「原版右鍵保留」不構成原版語意證據。
經濟移出／返回的remake回歸是處理端驗證；原版返回熱區未做逐相位配對。
BuildPlan十四顆造艦比例箭頭、月初清單其他控制、戰場資料列指向鏡頭及其餘八步仍不由本輪升格。
原版整份RAM未與remake保存逐byte相同；第57輪指標差與第56輪FF字鍵裁決保持。

## 重跑入口

全部工作走Docker包裝器，原版唯讀，輸出UID/GID1000:1000。

```sh
bash tools/left_hold_oracle.sh base
bash tools/left_hold_oracle.sh tab
bash tools/left_hold_oracle.sh star
bash tools/left_hold_oracle.sh tax
# 子案的完整旗標見固定補證；既有輸出拒絕覆寫。
bash tools/monthly_admiral_hold_oracle.sh parent
bash tools/monthly_admiral_hold_oracle.sh a
bash tools/monthly_admiral_hold_oracle.sh b
bash tools/monthly_admiral_hold_oracle.sh close-a
bash tools/monthly_admiral_hold_oracle.sh close-b
bash tools/gui.sh test ./internal/game -run '^TestLeftHold' -count=1 -v
bash tools/gui.sh xrun bash tools/left_hold_gui.sh
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
```

正式Game輸出使用LOGH3_LEFT_HOLD_OUT、LOGH3_LEFT_HOLD_COMPARE_OUT與LOGH3_LEFT_HOLD_MONTH_OUT；
各環境變數要在容器命令內設定，先建前兩者父目錄，個案輸出必須不存在。
GUI預設讀workplace/round58-hold-player-b/battle-ready模板；新批次輸出用LOGH3_LEFT_HOLD_GUI_PREFIX。
固定收據由[凍結工具](../../tools/scripts/left_hold_game_receipt.py)產生，既有JSON與差分拒絕覆寫。
舊月初規格清理由[產生器修正](../../tools/scripts/monthly_admiral_cursor.py)套用一次，重複套用拒絕。

最終保持回歸每跑8個主測試、6個子案例，兩跑無失敗或跳過；標準18產品套件通過。
測試只能證明已列範圍，整份088、正常索敵配對、音樂聆聽／macOS與正式交付仍待完成。
