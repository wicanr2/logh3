# 073：戰場升級的原版診斷

本輪等級：受控原版入口診斷。未做原版／remake同態或GUI對拍。
遠端#79 OPEN，108 DRAFT、完整062 DRAFT。

固定dosgolem f7523f9、乾淨工作樹，從ownership79-r82-entry-a完整machine／DOS保存
自然到185C:060D，再明示注入，停於185C:09BF。IRQ0每20000步，CycleClock=false。
每案開始設seed12345678，記錄原始程式、父保存、全部注入、按下／放開步數及末端seed。

| 核對範圍 | 結果 |
|---|---|
| 60案原版雙跑 | 兩份13489258-byte JSON逐位元組相同 |
| 經驗門檻 | 各級need-1、need、need+1；高word非零、多次升級及99級停止 |
| 成長 | 109次嘗試、47次升級；407個原版RND輸出獨立模型符合完整提督表 |
| 能力上界 | 正常能力與0／199／200／255注入；200以上不取成長RND |
| 報告 | 11則文字、頭像參數、完整640×480色號與16色、PNG保存；副官一張已目視 |
| 艦隊與艦艇表 | 每案完整前後SHA及全部差分保存；重算模型尚缺 |
| remake／正常玩家 | 未改正式路徑，未驗，不能升格 |

原版JSON SHA-256 `3d11f8b68703a5eefe18270821d4ac19dd94bf92a4274c3e86632dd412290d80`。
53來源與全部案例見[原版索引](../internals/data/battle-promotion.json)，SHA-256
`ab96e1d8d0213da5dfc8b34acc1f46bf1b83440eb3d552ed7f37efd8c2b0c661`。
分級原指令、21項語意及兩陣營12級文字見
[分級索引](../internals/data/battle-promotion-review.json)，SHA-256
`3d52c61e0e93433297f6240078614c11da38cf27e24347a4cfe926cfc4f451b6`。
重生與停止線見[研究418](../re/418-battle-promotion.md)。

本輪未改產品程式，未重跑第83輪標準產品套件；舊通過範圍維持原聲明。
原版全圖已保存，尚無remake圖、差分或像素統計，不稱畫面對拍通過。
