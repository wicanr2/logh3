# 第141輪：設定視窗左緣與三平台重新打包

## 目前狀態表

| 工作 | 狀態 | 完成條件 |
|---|---|---|
| 設定左緣 | 限定CONFORMED | 兩陣營22張正常設定／選單各530行外緣連續、內側陰影完整；標準15 PASS／4無測試與正常設定／戰場操作通過，README三圖更新及獨立核對 |
| 完整版與引擎包 | 本機交付通過 | 同一1584項凍結來源，606項非vendor與目前產品相同；七包內容／manifest／SHA-256、Linux15入口、Windows8條Wine入口、兩正式ZIP正常設定及530行框線通過；macOS3196簽章頁與雙架構通過 |
| Release與README | README完成，公開覆寫與回驗中 | 舊正式七包、Release、tag及五下載附件已備份；新版五附件待精確覆寫，發布後下載核對 |

起點`8d3d16c7`，工作樹乾淨。使用者要求重新打包完整／release，再追加帝國設定左邊框消失的修正，明確要求保留陰影。GitHub worklist快照102張、1 OPEN，282條verify零失效，#48不變。

已核對v.1.4.0-20261015正式發布。使用者在兩項明確選擇後回答「複寫已發布的」，授權本輪保留v.1.4.0-20261015版號並覆寫該Release、附件及tag，作為AGENTS不可覆寫規則的本輪例外。其他版本保持。先備份舊身份、完整包及來源，不因同版而混用舊雜湊。影片不重拍，從舊正式交付逐檔保留。

邊框修正依[148補充](../spec/148-clear-settings-font.md)，現有[正常驗證工具](../../tools/scripts/compact_settings_gui.py)可抽驗兩陣營；可寫輸出使用`workplace/settings141/`。後續封包沿[既有封包工具](../../tools/scripts/v140_platform_delivery.py)及[Windows凍結來源工具](../../tools/scripts/windows_hotfix_package.py)。歷程追加WORKLOG，發布與README結果依實際交付回填。

[左緣逐行核對與README原圖匯出](../../tools/scripts/settings_border_receipt.py)保存兩陣營正常操作、外緣與陰影的證據。

同版重包的備份及新收據由[重包工具](../../tools/scripts/v140_repackage.py)建立。舊正式目錄與身份保存在`workplace/settings141/published-before-overwrite/`，`MAPPING.json`提供原路徑對照。[本次來源索引](../internals/data/v140-repackage.json)保存新七包、兩正式ZIP正常GUI、框線像素及平台驗收身份。影片與16項證據逐位元組沿舊批保存，不重拍或重新混音。
