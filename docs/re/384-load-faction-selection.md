# 384 讀檔後陣營選擇與座位控制

2026-10-03，第46輪，工作[#89](https://github.com/wicanr2/logh3/issues/89)。
父提交`04ae7d659321b0b568287852b7afc1dba30121f1`，正式修改的檔案SHA保存在
[完整收據](../internals/data/load-seat-selection.json)。本研究不取代088完整入口驗收。

## 輸入與工具

| 輸入／工具 | 版本或SHA-256 |
|---|---|
| 原始bundle `银河英雄传说3SP.zip` | de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673 |
| GIN3.COM | 770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918 |
| dosgolem-logh3，乾淨 | f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c |
| IDA Pro／IDAPython | 9.4／940；UID1000 |
| 一次性042F COM輸入 | dc7d9ecd9fbe218720ecc85e29f46b858be9935b0aaa2b7d3c2487b8c879180a |
| IDA匯出 normal89-r46-seat.json | e3fabe15a275f0457afe923a820d78ca7db4966cf627262ddc8c9b137a51db2f |

沿用Go、GUI及IDA三個既有image，完整tag／image ID見JSON。原版唯讀，未改dosgolem。
原版冷啟動到S3陣營選單各保存229999999步的完整RAM及snapshot。
[包裝器](../../tools/load_seat_oracle.sh)限定固定HEAD及乾淨工作樹，拒絕覆寫批次。

```sh
bash tools/load_seat_oracle.sh normal89-r46-menu-a
bash tools/load_seat_oracle.sh normal89-r46-menu-b
```

再逐組執行`empire／alliance／versus／hover0／hover1／hover2／vsconfirm／vshandoff`的a及b。
每個後綴載入自身的冷啟動snapshot，絕不改RAM或重擲種子。
argv、tick20000、輸入時間、完整1MiB RAM、原生640×480畫面、有效色盤與SHA均保存於JSON。
Snapshot各自記雜湊；可重現判準是完整RAM、screen.bin、PNG、DAC，不靠snapshot序列化檔名。

## 已證實的選單

點S3存檔列後，WORK回合仍0，尚未載入S3；選擇陣營後才變成380。
右鍵三列各雙跑仍回合0，只有游標／反白改變。這不是取消選單的證據。

042F:0006之前的005為前一函式retf；IDA輸入由固定RAM線性042F0..04F7F加100h COM墊片產生。
IDA原始名稱`sub_10206`，邊界10206..10342，末尾042F:0141 retf。
本次IDA基底10200；原始CS:IP與dosgolem線性分欄，不把它們當同一位址空間。
3667:0004建立三列，341B:0145及3667:00F2等待選擇；未附加語意指令保持unknown。

DS:022E是四筆6-byte描述，每筆word旗標1及16:16 far pointer，不是連續字串。
固定DS54B3只屬本批記憶體；保存事實以WORK位移及原始bytes為主。

| 項目 | 原始標題bytes | 是否顯示 |
|---|---|---|
| 帝國 | `AB D2 20 20 B0 EA 00` | 是 |
| 同盟 | `A6 50 20 20 B7 F9 00` | 是 |
| 對戰 | `B9 EF 20 20 BE D4 00` | 是 |
| 取進度 | `A8 FA B6 69 AB D7 00` | 否 |

Big5解碼核對第四列是「取進度」，不能把未顯示的描述猜成取消按鈕。
已證實框(289,181,62,98)，三列內寬2字；共用既有3667選單幾何及FRAME.BLK。
十個單次輔助邊界收據：x302／303／336與y199／260不載入，x304／334／335與y200／259載入。
命中範圍[304,336)×[200,260)，每列20 pixels。此邊界證據明標單次，正式九組收據另雙跑。

## 九個控制bytes及雙人交接

WORK+0C..14為模式、兩座位配置CPU、兩座位目前CPU、兩座位陣營、兩座位首都。
完整三種值已回填[保存事實30](../internals/30-save-format.md)。
IDA非破壞性附加27個寫入語意，原始位址與操作數均保留，動態watch核對三個分支。
例如帝國042F:00A3簽章`a2208f`寫模式、00B6`a2258f`寫座位0陣營；
同盟00C9同模式寫入；對戰00EE`c606208f02`寫模式2。
Watch的CS:IP是寫後位置，例如帝國00A6對應實際00A3，不把下一指令當寫入者。

單人一律玩家座位0，CPU座位1；同盟交換陣營及首都。
雙人載入先帝國交接，確認直接命令畫面；帝國實行後換同盟，確認仍直接命令畫面。
回合保持380，原版FLEET3904及SHIP27456完整表不變，不重新開局或收一次月稅。

## 實作與正常驗證

依[088 READY子契約](../spec/088-normal-battle-player-entry.md)才接正式產品：
追加StateLoadFaction保留舊state整數；已使用列先保留索引，左鍵合法列才載入。
繪製／命中共用CommandMenuSpec，右鍵保留選單，游標最後畫。缺損存檔保留舊局面。
三選擇先定陣營，再用六檔載入，保存九個控制bytes；雙人財政由已載入的兩方資料建立。
FLEET是全域池，交接只換財政與視圖，移除舊陣營艦隊slice快照，造船百分比也保留兩方。
已載入這輪交接跳過新局月報，兩邊實行完才解除標記；不擴張成雙人跨月同態聲明。

正式[正常玩家探針](../../tools/scripts/load_seat_player.go)不注入RAM或局面；
bootstrap.Load／Game.New、取進度／138／S3／三選擇，各四種反白及完整存檔，55份輸出雙跑相同。
另[真實X11](../../tools/load_seat_gui.sh)五案驗右鍵、帝國、同盟、對戰及兩倍視窗，正常鍵盤138與滑鼠輸入通過。
觸控沿同一Click熱區，未新增獨立手勢或語言模式，不宣稱實機觸控驗收。

| 已保存比較 | 全畫面不同pixels | WORK／FLEET／SHIP／STELLAR不同bytes |
|---|---:|---|
| 四種選單，各整個6076 pixels popup零差 | 各16333 | 尚未載入，不把新局世界當S3 |
| 帝國載入 | 0 | 0／0／0／0 |
| 同盟載入 | 0 | 1／0／0／0；WORK+09 |
| 雙人未確認 | 2841 | 0／0／0／22 |
| 雙人帝國確認 | 0 | 0／0／0／0 |
| 雙人同盟確認 | 0 | 2／0／0／0；WORK+09／+0A |

每組全圖、有效RGB、最大通道差、bbox與差分均保存，沒有遮罩／裁切／縮放。
比較預先為nearby-state，零畫面差不事後升格。
選單背景差在載入前的新局初始化世界；雙人未確認有地圖首都反白與捲動箭頭差，
STELLAR的22個差異集中在每筆+7D，語意unknown。其精確時序尚未裁決，不能以popup零差掩蓋。

正常S3首次實行的17份輸出也雙跑相同，事前固定D3DC167E，不重擲。
首戰仍7艦隊[12,16,17,28,29,30,31]、55筆池、結尾seed B6E4B78A；
FLEET120／SHIP865／STELLAR44 bytes差保持，WORK由36減到33。
全圖仍5863 pixels差，日期原版380／remake381未修；完整入口及正常索敵仍待驗。
第45輪一次性AI閘門收據保持原始來源雜湊，不重生或改寫當時證據。

21個限定PASS、零skip，含兩AI正常雙陣營與旗標第二次呼叫、陣營保存、損檔不改局面、全域池交接。
標準18套件初跑的舊正常讀檔測試未選陣營，補步驟後相同命令乾淨重跑通過。
首次新測試把Fleet欄位誤寫Slot，修為既有結構後重跑；原版首次輸出掛載及prototype字串誤判屬腳本失敗，
兩者失敗輸出不當產品缺陷或正式收據。日誌與SHA在JSON。

## 範圍與下一步

088讀檔選單子契約已驗，整份仍DRAFT，#89與#79保持OPEN。
沿研究251已知主迴圈，下一步修正式日期在所有戰場結束後才進位，並同步活動陣營／座位的階段。
不只改Date文字或存檔偏移來掩蓋，先以原版寫入邊界及READY契約核對。
原Goal、其他八步、音樂人耳／macOS／發行仍未完成。
