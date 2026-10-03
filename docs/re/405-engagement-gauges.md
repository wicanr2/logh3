# 405：交戰量表的資料與呼叫者

2026-10-04，第68輪；沿[#48](https://github.com/wicanr2/logh3/issues/48)。
父remake提交14e9a6a5199e26c5d151bdaf47c72552ef7687d2，研究時工作樹含本輪修改。
入口：[計畫](../goal/2026-10-04-engagement-gauges-round-68.md)、
[094](../spec/094-engagement-gauges.md)、[事實37](../internals/37-ani-pnl-a98.md)。

## 輸入與工具

| 輸入 | SHA-256 |
|---|---|
| GIN3.COM | 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918 |
| 原版bundle | de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673 |
| combat79-r63-laser-submit-a/selected.state | af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d |
| normal79-r61-a/end-mem.bin，IDA切段來源 | 6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4 |
| 本輪原版Go探針 | 8421358bcd560707dbd0b44e7b7302521893a81569affd861a25d35d60a08e0c |
| 凍結JSON | e1ed3aad4d7577f82df4449c097cee0461144ef0f5e555b669fac5c9a2914d86 |

dosgolem專用副本f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean，Go1.24.13；
Go容器映像sha256:1a6d4452c65dea36aac2e2d606b01b4a029ec90cc1ae53890540ce6173ea77ac。
IDA9.4 locked-v1映像sha256:6f6d59af49d0008c4109a5295b5f374bdc007e2d1ab28cb9de08779584de2780。
COM複本為256-byte前綴與完整64KiB段。各輸入雜湊及IDA EA基準保存在JSON；
本文定位皆為原版CS:IP，線性與IDA EA在逐指令資料分欄，不相混用。

## 方法與裁決

從完整state運行950,970步到自然321F:0D11入口，base step496550969，FLEET12線性6115C。
沿用原版640×480畫布與原版色盤；重置完整Machine snapshot後，分別直接呼叫0D11／0DF7。
每案只明示改FLEET+1與+3C，或PLANET+14；far參數及返回FFF0:0000也在報告記錄。
父state既有玩家輸入及局面不冒稱本輪正常路徑，首行星指標54560只供直接常式診斷。
基準seed4EA64DA8在執行前已固定於父state；量表沒有RND呼叫，不取巧挑結果。

`bash tools/engagement_gauges_oracle.sh a`與`b`各32案，68檔逐byte相同，報告除命令相同。
輸出在workplace/oracle/gauges48-r68-{a,b}；140來源指紋與完整原指令凍結於JSON。
兩側各7完整度及9行星防衛力／分母樣本，驗空、滿、整數捨去及word倍增溢位。

已證實0D11取FLEET+3C×66÷100。已證實0DF7先以uint16(分母×2)比較PLANET+14，
達門檻滿66，否則乘33再除分母。0／1防衛力與32768分母均滿66；分母0由原始分支可證不除零。
實際繪圖在39D7:0647／0854，參數y23..88，framebuffer增40；端點與三欄右邊色均由原生色號核對。

已證實20B4:08FB及09E9呼叫203A:0424，給防衛攻擊前後的行星量表分母。
20B4:069F取同一攻擊力，在06A6保存，0754／086E沿用於登陸量表。
因此舊「防衛力／艦數比例」說法不足，不能拿文字上限作量表分母。

## 實作與限制

094證據審查後READY，再實作單一screen幾何與正式防衛分支快照。
32例直接讀凍結原生像素，23,232像素零差；規則快照不改戰損與seed，雙側Game結果幀及動畫關閉通過。
完整圖、色盤與正常路徑仍未閉合。Game.Palette目前回固定主選單色盤，與原版base色盤不同；
局部RGB對照使用同一原版palette，不能當正式Game RGB一致證據。
完整056、092保持DRAFT；其他標題及下方方塊仍缺。驗證見[057](../playtest/057-engagement-gauges.md)。
