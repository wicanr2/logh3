# 378：小地圖尾列來自 bitmap 的越界來源取樣

工作項[#79](https://github.com/wicanr2/logh3/issues/79)，[第38輪](../goal/2026-10-02-battle-residue-source-round-38.md)。
父提交7cac62f6ba7ff2278b2a376948b15c3207cd0fe5。
本輪原版執行器未改，固定dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c、clean。
GIN3.COM SHA-256 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
入口RAM SHA-256 9095dfa08ab36201784e6d4555b5e42312ac76bb20049614cf53bf93bc09e7ec。
所有bytes、分級定位、切段／輸出雜湊、完整探索旗標由[索引](../internals/data/battle-minimap-bitmap.json)保存，
重生工具為[battle_minimap_evidence.py](../../tools/scripts/battle_minimap_evidence.py)。

## 最後寫入者

從固定750M戰略父狀態續跑，相同三次600,200點擊，觀察VRAM byte3FAE，步數止於900M。
最後一次寫入在861624648步，trace的下一指令2376:0B93，實際指令2376:0B90 `26 20 05`。
mode3、前景1、值CF，緊接背景填色3。格表字形確曾寫入，但它不是最後寫入者。
同一父狀態的寄存器探測在0B90讀到DS=706F、首byte讀後SI=7、DI=17FC。
所以來源706F:0006，目標(480,76)，128列後最後一列是y203。
堆疊返回1727:0051，x480、虛擬y36，前景1、背景3；原版helper已把y加40。
寄存器上限2000只覆蓋前2000次讀取，尾48次未捕捉，不把它冒充最後一列trace。

## 建立流程

IDA Pro9.4既有image，由入口RAM另切1727與2376段，前置256個零byte，offset0對應IDA0x10200。
1727切段SHA-256 d611cce37054e0a605def01a178c1ec350c5d7d0bf6e30ec9488ba24978d9421；
2376切段SHA-256 6f8f3986efe5936572c28a40dd26496d5fca3b8ed2303a69bb5f1e61a2106a3e。
第一次查詢000A／005A落在far call內部，解碼無效；保留診斷輸出，改以新資料庫查0004／002F／05C6。
不把自動產生的函式名或邊界當證據。完整blit重新由真正入口0B44查詢。

| 原始定位與bytes | 已證實語意 |
|---|---|
| 1727:0004 `b8 00 08`、0008 allocator far call | 配置2048-byte bitmap，回傳指標存DS:A422／A424 |
| 0026 `9a c6 05 27 17` | 以DS:3C1A／3C1C地圖指標建立bitmap |
| 0607 `c4 5e 12`、060A `ff 46 12` | 每像素讀來源，再將來源指標加1 |
| 0612 `8a 8f 8c 3c` | 用來源byte查DS:3C8C，沒有64-entry邊界判定 |
| 0616／0618、062B | AND單bit mask，累積八bit，寫一個bitmap byte |
| 063C `83 c0 08`、0644 `f7 f1`、0653 `83 46 12 80` | 每列餘數加8模10；餘數零或末列回退128 |
| 2376:0B85／0B88／0B8F／0B90／0B97 | DS:SI讀128×16 bytes，AND前景至VRAM，列stride80 |

輸出列j的來源列是floor((j+1)×4/5)。最後j127讀第102列，超過128×102地圖。
本次實際配置為地圖13056 bytes、0208標頭、bitmap2048 bytes。最後一列讀標頭兩byte與bitmap前126 bytes。
該bitmap不是局面固定常數；CMP不同，前列不同，尾列亦不同。
DS:3C8C前64項除0／1／5為00外均FF；其後是八個bitmask與UI字串，尾列取樣可能讀到它們。
完整256-byte投影SHA-256 7d8ca2d554b769865a290bd0428955e46d63c1a58613ca18295c9ee10cb3b2ea。
不追查與此玩家差異無關的通用allocator；其他配置次序的鄰接結果unknown。

## 從原始CMP重生

CMP只在空地0的格線位置替換2／3／4，橫線每12列、直線每15欄。
從MAP08／MAP26解碼後套此規則，與兩側原版完整13056-byte地圖相同；其餘值未改。
用上述算法與明示鄰接資料生成bitmap，S1全部43邊界及兩份舊S3共45份完整2048 bytes相同。
S1 SHA-256 96cdcf5dccdfb7e1293a4f2f0f6675505364f14cc20b6bb7c334fc040e2bd6e6，
最後列CE CE CF 56 CF 56 CF 56 CF 56 CF 56 CF 56 CF 56。
S3 SHA-256 fe5a5b2b2c2c06360272780fea381db9f7ea69d022a5a5248f2a5407992dc664，
最後列CE CE C5 CE C5 CE C5 CE E5 CE E5 CE E5 CE C5 CE。

## 勘誤與停止線

研究103／規格046先前把格表多列寫入當成尾列來源，只看到較早寫入者，未追最後一次bitmap覆蓋。
本輪完整呼叫參數、來源指標、最後寫入trace及兩份地圖的完整重生足以推翻該解釋。
第37輪「缺-vram-sites」亦需勘誤：CLI的-vram-at本身就啟用統計，空結果源自父狀態過晚。
機器Write16經兩次Write8，沒有漏掉word寫入的證據。錯誤定位與解碼均屬研究腳本問題。

證據足夠後依[086](../spec/086-battle-minimap-bitmap.md)先DRAFT、審查READY，再實作。
對拍結果追加於原[驗證045](../playtest/045-battle-movement-render.md)，不另建競爭狀態表。
本輪不改seed、golden、遮罩或允許差異，不宣稱完整八步、任意heap、AI、音樂或發行完成。

## 正式驗收

新增全畫面斷言先讓舊程式43幀各失敗38點；修正後43幀完整307200像素及有效色盤相同。
既有S3全35幀、兩份小地圖與戰場、三份正常玩家存讀及18標準產品套件通過。
正常主選單到S1移動GUI已檢視，與同狀態診斷入口分列。
086限定CONFORMED，046來源解釋SUPERSEDED；完整來源及前後收據雜湊見驗證045第38輪。
