# 083：戰術階段面板

狀態：**READY**。工作項：[#79](https://github.com/wicanr2/logh3/issues/79)。
接續[第三十四輪](../goal/2026-10-02-battle-phase-panel-round-34.md)、
[080](080-battle-movement-phase.md)及[研究339](../re/339-battle-turn-eight-steps.md)。

## 有限契約

第57輪摘要／視野框限定子契約：**CONFORMED**，先經READY審查，依[研究395](../re/395-battle-data-scroll.md)及原版113A:20A9、1727:035E／03CF。
Input摘要第一行軍名使用玩家陣營色0為6、1為5；計分兩行先玩家、再對手。
Victory呈現為玩家分數減對手分數，既有BattleVictory的帝國基準規則不改。
小地圖視野框左上為Minimap.X+世界鏡頭X/16、Minimap.Y+1+世界鏡頭Y/16，整數截斷；框28×22。
這是已知DOS座標契約，不能由bitmap縮放尺寸反推。其他階段文字與八步規則保持原範圍。
正常S3同盟資料視窗三張完整畫面及兩倍GUI驗證見[對拍047](../playtest/047-battle-data-scroll.md)，不升格整份083或088。

原版181B:003F先印DS:3CD4補給階段，再印DS:3CE2修理階段；
181B:00FA進入後印DS:3CF0移動階段，之後才初始化移動額度。
字串0C控制碼清狀態窗，內容只留階段文字；不顯示正常輸入的日期、分數與星域。
原bytes、字串、函式順序、清窗契約與雙跑已由[研究375](../re/375-battle-phase-panel.md)審查。
定位、工具／位址空間、雜湊與分級見[語意索引](../internals/data/battle-phase-semantics.json)。
原版35邊界與7艦隊／55分隊全紀錄見[逐圈證據](../internals/data/battle-phase-frames.json)。

限定補給／修理／移動呈現及回到正常輸入的還原，不改資源或八步次序、
不猜後續未驗階段標籤、不增加硬體wall-clock門檻。
驗收依本輪事前宣告的same-state固定局面與正常玩家抽樣分列，既有nearby-state不升格。

| typed狀態 | 切換時機 | 輸出 |
|---|---|---|
| Input，零值 | StartBattle與正常回合結束 | 完整輸入狀態文字 |
| Supply | beginBattlePreparation，在任何補給前 | DS:3CD4四字 |
| Repair | 第一筆Repair結果前；無結果亦在士氣／防衛力恢復前 | DS:3CE2四字 |
| Movement | beginBattleMovement，在newBattleMovement初始化前 | DS:3CF0四字 |

非Input由同一StatusSpec框重畫，只印四字，位置LabelX／Rows.Phase、色號Ink。
日期、分數、星域及unit情報不畫；既有Supply／Repair結果面板仍疊在階段底圖之上。
後續未驗階段不加新文字，直到正常Input恢復。階段只作執行期呈現，不寫存檔。

## 驗收

- 固定局面只注入一次；正式規則初始化後逐圈執行，不以原版每幀資料餵引擎。
  35邊界各比完整7艦隊／55分隊、seed、鏡頭、色號與有效RGB。
- 保存原版／remake／差分PNG及全畫面、戰場、小地圖、面板統計；不遮罩未知差異。
- 正常主選單→S3→命令→實行，抽結果底圖、移動標籤、Input恢復及完整存讀。
  正常產品抽樣與診斷同狀態比較分列。
- 標準產品測試及正常GUI通過；缺原版字庫／框則skip，不自製代用品。

第35輪沿正常UI載入三份隨附存檔，均實行並恢復Input、完整SHIP／FLEET存讀相同。
S1有47個移動中間幀，正式S1移動與S2第2回合畫面已檢視，見[驗證045](../playtest/045-battle-movement-render.md)。
第35輪三份首戰補給／修理結果均0幀；當時正常資源結果與原版第二戰場同狀態未驗，維持READY。
第36輪原版S1第二戰場43邊界雙跑相同，階段面板與移動完整資料已驗；
唯一入口注入下方向圖示18920像素相同。當輪全畫面仍差434到664，未通過完整第二戰場呈現。
正常資源結果仍缺；080／083保持READY，詳研究376與驗證045第36輪。
第37輪分隊排序依085修正後，43幀Field、狀態面板及小地圖零差異；全畫面仍各差203列38點。
085限定CONFORMED，正常資源結果仍未驗，080／083保持READY，詳研究377與驗證045第37輪。
第38輪bitmap尾列來源已解，依086先READY再實作後43幀完整全畫面及有效RGB相同。
既有S3全35幀、兩份小地圖、三份正常玩家存讀及正常S1 GUI維持，詳驗證045第38輪。
正常補給／修理結果仍缺，083保持READY，不把移動全畫面通過當作正常結果底圖驗收。
083只涵蓋上述呈現。080、062及#79完整八步另驗，不外推其他局面或硬體時長。
