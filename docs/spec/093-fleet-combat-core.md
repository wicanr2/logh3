# 093：艦隊交戰戰損與排序核心

狀態：READY。沿[#79](https://github.com/wicanr2/logh3/issues/79)及[#48](https://github.com/wicanr2/logh3/issues/48)。
完整[092](092-fleet-combat-entry.md)仍DRAFT；此子契約只處理已提交命令的單擊資料與交戰排序。
正式Game尚未呼叫核心，命令、演出、等待、全回合與存檔接線仍由092驗收。

## 證據

- [研究402](../re/402-fleet-combat-damage.md)、[戰損收據](../internals/data/fleet-combat-damage.json)，
  SHA-256 360277c8dcbf6214da429baead75d4a7bf20b09b03f1ca2af3c0313df4171b30。
- [研究403](../re/403-weapon-candidates.md)、[候選收據](../internals/data/weapon-candidates.json)，
  SHA-256 59a5ad2bd68a54c888179be3d85a6b99c15f16a1892b7d2b05c4de0667ad601f。
- [研究404](../re/404-combat-sequence.md)、[序列收據](../internals/data/combat-sequence.json)，
  SHA-256 199f328b4463d49a817e54c2ade49dd148a24d45e94e94d135488d699c3205a8。
- 原版GIN3.COM SHA-256 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918；
  bundle SHA-256 de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
- dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean，IDA9.4，輸出UID1000。
  位址為原版CS:IP；IDA EA及dosgolem線性位址在各收據分欄。

下列契約proven限固定原指令與受控輸入；零距離是使用者核准差異。

2026-10-04證據審查：三份凍結收據SHA-256符合；重新讀取第61輪原版RAM
`workplace/oracle/normal79-r61-a/end-mem.bin`，SHA-256
6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4，
確認12組武器全六點、五筆排序基值及11筆機動係數。第64／66輪直接暫存器、
堆疊及RAM輸出涵蓋本文算式與排序；允許實作純核心，不擴張092的玩家路徑閘門。

## 型別與資料

- 攻方、目標沿既有BattleUnit。武器0／1／2／3／FF與目標池由命令層提供，不更改存檔格式。
- 艦種武器type及每艘件數沿ShipKindTable的+0A／0B、+0E／0F、+12／13。
- 武器表為DS4FDA的12組power／a／b及DS477C指向的六點曲線。
  X為0、0.25、0.5、0.75、1.0、0x7fffffff，Y為原版u16。type4只保存原版表，不稱正常選單已驗。
- 攻／守三能力對應FLEET+3D／3E／3F，朝向+4A，目標防禦3由命令層提供。
- 目標提督ADML+12與操作模式只作傷害輸入；本核心不替玩家選武器、不啟用CPU攻擊。
- 排序輸入按原版FLEET編制順序提供全部一般分隊，不預先刪FF／3或零艘數。
  每筆帶SHIP+06、提督ADML+0B及機動力；kind+09查DS4426基值。

## 單擊

1. FF／3、攻方或目標零艘數不執行傷害，不取傷害RND、不扣庫存。
2. q=FixedDist／[a×(1+e)]，依六點曲線插值；下降斜率取signed32乘積高word。
   q=0直接取首點，依使用者2026-10-04裁決；不讀原版殘留stack。
3. 曲線Y×power／100，再乘攻方對應能力／100。防禦依目標裝甲橢圓、來襲方向、朝向與對應能力。
4. 每艘件數×攻方艘數。飛彈夾庫存並全扣；戰鬥機夾庫存、先扣20%，艦種3傷害用件數再除二。
5. 百分比為signed32(shots×attack)／defence／750的低word，夾100。
   防禦3減半。旗艦屬性大於100按100／屬性縮減；其餘旗艦pct≥80且CPU模式1減10。
6. requested=int16(N×pct／100 + RND()%200 − 屬性)；N<100只將亂數差額除二往零截斷。
   負值夾0，零值另取RND()%10。沿ApplyShipLoss扣艘數、修理上限與彈藥。
7. 回傳實扣／20分數、實扣／2提督及實扣／4副官功績增量，逐擊除法。
   核心不直接寫總分／提督表；caller依目標敵方陣營累加並刷新艦隊完整度。

本核心回傳攻／防係數、曲線Y、q、射擊量、pct、requested、實扣及增量，供演出與保存共用。
同目標後一擊必須讀前一擊已修改的BattleUnit，不重建未受損目標副本。

## 排序

key=uint16(RND()%base+base+RND()%提督+0B+提督+0B+SHIP+06+機動力)。
對全部一般分隊先取兩個RND，再以unsigned word鍵降序梳狀排序。
回傳原輸入索引的排列及鍵，不修改caller名冊順序、不篩命令。
機動力為byte(100 − signed16[(270−2×ops)×DS4E2E[groups]]／100)。
ops為提督／副官+17較大值，groups限0至10。

## 失敗與驗收

非法武器／艦種／缺表／零防禦／缺RND回明確錯誤，不猜fallback。
先驗證再扣庫存或取RND；排序先驗全部條目，失敗時不消費RND。
原版會除零的未合法輸入不仿效崩潰，保留為明確資料錯誤，不稱該範圍parity。

- 原版資料缺檔時oracle測試skip，不放合成替代品。
- 直接核第64輪有效單擊、11種近距離、合法戰鬥機及第66輪兩擊。
  指定係數、逐擊庫存／艘數／修理上限、分數／功績增量與seed須符合原版。
- 八筆排序鍵及完整順序須符合原版排序邊界，不能只驗活躍兩筆。
- 零距離只驗首點及其已核准差異，不宣稱原版殘留stack parity。
- 標準產品測試通過；正式Game／UI／保存尚未接線，維持READY，不升格完整092／062。

原版RAM、state與PNG留本機；保存數值規則及可重現測試，不內嵌原版執行檔／音畫。

## 第67輪實作與限定驗證

[核心](../../internal/rules/fleetcombat.go)提供ApplyFleetCombatDamage、FleetCombatOrder及
FleetCombatMobility。先驗輸入再變動庫存或亂數；不新增保存欄位。
[測試](../../internal/rules/fleetcombat_test.go)直接核凍結索引的原版RAM及暫存器，
不讀第64／66輪probe計算值作期望。原版缺檔skip，雜湊不符失敗。

- 31個戰損案例及11種近距離：係數、件數、百分比、requested、actual直接對原版AX／BP。
- 合法戰鬥機：原版收據只有入口與末端觀測，核呼叫數、庫存、戰損、分數、功績及seed。
- 兩組連打：八鍵、完整順序、攻守指標、逐擊係數、中間分隊狀態及末端資料符合。
- 12個零距離及stack反例：依原版表首點驗核准規則，不比原版殘留stack損失。
- 原表12組六點、排序基值、機動係數與艦種表均直接核完整RAM。
- 無效命令、零防禦、缺表、零滿編及排序後段非法輸入不變動資料或RND。

六組核心測試無skip通過，標準18產品套件正常結束。驗證索引保存於
[純核心收據](../internals/data/fleet-combat-core.json)，由
[凍結器](../../tools/scripts/fleet_combat_core_receipt.py)重建。
正式Game尚未呼叫核心，UI、完整演出與保存接線仍缺；093維持READY，092維持DRAFT。
