# 094：Windows 正常保存、退出與重新啟動讀回

日期：2026-10-05，第106輪。沿 [#69](https://github.com/wicanr2/logh3/issues/69)，新發現另先開 [#90](https://github.com/wicanr2/logh3/issues/90)。
結論：修正正式 GUI 漏讀 Game.Quit，Windows 兩次與 Linux 一次正常玩家流程完成公開保存、退出、新程序讀回、重存與再次公開退出。
48 張完整 640×480 視窗、14745600 像素及300項真實 X11 操作，與同來源 Linux renderer 對照零差。
完整身份與命令保存在[來源索引](../internals/data/windows-save-load.json)，由[核對器](../../tools/scripts/windows_save_load_receipt.py)重驗。

## 來源與範圍

開工 HEAD `91d85785b937145f060066cf9bb29edc643a7c05`，工作樹乾淨；原版 bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`，原版164檔完整重核。
知識路由為 Wine／Xvfb／正常玩家路徑，入口 `~/.codex/knowledge-base/sources/claude/retro-cht/retro-game-playtest.md`。
規格沿005／006／007／009／067與[117](../spec/117-player-public-quit.md)。新編艦隊沿013既有規則；沒有實行回合或測試亂數判定。

| 程式 | SHA-256 | 用途 |
|---|---|---|
| 修正前 Windows | `6af940c72f340a5ecc49b2c6dc092e0e6dca077c1136abada3cad09b49e8ffa9` | a：正式正常 GUI 失敗觀測 |
| 修正後 Windows | `abab7e58f88b0f37dc6a7e69bf326426ec95c94aec4fe8f899a2781621292343` | b／c：正常公開存讀 |
| 修正後 Linux | `fa2995765c7ee2ed5d458e46f7b3536a93a1f82b408130320d6f5b2c9cab29fb` | refc／refd腳本對照、linuxb真實 GUI |

Go 1.24.13、固定第103輪 vendor，無網路離線建置，兩平台各兩次完整二進位相同。
實際建置副本的 cmd／internal／go.mod／go.sum 等於現行來源。
唯一正式產品變更為 `cmd/gin3/main.go` 的四行退出檢查；新回歸為 `quit_test.go`。
第103輪來源包保持歷史版本，本輪新程式尚未加入該包，不宣稱它仍涵蓋最新玩家程式。

## 失敗與修正

修正前 a 完成正常新遊戲、編艦隊與公開保存，六檔及 IDX 與 Linux 相同。
系統指令 → 結束 → 決定後，程序10秒未結束。九張視窗零差，保存沒有失敗；退出觀測仍標 FAIL。
`workplace/music69-windows/saveload69-r106-a/` 的 driver、START、TRACE、PNG與保存完整保留，沒有 CHECK 通過檔。

已證實根因：Game 在 StateQuitConfirm 設定 quit，正式 app.Update 沒有讀取。
原版確認走 DOS `AH=4C AL=00` 的既有證據在[研究341第四節](../re/341-fleet-number-list-and-quit.md)，原版 GIN3.COM／dosgolem身份由該紀錄回查。本輪沒有新增原版執行器或改動它。
117先由草稿審查至READY，回歸先失敗「正式 app 未結束 `<nil>`」，再在 Update 推進 Game、音訊與輸入之前讀取旗標、回傳 `ebiten.Termination`。
取消繼續、確認結束的正式 app 回歸通過，真實 GUI 再驗退出。

## 正常路徑與完整比較

Windows 每批兩個新程序，預設片頭，真實 X11 輸入；沒有 `-script`、`-intro=false`、狀態注入、自動截圖退出旗標或外部 kill 的成功路徑。
玩家與資料目錄 `/player`／`/orig`／`/data` 唯讀，cwd `/player`。
存檔在 `玩家 保存 空間` 中文與空白路徑，完整比較讀取的實際 Windows 產物。
每張視窗全圖回授確認後才送下一組輸入，沒有遮罩、裁切或縮放。

| 批次 | 平台及入口 | 操作 | 完整視窗 | 結果 |
|---|---|---:|---:|---|
| b | Windows 正常新遊戲及主選單讀回 | 100 | 16 | 全圖0差，兩次公開退出0 |
| c | Windows 獨立重跑，使用獨立refd | 100 | 16 | 全圖0差，兩次公開退出0 |
| linuxb | Linux 正常新遊戲及主選單讀回 | 100 | 16 | 全圖0差，兩次公開退出0 |

新遊戲選帝國、通過防拷、關閉月初及訂造報告、新編萊因哈特艦隊，公開存第0格；退出後重新啟動，主選單取進度、通過防拷、選第0格及帝國，再公開存第1格及退出。
六檔每格共44560 bytes，IDX528 bytes。三批兩次保存，共267360 bytes六檔及3168 bytes IDX與平台對照全同。
IDX為同一檔，第二次保存會更新索引；CHECK保存首次實際讀取與對照的雜湊，最終檔保存两格的索引。核對器依保存時點核對，不拿最終IDX與首次雜湊比較。
現況艦隊原池為0，FLEET3904及SHIP27456 bytes全檔保存並讀回。

| 存讀差異 | 結果 | 依據 |
|---|---|---|
| STELLAR／PLANET／ADMIRAL／FLEET／SHIP | 第0格與重存第1格全檔相同 | 所有比較逐位元組，無排除 |
| WORK | +0E、+10、+12：0→1；+14：0→31，其他226 bytes相同 | 正常選帝國時 `chooseLoadFaction` 寫控制九bytes，既有載入契約 |
| 跨平台WORK | 完整230 bytes符合各階段Linux對照 | 不以欄位遮罩通過 |

refc／refd各15張獨立腳本對照，全圖色號及PNG、保存六檔及IDX均可重現。
Linux腳本只提供同來源 renderer／公開輸入對照，正式玩家路徑由真實X11批次補足。
初次linuxa亦完成相同操作；CHECK.scope文字誤寫Windows，START.platform正確。保留原批次，修正文案後重跑linuxb，正式48張數字不包含linuxa。

## 重生入口

在乾淨輸出與已載入第103輪固定 vendor 的工作區執行；工具拒絕覆寫既有批次。
原版由玩家自備，唯讀掛載；腳本、工具鏈與所有身份見來源索引。

```bash
bash tools/gui.sh xrun python3 tools/scripts/windows_save_load.py build
bash tools/gui.sh xrun python3 tools/scripts/windows_save_load.py references --revision refc
bash tools/gui.sh xrun python3 tools/scripts/windows_save_load.py references --revision refd
bash tools/windows_save_load.sh gui --revision b --reference refc
bash tools/windows_save_load.sh gui --revision c --reference refd
bash tools/gui.sh xrun python3 tools/scripts/windows_save_load.py native-gui --revision linuxb --reference refd
bash tools/gui.sh test ./internal/... ./cmd/... -count=1
bash tools/gui.sh xrun python3 tools/scripts/windows_save_load_receipt.py --check
```

Linux image `logh3-ebiten:2`，ID `sha256:6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`。
Wine image `colonization-windows-verification:20261001-r1`，ID `sha256:7d3c7ff554db39300e3e7a92a7b79c2d5fcd647289596d22fa10899baac61201`，Wine9。
有界容器採無網路、UID／GID1000、CPU／記憶體／PID／log限制及清理trap。

## 測試、勘誤與限制

正式18套件13通過、5無測試；Game131.753秒，gin3 55.481秒。
首次標準命令誤用 `./...`，掃進研究腳本及暫存vendor，保留錯誤日誌；改成正式產品範圍、同一image乾淨重跑通過。
擷取器初次Go建置自動把go.mod的indirect註記整理，已還原原檔；重生器改在固定vendor副本建置，不改依賴版本或正式go.mod。
Python語法、完整檔案身份與來源審查通過，沒有把研究腳本的setup錯誤寫成產品缺陷。

舊第103輪20469來源保持；104／105各只在main.go有預期變更，舊內容按91d85785歷史提交核實，其他1095／811來源保持。
舊收據與索引不覆寫，不把舊Windows配樂觀測外推成新二進位的音訊實測。
輸出擁有權1000:1000，root-owned與錯誤`.md`目錄檢查及專案容器清理另記WORKLOG。

117限定正常GUI退出接線CONFORMED，#90可關閉。#69保持OPEN。
本輪不是Windows實機、原版新同狀態或正式發行驗收。完整音樂播放／循環、人耳、macOS、最新來源包及正式版號／dist-all仍缺。
