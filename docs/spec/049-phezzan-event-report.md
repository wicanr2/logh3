# SPEC 049 — 費沙事件雙頁報告

狀態：**CONFORMED**（事件雙頁主 overlay；限制見末節）

證據：[`docs/re/308`](../re/308-phezzan-report-receipts.md)、
[`docs/re/296`](../re/296-phezzan-event-trigger.md)。

## 契約

1. 費沙第一次易主、既有易主報告全部關閉後，先依序為人類座位排兩頁
   `StatePhezzanReport`，再執行費沙經濟轉移、邊境叛離與懲罰。
2. 單人只為目前玩家顯示；對戰依原版順序先新擁有者、再對手，各兩頁。
3. 四組文字依 `(owner, viewer)` 選 `owner_0`、`owner_1`、
   `opponent_0`、`opponent_1`；原文來自 `phezzan-event.json`。
4. 畫面為 24×5 內容格的 style-1 frame、attribute `0x97`；頭像取 viewer
   的軍事專任參謀，完整 80×100 畫在 `(432,180)`。
5. 每次滑鼠事件關一頁；最後一頁關閉後才呼叫既有 `runPhezzanEvent`，
   然後接回月底／一般回合收尾。新遊戲與讀檔清空 pending queue/callback。

## 驗收

- 帝國 owner 的兩頁 overlay 各 58,240 像素對原版 0 差異。
- 遊戲層驗第一頁與第二頁期間 `WORK +0x34` 尚未設；第二頁關閉後才設旗標
  並套事件本體。
- 左鍵與右鍵各關一頁，callback／事件本體只執行一次。
- 四組訊息各有兩頁，產生器從既有權威 JSON 重生。

## 限制

- 同盟 owner 與對戰雙人兩邊觀看使用同一支已證實原版分支及固定資料，
  但本輪沒有各自的 framebuffer 收據，不擴張逐像素聲明。
- overlay 外的「確認」提示未納入像素矩形；不影響兩頁內容與事件前阻塞。

