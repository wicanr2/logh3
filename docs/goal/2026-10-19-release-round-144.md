# 第144輪：README新版實錄與公開發行

使用者要求更新README截圖並發布Release。版本沿已驗本機交付`v.1.5.0-20261019`，新建tag，不移動既有版本。

## 目前狀態表

| 工作 | 狀態 | 證據 |
|---|---|---|
| README | 19張新版正式包實錄已就緒，下載連結切換至v1.5 | [圖片索引](../images/readme-v150.json) |
| 發行輸入 | 七包及影片交付收據獨立重核通過 | [交付131](../playtest/131-admiral-editor-delivery.md)、[本機索引](../internals/data/v150-delivery.json) |
| 公開附件 | 僅三平台引擎、來源及SHA256SUMS.txt五附件 | 完整版與影片留本機；不擴大素材公開範圍 |
| 提交與tag | main與新tag已推送，遠端回讀一致 | `5ec56c683688c9c540e27ecde77352a36f6bf0b4`逐檔符合612項凍結產品來源；新tag為`v.1.5.0-20261019` |
| Release | 已公開並設為最新版，五附件實際下載全部通過 | [正式Release](https://github.com/wicanr2/logh3/releases/tag/v.1.5.0-20261019)、[公開核對收據](../internals/data/v150-release.json)，既有三個Release保持 |

本輪只做文件與發行，不重開已驗規則。Issue快照102張、#48 OPEN；282條verify無失效。三平台完整驗收與來源限制沿131。Windows為Wine驗收，macOS為雙架構、依賴及ARM簽章檢查，沒有新增平台真機聲明。

發布後沿[既有核對器](../../tools/scripts/v140_release_verify.py)，指定`--version v.1.5.0-20261019 --batch /src/workplace/admiral144 --index docs/internals/data/v150-release.json`，五個實際下載附件的大小、SHA-256與校驗清單全部通過，新tag的612項非vendor來源全同。核對器在既有Docker Python入口執行。本機交付索引維持凍結，公開收據另存並由本表索引。

使用者於2026-10-20允許本次主機gh上傳與核對，並另要求git push。沿用其核准的既有版本、tag及封包，實際公開時間為2026-10-20 07:16:29 UTC。先核草稿五附件的GitHub SHA-256及說明，再公開、實際下載並核對。三個舊Release的tag、本文、狀態、附件身份與雜湊均保持；新版本未上傳完整包或影片。

## 發行說明

從「系統指令 → 設定 → 編輯提督」編輯帝國40位、同盟42位將領。介面沿用遊戲配色、立體框與人物頭像，提供十三項能力及四項特殊能力。

- 光線砲、飛彈與戰鬥機能力可分別修改，三項內部能力附用途標籤；原版欄名仍未知。
- 萊因哈特、楊威利的全軍攻擊預設+30%，全軍移動維持+50%。
- 坎普、波布蘭任提督或參謀時，本艦隊戰鬥機攻擊預設+30%。同類取最高值，全軍與戰鬥機倍率各自乘算。
- 編輯結果全局保存，立即套用目前遊戲，重開、新遊戲及讀檔也套用；原版六表存檔格式保持。
- README更新19張正式Linux ZIP的正常操作實錄，包含兩陣營設定、編輯器、星圖、劇本與三種武器交火。

公開附件只有Windows amd64、Linux amd64、macOS Intel／Apple Silicon通用引擎、引擎來源與SHA-256校驗檔。玩家須另行匯入合法取得的原版資料。完整版、原版／HD音畫與推廣影片留本機。

已驗Linux及Windows Wine的正常編輯、取消／套用、F7、重啟與讀檔。macOS雙架構、最低11.0、SDK15.5及ARM簽章通過，Windows／macOS真機未新增驗收。新增加成是remake擴充，不宣稱DOS原版一致性；戰術讀檔續戰仍由#48追蹤。
