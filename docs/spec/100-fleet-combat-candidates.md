# 100：艦隊交戰候選與原池命令

狀態：CONFORMED，僅本文候選／原池命令資料介面。沿[#79](https://github.com/wicanr2/logh3/issues/79)與
[#48](https://github.com/wicanr2/logh3/issues/48)，完整[092](092-fleet-combat-entry.md)保持DRAFT。

## 來源與證據

- [入口收據](../internals/data/fleet-combat-entry.json)保存036D／037C／03D9原bytes，
  SHA-256 132e4ae3c6ac38ddaf62075d3bd450c3537dda5c8a4a12a4d376424a3f239121。
- [研究403](../re/403-weapon-candidates.md)與[055](../playtest/055-weapon-candidates.md)，
  [候選收據](../internals/data/weapon-candidates.json)SHA-256
  59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f。
- [研究401](../re/401-fleet-combat-command.md)與[053](../playtest/053-fleet-combat-command.md)，
  [命令收據](../internals/data/fleet-combat-command.json)SHA-256
  e3b39582bfda07105d27837747d46bca19eaf2f1fb5d6ff383500870e4a6870d。
- 原版GIN3.COM SHA-256為770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918，
  bundle為de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
- dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c、clean，IDA Pro 9.4。
  原版CS:IP、IDA EA與線性位址分欄保存於收據，不混用。

下列proven限固定bytes及48個候選案例、已驗提交案例。

2026-10-04證據審查：兩份凍結來源SHA符合；原版036D、037C、03D9分別為
2688472e、2689472c、26c6472e03。防禦只寫3，FF取消不寫目標word。
候選48案已包含射程相鄰邊界、庫存高word、遮蔽與後續查詢覆寫。
允許實作本文限定資料介面，不授權以未知自動選擇或演出時序補完整092。

| 原始定位 | 契約 |
|---|---|
| 1F6F:02D2／17A0:00AC | 敵方目前SHIP+34／36、Out bit7未設、FixedDist≤300 |
| 1F6F:032D／0351／035A | Detail 0及Kind 7排除；零艘數不因此排除 |
| 1F6F:0398／056C／0615 | 攻方武器type、朝向+4A與方向橢圓，等於半徑可用 |
| 1F6F:04D0／0520 | 飛彈／戰鬥機32位庫存非零；FF type未搭載 |
| 1F6F:04B3／091C | 僅光線砲驗雙方座標+7整數射線，每第8步驗地形與分隊 |
| 1F6F:0B0A／0C11 | 其他物件原點在雙方原點包圍盒，Out bit7未設；16×16半開命中 |
| 1588:0571／0621 | X/8、Y/10，再各/2取128欄tile與四子格mask |
| 1ED2:036D／037C | 先寫武器+2E；非FF才寫目標池word+2C |

## 型別、規則與失敗

候選函式接受攻方物件索引、按caller順序排列的BattleUnit、攻方朝向、ShipKindTable、
BattleTerrain與128×102地圖。輸出目標物件索引與三bit武器mask，不修改分隊、地圖或RND。
候選順序沿輸入物件，不聲稱Game名冊已符合原版顯示物件排序。
目標+3C是查詢暫存，不新增永久候選欄位，也不讀上次查詢殘值。

- 使用目前X／Y，不使用TargetX／TargetY。目標Out與Action bit7任一成立均排除。
- 查詢先於攻方Action恰為7F的下令門檻；查詢不因攻方Out或零艘數提前返回。
- 三武器用093同一份12筆半軸表。光線砲射線不改成地形LineOpen或連續相交。
- 原版診斷有修改半軸表的案例。共用內部演算法接受獨立表以核這些受控案例，
  正式入口固定用093凍結表，不修改全域表或玩家資產。
- 射線從較小主軸端開始；每第8個整數步檢查，其他步不檢查。
  兩點相同時沒有步進，光線砲射線可用；q=0傷害另依093核准首點。
- 缺艦種、非法type、朝向、陣營、byte欄位、32位庫存、地形或圖塊索引，回error及nil清單。
  不以缺地形當開放射線。世界域限X 0..2047、Y 0..2039，非法外部存檔不外推原版。

Game介面用既有battleUnitPool把物件索引轉原池，並從BattleFleet取朝向。
命令直接保存在現有Slot.Ship的+2C／+2E，不增加JSON／存檔欄位。
重新查候選後才允許武器0／1／2與有效目標；防禦3不需要目標，FF取消保持原目標word。
防禦亦保持原目標word，沿原版只寫+2E的契約。
自身非一般分隊、Action不恰為7F、無合法池或061單人CPU被動均不允許提交。
錯誤不寫任何byte，不取RND、不扣彈藥。

## 垂直鏈與停止線

typed候選從已載入艦種／地形及戰場分隊生成，命令保存原池。
既有SaveSlot沿原池保存並重讀，不增加格式。
目前僅提供正式Game資料介面，完整目標／武器UI與第④步呼叫仍未接，不能宣稱玩家交戰完成。
不替玩家選擇「自動」，不啟用單人CPU攻擊，不把資料介面插進舊ResolveBattleTurn。
時間線、音效、傷害續接與正常UI的完整驗收仍由092／056處理。

## 驗收

- 凍結收據與原版輸入雜湊符合；缺原版skip，雜湊錯誤失敗。
- 48案逐次比較02D2查詢、0382寫入mask與0097返回AX。
  後續查詢覆寫mask的案例須逐次核，不拿末端mask替代。
- 正式Game用原版池資料查詢、提交0／1／2、防禦與取消，原始欄位變動限+2C／+2E。
- SaveSlot及Slot重讀保存目標／武器；非法輸入、缺地形與被動CPU不變動資料。
- 正式bootstrap、主選單及S3到自然首戰，48個一般分隊可查候選且不變動畫面、原池與RND。
  這項只驗正常資料域與唯讀性，不把測試直接呼叫當成玩家武器UI。
- 標準18產品套件通過。100不因局部通過升格完整092或關閉#48／#79。

## 第74輪限定實作結果

[FleetCombatCandidates](../../internal/rules/fleetcandidates.go)與
[Game命令資料介面](../../internal/game/fleetcombatcommand.go)已實作。
48案111次查詢AX、61次mask寫入直接符合原版觀測，不讀舊probe期望。
五種提交／取消／防禦的SaveSlot與Slot重讀、十種非法提交均通過。
正常bootstrap／主選單／S3首戰48個分隊查詢不改寫原池、畫面或RND。
18標準套件13PASS／5無測試；[063驗證](../playtest/063-fleet-command-data.md)保存命令及限制。
[完整索引](../internals/data/fleet-command-data.json)SHA-256
4b42447bfe2bbe899bc1c51987501b9b2c21cd95a1323452c38383f490acea80。
完整武器UI、自動分支、八步結算與Game艦隊演出仍未接，本規格不替它們驗收。
