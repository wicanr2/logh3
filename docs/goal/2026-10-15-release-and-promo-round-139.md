# 第139輪：v.1.4.0三平台發行與推廣片重拍

## 目前狀態表

| 工作 | 狀態 | 完成條件 |
|---|---|---|
| 設定頁147 | 限定CONFORMED | 正式Linux ZIP帝國與Windows ZIP Wine同盟，敵軍預設關閉、AI五值、四字型、F7、兩規則、實際圖示及戰場AI切換通過；README三圖更新及獨立核對 |
| Windows完整版／引擎／來源 | 修正後重建通過 | 024a851d產品來源，實際ZIP八條Wine入口及同盟正常GUI；舊候選及凍結來源完整歸檔 |
| Linux完整版與引擎包 | 修正後重建通過 | 實際ZIP逐檔／權限、11條完整與4條引擎入口及帝國正常GUI通過 |
| macOS完整版與引擎包 | 修正後結構檢查通過 | 雙架構／依賴／ARM3180簽章頁及ZIP重新檢查，未真機驗收 |
| 推廣片重新拍攝 | 本機交付通過 | 新拍23場艦隊及2支分隊要塞演出逐格通過；147設定／AI重新錄製；36段102.13秒，35剪接點音樂連續、AAC SNR34.23dB、零黑段，總覽及原尺寸抽樣已目視 |
| GitHub Release | 已發布並下載核對 | 正式tag指向d2a59fc7；三平台引擎、來源與校驗共五附件，實際下載SHA-256及598項tag產品來源全部相符 |

使用者已授權本輪全部工作及commit／push。版本固定`v.1.4.0-20261015`；依追加設定頁修正重建尚未發布候選。舊候選及凍結目錄完整保存於`workplace/promo139/precompact/`，`MAPPING.json`綁原路徑及七包雜湊；既有已發布身份不動。三平台使用同一1576項凍結程式。原版、HD、完整包及影片保持本機，不上傳Release。公開儲存庫的素材／decomp授權沿AGENTS既有決定，不擴張為完整包或影片公開授權。

未完成項權威仍為GitHub worklist Issue；開工快照102張、101 CLOSED／1 OPEN，282條verify零失效，#48正式戰術讀回重建與DOS整段同狀態保持OPEN。本輪不擴張規則或變更既有完成門檻。

入口：[封包工具](../../tools/scripts/v140_platform_delivery.py)、[新錄影輸入](../../tools/scripts/promo_v140_prepare.py)、[要塞正常錄影](../../tools/scripts/fortress_presentation_gui.py)、[本輪驗收](../playtest/128-v140-release-and-promo.md)。歷程追加WORKLOG，首頁只放穩定成果。

影片合成與連續取樣驗證：[本輪製片工具](../../tools/scripts/promo_v140_movie.py)。

公開後的獨立核對：[Release下載與tag驗證器](../../tools/scripts/v140_release_verify.py)。

使用者追加的設定頁修正依[147限定CONFORMED](../spec/147-compact-system-settings.md)實作，正常驗收與新錄影入口為[compact_settings_gui.py](../../tools/scripts/compact_settings_gui.py)。敵軍預設關閉，README展示開啟與實際戰場效果；現有AI五值已接系統設定。

README正式ZIP截圖沿[匯出工具](../../tools/scripts/readme_compact_settings.py)產生並獨立核對，不修改來源PNG。

正式入口：[v.1.4.0-20261015 Release](https://github.com/wicanr2/logh3/releases/tag/v.1.4.0-20261015)。產品修正`024a851d`、交付及tag`d2a59fc7`均已推送；[唯一交付索引](../internals/data/v140-release-promo.json)的`release`保存實際下載與tag核對收據。既有發布身份未移動。本輪授權工作已完成，Windows／macOS真機、新混音人耳及#48既有未完成範圍保持明示限制。
