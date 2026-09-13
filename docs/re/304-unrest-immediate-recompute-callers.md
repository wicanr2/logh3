# 304 — 玩家編成、AI 編成／整補與費沙事件要立即重算未安定度

日期：2026-09-14  
原版證據：[`299`](299-region-unrest-writers-and-recompute.md) 已驗證的
`090A:0E6D` 十八個呼叫點與 `urx299w` 時間線  
工具：IDA Pro 9.4 headless、dosgolem-logh3 `e0b6a03`；位址採原版執行期
`segment:offset`  
狀態：三類呼叫點 **proven**（原始 far call 與動態時間線）；remake 接線
**verified**（正常玩家路徑與事件／AI 回歸測試）

## 原版時點

- `297B:0546` 的尾巴在 `297B:05A6` 呼叫 `090A:0E6D`。玩家按下艦隊
  編成的「決定」與 AI 建艦隊、整補都走這支艦隊能力重算；`urx299w`
  實際量到 AI 每建一支就重算一次。
- 費沙事件包裝器 `1213:023B` 在事件本體回來後，於 `1213:0256` 呼叫
  `090A:0E6D`。事件會改費沙與對手首都的經濟力、數格民心與叛亂結果，
  所以不能延到下一個座位回合。
- 原版新艦隊剛建成時 `+0x08 == 5`；AI 後續下移動命令才改變這個狀態。
  未安定度 f6 正好會計入「同陣營、同星域、`+0x08 == 5`」的艦隊。

## remake 接線

- 玩家正常路徑在 `StateFleetComposition` 按「決定」後，把新艦隊設為
  `Parked` 並以目前回合計數器重算。
- `buildFleets` 每成功加入一支 AI 艦隊就重算；`refitFleets` 每處理完一支
  就重算。兩者發生於原版計數器進位前，所以傳 `g.turn-1`。
- `runPhezzanEvent` 在事件本體完成後、保存報告清單前立即重算，同樣使用
  事件當刻的 `g.turn-1`。

`TestFleetDecideRecomputesUnrestLikeOriginal` 從正常玩家路徑走到編成，先把
32 格寫成不可能被「沒呼叫」掩蓋的哨兵值，再驗證建成艦隊為停駐狀態且
整張已重算。`TestPhezzanEventRecomputesUnrestImmediately` 直接釘住事件包裝器
完成後的接線；既有 `TestFleetDecideMatchesOracle` 仍與 `dec1` 畫面相同。

## 尚未由本篇解決

這不是 #19 的完成宣告。下列仍保持 open：

- 逐格戰鬥與叛亂效果之間的重算順序；
- `2E54:04FB` 與其餘尚未辨識的呼叫點；
- 載入後第一次回合尾巴之前所讀路徑表的型態；
- 回合尾巴逐格更新 `+0x13` 與未安定度的交錯差異。

這些會改變同回合中間狀態，必須先取得同狀態收據，不能用本篇三個已知
呼叫點外推。

