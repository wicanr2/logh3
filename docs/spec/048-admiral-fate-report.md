# SPEC 048 — 失蹤提督生死判定與報告

狀態：**READY／已實作，尚未 CONFORMED**（缺可用的原版 framebuffer 收據）

證據：[`docs/re/307`](../re/307-admiral-fate-pass-and-report.md)。

## 範圍與順序

1. 戰略 AI 完成後、未安定度／維持費／民心／移動之前，掃描 82 筆提督。
2. 只處理 `+0x00 == 0x80`；抽籤、特殊提督與人類／電腦死亡門檻完全照
   `056A:0219`，並使用共用 `ds:94DC` 對應的 `strategyRand`。
3. 死亡與生還依原版精確寫回提督記錄；每一位人類座位提督排一則報告。
4. 報告逐則模態顯示，滑鼠左右鍵皆關閉；全部關完才續跑本回合其餘階段。

## typed contract

- `rules.AdmiralFate` 保存 `Admiral`、`Faction`、`Dead`；
  `ResolveMissingAdmirals` 就地更新原始 28-byte 記錄。
- `AdmiralFateReport` 保存事件時索引、陣營與結果；pending queue 與 callback
  不和占據／叛亂／易主報告共用。
- 單人只顯示 `g.faction`；對戰把兩個座位都視為人類。電腦提督仍套狀態，
  但 `056A:0311` 的座位檢查會跳過畫面。
- presentation 使用 `FlipReportPanel.DrawAdmiralFate` 三行組字，頭像直接取
  事件提督索引；原版錯字「成央B」逐字保留。
- 新遊戲與讀檔清空 reports、pending queue 與 callback。

## 自動驗收

- seed 0：特殊提督 0 生還，`+0x00=FF`、`+0x0A=帝國首都`。
- seed `0x60F`、`+0x0B=255`：一般提督兩次亂數為 117、2817，死亡；
  驗 `+0x09|=80`、`+0x00=0`、`+0x02…+0x05=0`。
- 未通過首道門檻者保持失蹤。
- 兩則報告需兩次滑鼠事件，callback 只執行一次；左右鍵各覆蓋。
- 原始訊息表自動產生四段固定字碼，畫面測試確認全部存在。

## 尚未通過的 gate

- dosgolem 的此 compositor framebuffer 尚不可用，故沒有文字、本人頭像與
  完整畫面的逐像素比較；目前幾何只繼承同 compositor 的既有證據。
- 未建立自然遊戲狀態造成失蹤提督的正常玩家路徑；本規格只從已存在的
  `+0x00 == 0x80` 狀態開始，沒有猜測失蹤是如何發生。
