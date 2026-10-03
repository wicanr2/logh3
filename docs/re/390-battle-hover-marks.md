# 390：正常Input摘要的小地圖集結點

沿[#89](https://github.com/wicanr2/logh3/issues/89)，[第52輪](../goal/2026-10-03-battle-hover-marks-round-52.md)。
GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，clean且未改。

## 原版重播

使用第51輪各自獨立的空白Input snapshot，先前完整RAM與畫面已雙跑相同。
[重播工具](../../tools/battle_hover_marks_oracle.sh)保存父snapshot雜湊、完整旗標與新輸出。
四案各a／b獨立重跑，完整1MiB RAM、640×480色號、PNG及DAC逐位元組相同。

| 案例 | 純移動的原版滑鼠座標 | 457M明示注入 | end.bin SHA-256 |
|---|---|---|---|
| flagship | 248,208 | 無 | e87697e47c1558d21d81118b07b1b623a633ed6112a08fc6dd611b0833d82b65 |
| squad | 249,184 | 無 | 9cb2bcfe0208a1fcbaaa6dece017cd16a23340b66db7c715b4a46c93936fa8a4 |
| fleet-target | 248,208 | lin:611C0寫40 02 90 01 | efa72e6c10c6ac92b6899f5d203576904aeda6e2eaf5f86ef06e7f62e5d016d9 |
| squad-target | 249,184 | lin:681D6寫40 02 90 01 | f602b9ee4f944364cab7819b4a0296fd2a8165fe827f345eaf0c230bf76c1d42 |

注入只改FLEET 12的+64／66或SHIP 82的+28／2A，集結點576,400。
report確認實際線性位址等於要求值。460M移動，474999999傾印，475000000保存。
兩個非零集結點案例是窄幅診斷，不當正常玩家路徑。原版檔案維持唯讀。

## IDA資料流

使用IDA Pro 9.4 locked-v1、UID1000，從第51輪正常入口RAM切1727段，前置256零bytes。
COM SHA-256 `a896c5c048c7c762f8648ff868097f6a974ca60605682ae3bf4abd51ac638681`。
IDA實測基底10200，與整段bytes相同；原版CS:IP與dosgolem線性地址另外保存。
原始RAM SHA-256 `ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97`。
[匯出工具](../../tools/scripts/battle_hover_marks_ida.py)不更名，未審指令保持unknown。

| 原版定位及bytes | 已證實語意 |
|---|---|
| 25DB:0BB0 `9a77062717`／0BBF `9a36072717` | 旗艦呼叫0677，其餘分隊呼叫0736 |
| 1727:067F `268b475c`／068E `268b575e` | 旗艦讀FLEET+5C／5E現在座標 |
| 069B `268b7764`／06A9 `268b5f66` | 旗艦讀FLEET+64／66集結點 |
| 0686／0692／069F／06AD | unsigned右移4位，X加480、邏輯Y加36，實體Y再加40 |
| 06D1 `b80400`／06E5 `9a3609d739` | 色號4、遮罩FFFF，呼叫39D7:0936畫XOR實線 |
| 071E `9a3609d739` | 退出以相同參數再XOR擦除航線 |
| 073C `268b4728`／074B `268b572a` | 一般分隊只讀SHIP+28／2A集結點 |
| 0758 `b90400`／075E `9af408d739` | 色號4，呼叫39D7:08F4點集結點 |
| 0780 `9af408d739` | 退出以相同參數再XOR擦除集結點 |

動態旗艦與分隊各只命中對應owner及畫線／畫點一次。現在等於集結點時航線也畫一點。
非零旗艦航線共11點，完整小地圖相對空白原圖每個改變值均XOR 4。
既有[畫點／畫線契約](../internals/52-plot-primitives-and-canvas.md)與ClipRect.XorLine沿用，
不重新研究EGA硬體。退出後清快取FFFF，正式重建Frame只在摘要快照有效時疊一次。

## 戰場剩餘像素

第51輪初始差(255,255)原版12、remake5；它超出13像素寬游標，不是游標點陣責任。
原版SPRITE10物件圖形128 bytes與原素材相同，該角落是5。
完整顯示清單中現在反白物件+3=FF、word+4=56319，艦隊12方向物件+3=FF、word+4=64。
原版比較器按+4由大到小，方向在反白後畫，角落由方向圖塊覆蓋為12。
這個實例的覆蓋來源proven；25DB:0BD2未初始化word+4，任意局面的排序契約仍unknown。
不把單點改成常數12，也不因一份清單就重排所有優先FF物件。本輪先完成小地圖owner。

## 工具問題

第一次探針把7FC當掃描終點，恰落指令中間，重跑日誌確認assert失敗。
回查規格路由後分開掃描上限與實際指令終點，另建新資料庫探索，再固定四段返回邊界。
正式範圍0677..06F0、06F0..0736、0736..0767、0767..0792，各前一byte及最後一byte均CB。
失敗輸出保留，不作函式證據。

正式IDA完整命令：

```sh
timeout 3m bash tools/ida.sh raw idat -A -L/work/normal89-r52-1727-graded.log '-S/idascripts/battle_hover_marks_ida.py /work/normal89-r52-1727-graded.json 1727 677 792 graded' /work/normal89-r52-1727-graded.com
```

## 正式限定驗收

先DRAFT，再審四個邊界、105條原指令與12項語意及八組獨立雙跑，088子契約轉READY後實作。
正式摘要快照取活BattleFleet現在／集結點；renderer沿唯一MinimapSpec與既有XorLine。
負數或安全矩形外座標略過，不推論原版非法輸入行為。

五張正常小地圖各0／16384，初始全圖2減1，旗艦及分隊1減0，空白及右面板仍全圖0。
完整原圖／remake／並排／差分已目視，無遮罩；初始剩餘方向覆蓋像素未改。
55份正常入口輸出雙跑相同，WORK／FLEET／SHIP純移動保存相同，RNG不變。
45限定PASS零skip、18產品套件、真實cmd/gin3正常首戰一倍／兩倍通過。
四案原版RAM疊圖回歸僅驗primitive，正常Game五案另列；兩種證據不互換。

[完整收據](../internals/data/battle-hover-marks.json) SHA-256
`91ec1901d3232702c3612d4edd459d72ff60c4633483bc578339b0e3f5c56250`。
原第51輪JSON仍SHA c1f387c1e3729450f4b7b3f9f8babb21b0b8f49d1b0762b1f6e0cca648cfb5c6。
WORK14／FLEET27／SHIP276差保持，完整088 DRAFT、#89／#79 OPEN，nearby-state不升格。
