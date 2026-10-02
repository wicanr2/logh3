# 376：S1第二戰場移動與方向圖示

工作項[#79](https://github.com/wicanr2/logh3/issues/79)，第36輪。
原版GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`、clean；原版未修改。
正式比較範圍在[本輪計畫](../goal/2026-10-02-battle-second-oracle-round-36.md)宣告。

## 原版局面與資料

S1正常載入與戰略過場可進星域8、第1回合。18艦隊／135分隊，鏡頭58,1。
第35輪remake正常首戰為16／103，兩者不能直接作同狀態；此次remake明示入口注入。
首次實行前固定12345678，第一階段消費亂數後移動入口為ECC6CD87，不重擲。
兩次獨立原版43邊界／129檔全部相同，詳[全紀錄](../internals/data/battle-second-frames.json)。
41圈及終態的完整18筆FLEET與135筆SHIP已由正式引擎演算比對，零差異。
唯一入口RAM SHA-256 `9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec`。
終態RAM SHA-256 `4d415a8398cf8af855b62276507d8c4f79e7840f9cc4eb08ba5d45c085c3c6aa`。

## 呈現差分

未修正全畫面43邊界各差874到1104像素。修正前收據SHA-256
`d28e0faa86bf1eb3a456be55cfa2ab55b35d98bb3a9d657e2c382fac4e732b13`，
`workplace/battle79-round36-before-render-receipt.json`。
方向圖示缺漏為proven；重疊分隊的繪製順序及小地圖下方203列差異仍unknown。
不遮罩、不裁切，不把完整資料相同稱為全畫面通過。

## IDA原始定位

工具IDA Pro 9.4、`ida-pro-9.4-idapython:locked-v1`、UID1000。
從入口RAM切段，前置256個零byte；切段offset0對應IDA 0x10200。
原版CS:IP、dosgolem線性與IDA位址在JSON分欄，不混用。
25DB切段SHA-256 `d1004845670387a15986f585658051965a2cb899f1f499d062b7b7ba8bd34b7a`；
3982切段SHA-256 `d80d91d9ce32708d3084e52b0c696375cee250f3c5627b737db0c2c2ba18d407`。
使用既有[非破壞性匯出器](../../tools/scripts/battle_rules_ida.py)，一次性資料庫，無改名。
原始輸出`workplace/ida/re79-second-r36-{25DB,3982,35D1}.json`；完整指令仍unknown，以下逐項審查。

| 原版定位 | 語意 | 等級／證據 |
|---|---|---|
| 25DB:0408..0438 | 強制完整顯示或SHIP+17為3，旗艦SHIP+3為0，呼叫04E7旗標1 | proven，原始分支與far call |
| 25DB:0466..0487 | 非完整顯示的旗艦呼叫04E7旗標0 | proven，原始分支與far call |
| 25DB:04E7..0513 | 旗標非0將Facing+4A轉方向寫Fleet+6A，否則寫9 | proven，`26 88 47 6a`、`26 c6 47 6a 09` |
| 25DB:0670..0713 | 方向位置以Fleet+5C/+5E減鏡頭，另加DS:4A3C方向位移 | proven，兩次world座標與表加法 |
| 25DB:0865..08A1 | Fleet+6B推送圖示物件+3 | proven；等順位的排序仍unknown |
| 25DB:08A2..0914 | Fleet+6A選圖塊，3982:0001指標寫入物件+0E/+10 | proven |
| 3982:0001..0021 | DS:F12C/F12E為圖塊池，每塊128 bytes | proven，索引左移7 |
| 3982:0022..0037 | direction=((Facing>>4)+1)>>1 &7 | proven，原始整數指令 |
| 35D1:0009..005C | 圖示加入DS:EAE0清單，再呼叫028D | proven；028D排序與重畫次序待續 |

DS基底為dosgolem線性0x54B30。DS:4A3C八筆有號word座標為
`(16,0),(8,8),(0,16),(-8,8),(-16,0),(-8,-8),(0,-16),(8,-8)`。
原版方向圖塊池線性0x401BE，與SPRITE.SPR前八塊逐位元組核對。
未達proven的指令保留unknown，不從本段散文自動提升索引。
19筆逐項審查已由[證據工具](../../tools/scripts/battle_second_render_evidence.py)核對原始RAM bytes並
寫入[分級索引](../internals/data/battle-directions.json)。完整IDA匯出保留原始函式名與未知指令。

35D1:028D..02AA已另建一次性切段核對，呼叫3C6A:000D，元素4 bytes、
筆數DS:EADE、陣列DS:EAE0、callback35D1:02AB。完整匯出
`workplace/ida/re79-second-r36-sort.json`保留；同順位效果未驗，不為此深入通用排序helper。

## 方向圖示限定修正

084先DRAFT再READY，原版1024 bytes及43邊界344筆方向與Facing相同後接正式renderer。
43幀各440個方向像素與原版物件所指的位置相同，合計18920。這些每幀RAM只用於斷言，
沒有餵回Game。剩餘全畫面差434到664，繪製順序與y203仍unknown。
原版整份FLEET／SHIP仍全部相同，S3原有35邊界仍通過。
測試改用明確的Records名稱，只對完整資料及方向下通過斷言；未通過的全畫面逐幀記入log及收據。
舊完整畫面失敗log保留，不更新golden、不遮罩、不將有限通過冒稱整畫面一致。

## 重生入口

`bash tools/battle_second_oracle.sh battle79-second-r36-a`，b為另一次獨立重播。
`bash tools/py.sh /work/tools/scripts/battle_second_plan.py export`核完整雙跑。
父快照與旗標見計畫及工具，探索輸出另列、不升格；原版檔與dump不入Git。
方向圖示規格見[084](../spec/084-battle-fleet-directions.md)，完整畫面仍依[驗證045](../playtest/045-battle-movement-render.md)管理。

在沒有同名輸出與資料庫的新工作區，以下重生原始IDA證據；已有DB時不連續重開。

```sh
bash tools/py.sh /work/tools/scripts/battle_second_render_evidence.py prepare
timeout 60 bash tools/ida.sh raw idat -A '-S/idascripts/battle_rules_ida.py /work/re79-second-r36-25DB.json 25DB 03E9:auto 04E7:auto 0670:auto 0865:auto 08A2:auto --index=/work/re79-second-r36-25DB.index.json' /work/re79-second-r36-25DB.com
timeout 60 bash tools/ida.sh raw idat -A '-S/idascripts/battle_rules_ida.py /work/re79-second-r36-3982.json 3982 0001:auto 0022:auto --index=/work/re79-second-r36-3982.index.json' /work/re79-second-r36-3982.com
timeout 60 bash tools/ida.sh raw idat -A '-S/idascripts/battle_rules_ida.py /work/re79-second-r36-35D1.json 35D1 0009:auto 0194:auto --index=/work/re79-second-r36-35D1.index.json' /work/re79-second-r36-35D1.com
bash tools/py.sh /work/tools/scripts/battle_second_render_evidence.py
```
