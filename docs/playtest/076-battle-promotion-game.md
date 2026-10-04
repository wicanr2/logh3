# 076：戰場升級的正式規則、報告與存讀

2026-10-05，第87輪，遠端[#79](https://github.com/wicanr2/logh3/issues/79)。
父提交b501420f46eb55314b1d9430fc931ce700ae1a79；108 READY，完整062 DRAFT。
原版與快取來源沿[075](075-battle-promotion-global.md)及[研究421](../re/421-battle-promotion-save.md)。
完整工具、程式與產物SHA-256見[產品來源](../internals/data/battle-promotion-game.json)。
索引216來源，SHA-256 `1f2f9b252fc73c4d8091c9483b383b7f170aeac90abacfb259b7fc9a536517d6`。
原版dosgolem固定f7523f9、乾淨；本輪沒有改動執行器或重寫歷史收據。

## 實作與受控原版對拍

108審查READY後才接正式8-4。從五張完整原池取得成長、能力、速度、
分隊士氣及全局重算。逐次提交後同步typed欄位，保持原有位置、命令、庫存及行程。
星域換手後沿戰略處理開始的路徑快取；LoadSlot清除快取。
保存與提督UI共享本局提督表，不取舊模板；十二級軍階依提督自己陣營選文字。
軍事費以當下Game欄位為準，WORK底稿到保存才同步。

| 驗證 | 結果 | 完成聲明範圍 |
|---|---|---|
| 66案、121次嘗試、53次升級 | 成長旗標、完整五表及末端相同 | 固定8-4入口的資料same-state |
| 473次RND | 回值與末端seed相同 | 每案預先固定12345678，不挑種子 |
| 11則人類報告 | 文字、頭像、等待與逐人重試相同 | 提督／副官、連升三級、左右按下 |
| 11張完整報告矩形 | 每張58240像素、有效16色零差 | 矩形[32,180,416,140]，沒有遮罩 |
| 診斷全畫面 | 每張248681像素不同 | 空白底圖，保留全圖差分，沒有宣稱全戰場相同 |
| 九案六檔保存及讀回 | 完整五表及行星payload相同 | 明示StateSaveGame診斷，不宣稱戰場續戰 |
| 兩案路徑快取 | 完整星域／WORK相同 | 換手案與同值控制，沒有提早重建 |
| 非法資料 | 七類錯誤原池與RND不變 | 池、索敵池、分母、陣營、路徑、規格及提督 |
| 永久等待／當下經費 | word回繞與現況預算通過 | 額外產品資料鏈回歸，不稱新原版收據 |

## 正常玩家與GUI

正式bootstrap經主選單、138、S3帝國與實行到達首戰。
再用公開戰術選單實行一次，5000個1/60秒tick後回Input，
由公開存進度選單保存第0格。WORK+06為1，戰略回合仍380。
沒有注入功績、位置或種子；兩次獨立重播的三張完整畫面、六檔保存及輸入紀錄相同。
這證明8-3→8-4→既有8-5入口可正常續接，不證明自然功績升級或原版整段同態。

cmd/gin3真實X11鍵鼠另抽首戰1×／2×，邏輯640×480完整色號均與同版本正常Game相同。
視窗分別640×480／1280×960；二倍使用有界1600×1200 Xvfb，沒有裁掉下方。
本抽樣限首戰入口回歸，尚未驗正常升級GUI。

## 重跑入口

以下均由專案Docker包裝器執行。輸出目錄須獨立，不覆寫固定來源。

```sh
bash tools/go.sh test ./internal/rules -run '^TestBattlePromotion' -v -count=1
bash tools/gui.sh xrun env LOGH3_PROMOTION_OUTPUT=/src/workplace/promotion79-r87-game-final go test ./internal/game -run '^TestBattlePromotion' -v -count=1
bash tools/gui.sh run tools/scripts/battle_promotion_player.go -out workplace/promotion79-r87-normal-b
bash tools/gui.sh run tools/scripts/battle_promotion_player.go -out workplace/promotion79-r87-normal-c
bash tools/gui.sh test ./cmd/... ./internal/... -count=1
bash tools/gui.sh xrun python3 tools/scripts/battle_promotion_game_receipt.py
```

凍結器讀固定命名的來源、標準及目標log；正常GUI沿tools/turn_counter_gui.sh，
設定LOGH3_TURN_GUI_PREFIX=promotion79-r87-gui與LOGH3_TURN_PLAYER_DIR=promotion79-r87-player-a。
重建本機原版診斷來源見075及421，不以自製資料替代。

## 限制

自然累積功績後的正常升級、正常升級GUI、8-1清空分隊到8-4的原池輸入鏈，
8-5原版結束判定與完整062同態仍缺。108保持READY，#79保持OPEN。
此輪不改音訊與核准的零距離、清單速度或OGG門檻，不重開既有使用者決定。
標準18套件13 PASS／5無測試；輸出擁有權、來源與Docker清理在收尾核對。
