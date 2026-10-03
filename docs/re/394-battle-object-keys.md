# 394：戰場顯示物件指標與排序鍵

日期2026-10-03，沿#89與[第56輪](../goal/2026-10-03-battle-object-keys-round-56.md)。
父提交c9d4ce3fd8b5fbbb37743684bcbf13449f7d7dfd，完整088維持DRAFT。

原版GIN3.COM SHA-256為770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c／clean，未更動執行器。
固定S3兩跑入口RAM SHA-256均為ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97，
來源[第55輪收據](../internals/data/battle-work-control.json)。

已觀測：DS:EADE為64，DS:EAE0指向55筆分隊、7筆方向物件與兩個反白物件。
清單按物件+3升序、+4 word降序；初始目前位置反白的鍵56319，艦隊12方向的鍵64。
兩者順位均FF，方向物件在反白之後；remake目前先畫方向、再畫反白。
原版(255,255)為12，remake為5。以上為原始bytes／清單的proven觀測；
鍵的初始化來源、是否可概括成固定類別順序仍unknown，待完成讀寫鏈。

原CS:IP、IDA位址與dosgolem線性位址分欄。新匯出只加分級語意，不覆蓋原名。

## 已證實資料鏈

[結構化證據](../internals/data/battle-object-keys.json) SHA-256
`a84cbb1963b31b9edfe497e9e0aef730768c9bb5d31bcb3a93dc432e4b5085ad`。四段IDA範圍592條原始指令全部符合固定入口RAM，33項附加語意proven。
新35D1切段COM SHA-256為3edf48c231da25109b8f20bc208c33e517a005f928c98ba973593ea7ab15e68b；
IDA Pro 9.4基準10200，原CS:IP及線性位址另列。其餘未審查語意保留unknown。

| 原CS:IP | 已證實內容 |
|---|---|
| 25DB:0109／010D | FLEET+6C／6E保存方向物件的offset／segment |
| 25DB:0198／019C | SHIP+38／3A保存分隊物件的offset／segment |
| 25DB:01D0／01DC／01E7 | 一般分隊明確以SHIP+03高byte、+01低byte寫物件+4鍵 |
| 35D1:0029／002D／0031 | 遠指標加入DS:EAE0，DS:EADE增加 |
| 35D1:0053／028D..02A2 | 註冊後依同一個02AB比較器排序清單 |
| 35D1:02B4..0322 | unsigned byte+3升序，再unsigned word+4降序，相等返回0 |

全部7筆FLEET指標與55筆SHIP指標均指向清單中唯一物件。
分隊的圖塊、順位、分隊鍵與視角相對座標均符合原池。
方向物件圖塊與FLEET+6A快取相同；正常S3只有艦隊12方向可見，其餘六筆為空白圖塊9。
分隊+4有明確建立來源；方向建立器00FB、位置更新0670與反白建立器0BD2沒有設定+4鍵。
註冊鏈也沒有將這兩類字鍵正規化。

正常雙跑監看56筆變更寫入。兩個鍵最後一次正常變更早於物件建立，來自2376繪製服務；
方向低byte最後在389104136步由00恢復40，反白兩byte在388973956步恢復FF DB。
物件建立至首戰仍保留0040與DBFF；對這些值的最早完整來源不再深入追查。
本結論限定已驗正常局面，不把相等鍵返回0外推成通用排序器穩定。

## 可重播反例

[原版工具](../../tools/battle_object_keys_oracle.sh)沿固定S3父state、tick20000與同一正常輸入。
base-a／b完整RAM、色號、PNG及DAC相同。明示key-high-joined-a／b只在407194660步
將線性71022的word+4改FFFF，也各自完整雙跑相同。
兩筆注入必須在同一-poke參數內以分號分隔，report分別證實seed及鍵的實際寫入位址。
兩側最終RNG仍B6E4B78A，WORK／FLEET／SHIP／STELLAR／PLANET完整表皆保持。

反例RAM SHA-256為0281f8c208ab65e9fb6b15346c8787723bec994c6fac86e81b260b526f3d16e2。
FFFF使方向物件在DBFF反白前畫，完整640×480只差(255,255)一像素，12改5。
這證明覆蓋由原版殘留字鍵決定，不是原池座標、游標或素材解碼錯誤。
注入反例與未注入正常收據分列，不升格正常玩家全局面一致。

首批key-high探索重複傳入-poke，CLI只保留最後值，seed寫入未出現在report。
後續查CLI及parsePokes契約，改單一分號腳本並以新目錄重跑；舊探索保留但不收入正式雙跑。
其結果碰巧相同不能證明遺失的注入存在。第54輪多筆-poke探索也有此限制；
它本來就未列通過，相關勘誤追加至研究392。執行器未更動，不令舊正常收據失效。

## 可丟棄呈現探針與裁決前沿

測試期覆蓋workplace/round56-input-before-directions.go只把Input反白畫到方向圖示之前，
正式internal/game/game.go沒有修改。正常主選單→138→S3帝國→實行→首戰及四次純移動，
五張完整307200像素相較原版均零差，只有entry一像素改變，其他四張保持。
這只證明目前樣本。其他方向與反白的殘留字鍵可能造成不同覆蓋，固定類別順序不能冒稱全部exact。

忠實度決策已提出：採明確固定次序並保留刻意差異，或重現原版殘留記憶體排序。
尚未收到裁決，正式呈現分支不進READY／production；完整088仍DRAFT。
原始全表仍WORK0／FLEET27／SHIP214／STELLAR0／PLANET0，沒有排除或填造far指標。
顯示指標比較契約另待核定，正常索敵同態配對與完整八步未完成。

重播需父state與本輪來源，各輸出拒絕覆寫；所有完整旗標、輸入／輸出／prototype雜湊在JSON：

```bash
bash tools/battle_object_keys_oracle.sh normal89-r56-base-a
bash tools/battle_object_keys_oracle.sh normal89-r56-base-b
bash tools/battle_object_keys_oracle.sh normal89-r56-key-high-joined-a
bash tools/battle_object_keys_oracle.sh normal89-r56-key-high-joined-b
bash tools/gui.sh xrun python3 tools/scripts/battle_object_keys_receipt.py
```

兩個Python與shell語法驗證通過，新收據拒絕覆寫且SHA保持。
正式產品無變動；本輪重跑標準18產品套件，13個測試套件通過、5個無測試，零失敗。
workplace/round56-standard.log SHA-256為ff6417d27c7cd60aa05cdc8d1d6d30fa92ad9dcf7743ce3eb4b9f747cc4565ff。
第55輪凍結來源及160限定回歸收據保持；不由產品全測通過宣稱原版對拍完成。
