# 第172輪：新遊戲第一場戰鬥崩潰

使用者提供Windows崩潰截圖，要求修正。起點main `5b7c6c7e`，工作樹乾淨。使用者要求驗證完後commit、push，再重打完整版。

## 目前狀態表

| 項目 | 狀態與契約 |
|---|---|
| 崩潰 | PASS，新遊戲沒存過檔進第一場戰鬥不再崩潰；v1.6.4兩平台兩陣營都重現，修正版同一條路徑打完整場、存讀檔後再開戰 |
| 配置 | PASS，新編艦隊依`297B:0ABF`配置；首都的中心與航向符合原版傾印 |
| 同格重編 | PASS，解散當場作廢原始表那一列，同步不再失敗 |
| 同盟鏡頭 | PASS，新遊戲依模式寫座位bytes，進場鏡頭對著玩家艦隊；由第一輪封包的目視抓到 |
| 標準測試 | PASS，21套件16PASS／5無測試／0失敗 |
| Windows | Wine通過，兩陣營存檔與Linux逐位元組相同；沒有實機 |
| 第169輪閃退 | 使用者確認發生在進戰場，與本輪同一個崩潰 |

入口：[規格159](../spec/159-new-game-first-battle.md)、[研究450](../re/450-new-fleet-placement-and-new-game-slot.md)、[驗收143](../playtest/143-new-game-first-battle.md)、[機器收據](../internals/data/new-game-battle172.json)。
