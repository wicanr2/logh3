# 409：艦隊投影與下方場景接線

2026-10-04，沿#48／#79。父提交65937900af4559fbf1fcb871fd4694c536dcf8f1。
問題是095使用原版已投影參數，尚不能從目前編制來源產生完整下方場景。
本輪先核原指令與30案，再寫097 READY及正式核心；完整092仍DRAFT。

## 固定輸入與工具

| 項目 | 固定身份 |
|---|---|
| 原版bundle | `org_game/银河英雄传说3SP.zip`，SHA-256 de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673 |
| GIN3.COM | SHA-256 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918 |
| 提交父state | `combat79-r63-laser-submit-a/selected.state`，af0019ce8a29ea2e0026a0ba15ac4f37d870f848000247a07250ee10190de19d |
| dosgolem | 專用logh3-oracle分支，f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，乾淨 |
| Go | golang:1.24-bookworm，實際go1.24.13；IRQ0Every20000／Pinned，CycleClock關閉 |
| IDA | ida-pro-9.4-idapython:locked-v1，實際9.4，UID1000 |
| remake | golang:1.24-bookworm及既有logh3-ebiten:2；實作與測試時工作樹含本輪修改 |

從父state沿原版實際呼叫停在321F:000A，沒有換種子或補寫編制。
入口步數499017677；SS:SP及兩個FLEET far參數保存在report。
入口RAM SHA-256 9dfbb2cbec5d93849e0f100511cad65b342b41f68198378cd9cdc103ebeb9279。
受控案例只在此入口寫兩側目前XY，每個linear位址與前後值均保存。
seed兩word保持；30案皆包含14分隊，沒有靠重擲取得局面。

## 原版契約

proven限[604條原指令與九項分級語意](../internals/data/fleet-projection.json)及固定案例。
原版CS:IP、IDA EA與dosgolem線性位址分欄，未覆寫原始名稱或運算元。

- 000A先依第一艦隊編制生成列表，再附第二艦隊；01CA不過濾零艘、索敵或艦種。
- 0223用兩側編制首筆的現在XY。負的反正切低byte是旋轉角。
- 中點用word加法後無號除二。距離不大於200，先截去中點距離小數再乘3除2，
  左右分別從320減、加；大於200則固定170／470，Y均298。
- 0513相對座標乘3除2、旋轉後保存Y到SHIP+44；半Y由X扣除，再加側別基底。
- 可見旗標+42為signed X8..568、Y220..376，含邊界。
- 0643比較unsigned投影Y，然後unsigned投影X。它不以+44排序。
  3C6A:000D沿既有梳狀排序契約，同鍵不能改用另一種穩定排序。

零距離的投影角為192，沿既有AtanTable256(0,0)；這與093零距離傷害裁決分開。
範圍0..2047內的旋轉乘積由3A92:0062／0081加512再捨去10位，既有Rotate256符合。
角度的高byte不消費；未外推任意word溢位。

## 獨立重生與實作

執行`bash tools/fleet_projection_oracle.sh a`及`b`，拒絕已有輸出目錄。
兩側61份RAM逐byte相同，report只移除命令末端a／b後相同。
30案涵蓋0／1／20／199／200／201、八方向、奇數斜向及可見邊界。
`fleet_projection_receipt.py`逐條核IDA bytes與本輪入口RAM，再凍結原版期望值。
131來源、604原指令，SHA-256 cd191199bf5297fd25d41891dc89fe4108709c6426bde78fa44ce95aa8f471f0。

正式`ProjectFleetScene`保留來源順序與完整繪製順序，不改交戰鍵、來源或RND。
30×14投影word、可見性、深度、鏡頭及排序直接比原版全部符合。

`BuildFleetSceneOps`從typed來源生成06FC完整下方操作，使用095素材解析與合成器。
圖塊索引為陣營乘二，再加未知一；標籤、亮標、連線與八條朝向箭頭沿原caller。
2376:083B的jcxz使零塊不畫；0835的素材移位setup沒有VGA I/O，不改遮罩。
095未閉合裁切回錯誤且不提供部分計畫，正式可達性仍unknown。

兩個舊原版06FC入口及結果凍結到[fleet-scene-plan.json](../internals/data/fleet-scene-plan.json)，
SHA-256 4988a5b95d2167bc14166a8fb173fb26bc0b6d4b9f6e7f070c86c9bcc3b126f4。
兩張下方完整操作及全圖均零差；上方明示保留原版入口，不冒稱整場Game。
可見產物與測試見[對拍061](../playtest/061-fleet-projection.md)。

完整等待起點、正式Game、正常GUI及保存尚未驗收。061電腦被動政策保持。
本輪沒有改dosgolem，也沒有重生或改寫69／70／71的凍結收據。

## 同輪續證：結果等待

由第70輪`scene-01.state`沿05D7實際場景到3765:006C，父SHA-256
07cebb1872b6b4c2af91158ee932299a34aed59679afab2dfe43784318f7d8ef。
同一dosgolem／Go／計時旗標，原版caller stack傳90。首跑等待上限800萬步太低，
既有第69輪原版已量得17969972步；失敗停在36AD:0270，入口輸出與abort分類保留。
依既有範圍調到2000萬步後，11案正常返回，不把診斷限制當成產品缺陷。

proven限[等待收據](../internals/data/fleet-wait.json)的108原指令與11案雙跑：

| 案例 | 等待步數／原版事件 |
|---|---|
| 無輸入90倒數 | 17969971步；timer到0、event0 |
| duration0／1直接診斷 | 105／170008步；timer到0 |
| 演出中左／右按下並放開 | 101／103步；等待讀到排隊事件1／2，timer仍90 |
| 等待開始後200000步左／右按下 | 200232／200234步；讀到1／2，timer89 |
| 空白／ESC於演出中 | 兩案17969971步；event0，不縮短 |
| 空白／ESC於等待中 | 17969945／17969947步；event0，不縮短 |

全部返回前清timer、AX回1。11案44份RAM／色號兩跑逐byte相同，正規化report相同。
按下／放開的步數及原版滑鼠XY完整記錄；等待中按下造成提早返回，預定較晚的放開未送出。
3765:00BF先取12-byte事件佇列一筆，空佇列才用當前滑鼠狀態並清event word。
006C只在timer0或event1／2時結束。沒有研究ISR、driver或把步數換算成實機時間。

95來源、108原指令，凍結SHA-256 8ff102acff0607b23524e0b8678fc3eef3817aa171a6ce779d5d58c56392c566。
[098 READY](../spec/098-engagement-wait.md)後以正式Game的共用輸入與等待predicate核11案，全部通過。
既有左右排隊／空白與ESC忽略／倒數滿返回符合，不需修改產品輸入。
此診斷不把原版步數映射成Game整段動畫tick；返回戰場後的放開事件、等待起點及完整場景仍待驗。
