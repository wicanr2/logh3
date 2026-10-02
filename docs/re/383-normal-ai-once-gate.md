# 383 正常S3首戰的AI一次性跳過

工作[#89](https://github.com/wicanr2/logh3/issues/89)，父提交c05be3dd1afa8bfb745a6cda94ce02e46b6a178a。
方法：[第45輪](../goal/2026-10-03-normal-ai-gate-round-45.md)；規格：[088](../spec/088-normal-battle-player-entry.md)。
[完整收據](../internals/data/normal-ai-once-gate.json)保存全部argv、輸出、工具與來源SHA-256。

## 最早分歧

已證實：正常S3原版WORK+32為1，首次AI清0後返回。remake原先未消費旗標而執行AI，造成首戰名單7對3。
相同原版狀態只將旗標改0，也得到3支；remake可丟棄覆蓋只增加一次性消費，恢復7支。
029與研究175已有閘門證據，本次補正常觸發與反事實，再審查088 READY子契約，最後修改正式產品。
本次不猜測其他設旗來源；完整入口仍不同。

## 固定來源與雙跑

- GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`。
  bundle SHA-256 `de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673`。
- dosgolem f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c、工作樹乾淨；原始遊戲唯讀。
  版本、Go／GUI／IDA image身份完整保存於JSON。
- 冷啟動S3來源fcmd-s3-base.state SHA-256
  `0155f63e373582b76a3b7ee54eccbef5ae2626481e09a68550b5fddc580146d7`。
  冷啟動命令在workplace/oracle/fcmd-s3-base/meta.txt，其SHA在JSON。
  由tools/fleetcmd_receipts.sh保存299999999檢查點，與第44輪相同版本及玩家輸入。
- 此次載入後FLEET3904／SHIP27456／STELLAR4032與第44輪冷啟動RAM及正式Game三張全表相同。
  300000000事前固定D3DC167E；自然案不注入旗標，受控案另寫lin:5DA76=00。
  300000001兩案完整1MiB RAM只有5DA76一byte不同，沒有重擲。
- normal89-r45-natural-a／b與zero-a／b由tools/normal_ai_gate_oracle.sh重生。
  每分支四份完整RAM、色號、PNG及DAC七份輸出雙跑相同。全部注入與時間戳見JSON。
  本次是正常檢查點續跑，不替換第44輪冷啟動完整玩家入口畫面。

| 比較點 | 自然旗標1 | 受控旗標0 |
|---|---|---|
| AI進入步數 | 306103454 | 306103454 |
| 2D10:000C／0017／0021 | 各0次 | 各1次 |
| 清除觀測 | 306103457，1→0 | 初始注入0，完成後仍0 |
| 返回 | 306103458 | 324822153 |
| 首戰名單 | 12／16／17／28／29／30／31 | 12／28／31 |

自然分支清除前後FLEET／SHIP／STELLAR全表及seed D3DC167E不變。
跳過包含建隊、整補及命令，不能只略過Orders。
watch的2D10:0035是寫入後IP，實際寫入指令是0030，由IDA原始bytes確認。

## IDA定位

從自然案loaded-mem.bin切dosgolem線性2D100至2D75F，前補100h零bytes供COM載入。
輸入SHA-256 `e88b68361851a9fc8214b88f56ec352c44b6b83cc267c2c6195dbe2d87c8da1d`。
tools/ida.sh沿locked-v1 image，IDA9.4、SDK940、UID1000；只建一次性DB。

| 原版CS:IP | dosgolem線性 | IDA COM位址 | bytes | 已證實語意 |
|---|---|---|---|---|
| 2D10:0005 | 2D105 | 10205 | 803e468f00 | 比較DS:8F46與0 |
| 2D10:000A | 2D10A | 1020A | 7524 | 非零跳至0030 |
| 2D10:0030 | 2D130 | 10230 | c606468f00 | 清DS:8F46 |

DS:8F46此次線性5DA76，等於WORK+32。可重用識別保存結構位移及簽章，不固定執行期段值。
0004是前支retf、0037是本支retf、0038是下一支push bp。原始函式名sub_10205保留。
其他指令仍unknown；切段無外部xref不代表沒有caller。
匯出JSON SHA-256 `85c7231d753489164b055886a2dd96eda07e07efab97351d1b2b8e43dce5186f`。
第一次誤用SDK／版本API，沒有輸出，不算成功；修查詢後以相同bytes建新DB，schema／版本／UID／SHA通過。

## 正式實作及限制

runBrain既有單人／資源閘門之後消費WORK+32，非零清0返回；其他bytes保留。
mutableWork及正式SaveSlot／載入共用狀態。原版與增強AI下一次仍執行，不永久停用。
正常bootstrap.Load／Game.New、主選單、138、S3與實行；測試期Go覆蓋只觀測並固定seed。
沒有原版RAM回注、EnterBattle或局面注入。17份輸出雙跑相同，省略實行前診斷存檔時七份表／畫面也相同。

| 首戰資料 | 修正前 | 修正後 | 原版 |
|---|---|---|---|
| 參戰艦隊 | 3支 | 7支，次序相同 | 7支 |
| 戰場池，含未放出索敵隊 | 27筆 | 55筆 | 55筆 |
| 結尾seed | ECDD600B | B6E4B78A | B6E4B78A |
| WORK回合 | 381 | 381 | 380 |

修正後48個活動單位，不混作55筆池資料。全局FLEET差120／SHIP差865／STELLAR差44／WORK差36 bytes。
完整前後hex及偏移數字保留；執行期指標及其他未分類bytes不宣稱一致，也不直接當作玩法缺陷。
載入後控制區+0C／+0E／+10、日期及完整畫面待追，088保持DRAFT、#89 OPEN。
完整640×480仍差5863像素，Field差73，有效RGB同口徑。完整並排、原圖及差分已目視，沒有裁切、遮罩或縮放。
nearby-state不升格；087的151個受控UI幀保持原成果，正常索敵配對與完整八步未完成。

限定38個PASS事件、零skip；兩種AI、非零邊界1／2／255、消費後存讀及第二次AI、缺資源不消費，
正常增強AI雙陣營均通過。S1／S2／S3正常戰術實行及完整SHIP／FLEET存讀通過。
標準18套件初跑因S1／S3舊remake抽樣數量失敗，更新實測數量後同命令乾淨重跑通過。
S1／S2只聲明remake抽樣，不新增原版同態聲明。log與SHA見JSON及046。
