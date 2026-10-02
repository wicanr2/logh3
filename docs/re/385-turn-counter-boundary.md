# 385 回合計數器在全部戰場與尾段結束後進位

工作[#89](https://github.com/wicanr2/logh3/issues/89)，方法與範圍見[第47輪](../goal/2026-10-03-turn-counter-round-47.md)。
原版GIN3.COM SHA-256 `770ed52607bf02584c125043830f22b19cc4b9a21f4909ad2f5f02d8fc8fd918`；
dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`乾淨；沿用既有Go1.24工具 image。
完整旗標保存在[重播工具](../../tools/turn_counter_oracle.sh)及各批次flags.txt，原始記憶體保持本機。

## 已證實

- 無戰鬥正常實行：435711455步WORK+04為0，435711456步變1。
  432905435進全部戰場掃描，433372923進回合尾段，435711456才寫回合。
  `normal89-r47-clock-a/b`各四份完整1MiB RAM及色號／PNG／DAC共七份雙跑相同。
- 正常S3固定seed D3DC167E，正常實行與叛亂報告輸入後進首戰。
  300000001步及434999999步WORK+04都是380，watch無寫入；兩份RAM與色號／PNG／DAC共五份雙跑相同。
- IDA Pro9.4／SDK940一次性COM輸入SHA-256 `4d3f044955ce8e52c79b5b0c52060cd1b0474b57bddc2aa76f871f283799d72d`。
  原始名sub_10543，IDA範圍10543..105E8；COM基底10200、原版CS:IP0374分欄。
  前一支尾段033E `89ec5e5dcb`，0343才是主迴圈。未做語意改名。
- 0374:03B3 `9ae8037403`呼叫全部戰場；03BE `7407`跳過非零提前retf。
  03C7 `9a8e076a05`呼叫回合尾段；03D0 `7407`同樣只讓零值走到進位。
  **真正寫入端0374:03D9 `ff06188f`**，dosgolem線性3B19，IDA105D9。
  下一條0374:03DD `b001`是mov al,1；兩種位址不能合併。

## 舊紀錄勘誤

[研究251](251-turn-loop-and-battle-nesting.md)寫0374:03DD／線性3B1D為inc。
本次watch與完整RAM重現同一順序，原始bytes與IDA指出該值是執行後IP。
真正指令在03D9／線性3B19。這項勘誤不推翻全部戰場嵌套與末尾進位的控制流結論。

## 玩家路徑影響

原版與remake的首戰計數380／381，來自Game.beginTurn過早增加g.turn。
既有AI、未安定度與費沙事件使用turn-1補償，Date與SaveSlot直接保存提早增加的值。
先用可丟棄覆蓋將進位移到全部戰場／回合尾段報告／勝負之後，再建立088 READY子契約。
正常bootstrap／Game首戰恢復799/3/1；19份輸出雙跑相同，7艦隊／55筆池與亂數結果保持。
尚有完整畫面5853 pixels及WORK32／FLEET120／SHIP865／STELLAR44 bytes差，父088保持DRAFT。

## 正式驗收

[完整收據](../internals/data/turn-counter-boundary.json) SHA-256 `09d28a9e6d2b5c5cc49fd4f1ffa8d7c057170e4eb7f9e022baf6c9593d3a5a7b`。
48限定PASS零skip，18標準套件通過；標準命令`tools/gui.sh xrun go test ./internal/... ./cmd/... -count=1`。
三份隨附存檔正常實行、戰術回合與完整保存保持，首戰戰略回合分別93／346／380。
兩種AI建隊／命令都讀實行前回合；兩個戰場與費沙模態期間不進位，全部完成後只進位一次，終局不進位。
既有#85兩組正常星域4032／行星6642 bytes與全圖對拍保持；不由這些樣本推論完整S3資料已同態。

正式cmd/gin3真實鍵鼠從主選單、138、S3與帝國選擇至首戰，一倍與兩倍完整640×480色號與正常Game相同。
PNG目視確認日期7990301與達希利第一回合。GUI未使用script、direct-entry或記憶體注入。
前兩次外部resize失敗分別為選錯命令及未形成兩倍視窗；輸出留workplace/round47-gui-*。
路由重載Xvfb試玩入口後，改用產品既有-scale及PID選視窗，同image與同命令重跑通過。
這是抽驗腳本問題，未因此改產品或放寬視窗尺寸／全圖相同斷言。
