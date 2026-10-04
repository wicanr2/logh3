# 416：戰術回合的星域歸屬與通知順序

2026-10-04，第82輪，沿[#79](https://github.com/wicanr2/logh3/issues/79)。
remake父提交58291ecad91f4efbfdfce26a3988707200044e51。
本輪只研究062子步8-2，不開8-3–8-5。入口為[106](../spec/106-battle-region-ownership.md)。

## 來源與方法

原版GIN3.COM SHA-256
770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918。
bundle SHA-256 de905aa0b5a666aa9d31dce36f43a5584f1842950c39d7e1b1f97c72e0d79673。
dosgolem專用logh3-oracle工作副本HEAD
f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c，乾淨，未修改執行器。
GUI工具logh3-ebiten:2，image SHA-256
6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8。
IDA Pro 9.4使用ida-pro-9.4-idapython:locked-v1，以唯讀RAM另製COM分析。
自然續接父保存檔dosgolem專用副本workplace/fcmd/eng-t5.state，SHA-256
c6ce5af3a8d52acc415ac1029a29a0905e9c95e02ac0e0ba547e71300fd947c1。

[凍結資料](../internals/data/battle-ownership.json) SHA-256
c4b36a1280a713f818716ce67b69b25a7c5e65ba4abdfe79667e3aa46e081ec1，
保存177來源的路徑、大小、完整SHA-256、22項原指令語意、156案全部注入與完整星域記錄、
104則報告文字、10組自然呼叫收據完整旗標與保存狀態比較。
來源RAM為battle79-move-entry-a/mem-in.bin，1 MiB，SHA-256
9c4a4c28a1f53056962f613b1b643e5a5ba199d9231df291918d9499111be38d。
來源保持唯讀；所有控制案例在副本寫記憶體，不改原版程式碼。

## 函式與位址基準

本節段址為dosgolem載入後的原版CS:IP；DS為54B3，線性位址為DS×16＋位移。
另製COM在檔案前置256 bytes；IDA EA與原CS:IP對照在凍結資料，不當作同一位址空間。
三份COM的完整雜湊如下。

| 原段 | COM SHA-256 |
|---|---|
|185C|1b6b4aae3a0b652f48e8d3c24f863c39fde0b4c386ddf9afe7ab8c425d206bd7|
|1213|185b13e49a43df742326dd25866d46b8f17cd2d127b11677e8651c1f29edd86b|
|3982|e9eee2f384dd3c900d8002d4d2d9661b0ee5ee4dc0b96e10b4cbf44b7bebbc26|

185C:0001的0002／0007／000C／0011／0016依序呼叫001D／02DA／0491／060D／09BF。
02DA前八byte為ec83c4405f5e5dcb，進入點55，避免把前支函式尾端當入口。
185C:02DA與03C6在IDA保留原始sub_104DA／sub_105C6及EA邊界。
已證實語意另附原始定位、bytes及證據，不改函式名稱。其他指令保持unknown。

## 已證實的規則

| 定位 | proven語意 |
|---|---|
|185C:02DA|讀DS8F17星域，STAR是DS:A79C起的直接陣列，每筆126 bytes|
|STAR+16／17／18|先清零，再依當前戰場行星Owner累加帝國／同盟／無主數|
|DS:CFF8／CFFC|戰場行星遠指標及byte筆數，每筆82 bytes，Owner在+11|
|185C:02DA分支|同盟與無主數皆零取0；帝國與無主皆零取1；其餘2。零顆也取0|
|185C:03C6|新舊不同才通知陣營0、1；3982:022D取正式座位映射，只通知人類|
|185C:03BD|兩個通知返回後才寫STAR+12，等待期間三方數已新、擁有者仍舊|
|3826:0405第三參數0|初稿推論無頭像，已被本輪完整繪製推翻，見下方追加勘誤|
|1213:007B|自己拿下的人類分支在顯示報告前登記；STAR+11為14且未發生才設DS988E=1|

四個原始格式位元組在凍結資料report_formats，來自DS3D58／3D71／3D92／3DAF。
部分占領時以舊STAR+12選「部分地區已脫離我方統治」或「我軍已進駐部分行星」。
原版曲線與亂數不參與這支常式。

## 156案常式診斷

11組行星Owner排列×3舊擁有者×4人類遮罩＝132案；
費沙3新擁有者×4遮罩×2事件已發生旗標＝24案。
固定seed12345678。兩次完整JSON逐byte相同，SHA-256
57d7ee092bee49073f053e3ab19dc510d29f8ae95c14e60cb14ad951189280fd。
每案完整126 bytes只有+12及+16–18改變；104報告當下仍保留舊Owner，費沙待辦時點正確。
seed每案保持78563412。此批冷VGA、tick0、訊息速度0，屬控制常式診斷。

重跑入口：在專用dosgolem副本，以DOSGOLEM_ORIG指定原版唯讀目錄、
DOSGOLEM_EXTRA_MOUNT指定本專案:/task，透過tools/go.sh run
/task/tools/scripts/battle_ownership_probe.go -output /task/workplace/<新批次>.json。
工具與全部注入明細由凍結資料sources／cases回查。

## 自然呼叫與等待

[tools/battle_ownership_oracle.sh](../../tools/battle_ownership_oracle.sh)
接受ownership79-r82-<案例>-a或-b。10組為entry、control、capture0／1、partial0／1、
capture0／1-second、capture0／1-finish。每組原版兩次RAM、640×480 framebuffer、色盤及PNG完全相同。
entry以正常執行保存狀態續進8-1，再自然到8-2；換手案例完整列出行星、座位、舊Owner注入。
此路徑證明原版控制局面的自然呼叫，不能取代完整正常玩家流程。

第一張報告等著時03BD及0491均未呼叫；一次點擊後第二張等待時仍未呼叫。
第二次點擊後03BD、0491、060D、09BF及下一回合12AA:0245各命中一次。
三方數已更新，Owner只在最後一步更新。控制組無報告直接續接。
原版此局面為星域26達希利。339標題的巴米利恩是劇本名，不能推成當前星域。

## 保存狀態與工具失敗

原始.state雜湊不同，原因為gob map及DOS handle切片列舉順序。
依固定版本production schema解碼全部46個machine欄位與24個DOS欄位，
只按已證實來源排序Handles／EMSHandles；map用完整鍵值比較。
10組全部相同，每組故意改RAM一byte的負對照皆被攔下。
來源schema雜湊、原檔雜湊及全欄位canonical JSON雜湊都保存在凍結資料。
保存容器internal/state/state.go來源SHA-256
299155394888f9b50de2171e580ed6d6a3415fe4e91facebc504c8ee9ad65d54。
不宣稱.state原檔逐byte相同，不丟棄CPU、時鐘、VGA或DOS欄位。

先前比較器漏掉保存容器的8-byte段長，修正後乾淨重跑。
IDA新COM自動分析將指定範圍當資料；明確清掉新資料庫該範圍再建立指令，原bytes不變。
這些失敗屬工具設定，未改遊戲或執行器。探索v1／v2輸出保留本機，不當作正式收據。

## 規格審查

156案完整記錄、seed、通知文字與時點均核對；10組自然caller證明等待與續接；
IDA進入點、原指令bytes及控制流相符，初稿106限定升READY。
remake產品實作、保存、視窗區域及正常GUI驗證另記對拍071。
完整062、整段戰術同態與正常換手GUI仍未完成。

## 同輪追加勘誤：頭像與首報中間幀

初次產品畫面差分推翻「第三參數0不畫頭像」。v1凍結資料中的face=-1是錯誤推論，
不是量測值。原版資料保留，106退回DRAFT補證，再依v2重新審查。
3826:040C先讀第二參數；045E比較FFFF，非FFFF在0465取DS8F4D情報參謀。
只有FFFF才在046D取第三參數提督低byte。0480畫頭像，048B才呼叫3765:006C等待。
本批原版情報參謀為13／58。另製3826 COM SHA-256
bec30e4885bdbd42ac0fdf512cc0e1c271fdab0bc159d09bbcf835bb73d2d2ad，
IDA 9.4原指令、caller參數及四組complete雙跑互證。

第一張capture0原採樣1188500000步尚未畫完頭像；3958:0215在1188165561進入，
3765:006C在1188565812才進入。原採樣比等待早65812步。
僅確認通知尚未返回不能證明畫面已完成。capture0／1與partial0／1-complete
固定1189100000步，四組均已命中頭像與訊息等待，兩座位通知續接未返回。
六張正式報告矩形改採四組complete及兩組second。游標完整保留，不加遮罩。

新增後共14組RAM、framebuffer、色盤、PNG雙跑逐byte相同；保存全70欄位亦相同。
v2保留v1雜湊與推翻原因，將報告頭像、3826參數及完整採樣加入來源索引。
冷VGA常式診斷不當作頭像顯示證據。

## 本輪產品驗證與現行來源

現行v2凍結244來源、14組自然雙跑與28項proven語意，SHA-256
306f7612cb536cdb441bba2e245baecb9786fbc84a9210b066b06a09934ac262。
v1的無頭像推論已推翻，保留歷史來源不供產品測試使用。
156案產品完整星域、104則報告與費沙時點符合；六張完整報告矩形各58240像素零差。
原版直接改行星記錄；remake typed副本保存前未同步，使兩案保存舊Owner。
補SaveSlot戰場分支及進入公開保存選單前同步後，六檔重讀通過。
正式首戰兩倍率GUI及標準套件通過，限定範圍見[071](../playtest/071-battle-ownership.md)。
產品307來源SHA-256為4ea853fd7d5fa0e97d19745ae21b38ba19ab6148141fe8d79d737be02fc3e3e3。
106保持READY，正常GUI換手與完整八步同態尚未完成。
