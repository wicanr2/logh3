# 第145輪：前線駐留後重新出擊與人事選單點擊

使用者回報Windows v1.5同盟試玩。前線駐留後領不到當地艦艇，更換參謀或副官
常要重複點擊。本輪修正既有契約，不改遊戲規則、存檔格式或已發布版本。

## 目前狀態表

| 工作 | 狀態 | 證據 |
|---|---|---|
| 前線編成來源 | 兩陣營確認／取消與保存讀回通過 | `TestFrontlineGarrisonSortieStock`；[015](../spec/015-fleet-composition.md) |
| 人事渲染 | 有界字模快取已實作，冷／熱快取與純量逐像素相同 | `TestClassicGlyphCacheMatchesScalar`；[135](../spec/135-original-style-hd-ui.md) |
| 正式測試 | 19套件，15 PASS／4無測試 | `tools/gui.sh test ./internal/... ./cmd/...` |
| 真實人事點擊 | 最終Linux兩陣營、Windows Wine同盟各28張完整視窗零差，共58060800像素 | [正常劇本](../../tools/scripts/personnel_click_plan.go)、[GUI核對器](../../tools/scripts/personnel_click_gui.py) |
| Windows交付 | v.1.5.1-20261024本機替換包已驗，8048235 bytes | [封包腳本](../../tools/scripts/personnel_windows_patch.py)；只替換gin3.exe，原1.5素材與存檔沿用 |

最終收據由[獨立核對器](../../tools/scripts/personnel_patch_receipt.py)產生到
[來源及GUI索引](../internals/data/personnel-frontline145.json)，核對封包、PE架構、逐檔manifest、
凍結來源、Wine實際執行EXE與三批各28張完整視窗。

交付：[Windows替換包](../../dist-all/v.1.5.1-20261024/release/LOGH3-v.1.5.1-20261024-windows-amd64-patch.zip)。
SHA-256：`104d549cfa0f938603b49a1fd486540d4ce6c9ddb1f13627c2b3cbf244ee1407`。
實際EXE的`-version`輸出一致。ZIP逐檔、PE amd64、UTF-8 BOM／CRLF與無原版素材邊界通過。
只在本機交付，未建立tag、Release或推送。Windows為Wine，尚待使用者真機驗收。
本輪容器皆已退出並自刪；整個工作根無root-owned檔或誤建Markdown目錄，抽查均1000:1000。

根因：玩家出擊在取旗艦、開編輯器、取消退船、新艦隊位置與清單選區多處使用
`CapitalRegion`。駐留原本已將船艦退回所在地。修正使用`ADMIRAL +0x0A`，
沿研究[215](../re/215-ship-allocation-from-region-stock.md)與
[329](../re/329-fleet-disband-is-garrison.md)既有證據。

效能探針`LOGH3_PERSONNEL_TIMING=1 tools/go.sh test ./internal/game -run
TestPersonnelRenderTiming -count=1 -v`在`golang:1.24-bookworm`、4CPU／4GiB量測。
原先參謀清單／詳細、副官清單／詳細約34.36／92.90／134.25／91.77ms；
修正後32.71／36.98／38.46／34.17ms。這是開發環境診斷，不宣稱Windows真機手感。
CPU收據在`workplace/personnel145-before.pprof`，不是原版oracle。

字模快取後的短脈衝壓測仍能漏掉按下事件，因此另補靜態背景、游標及副官小選單快取。
最終正式HD人物資料探針四項為8.25／8.58／10.48／9.27ms，包含首個背景建立。
實際視窗移動游標常見2至3ms。短脈衝壓測的失敗收據仍保留，不能宣稱所有80ms
事件必定被虛擬顯示收到。最終三批GUI採既有事件回授方法：一次按下、觀測到後放開，
沒有重送。它驗證正常換人、取消、確認與完整呈現，不替代Windows真機主觀手感。

參謀背景舊路徑重畫游標造成27像素抗鋸齒差異，已按064的最上層游標契約修正。
差分及原因見135，不改原色畫面，不將舊基準聲稱為全圖零差。最終基準與整張X11核對
使用單次游標，沒有遮罩、裁切或排除區域。

本輪沒有新增DOS同狀態收據；前線回歸用明示的非首都診斷局面，重新出擊與編輯均走
Game公開點擊。正式GUI由主選單開始，不注入局面、不使用`-script`。
GitHub快照102張、#48 OPEN；開工282條verify零失效。未寫Issue、tag或Release。
