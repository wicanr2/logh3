# 398：可見艦隊數與無戰鬥清理

2026-10-03，第60輪，#59。父提交22af102d8049f19d54699dd24bf88fb35f988eaf。
入口：[工作快照](../goal/2026-10-03-region-visible-fleets-round-60.md)、
[091](../spec/091-region-visible-fleets.md)、[對拍050](../playtest/050-region-visible-fleets.md)。

## 已證實的寫入生命週期

2703:0214重數編制內至少一筆索敵等級非零的艦隊。帝國寫星域+2A，同盟寫+2B；
一支有多組可見仍只加一。不以艘數、旗標、星域擁有者或觀看座位篩選。
043B掃32星域；04F8:003E在玩家命令入口先重算實數，再重算可見數。
新遊戲初始化01F0也叫0214。保存照當時快照，正式讀檔經共同入口重算。
第59輪LoadSlot保留注入值只證明當時remake行為，不能作原版契約。
FF顯示占位沿090；有效艦隊最多32支，正常重算不會導出FF。

field_xref掃描只提供候選。272EF的add bx,2A先取址，之後才加陣營並inc es:[bx]。
不能從直接寫入少判定沒有寫入者。xref_linear按同線性目標掃出0214兩個far-call、
043B一個far-call，再由IDA核對。未以掃描數量否定未覆蓋的間接路徑。

## 垂直補驗發現的缺口

0374:03E8對沒有戰鬥的星域走0556，在攻擊行星或被動效果前，呼叫0AF2。
它將所在星域的艦隊+48完整byte設4，編制內各艦艇+17清零。
正式scanRegionTail漏接，已依READY契約補上，typed與Slot兩表同步。
獨立偵察隊、衰減計數、其他索敵旗標與座標保持。戰鬥分支不走這支清理。

早注入索敵1時，原版先由297B:097F到達配置在430533538清零。
舊reachFleetDecided資源缺戰術配置資料，單一早注入失敗不能宣稱正式到達配置壞了。
另在432932986於無戰鬥清理入口後注入Detail1／Action2，隔離0AF2；
peace兩跑完整RAM相同，正式helper的完整艦隊／艦艇表與原版相同。
432933022寫索敵0；watch記的是寫完後IP0B62，IDA實際指令是0B5D。
補驗不改golden遷就程式；原失敗保存於round60-focused-missing-clear.log。

## 來源與位址空間

- 原版bundle：org_game/银河英雄传说3SP.zip，SHA-256 de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
- GIN3.COM：770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
- 父RAM：hold60-r58-base/mem.bin，d1c815daa08404666898036520a6fe32210e20880725f47fa48bb20ce16f71be。
- dosgolem專用副本：f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，clean，tick20000，mode12h。
- IDA9.4 locked-v1：6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。
- Go1.24-bookworm：1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。
- GUI logh3-ebiten:2：6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8。

原版CS:IP、dosgolem線性與IDA COM位址分欄。固定RAM切段加256零bytes，
IDA基底取第一段起點加100h。三份匯出327條指令，31項附加語意proven。
保留原始定位、operand與bytes；未審查的附加語意標unknown。
關鍵bytes在091，逐檔雜湊與完整旗標在[凍結收據](../internals/data/region-visible-fleets.json)。

normal不注入，原版與Game正常一回合後清單全圖零差，1×／2×真實GUI也相同。
late1／late255在043B入口後注入，只驗全部64格導出值；clear／peace分開驗兩個清理者。
保存／讀檔與雙人換邊驗資料接線，不宣稱雙人完整GUI或原版與remake全RAM相同。
索敵戰術本身仍由087／#79驗；#89的一像素待裁決保持。

凍結收據SHA-256：07b5c05e76c3435d84822308bc17667157c91e062f6d3dc31dded80a84c7b3c0。
18產品套件通過，focused零FAIL／SKIP；原版與Game雙跑及正常GUI均通過。
