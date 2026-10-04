# 419：戰場升級後的能力、速度、完整度與費用

2026-10-05第86輪勘誤：[研究420](420-battle-promotion-global.md)證實下文STELLAR觀測起點A7BC偏32 bytes，正確為A79C。
舊STELLAR完整性聲明撤回，已由新雙跑取代；三表與費用結果保持。下文保留第85輪歷史。

沿[Issue #79](https://github.com/wicanr2/logh3/issues/79)與[108草稿](../spec/108-battle-promotion.md)。
[研究418](418-battle-promotion.md)已核成長與通知，本輪補重算資料鏈。
所有結論限於原版診斷；正式8-4尚未接線。

## 輸入與重生

| 輸入／工具 | 固定身分 |
|---|---|
| GIN3.COM | SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| 第82輪父RAM | `workplace/oracle/ownership79-r82-entry-a/mem-in.bin`，SHA-256 `b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900` |
| 完整machine／DOS父保存 | `workplace/oracle/ownership79-r82-entry-a/entry.state`，SHA-256 `598e10e5f9457842275fa5639219636f322d8bb2e0f8393fe25b93bfbdef0cfe` |
| dosgolem | 專用副本 `dosgolem-logh3`，HEAD `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，乾淨 |
| Go | `golang:1.24-bookworm`，固定IRQ0每20000步；每案入口seed `12345678` |
| IDA | Pro 9.4，`ida-pro-9.4-idapython:locked-v1`，UID1000，無網路 |
| remake | 父提交 `78ee450c34485a09109879de0ae548c1f4f66c4e`；本輪沒有產品程式變更 |

命令與原版雙跑入口見[對拍074](../playtest/074-battle-promotion-recalc.md)。
觀測器由固定第84輪Go來源產生，不改舊收據。
自然續行到 `185C:060D` 後才注入，停止於 `185C:09BF`。
所有注入以 `lin:` 記錄，實際報告按下／放開與步數保留在原始JSON。

本輪保存完整ADMIRAL 2296 bytes、FLEET 3904 bytes、SHIP 27456 bytes、
WORK 230 bytes、STELLAR 4032 bytes，以及每次重算返回的五表快照。
原版DS為54B3；表基底分別為DS:B77C、C074、SHIP遠指標DS:CFB4／CFB6、8F14、A7BC。
這些是本次dosgolem執行期定位，不能當成其他版本的固定線性位址。

兩份原版JSON各20108439 bytes，完全相同，SHA-256
`5049465c845318759894b9b1d957c3a2fbc807998608eae1eb0776fb84988c97`。
第二版重算來源索引含22項來源，SHA-256
`cda14b5b122e4e875a825948c61194c96b534a07c7a563c770b26570b35291e6`。
分級審查索引含8項來源，SHA-256
`844038b4feed02e0c1ecb87cfd18f06832b2ecb37639946d9e8db916eeb29a9a`。

## 已證實的重算

以下均為 `proven`。出處是固定RAM的IDA原始bytes與[第二版獨立模型](../../tools/scripts/battle_promotion_recalc_receipt_v2.py)。
模型先驗473個LCG輸出與原版逐次回值，再驗121次嘗試／53次成功。
每次重算後完整三表及最終三表皆相同，不只驗改變的欄位。

| 原始定位 | 寫入契約 |
|---|---|
| `297B:0546` | `05B2`能力／容量 → `0E8B`速度／完整度 → `090A:0E6D`全局重算 |
| `297B:0209..025C` | 可帶組數下限2／6／16／20／24／29／34／39，值1／2／3／4／6／7／8／9，末段10 |
| `297B:05B2..083E` | FLEET+0E與+3D..47；前五項副官較高時取兩者平均，後六項取最大 |
| `297B:0F38／0F64／0F8D` | SHIP+1F／21／22先寫基礎速度低byte，再套折扣 |
| `297B:0FB7..0FBD` | 折扣乘積先取低word，以signed16除100，向零取整 |
| `297B:0FD0／0FE2／0FF7` | 速度byte乘折扣後取低word，再除100；減商低byte保留回繞 |
| `297B:1059..106E` | FLEET+4D..50取各組SHIP+1F／21／22／23最小值 |
| `297B:10B0／10BC／10C8` | 獨立索敵池速度直接取自身陣營的艦種7 |
| `297B:1051／10D0` | 先寫FLEET+3C=100，再呼叫1261:032B重算完整度與旗標 |
| `1261:0491` | 每個仍有艘數的分隊，SHIP+06寫新提督領導力導出的完整度 |
| `1261:03C2／03D0／03F0` | FLEET+3C取存活分隊平均；bit0以50為門檻，跨界設bit2 |
| `1261:03A5` | 沒有存活編制時設bit3，保留重算前剛寫的+3C=100 |
| `297B:116C..12FA` | 全局現役艦隊的+0A／0C費用均重算；SHIP自身陣營／艦種／艘數決定成本 |

速度公式保留原版寬度：

```text
base = byte(kind[08]*ability[6]/100),
       byte(kind[06]*ability[5]/80),
       byte(kind[07]*ability[5]/60)
p = trunc(signed16(lowword((270-2*ability[9])*mobility[groups]))/100)
speed = byte(base - lowword(base*word(p))/100)
```

`mobility` 的11個有效組數值為0、0、7、14、21、28、35、42、49、56、63。
費用先以unsigned32加總成本byte×艘數並除1000，再乘signed的200-營運，
乘積取unsigned32後除100，商低word保存；維持費回傳signed16低word。
能力255診斷會與未截斷的普通整數算式不同，不能偷偷把它夾成200。

新增六案分別把兩陣營現役編制設為半數、一成與零艘，FLEET+3C注入42、bit0設1。
同盟半數案例跨回50以上，原版旗標變4；一成案例維持低完整度，旗標仍1；
零艘診斷兩側旗標變9、完整度100。零艘只驗helper邊界，不稱正常8-4入口。
完整三表增量為ADMIRAL 280、FLEET 344、SHIP 243 bytes。

## 全局預估與剩餘驗收

`090A:0E6D..0E87`依序呼叫2703:02E5與兩個陣營的090A:0E88。
後者經297B:10DE重算兩項費用，再寫WORK+5E維持費及+46造艦預估。
297B:1131呼叫12FB，以signed16加總狀態非2／5的FLEET+0C；1158／115C寫WORK+6E移動費。
可用款是signed32的軍事費＋預估稅－維持費－移動費。
小於0時造艦預估為0，達21474836時先除100再乘百分比，其餘先乘再除100。
本輪用原版重算後預估稅作已觀測輸入，獨立核53次兩陣營維持費、移動費與造艦預估。
不將這項核對稱為完整WORK獨立模型。

STELLAR有637個、WORK有218個改變bytes，完整原版前後雜湊與差分已保存。
經濟力、預估稅與未安定度仍須沿[研究299](299-region-unrest-writers-and-recompute.md)、
[研究304](304-unrest-immediate-recompute-callers.md)及042契約驗證本局面。
不重開已證實的通用公式；下一步只核本次升級caller及typed回寫。
正式報告、兩陣營12級軍階UI、存讀與正常玩家路徑仍缺，108維持DRAFT。

## 原始指令與語意索引

新COM從固定父RAM重生，297B輸入SHA-256
`35c5c5942827225474a3aae0712c5d23f50051a81567bbd513867e3ba49c8517`，
090A輸入SHA-256 `ffd508627b9fb24a08aeea30aadd33dcbac5e36492ebcd64a01a67966e073050`。
兩者均有256-byte前綴，IDA指令基底0x10200；原版線性定位為CS×16+IP。
兩種位址基準在JSON分欄，不互換。

[第二版分級定位索引](../internals/data/battle-promotion-recalc-annotations-v2.json)
SHA-256 `0bcd90c54f828f4cc4b96fcf9744f47d32287e3b255d08c45bf597cc8dd43fd5`。
匯出器自動合併42項已審查語意並核原始bytes；808條原指令的其他766條仍有unknown警示。
保留自動函式名、運算元與原始定位。
05B2正式邊界止於083E retf；初步較大查詢包含下一支083F helper，不能把它算進能力函式。
1261完整度沿研究413既有固定雜湊匯出，本輪沒有再改其資料庫。

## 本輪第一版勘誤

第一版只追090A:0E88的直接寫入，漏看297B:10DE內的12FB呼叫，
誤記「這個caller不重新計算移動費」。
回查[研究393](393-battle-work-control.md)已保存的67條原指令，
確認1131與1158／115C；全部bytes又與本輪父RAM相同，足以推翻該說法。
第二版另核每個返回點的移動費；本批皆0，不能單靠0值證明沒有寫入者。
其他運行狀態的分支契約沿研究393，不宣稱本輪重生所有非零移動費局面。

第一版三表與亂數結果仍有效，但全局費用與090A:0EC5語意以第二版取代。
保留[第一版來源](../internals/data/battle-promotion-recalc.json)、
[第一版分級索引](../internals/data/battle-promotion-recalc-annotations.json)與
[第一版審查](../internals/data/battle-promotion-recalc-review.json)，不改其bytes或雜湊。
其SHA分別為7d39818d、065c1366、3c334573，完整SHA仍記於各自來源及第二版來源鏈。
錯誤在證據解讀，沒有進正式程式。
