# 421：升級的路徑快取與保存回寫邊界

沿[108](../spec/108-battle-promotion.md)與[#79](https://github.com/wicanr2/logh3/issues/79)。
父提交b501420f46eb55314b1d9430fc931ce700ae1a79。
工具及原版輸入沿[420](420-battle-promotion-global.md)，本輪沒有新IDA匯出。
原版DS位移、CS:IP均為固定dosgolem f7523f9執行期定位。

## 既有路徑快取

第86輪觀測器新增兩個明示診斷，不改舊來源。
第一案在8-4入口把星域13的+12從0寫1；第二案把星域28的+12寫0，與既有0相同，作控制對照。
PLANET與三表不改，兩案seed12345678預先固定。
兩份完整原版JSON相同；一次升級後，三表及RND均與第86輪相應無注入案相同。

`proven`：8-4的090A:0E6D不重建路徑，只沿既有四矩陣算騷動。
從父狀態原始歸屬建立的6144-byte四矩陣，與兩案返回後EMS逐byte相同。
完整STELLAR及WORK均符合快取模型。
第一案若以換手後歸屬提早重建，684個距離／中點矩陣項不同；控制案為0。
收據見[快取來源](../internals/data/battle-promotion-cache.json)。

模型第一版錯把第二案也當作換手，反對照斷言失敗。
查原始歸屬後明列第二案是同值控制，第二版保留這個分類，不挑種子或改原版結果。

042原先「任何重算都可用目前歸屬重建」的敘述只適用未發生戰術8-2換手時。
108須保留戰略處理開始時已建好的路徑，跨該次處理的戰場與戰術回合沿用。
下次戰略處理重建；LoadSlot清除該暫存，無存檔格式變更。
診斷入口可明示提供快取；未提供時從診斷入口歸屬建立，不能宣稱已覆蓋此前換手。

## 現行保存程式的影響

以下為目前程式可回查的`proven`事實，並非新的原版oracle結論。

| 消費端 | 升級後必須同步 |
|---|---|
| rules.Slot.SyncFleets | Fleet.Flags、Expected、Morale、Operations、Composition每組Morale；否則會蓋回FLEET+00／0E／3C／47與SHIP+06 |
| Game.syncFleetTables | g.res.Admiral與g.slot.Admiral須同源，避免保存取舊表 |
| Game.SaveSlot／rules.Slot.Save | g.taxTotal及Slot.ShipBudget／Upkeep／MoveCost須承接新WORK，未知byte仍從底稿保留 |
| 戰術消費端 | BattleFleet.Defence、Ability0／3、Integrity、AdmiralLossFactor、Flag0、TurnSpeed；BattleUnit.Morale、TurnSpeed、CruiseSpeed、FightSpeed |
| 軍階UI | 每位提督自己陣營的12級文字，不以當前玩家陣營代替敵方資料；所有軍階入口共享同一來源 |

本步不重新部署，既有位置、目的地、行程餘額、命令、標記及索敵冷卻保持。
每次原版重算後只回寫已證實的能力／完整度／速度及全局欄位。
報告先呈現該次結果，關閉後才重試同一人，最後才換副官及下一艦隊。
完整契約與驗收列在108，READY後才進正式路徑。

## 正式接線與限定驗證

第87輪依108 READY實作；[076](../playtest/076-battle-promotion-game.md)記錄固定原版與正常玩家的不同範圍。
66案121次嘗試、53次完整五表及473次LCG均與原版相同。
11報告矩形各58240像素與有效16色零差；空白診斷全圖248681差異保留。
九案六檔保存及回讀保持，LoadSlot清快取並恢復本局提督表；兩案路徑快取控制通過。
正常bootstrap經S3公開實行一次及存進度，兩次完整畫面／六檔／輸入相同。
正式首戰X11兩倍率與標準18套件通過，尚不稱正常自然升級GUI或完整062同態。

現行Game在戰略處理直接扣g.military，WORK軍事費底稿到SaveSlot才同步。
8-4預算使用g.military；新增現況經費回歸同時驗WORK底稿未提前被覆寫。
這是產品輸入轉接的事實，不另宣稱原版硬體或正常整段一致。

首輪標準測試三個失敗屬舊測試契約：41級仍期望上將；兩個局部戰場少快取／Rand，且一個把未使用#1槽當敵隊。
依S3實際身份與明示診斷設定修正，沒有放寬正式驗證；兩次乾淨標準重跑皆通過。
正常重播首版因底部選單向上夾限而點錯列，屬工具；改用產品單一幾何後兩跑相同。

完整產品216來源見[索引](../internals/data/battle-promotion-game.json)，SHA-256
`1f2f9b252fc73c4d8091c9483b383b7f170aeac90abacfb259b7fc9a536517d6`。
尚待8-1清空分隊後的原池輸入鏈、自然升級及8-5，108保持READY。
