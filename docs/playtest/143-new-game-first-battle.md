# 新遊戲第一場戰鬥崩潰

2026-11-02，第172輪。來源為main `5b7c6c7e`加本輪未提交修改。規格見[159](../spec/159-new-game-first-battle.md)，根因與原版證據見[研究450](../re/450-new-fleet-placement-and-new-game-slot.md)，雜湊見[機器收據](../internals/data/new-game-battle172.json)。

## 結果

八次都走正式執行檔的正常入口，輸入都是滑鼠與鍵盤：主選單開新遊戲、答題138、編第一支艦隊、下令占領，然後一直按實行並點掉報告。帝國往星域15，同盟往星域13，電腦第4回合都會到那裡。沒有注入狀態、種子或時間，也沒有用script旗標或存檔。

| 執行檔 | 平台 | 帝國 | 同盟 |
|---|---|---|---|
| v.1.6.4-20261031 `gin3` | Linux | 點掉「發現敵方艦隊」後崩潰 | 同左 |
| v.1.6.4-20261031 `gin3.exe` | Windows，Wine | 同上，`0xc0000005 addr=0x40` | 同左 |
| 修正版 `gin3-fixed` | Linux | 通過 | 通過 |
| 修正版 `gin3-fixed.exe` | Windows，Wine | 通過 | 通過 |

修正前四次都以結束碼2退出，堆疊都是`BattleFleetsFor`、`startBattle`、`scanRegionsFrom.func1`、`dismissTurnReport`，與使用者的Windows截圖相同。

修正後四次都走完同一條路：進戰場、打完九回合、回指令畫面、存進度、取進度、再按實行，並進到下一場戰鬥。兩個陣營的七個存檔檔案在Linux與Wine之間逐位元組相同。

驅動器在進場時檢查鏡頭：玩家旗艦要在28×18格的視窗內，四次都通過，帝國(256,1740)對視窗(2,78)，同盟(1552,1880)對視窗(83,84)。這項檢查是目視抓到問題才加的，第一輪同盟進場時鏡頭對到帝國艦隊，原因見研究450。

戰場第1、9回合、戰後地圖與讀回後的畫面已目視：地點名稱正確（帝國艾爾・法西爾、同盟伊謝爾倫），回合數前進，讀回後日期與艦隊都在。敵方在索敵範圍外，所以進場時看不到。

## 測試

標準命令`go test ./cmd/... ./internal/...`在Xvfb下完成21套件：16 PASS、5無測試、0失敗。新增的測試在`internal/game/newgamebattle_test.go`與`newgamebattle_normal_test.go`：

| 測試 | 驗什麼 |
|---|---|
| `TestNewGameFirstBattleEntersWithBothSides` | 從星域掃描進戰場，雙方都有分隊，全部不在(0,0)且可通行 |
| `TestNewGameFirstBattleFromExecuteSavesAndResumes` | 按實行進戰場，存檔裡的中心等於入口，戰後同步成功 |
| `TestNewGameNaturalFirstBattle` | 原版AI下命令，星域15、2支艦隊，與截圖相同 |
| `TestNewGameContinuesAfterFirstBattle` | 正式資源、純玩家輸入、兩陣營：九回合走完正常收尾，存進度、取進度後再玩20回合，存檔與typed艦隊一致 |
| `TestNewGameSaveBetweenArrivalAndBattleKeepsEntry` | 到達後、開戰前存檔，開戰仍用來源入口 |
| `TestVersusFirstBattleEnters` | 對戰模式開局一樣能進第一場 |
| `TestNewFleetAtCapitalUsesOwnRegionEntry` | 中心與航向符合原版傾印：(336,280)／34、(1696,1740)／162 |
| `TestNewFleetAwayFromCapitalUsesStandbyPlanet` | 非首都新編艦隊用待命行星座標 |
| `TestLoadedGameFreshFleetArrivalIsPlaced` | 讀檔局新編艦隊到達後有座標與池索引 |
| `TestPlacedFleetsKeepTheirPositions` | S3已配置的艦隊不被移動 |
| `TestDisbandThenRebuildSameSlotSyncs` | 解散後同格重編可以同步 |
| `TestNewGameWritesSeatControlBytes` | 三種模式的`WORK+0C..14`與保存事實30相同，帝國列也等於原版傾印 |
| `TestNewGameBattleCameraFollowsPlayer` | 兩陣營進場鏡頭都對著玩家艦隊 |

負向對照：拿掉修正後，`TestNewGameNaturalFirstBattle`以相同堆疊崩潰；拿掉`DisbandFleet`後，`TestDisbandThenRebuildSameSlotSyncs`回「艦艇池索引0越界或重複」；沒有寫座位時，座位測試三列都失敗，鏡頭測試同盟失敗。

## 限制

- Windows只在Wine驗證，沒有實機。專案預設的`kol-wine-verify`映像本機已不存在，這次用既有的`yuan-wine-verify:noble-r2`（wine-9.0），圖形用`original`。Linux用預設HD。兩平台都關音效。
- 比較等級是`nearby-state`。同一星域多支新編艦隊的避讓順序與原版不同，非首都的行星分支只有靜態證據，見規格159。
- v1.6.4封包仍含這個崩潰，修正版封包見[144](144-v165-full-delivery.md)。
- 使用者確認第169輪回報的閃退發生在進戰場，與這次是同一個崩潰。當時走讀檔S1，所以沒有重現。

## 重播

```sh
P=/src/workplace/release171/extracted/squashfs-root/usr/bin
LOGH3_GUI_IMAGE=logh3-ebiten:20261030-r169 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh \
  tools/gui.sh xrun python3 tools/scripts/new_game_battle_gui.py \
  --package $P --binary $P/gin3 --out /src/workplace/newgame172/before-linux-empire
```

`--out`必須是尚不存在的目錄。同盟加`--faction 1 --menu-y 219 --target 151 216 --load-y 228`。修正版把`--binary`換成建置結果。Windows加`--windows --graphics original`，在Wine映像內以`tools/music_wine_entry.sh`執行，`--package`掛Windows封包目錄。

本機輸出在`workplace/newgame172/`，每次一個子目錄，含`RESULT.json`、`FRAMES.jsonl`、`game.log`與截圖。機器收據由`tools/scripts/new_game_battle_receipt.py`產生。
