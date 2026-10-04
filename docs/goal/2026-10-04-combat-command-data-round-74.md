# 第74輪：交戰候選與命令資料鏈

日期：2026-10-04。父提交dfccda34441a6271b18fd3594bfa340f177fa6ea。
遠端89張worklist、9張OPEN；269條核對零失效。
工作沿[#79](https://github.com/wicanr2/logh3/issues/79)與[#48](https://github.com/wicanr2/logh3/issues/48)。

正式Game在索敵後尚未呼叫艦隊交戰。這輪處理最前端的依賴：
原版SHIP／艦種／地形 → typed候選 → Game原池映射 → 提交／取消 → 原存檔保存。
證據沿研究401／403，不再為已充分的候選契約新增原版逆向切片。

- 審查[100](../spec/100-fleet-combat-candidates.md)，READY後才實作。
- 直接核48案原版逐查詢AX及mask，不用可丟棄probe結果作期望。
- 核原池命令讀寫、FF取消、防禦3、非法候選及存讀，不消費RND或彈藥。
- 跑標準產品測試，更新目前狀態、Issue證據及索引，commit並push。

100不定義完整目標UI、自動選擇、排序艦隊、時間線或傷害接續。
這些仍由092 DRAFT處理；不將局部資料鏈稱正常交戰完成。
061單人CPU被動、核准零距離、音訊及清單速度裁決維持。
