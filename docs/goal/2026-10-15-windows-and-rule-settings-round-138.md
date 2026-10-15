# 第138輪：Windows啟動與可選規則

2026-10-15。未完成項仍以GitHub `worklist` Issue為準，本輪開工102張Issue、101已關、#48開啟；282條核對零斷言不成立。這是本輪唯一狀態表。

| 項目 | 目前狀態 | 下一步 |
|---|---|---|
| Windows直接雙擊找不到GRAPH.IMG | 已修；舊版負向重現與新版Wine四案通過 | 新封包實際啟動 |
| 轉移旗艦 | 使用者確認，規格146 READY | 實作三條傷害路徑、保存及測試 |
| 名將移動加成開關 | 使用者確認預設開啟，規格146 READY | 設定／偏好及實際移動抽測 |
| 文件、commit與push | 進行中 | 驗收後更新README及WORKLOG並推送 |
| #48既有餘項 | 保持OPEN | 本輪不以新規則或Wine驗證結案 |

契約為[規格146](../spec/146-flagship-transfer-and-rule-settings.md)。Windows驗證腳本為[windows_resource_paths.py](../../tools/scripts/windows_resource_paths.py)，封包由[windows_hotfix_package.py](../../tools/scripts/windows_hotfix_package.py)建立。成果與限制完成後掛入README及驗收紀錄。
