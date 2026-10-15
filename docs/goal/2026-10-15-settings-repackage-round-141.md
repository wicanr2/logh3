# 第141輪：設定視窗左緣與三平台重新打包

## 目前狀態表

| 工作 | 狀態 | 完成條件 |
|---|---|---|
| 設定左緣 | 限定CONFORMED | 兩陣營22張正常設定／選單各530行外緣連續、內側陰影完整；標準15 PASS／4無測試與正常設定／戰場操作通過，README三圖更新及獨立核對 |
| 完整版與引擎包 | 等待版號裁決 | 最新字型與邊框，同來源重建三平台完整／引擎及來源；逐包驗收 |
| Release與README | 等待版號裁決 | 已發布v.1.4.0不可自行覆寫；使用者裁決新修訂號或明確例外後執行 |

起點`8d3d16c7`，工作樹乾淨。使用者要求重新打包完整／release，再追加帝國設定左邊框消失的修正，明確要求保留陰影。GitHub worklist快照102張、1 OPEN，282條verify零失效，#48不變。

已核對v.1.4.0-20261015正式發布。AGENTS及版號契約禁止覆寫已發布身份，版本選擇已提出，尚未收到答案。邊框修正及不依賴版號的驗證持續進行；不擅自改發新版或覆寫舊版。

邊框修正依[148補充](../spec/148-clear-settings-font.md)，現有[正常驗證工具](../../tools/scripts/compact_settings_gui.py)可抽驗兩陣營；可寫輸出使用`workplace/settings141/`。後續封包沿[既有封包工具](../../tools/scripts/v140_platform_delivery.py)及[Windows凍結來源工具](../../tools/scripts/windows_hotfix_package.py)。歷程追加WORKLOG，發布與README結果依實際交付回填。

[左緣逐行核對與README原圖匯出](../../tools/scripts/settings_border_receipt.py)保存兩陣營正常操作、外緣與陰影的證據。
