# 117：正常玩家公開退出

狀態：**CONFORMED**，限定正常 GUI 退出接線。2026-10-05 沿 [#90](https://github.com/wicanr2/logh3/issues/90)，補既有 Game 退出旗標的 GUI 接線。

## 證據與審查

已證實的原版契約在[研究341第四節](../re/341-fleet-number-list-and-quit.md)。
輸入 GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`，dosgolem `f7523f9`，原版線性記憶體與 DOS 呼叫紀錄。
確認後 `int 21h AH=4C AL=00` 結束；取消回命令階段。既有 Game 與 `TestQuitConfirmButtons` 已涵蓋兩條分支。

第106輪來源 Windows 程式的正常 X11 路徑完成保存後，按公開「決定」仍未結束。
原失敗批次 `workplace/music69-windows/saveload69-r106-a/` 保留。
程式審查確認 `Game.Click` 設定 quit，而 `app.Update` 未消費旗標。
這是已核實行為的顯示層接線缺漏，不需新增原版規則或格式。

草稿審查只接受以下窄修正，據此標 READY。正式同狀態 GUI 及保存證據由[第106輪計畫](../goal/2026-10-05-windows-save-load-round-106.md)追蹤。

## 玩家契約

- 正常系統指令 → 結束 → 決定後，下一次 `app.Update` 回傳 `ebiten.Termination`。
- 檢查在推進 Game、音訊與輸入前。取消時不結束。
- 使用 Ebitengine 的正常收尾，保留現有播放器關閉流程。
- 不修改 Game 規則、保存格式、AI、音源或動畫時序。

## 驗證閘門

- 回歸先證實正式 app 缺接線而失敗，修正後通過確認與取消兩案。
- Linux 與 Windows 真實 GUI 經正常選單保存、公開退出，從新程序主選單讀回、重存及再次公開退出。
- 六檔完整跨平台比較；載入的陣營控制位元組更新逐項保存，不遮罩。
- 標準產品測試通過。原版退出證據沿用341，本輪不外推其他原版流程。

## 驗證結果

[094](../playtest/094-windows-save-load.md)：回歸修正前失敗、修正後取消及確認通過。
Windows兩批及Linux正式GUI各兩個新程序，48張完整視窗14745600像素零差，300項真實X11輸入；公開退出後重新啟動主選單讀回及重存通過。
全部六檔及IDX符合對照，WORK四byte既有陣營更新完整記錄，沒有遮罩。
正式十八套件13通過、5無測試。READY審查內容保留；沒有擴張Game規則或保存格式。
