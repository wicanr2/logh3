# 第67輪：正式艦隊交戰規則核心

沿[#79](https://github.com/wicanr2/logh3/issues/79)及[#48](https://github.com/wicanr2/logh3/issues/48)。
父提交627c73bae07c29b4d3d8662cf31d75147a43110e；開工89張worklist／9張OPEN，269核對零失效。
前輪為progress，已推送多分隊收據與零距離裁決，不重生既有原版收據。

先把已驗單擊與排序寫成[093子規格](../spec/093-fleet-combat-core.md)，審查READY後實作正式規則。
以原版完整RAM／原指令／逐擊觀測驗係數、庫存、戰損、分數與功績增量及排序。
再接完整092的命令、演出、Game與保存；子核心通過不能替代正常玩家鏈或宣稱八步完成。
061被動CPU、已核准零距離首點、音樂與清單速度決定保持；#89裁決pending不重問。

本輪結果：093 READY後實作純核心，31戰損／11近距離／合法戰鬥機／兩組連打直接原版資料對拍通過。
12零距離只驗核准首點，六組核心無skip及標準產品測試通過。完整092命令／演出／Game／保存待續。
證據入口為[純核心收據](../internals/data/fleet-combat-core.json)及[056追加驗證](../playtest/056-combat-sequence.md)。
