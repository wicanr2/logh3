# 414：交戰命令顯示

2026-10-04，第77輪。父提交e30f3869b58db2d484f6109931e5181e784026b7。
沿[#48](https://github.com/wicanr2/logh3/issues/48)與[#79](https://github.com/wicanr2/logh3/issues/79)。

## 固定來源

原版bundle SHA-256 de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
GIN3.COM SHA-256 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
dosgolem專用副本f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c、clean，未修改。
靜態1MiB RAM沿normal79-r61-a/end-mem.bin，SHA-256
6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4。
命令父快照及原版輸入沿[研究401](401-fleet-combat-command.md)。

由[IDA匯出器](../../tools/scripts/fleet_command_display_ida.py)在一次性COM資料庫匯出。
工具IDA Pro 9.4、UID1000。COM前256 bytes為包裝，IDA基址10200；
每條保留原CS:IP、IDA EA、dosgolem線性位址及bytes，不覆寫原始名稱。
自動函式名與未審查指令保持unknown。

## 顯示資料流

以下proven限固定原始bytes及新原版畫面。

| 原CS:IP | 原始定位與已證實用途 |
|---|---|
| 1ED2:0218 | 呼叫165F:034F，掛射程及十字 |
| 165F:0558／065A／0751 | 艦種+0A／0E／12的FF門檻；飛彈及戰鬥機另讀32-bit庫存 |
| 165F:0585／0694／078B／07EC | 中心SHIP+34／36各加8 |
| 165F:05A2／06B1／0800 | 原FLEET+4A朝向 |
| 165F:05DB／06EC／07A3／0809 | XOR色號6／1／5／7 |
| 165F:0815..0821 | 十字前、後、側皆0800，實線FFFF |
| 1ED2:03F7 | 子選單呼叫3870:01AC隱藏圖形 |
| 1ED2:0781 | 候選16×16半開命中；沿研究401 |
| 1ED2:084A／08CB | SHIP+17恰3完整身份、恰2可見艦種，其他未知 |
| 1ED2:0912..0943 | 兩支分隊目前位置算整數距離，DS456F格式尾接0000 |
| 1ED2:0960 | 離開候選時清後四列 |

DS4534保留未知艦隊號，DS4541保留八個問號。
DS455E保留未知艦種及艘數，不用己方完整面板補資訊。
DS456F原格式為距離、兩格空白、%6lu及字面0000。
原版總面板八列；目前分隊四列及敵方四列各佔80 pixels。

新實驗由[原版重播器](../../tools/fleet_command_display_oracle.sh)產生，
[注入計畫](../../tools/scripts/fleet_command_display_plan.py)記錄池102情報變因、父狀態、完整旗標與seed。
不稱正常原版整段玩家路徑，也不由重現性單獨宣稱remake通過。

七組前後RAM、色號、色盤及PNG雙跑逐byte相同，seed保持12345678。
gzip／gob的快照bytes因Ports／PortsIn的map排列不同。
完整機器欄位解碼後逐項相同，DOS段逐byte相同；來源是dosgolem
internal/machine/state.go及internal/state/state.go的既有格式，未改執行器。
這項差異屬序列化，不能誤記為玩家狀態或畫面不確定性。
103的READY限一般攻擊分隊、前後四列與已定位圖形；不推廣補給／修理／斥候分支。

## 圓形、編號與剩餘順位

新補39D7:008C..016D、3982:0038..0073及25DB:0E78..0F93完整相關常式，
其後的未完整範圍保持unknown。固定COM与JSON雜湊由Game索引收錄，原始bytes逐條核RAM。

| 原CS:IP及bytes | 分級語意 |
|---|---|
| 39D7:0098 `8b460a` | proven：半徑同時是x及誤差初值，y初值0 |
| 39D7:0136 `8b46fc`、0139 `ff46fc` | proven：先取舊y，遞增y後以舊值算2*y-1 |
| 39D7:0147 `790e` | proven：誤差負值才減x並補2*(舊x-1)，八對稱點逐次XOR |
| 3982:003E `3c0a`、0040 `7317` | proven：unsigned編號小於10用全形，其他兩半形 |
| 25DB:0F17 `26884703`、0F33 `2688470c` | proven：亮標priority FF、圖塊10；0EDD至0F62沒有word+4寫入 |

第一次診斷還原只讀Action bit7，漏掉注入前已建立、仍在清單的友軍顯示物件。
校正為依DS9896／98BA既有清單及SHIP+2F、FLEET+6A快取還原，不注入原版像素。
這是顯示診斷還原，不宣稱所有typed旗標與原池完全相同。

校正後第76輪原有繪法七圖差6404／6658／6237／6057／4670／4249／4069；
接射程、亮標、前後四列與全形編號後，七圖各剩一像素。原版及正式renderer有效16色相同。
現在亮標物件線性74C20的word+4為8224，方向物件7101E為64，priority皆FF。
原版按字鍵降序，方向覆蓋亮標，(255,255)原版12、remake5。
關係與未初始化根因沿[研究394](394-battle-object-keys.md)一致。
既有#89忠實度裁決仍pending；沒有改正式順位，也沒有排除該像素。完整全圖驗收FAIL，103仍READY。

正常Game及正式bootstrap的兩陣營命令畫面與六份保存另列[驗證066](../playtest/066-fleet-command-display.md)。
正常路徑無位置／命令注入，只在執行前固定測試seed；不與本段受控原版畫面互換身分。
