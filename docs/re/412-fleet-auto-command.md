# 第75輪：自動目標、武器與預估暫存

日期2026-10-04，父提交529b61a5deb64e7affbb8b66f43b70b069d1e7a2。
沿[#79](https://github.com/wicanr2/logh3/issues/79)與[#48](https://github.com/wicanr2/logh3/issues/48)。
玩家阻塞是手動選「自動」後，正式命令無法接下一分隊；不替未知自動選擇猜武器。

## 方法與來源

採[固定重播](../../tools/combat_player_command_oracle.sh)及[注入計畫](../../tools/scripts/combat_player_command_plan.py)。
父state為第63輪自動／防禦選單，SHA-256
db488d25851ef2af76f599b7057cb34df3931aa34e06fdb2283682f816727e13。
每次相同滑鼠350,250，步數495300000按下，100000步放開，提前200000步移動。
原版virtual400，PNG原生Y加40；不是remake正常玩家路徑。
每一筆lin注入、固定seed、祖先注入、完整旗標及工具版本在[凍結來源](../internals/data/fleet-auto-command.json)。
SHA-256 23627c23a9d2ef93991c842e00fbf43df75f00d2043c78788454e868d3a19401。

31案各雙跑的RAM、307200色號、色盤及PNG一致，無遮罩；567份來源保存。
原始GIN3.COM為770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918，
dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c未改且乾淨。
IDA9.4沿第61輪固定RAM6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4。
兩個有界切片匯出538條原指令，30項proven附加語意。
每條bytes與RAM回核，原名、CS:IP、runtime線性、IDA線性及xref分欄。
本文件位址均是本次原版CS:IP，不拿別的工具地址替換。

## 已證實結果

| 原始定位 | 結果 |
|---|---|
| 1F6F:004D | 命令前+3E存預估剩餘艘數，演出後同一欄改投影X。須按相位解讀 |
| 005F→061F→0714 | 重查候選，再按mask預估各武器。計算無噪音、旗艦修正或庫存消耗 |
| 0855／0867 | unsigned32除法，與實際傷害的signed32分開；OR清CF後JNB必跳 |
| 08D6 | 攻方電腦才加目標權重，非「陣營1」；旗艦×3，kind6×2，kind5低word×3後除2 |
| 024C | 值=低word(RND×FLEET+45/100)，8192／16384門檻選第3／2／1筆 |
| 011C／018B | 缺beam時best保持FFFF，自身+3E減FFFF成+1；武器沿最後候選預估留下的值 |

已驗門檻8191、8192、16383、16384及32767×255的低word回繞。
測試seed由已證實LCG反解，固定首次回值；執行前寫入，不試種子或重擲。
基本局面預估beam780／missile630，選0；有防禦時390／300。
三個真正一般池102／104／105排序780／560／170，分別可選第1／2／3。
缺beam選102時其mask只有2，原版仍可提交武器2。這是已證實的確定性錯誤，保持原版。
純規則與Game自動提交均保留，手動100仍拒絕不在mask的武器。
無候選保持目標word，改防禦3，不取RND；外層已計過手動入口，仍可能進0152。

## 診斷勘誤與實作

第一版凍結器假設所有取樣均未到0152。無候選實際立即返回，外層仍會開場。
同RAM／命令回核後改以RND次數及艘數／庫存不變限定相位，不改原版輸出。
第一次純規則測試誤用實際傷害fixture，拿未提交FFFF目標當池索引，造成測試panic。
改成從當次061F的目標far pointer帶入防禦資料，同一容器乾淨重跑。

早期多目標注入pool103為一般組；原表FLEET16+38指出103是獨立偵察池。
另外三案把旗艦+03改1而未改編制。這十案只驗局部原bytes，不能通過Game一般池檢查。
保留原有證據，另追加五案真正一般池104／105及三案真正非旗艦104。
沒有修改正式池映射來遷就診斷資料。此修正不推翻23案局部計算，但縮限其玩家資料聲明。

[101](../spec/101-fleet-auto-command.md)先READY，才實作SelectFleetAutoCommand與Game自動資料介面。
31案115筆係數、shots、百分比與結果直接符合原版；21案Game原池提交／SaveSlot重讀通過。
十二種純規則錯誤、六種Game錯誤及雙人另一方仍是人類的回歸通過。
061單人CPU拒絕保持；零距離沿核准首點，不讀stack。
完整092、命令UI、正式動畫時間線及正常玩家交戰仍缺，不用這些局部測試關Issue。

驗證入口：[064](../playtest/064-fleet-auto-command.md)、[完整Game驗證索引](../internals/data/fleet-auto-game.json)。
