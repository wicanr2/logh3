# 108：戰場提督與副官升級

狀態：**READY**。2026-10-05，原版成長、三表、全局及路徑快取證據審查通過。
遠端#79，完整062仍DRAFT。第87輪已接正式8-4，66案原版五表、11報告、九案六檔存讀、兩案快取及正常首回合續接通過。
正常自然功績升級、正常升級GUI與8-1清空分隊的輸入鏈仍待驗，保持READY；見[076](../playtest/076-battle-promotion-game.md)。

玩家在第8步索敵回收後取得提督／副官升級通知。
第85輪已補原版66案雙跑、53次完整ADMIRAL／FLEET／SHIP重算與473個LCG輸出。
第86輪已補53次完整全局STELLAR／WORK及四路徑矩陣，見[研究420](../re/420-battle-promotion-global.md)。
結構化資料回寫契約與保存消費端已核於[研究421](../re/421-battle-promotion-save.md)。
正式報告、軍階UI及診斷存讀已由產品驗收；正常公開實行與保存另抽樣，不由原版probe代替。
原版會改SHIP速度與分隊士氣，不能只更新ADMIRAL。

輸入為唯讀GIN3.COM，SHA-256
`770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
主工具為固定dosgolem f7523f9與IDA Pro 9.4，
一次性COM位址、原版CS:IP與線性位址分列。

[本輪計畫](../goal/2026-10-04-battle-promotion-round-84.md)、
[IDA輸入產生器](../../tools/scripts/battle_promotion_prepare.py)、
[完整062](062-battle-turn-eight-steps.md)。

已證實輸入為ADMIRAL 28-byte記錄、FLEET 122-byte記錄、SHIP 78-byte記錄，
參戰清單順序與固定受控亂數。成長門檻、12級軍階、旗標及通知見
[研究418](../re/418-battle-promotion.md)、[原版診斷073](../playtest/073-battle-promotion-oracle.md)。
原版來源索引SHA-256 `ab96e1d8d0213da5dfc8b34acc1f46bf1b83440eb3d552ed7f37efd8c2b0c661`；
分級指令索引SHA-256 `3d52c61e0e93433297f6240078614c11da38cf27e24347a4cfe926cfc4f451b6`。

原版提督通知固定旗標2。每次成功依序重算、報告等待、再試同一人。
副官使用實際成長旗標OR1。正式實作需保持此順序。
現有RankOf的28門檻與10級上限由新原指令推翻；修正納入本契約。

297B:05B2／0E8B的能力、容量、三項速度、四項最小值、完整度與旗標已補證。
090A:0E6D會更新全局費用與造艦預估，不只是升級艦隊本身。
公式、原始bytes與66案邊界見[研究419](../re/419-battle-promotion-recalc.md)、
[原版診斷074](../playtest/074-battle-promotion-recalc.md)。
第二版來源SHA-256 `cda14b5b122e4e875a825948c61194c96b534a07c7a563c770b26570b35291e6`。
速度先截byte、折扣乘積取signed16低word；費用保持unsigned32乘除與低word保存。
零艘診斷先寫完整度100再設bit3，不用remake既有完整度helper的零回值猜補。

本局面全局經濟力／預估稅／未安定度已由既有規則組合驗算，53次完整表相同。
STELLAR基底為DS:A79C，第85輪的A7BC偏32 bytes觀測已由新雙跑取代；原三表與費用證據保持。
模型使用重算前行星與歸屬自行建路徑，不使用返回後稅額或未映射路徑的EMS frame作輸入。
窄回寫與保存邊界見下方契約。
正式8-3返回後按參戰順序，提督升級成功即重算／報告，再試同一人，副官最後處理。
每則報告等待用原版word的訊息速度乘二，返回後恢復設定，不改玩家永久設定。
驗收需比較每次三表、全局更新、報告文字／矩形／頭像／時點、軍階UI及存讀；
正常公開入口補抽樣，不以功績注入的診斷代替正常玩家證據。
正常玩家與GUI缺口不由局部診斷代替。研究索引皆保留未證實語意與停止線。

## 輸入、規則與失敗模式

1. 使用40個已證實的u16功績門檻、12個軍階下限及兩陣營12個Big5軍階，
   資料由凍結的battle-promotion.json／battle-promotion-review.json產生，禁止另抄UI表。
2. 提督28-byte記錄：+2為u32功績，+6為byte等級；等級恰99停止。
   門檻取min(level,39)，足額扣一次並以byte加一，255回繞0；不足額不取RND、不寫入。
   成長順序、跨度與旗標完全沿研究418的獨立模型，每個RND用既有rules.Rand.Next。
3. 每次成功即以目前提督／副官重算指定艦隊；再依研究419及420重算全局。
   重算使用既有SHIP自身陣營與艦種，不重新配置池。
   能力255及零艘是明示診斷，保持已核的byte／word／u32回繞，不聲稱自然生涯域。
4. 全局騷動讀既有路徑快取，不能因8-2換手提前重建。
   第87輪固定換手及同值控制均符合舊快取；換手後重建有684個矩陣項不同。
   快取在戰略處理開始保存，跨本次戰場及戰術回合保留；下一次戰略處理重新建立。
   LoadSlot清除暫存，不改存檔格式。診斷入口明示給快取或從入口歸屬建立，後者不覆蓋此前換手。
5. 規則輸入為Slot的完整五表、Planets、ShipKindTable、mobility、UnrestParams、StarGraph、UnrestPaths、當下Military及Rand。
   Game經費以g.military為現況來源，不能讀尚未保存的WORK底稿；規則診斷從固定WORK解出相同輸入。
   表長、索引、艦種／陣營、編制數、存活分隊滿編分母與路徑尺寸須有效。
   不合法資料回錯誤，原池與RND不改；成功提交只寫已證實欄位，沒有新格式或猜補值。
6. 不足升級的合法記錄保持原樣；該階段不清其他步驟留下的旗標。

## 原池與結構化資料

| 原始輸出 | 必須同步的現行欄位 |
|---|---|
| ADMIRAL+2／+6與十項能力 | g.res.Admiral、g.admirals及Slot.Admiral共用本局可寫表 |
| FLEET+00／0E／3C／47 | Fleet.Flags、Expected、Morale、Operations，避免SyncFleets覆寫 |
| FLEET+3D..47／4D..50 | BattleFleet.Defence、Ability0／3、Integrity、AdmiralLossFactor、Flag0、TurnSpeed |
| SHIP+06／1F／21／22 | Composition.Morale及BattleUnit.Morale／TurnSpeed／CruiseSpeed／FightSpeed；索敵池只更新速度 |
| STELLAR+1A與+7D | 目前星域表就地寫入，其他星域byte保持 |
| WORK+46／56／5E／6E | Slot.ShipBudget／TaxTotal／Upkeep／MoveCost及g.taxTotal，其他設定／經費底稿保持 |

既有XY、目的地、Budget／PerTurn、命令、標記、艘數、兩格庫存、修理上限及索敵Mode／State不由8-4回寫。
保存只保存本步結果，不重跑升級、重算或取RND。

## 通知、軍階與玩家續接

8-3返回後，依參戰清單處理提督，成功重算後才顯示該次報告。
提督報告固定flags2，副官報告actualFlags OR1；沒有副官時跳過。
僅提督自己的陣營為人類座位才印，字串與頭像使用該提督索引的低byte。
等待返回後才重試同一人，該人不足或99時才換副官，再換下一艦隊。

格式使用原版DS3DE7標頭、3DEF軍階或3DFE能力升級文字、3E0F..3E96的十項能力文字。
DS3DE7的第一參數為DS60F4的帶色陣營字串，第二參數為DS6428的姓名；
已由185C:072E..075F原始指令核實，沿現有FleetWindow陣營字串來源。
bit1的軍階訊息優先於bit0；其後依位序追加十項能力訊息，保留原版「晉昇」字樣。
軍階下限為0／2／4／6／8／12／16／20／24／29／34／39。
帝國索引10為一級上將，同盟索引10仍為上將，兩側索引11為元帥。
所有清單、詳細資料、編成、副官、報告與敵方資料共用同一份兩陣營軍階來源。

本通知使用word低16位的MsgSpeed乘二，等待設定只附於該則通知，不修改永久設定。
零等待值沿原版診斷需實際滑鼠關閉；左右按下只結束目前一則，放開不額外推進。
既有報告矩形與renderer沿059／062，不另定幾何或豁免字庫差異。

## 驗收與停止線

- 66案每次成長flags、RND回值與末端seed、完整ADMIRAL／FLEET／SHIP／STELLAR／WORK須符合原版。
- 逐則文字、頭像、等待時點與重試順序須符合原版；提督與副官報告矩形逐區統計，無遮罩。
- 星域換手診斷須保留舊路徑快取，不能用新歸屬重建湊通過。
- 結構化欄位同步後，SaveSlot六檔及重新讀取應保持本步結果；不宣稱戰場LoadSlot續戰。
- 正常bootstrap／公開輸入抽測8-3→8-4→8-5入口，不以注入功績代替正常升級證據。
- 執行標準產品套件及受影響GUI最小抽樣。完整062、正常原版整段同態與其他八步仍另驗。

機制與字串證據為proven。素材沿本專案原版權利邊界，不因新資料表而公開原版素材。
規格足以實作即停止RE，未驗玩家／GUI保持READY，不先標CONFORMED。
