# 411：候選契約審查與原池命令接線

日期：2026-10-04，第74輪，父提交dfccda34441a6271b18fd3594bfa340f177fa6ea。
沿[#79](https://github.com/wicanr2/logh3/issues/79)與[#48](https://github.com/wicanr2/logh3/issues/48)。
本輪沿既有研究401／403，不新增原版硬體或演出逆向切片。

## 證據審查

候選收據SHA-256為59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f，
命令收據為e3b39582bfda07105d27837747d46bca19eaf2f1fb5d6ff383500870e4a6870d。
下表原bytes另直接取[入口收據](../internals/data/fleet-combat-entry.json)，
SHA-256 132e4ae3c6ac38ddaf62075d3bd450c3537dda5c8a4a12a4d376424a3f239121。
兩份凍結檔、48組原版a／b輸入RAM與report均重核SHA，原版bytes與分欄位址保持。
dosgolem為f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，未修改執行器；IDA9.4來源沿原收據。

| 原始CS:IP | bytes | 已證實行為 |
|---|---|---|
| 1ED2:036D | 2688472e | 寫武器+2E |
| 1ED2:037C | 2689472c | 非FF才寫目標+2C |
| 1ED2:03D9 | 26c6472e03 | 防禦只寫武器3，不清目標word |

候選依目前座標、300上限、情報與武器橢圓；光線砲每第8步驗地形與分隊。
查詢先於Action=7F命令門檻，不能先刪除無法下令的攻方。
search-open301後續查詢可覆寫目標mask，驗收必須逐查詢對應。
以上proven限原版固定bytes及研究403案例，不外推自動選擇與整個顯示名冊排序。

## 實作與驗證

100先DRAFT、核對來源及bytes後READY，才實作候選及Game資料介面。
直接取原版02D2、0382、0097觀測作期望，48案111次AX返回與61次mask寫入全符。
受控半軸變更以獨立表核共用演算法；正式入口共用093凍結表，不修改全域或玩家資產。

Game用既有原池映射讀寫SHIP+2C／+2E，提交前重新查詢；沒有新增存檔欄位。
五種命令保存／重讀及十種拒絕案例通過，錯誤不改原池、分隊或RND。
既有正常bootstrap測試補48個分隊的查詢，畫面、原池與RND保持。

首個編譯檢查誤把battleCells方法當欄位、passiveComputer當單值API；依實際介面修正後乾淨重跑。
這是接線時的編譯錯誤，不列為原版或已發行產品缺陷。

來源、程式與測試紀錄共201項凍結於[完整索引](../internals/data/fleet-command-data.json)，
SHA-256 4b42447bfe2bbe899bc1c51987501b9b2c21cd95a1323452c38383f490acea80。
重建入口：[凍結器](../../tools/scripts/fleet_command_data_receipt.py)，用既有GUI容器執行，步驟見[063](../playtest/063-fleet-command-data.md)。

100只限資料介面CONFORMED，Game武器UI、第④步、傷害／演出續接與音效尚未呼叫。
完整092／056仍DRAFT；061被動CPU及零距離首點裁決保持，不關#48／#79。
