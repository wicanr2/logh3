# 429：正常玩家旗艦全滅與戰後公開存讀

2026-10-05，第95輪，沿GitHub #79及既有109／110 READY契約。
父提交 `1bd9dfc1dfd26ee4e2b5d575a67f2f63f0149ec4`。正式程式無修改。
入口：[計畫95](../goal/2026-10-05-battle-lifecycle-player-round-95.md)、
[玩家驗證084](../playtest/084-battle-lifecycle-player.md)、
[完整來源索引](../internals/data/battle-lifecycle-player.json)。

## 方法與證據範圍

沿第94輪已凍結的S3同盟公開操作，新增唯讀報告與逐擊觀測。
原始資料由正式bootstrap載入，依序走片頭、主選單、防拷138、S3陣營選擇、戰略實行及首戰。
首次戰場實行前設定seed `0x12345678` 一次，輸入紀錄可核對唯一seed事件。
沒有注入位置、艘數、庫存、士氣、功績或勝負。
新增觀測保存六張完整原池、typed分隊及seed；typed分隊先複製，避免後續變動污染前一筆。

使用 `tools/gui.sh`、`logh3-ebiten:2` 與Go 1.24，UID/GID 1000。
原始遊戲輸入唯讀掛載，容器無網路。第94輪1005項來源逐檔核對，不改歷史收據。
本輪未啟動原版執行器或IDA，沒有新增原版位址或原版正常玩家證據。
原版單擊及回收契約沿既有收據；本輪結論限於remake正常玩家鏈。

## 自然結果

| 觀測 | 結果與等級 |
|---|---|
| 正常操作 | 6745項公開輸入、50次艦隊命令、19次交戰，remake已證實 |
| 逐擊資料 | 38筆攻擊前／戰損後觀測，19筆已提交戰損，remake已證實 |
| 末擊 | 星域26、第9戰術回合、命令序號21，敵艦隊12旗艦分隊0由6艘減至0艘 |
| 傷害 | Attack213、Defence122、NormalizedDistance22502、CurveY71、Shots16000、Percent35；Requested26、Lost6 |
| 計分 | Scorer1、Score0、AdmiralMerit3、AdjunctMerit1；沒有fallback或skip |
| 報告 | 先全滅、後敗退；全滅頭像參數-1，敗退頭像75；兩則seed均1360112191 |
| 終態 | 報告返回後正常命令階段，零占領報告 |

全滅文字保留Big5及控制碼，顯示為「帝國軍第13艦隊舒坦梅茲」「已經全滅」。
原始文字hex、兩張307200-byte色號畫面與PNG都由觀測器保存。
報告前後FLEET12的提督／副官欄位 `+04／+05` 由12／28變FF，陣營及Groups保持。
一般編制引用的SHIP池80、82、83、84、85、87、86、88解除艦隊歸屬。
ADM12／28的 `+00` 由FF變80、`+09` 由0C變8C。
這是旗艦全滅導致艦隊回收與提督失蹤；不是提督死亡判定。
下一戰略回合的生死處理另由 `admiral_fate.go` 執行，本次沒有驗到該分支。

兩次正式bootstrap獨立重播均符合全部38筆逐擊與2筆報告。
每筆六表44624 bytes，共40筆1784960 bytes；同時比較typed分隊、文字、頭像與seed。
JSON比較先轉為相同表示，保留全部欄位及數字精度，不略過typed或raw資料。
戰後由公開選單保存、返回，再公開載入與選陣營；六份玩家存檔與讀回後診斷保存逐byte相同。
六份儲存檔總長44560 bytes，與記憶體六表44624 bytes分開計量。

## 真實GUI

正式 `cmd/gin3` 以測試overlay設10 Hz，使用正式輸入及renderer。
1032個操作完成；64次額外鏡頭回授手勢均有記錄。
63張原生X11全圖共19353600像素，與獨立Game計畫零差，Game色號亦逐byte相同。
全滅及敗退的報告狀態、文字、頭像、typed分隊、六表與seed符合計畫。
六份實際玩家存檔符合計畫，公開載入後六份診斷保存符合，SaveErr為nil。
已目視實際全滅X11畫面，文字及視窗無溢出。

這不新增正式60 Hz長流程時序驗收。第93輪三批60 Hz失敗保持。
沒有原版正常流程配對，不能宣稱原版same-state。
一般非旗艦分隊自然死亡與空組搬移GUI、索敵燃料到期、自然升級及占領GUI仍缺。
109／110保持READY、完整062保持DRAFT、#79保持OPEN。

## 失敗分類及身份

- 首版正常探針未建立SaveDir，已完成戰鬥後在公開保存失敗。補輸出目錄後b版同條玩家鏈通過，屬探針設定問題。
- 首版重播準備器的Go文字錨點不符gofmt結果，修正精確定位後生成，屬工具問題。
- bootstrap a／c把typed結構與JSON物件直接序列化比較，欄位排列不同造成誤報。逐欄核對無差，改用UseNumber正規化全部欄位，d／e乾淨雙跑通過。
- 失敗生成檔、log及觀測均保留於來源索引，沒有改期望資料或正式產品。

| 檔案 | bytes | SHA-256 |
|---|---:|---|
| `workplace/lifecycle79-r95-normal-b/normal-1/inputs.json` | 576656 | `011ae9e105210b2982d6185efb0d13dc1139b362f0bf83f38f8d4763f8660077` |
| `workplace/lifecycle79-r95-normal-b/normal-1/lifecycle.json` | 6740016 | `5c343b93539766ba9a7cf3d777802897300a07d021af218586a5db3b3a7523d3` |
| `workplace/lifecycle79-r95-bootstrap-d/replay.json` | 459 | `def23cb50b8b0b8d340e06599e59207f670d12899d988e17b3c663f6cfe09e01` |

來源索引包含輸入、生成的overlay與Go／Python、log、報告、畫面、逐區數字及六檔保存。
完整重跑方法見084；新輸出拒絕覆寫既有批次。
