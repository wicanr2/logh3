# 第78輪：交戰音效來源

日期：2026-10-04。沿[#48](https://github.com/wicanr2/logh3/issues/48)及
[工作快照](../goal/2026-10-04-fleet-sound-round-78.md)。
父提交e927750a782af344d77e6a3dd866e18af753ad28。

## 標頭訂正

已證實：GIN3SP.VOC共321633 bytes，SHA-256
`b36a1b43f3199e776f70302671a6bc37698dd5255113995e7a7f55f78f699a9f`。
檔案前16 bytes為`MicroWaveSVFv1.0`，接續bytes為`1c000535a614`。
事實01按副檔名寫成Creative Voice，沒有標頭證據；本次直接bytes足以否定此格式判斷。
開機會開啟本檔的既有證據保持。下節補樣本索引、長度與消費端參數。

固定RAM沿normal79-r61-a/end-mem.bin，SHA-256
`6800de725851633cf2b817afad0cfda3be4a0d6b8bb7d1ebdacd57a92ce1d5b4`。
dosgolem固定f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c。
原始CS:IP、IDA EA及線性位址分欄，未審查語意保持unknown。
來源匯出由[IDA探針](../../tools/scripts/fleet_sound_ida.py)執行，輸出留workplace/ida。

## 樣本與選擇

已證實：2376:16EC開檔後讀16 bytes，172E只比對前12 bytes的`MicroWaveSVF`。
1743讀兩byte樣本數。本檔為28，18起逐段四byte標頭與payload，恰到321633 EOF。
2376:1644從18起跳過前段payload，讀四byte標頭：word長度、byte時間常數、byte DSP命令。
DS:92C2存長度、92C1存時間常數、92C0存命令。28段時間常數均166，命令均0x14。
1592把0x40與時間常數交DSP，再交0x14及長度減2。載入仍讀完整payload。

3B74:031A先讀2376:1585的busy byte到0，再把輸入低byte除28的餘數交1644。
030E／0314只呼叫1603／1626的64KiB緩衝區配置／釋放，不代表播放或停止樣本。
未深入DSP寫入helper、ISR、DMA或busy實機時鐘。

1DD6:0A17沿既有406原始指令：SHIP+03為0選表列7，否則取SHIP+04艦種。
攻擊表DS:43E0，中彈表439E，兩表各66 bytes。每陣營33 bytes，每列3 bytes。
攻擊第三維為武器0／1／2，中彈第三維為損失百分比<20／<40／其他。
靜圖及動畫關閉不播，音效byte為FF亦跳過。零損失不進守方動畫。

## 平台前提

[Creative Hardware Programming Guide](https://www.phatcode.net/res/243/files/sbhwpg.pdf)
第3章3-4／3-13及第6章6-6：0x14是單聲道unsigned 8-bit PCM，長度參數是樣本數減1。
時間常數166導出每樣本90微秒，等效取樣率1000000/90 Hz。
因此本版播放payload前N-1 bytes，時長(N-1)×90微秒。
此結果是hardware-spec approximation，不宣稱原版逐波形或wall-clock一致。
正式轉檔使用11111 Hz近似輸入，重取樣至44100 Hz立體聲OGG；相對速率差約0.001%。

## 正式接線與驗證

來源索引保存310條IDA 9.4原始指令、9項人工審查語意與28段完整payload雜湊。
未逐條審查的指令仍標unknown。[104](../spec/104-fleet-sound.md)先READY再接正式路徑。
SVF解析完整標頭與EOF，拒絕截斷、尾端多餘資料及不支援的DSP命令。
Game在攻擊／中彈首格前送一次性事件，前一樣本未完成時延後圖段，正式app取出播放。
一次性播放器與配樂共用Context，音效不循環；缺完整28段、無裝置或解碼失敗時明示停用。

音效13的FFmpeg解碼含尾端補樣本，未裁切時比OGG有效結束位置長約22毫秒。
轉檔器按OGG有效結束位置核時長及裁掉補樣本，原有5毫秒門檻保持。
正式播放器的28段解碼串流長度與有效結束位置相符，這不是人耳或裝置驗收。
轉檔雙跑OGG與裁切後PCM雜湊相同；完整命令的暫存路徑不同，各收據仍保存原值。

四案三種速度及動畫關閉核原版音效編號、原池庫存／戰損及seed。
首格等待與零損失守方不播另做局部診斷；零損失案明示將目標設為0艘，不冒稱正常原版局面。
兩陣營正常公開輸入各雙跑，正式bootstrap與app錄製器收到相同序列，六份保存與無音效父輪相同。
完整數字、雜湊、命令與範圍見[驗證067](../playtest/067-fleet-sound.md)。
本輪未新增dosgolem硬體播放能力；沿既有COMANI序列及原始bytes，不產生偽造的原版PCM收據。
正常原版整段同態、真實GUI交戰及聆聽仍缺；104保持READY，#48／#79保持OPEN。
