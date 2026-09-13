# SPEC 045 — 月底船艦接收報告的底板細節

狀態：**CONFORMED**（2026-09-14；首都區與兩個箭頭逐位元組差異 0）

證據：[`docs/re/55`](../re/55-ship-delivery-report.md)

## 範圍

只處理 `StateShipDelivery` 畫面上兩項已證實的呈現差異：結算中地圖不以
玩家色重塗首都，以及右側捲軸槽顯示上下箭頭。星域 24／28 因戰略行星
攻占尚未接上而產生的歸屬差異，不在本規格內。

## 契約

- `StateShipDelivery` 使用與 `StateVersusInput` 相同的 `plainMap`：兩方首都
  都是普通星域色，不把玩家首都改成帝國色 6／同盟色 5。
- 上箭頭佔 `(612,305)`–`(621,314)`，下箭頭佔
  `(612,385)`–`(621,394)`；兩張都是 10×10 色號圖。色號矩陣直接取自
  `turnmonth2/screen.bin`，由 `screen.DrawSettlementScrollArrows` 作為唯一
  幾何與像素來源。
- 箭頭在底板與右側面板完成後、船艦接收報告視窗之前繪製；不得出現在一般
  命令階段。

## 驗收

`internal/game.TestShipDeliveryMatchesOracle` 由正常玩家路徑跨過月底，分別將：

- 包含帝國首都全區的 `(0,40)`–`(79,119)`；
- 上、下箭頭各自的 10×10 區域；

與原版 `turnmonth2` 收據逐位元組比較，三區差異皆須為 0。收據不存在時
測試必須 skip，不製造替代 oracle。

## 實作與結果

`Game.playerStarMap` 在 `StateShipDelivery` 選用既有 `plainMap`；
`screen.DrawSettlementScrollArrows` 保存並繪製兩張 10×10 色號矩陣，呼叫點
位於右側面板完成之後。上述三個驗收區域均為 0 差異。

原版 `screen.bin` SHA-256：
`ae3622b8887b830ff1adf629d7223cd0a0be5e47a0862fb6a55057057ac30613`；
收據綁定 dosgolem `2833ec9` 與完整旗標（本機
`workplace/oracle/turnmonth2/meta.txt`）。
