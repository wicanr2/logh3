# 第138輪：Windows啟動與可選規則

2026-10-15。未完成項仍以GitHub `worklist` Issue為準，本輪開工102張Issue、101已關、#48開啟；282條核對零斷言不成立。這是本輪唯一狀態表。

| 項目 | 目前狀態 | 下一步 |
|---|---|---|
| Windows直接雙擊找不到GRAPH.IMG | 已修；舊版負向重現、新版Wine四案及實際ZIP八案通過 | 本機v.1.4.0-20261015完整／引擎包可交付 |
| 轉移旗艦 | 規格146限定CONFORMED；三條受控傷害、機率、連續擊毀、清理及六表存讀通過 | 自然遊玩觸發及DOS parity未列入本輪聲明 |
| 名將移動加成開關 | 預設開啟；實際額度／範圍、兩陣營正常GUI、Windows設定重開通過 | 原版AI模式不加成，名將戰術保持 |
| 文件、commit與push | 實作6ef56b26已推送main；README、驗收127與收據已完成，文件另作收尾提交與推送 | 完整包及原音樂只留本機 |
| #48既有餘項 | 保持OPEN | 本輪不以新規則或Wine驗證結案 |

契約為[規格146](../spec/146-flagship-transfer-and-rule-settings.md)，結果為[驗收127](../playtest/127-windows-and-rule-settings.md)，雜湊由[收據索引](../internals/data/windows-rule-settings.json)保存。Windows負向腳本為[windows_resource_paths.py](../../tools/scripts/windows_resource_paths.py)，[實際ZIP驗證](../../tools/scripts/windows_delivery_smoke.py)，[正常設定GUI](../../tools/scripts/rule_settings_gui.py)；封包由[windows_hotfix_package.py](../../tools/scripts/windows_hotfix_package.py)建立。
