# 418：第8步提督與副官升級

2026-10-04，第84輪，遠端#79。父提交4065d46f38b70fabc00783b91ed430d3902cbeba。
本輪完成原版診斷與成長模型補證；108仍DRAFT，正式路徑未改。

## 輸入與工具

| 項目 | 固定身分 |
|---|---|
| 原版GIN3.COM | SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918` |
| dosgolem專用副本 | f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，工作樹乾淨 |
| 完整父保存 | ownership79-r82-entry-a/entry.state，SHA-256 `598e10e5f9457842275fa5639219636f322d8bb2e0f8393fe25b93bfbdef0cfe` |
| 父RAM | 1048576 bytes，SHA-256 `b905eda112fd3a78292bc0d1cdc10068703e0291968640cbaa0d150404404900` |
| 原版雙跑 | 每份13489258 bytes，SHA-256 `3d11f8b68703a5eefe18270821d4ac19dd94bf92a4274c3e86632dd412290d80` |
| IDA | Pro 9.4，locked-v1，UID/GID 1000:1000，一次性COM資料庫 |
| 原版來源索引 | 1462618 bytes，SHA-256 `ab96e1d8d0213da5dfc8b34acc1f46bf1b83440eb3d552ed7f37efd8c2b0c661`，53來源 |
| 分級指令索引 | 542746 bytes，SHA-256 `3d52c61e0e93433297f6240078614c11da38cf27e24347a4cfe926cfc4f451b6`，6來源 |

父保存與RAM在`workplace/oracle/ownership79-r82-entry-a/`。雙跑在
`workplace/promotion79-r84-native-v2-{a,b}.json`，均為本機oracle。
完整來源與注入保存在[原版索引](../internals/data/battle-promotion.json)，
分級指令與兩陣營軍階文字保存在[分級索引](../internals/data/battle-promotion-review.json)。

位址分三欄：原版CS:IP、IDA COM位址、dosgolem執行期線性位址。
COM前256 bytes為零，各自取父RAM的185C、28E0、297B段後4096 bytes。
185C COM SHA-256為`1b6b4aae3a0b652f48e8d3c24f863c39fde0b4c386ddf9afe7ab8c425d206bd7`；
28E0為`2987d4970b51dda7262ed9834e5a25fa57a8bc490ec07cc42dfcd0a32866345d`；
297B為`45adb5a8ae308d85bc0b06c88ea1c3fca6d4a5dd07ed82637c7f0d1a367fb309`。
兩份分級匯出共836條原指令；21項直接語意已證實，其餘保持unknown與醒目警示。
原始函式名、運算元與bytes均保留。095C初次匯出只到提前返回，尾段0979另取；
最終分級版用095C:09BD包含完整共同返回。

## 已證實規則

1. `185C:060D`逐參戰清單處理提督，再處理非FF副官。每人成功後重試，直到回0。
   每次成功先呼叫297B:0546重算，再通知；通知返回後才能再升下一級。
2. `28E0:05F9`讀ADMIRAL的功績u32與等級byte。僅等級恰為99直接拒絕。
   門檻取`DS:4DBE[min(level,39)]`；功績不足回0，否則扣門檻、加等級，回旗標bit0。
   等級100與255只為診斷注入，不能當正常生涯支援證據。
3. 功績高word非0仍會升級。`28E0:0669`接unsigned分支；70000功績案連升三級。
   研究339所寫「高位word必須是0」不成立，保留原文並追加本輪勘誤。
4. 軍階下限依序為0、2、4、6、8、12、16、20、24、29、34、39。
   原版有12個軍階索引；帝國索引10印一級上將，同盟仍印上將，索引11均印元帥。
   這推翻現有`rules.RankOf`從28起為上將且只有10級的模型。正式修正須待108契約閉合。
5. `28E0:095C`在能力>=200時回0且不取RND。其餘取兩次RND，各以跨度+1取餘數，
   相加後整除100。能力<100用第一跨度，其餘用第二跨度。
6. 成長先試統率力300／50，再以RND選占領或防禦200／100，
   再選光線砲、飛彈或戰鬥機100／50。若已有三項非零成長便停止。
   否則選移動或回轉200／50；仍不足三項再選營運100／50或運用200／50。
   byte加法與旗標位元按原版保留。
7. 提督報告固定傳2，即使26→27未換軍階仍印「的官階晉昇為中將」。
   副官傳實際旗標OR1，bit1優先於bit0，後續列出成長能力。
   11則報告已捕捉原文、頭像參數、等待輸入、完整640×480色號、16色與PNG。
   原版程式是否寫錯仍是hypothesis；玩家實際通知內容已證實。

成長模型使用原版實際407次RND輸出，獨立驗算109次嘗試、47次成功，
每案完整2296-byte ADMIRAL均與原版末端相同。
本輪不驗remake RNG、不以相同seed數字宣稱兩側亂數等價。

## 尚缺的玩家鏈

297B:0546不只改提督表。60案完整FLEET共有290個byte變動，SHIP共有200個byte變動。
SHIP涉及`+1F/+21/+22`速度；FLEET涉及可帶組數、能力、費用等重算。
目前只保留每案完整表前後雜湊與全部差分，尚無獨立艦隊重算模型。
297B:05B2、0E8B及相應typed回寫仍須補證，不能猜補或跳過速度更新。

108保持DRAFT。正常玩家取得經驗、連續回合、remake報告矩形／GUI與保存均尚未驗。
062的8-5與完整回合同態也未完成。

## 重生

在專案Docker包裝器內執行：

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_prepare.py
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_native_prepare.py
bash tools/battle_promotion_oracle.sh a
bash tools/battle_promotion_oracle.sh b
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_receipt.py
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_review_receipt.py
```

腳本均拒絕覆寫既有收據。重生須另開新批次名稱，保留本輪固定來源。
IDA的完整命令與一次性DB輸入保存在WORKLOG；
[分級匯出器](../../tools/scripts/battle_promotion_review_ida.py)自動合併
[21項定位索引](../internals/data/battle-promotion-annotations.json)，並核對bytes。
定位索引先複製到`workplace/ida/promotion79-r84-annotations.json`供唯讀語意輸入。

初次工具呼叫錯誤與v1生成器入口置換錯誤均屬驗證環境／腳本問題。
成長模型首次假定SHIP不變，被原版差分推翻；已改為保存未知重算差分。
失敗未改產品程式、既有收據或原始遊戲資料。
