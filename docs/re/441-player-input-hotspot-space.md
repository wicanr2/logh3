# 441：完整版滑鼠阻塞與熱區座標勘誤

2026-10-06，第120輪。原封包 v.0.1.2，來源提交66d7ad7；Go1.24.13、logh3-ebiten:3、Xvfb／xdotool。正式程式由片頭進入帝國，未使用 script 或直接入口。

## 已證實

會計報告可見變更與確認小框為 x430、y340／360、36×20；JSON 熱區為 x432、y300／320、32×20。訂造及建造計畫同樣差40。Game.Click 承諾畫面座標，卻直接把 y 交給這三份舊滑鼠熱區。實際點擊448,350後仍在會計報告；448,370亦無作用。

三份下拉選單 Mouse 固定600,300，Frame 的 liveCursor 狀態白名單不含選單。實際移到選單項目後可見箭頭仍停在600,340，Game 本身持續接收滑鼠。訂造與建造畫面也沒有動態游標接線。訂造確認及建造結束沒有左鍵分支。

研究52對 Hotspots 全部是畫面座標的敘述不能套用此三份歷史 JSON。研究259末尾已指出其熱區是畫面減40座標，早期測試卻仍餵300／320附近的座標，因此未覆蓋玩家眼前的按鈕。保留歷史收據；撤回 v.0.1.2 已驗正常月初左鍵操作的暗示。原版位址與熱區 bytes 仍以研究09／12／259原文為準，本輪不新增反組譯定位或原版 oracle 收據。

修正範圍與新收據見[121](../spec/121-player-input-hotspots.md)、[107](../playtest/107-player-input-hotspots.md)。完成後的來源、圖片及封包雜湊由[索引](../internals/data/player-input-hotspots.json)保存。

原封包重現圖片均在 `workplace/input96-r120-old/`：

| 檔案 | SHA-256 |
|---|---|
| 01-account-x11.png | e125fdc0a3297f33d47b9341022d263fc66f0e0f9c2dd8535ab46dca577d507d |
| 02-change-visible-x11.png | c1290d6e58b341440cc695bf95b2935efffd69b07834ff21d599c233935a6c5c |
| 03-confirm-visible-x11.png | a91321ec09ddaa0fa93530481b5b25eccf8dc2b248dbe09a846cf3d7e34972fd |
| 07-system-moved-x11.png | b397897ef9055c82cbe7b4834eeec90991aa78ecdc661e8382e4f40080164508 |
| TRACE.json | 7f6b9abda4b3f4ce34c5503a8facfc4ea1cb66c4fe5ca02199c75de5d6e5b530 |

TRACE保存正式參數、X11座標、按鍵及時間。永久的player_input_gui.py亦可將binary／assets指向舊完整包，沿新規格的預期畫面重現於可見變更按鈕失敗，不以直接Game入口代替。
