# 450：新編艦隊的戰術配置與新遊戲的原池

2026-11-02，第172輪。起點是使用者提供的Windows崩潰堆疊。規格見[159](../spec/159-new-game-first-battle.md)，驗收見[143](../playtest/143-new-game-first-battle.md)。

## 崩潰位置

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x40 ...]
rules.(*Slot).fleet(...)                         save.go:577
rules.(*Slot).BattleFleetsFor(...)               resolve.go:275
game.(*Game).startBattle(..., 0xf, 0x1)          battle.go:75
game.(*Game).startPlacedBattle(...)              battle.go:23
game.(*Game).scanRegionsFrom.func1()             rebellion.go:116
game.(*Game).dismissTurnReport(...)              turnreport.go:86
game.(*Game).Click(...)
```

`addr=0x40`是`Slot`裡`Fleet`切片的長度欄位：`Stellar`佔0x18、`Planets`佔8、`Admiral`佔0x18，`Fleet`從0x38起，長度在0x40。所以接收者`g.slot`是nil，星域是15，場景是點掉「發現敵方艦隊」報告之後。

`startNewGame`把`g.slot`設成nil，只有存檔或讀檔才會建出來。沒存過檔的新遊戲一路走到第一場戰鬥，就在這裡崩潰。**proven**：純滑鼠鍵盤操作的正式v1.6.4執行檔重現相同堆疊，見驗收143。

## 三個互相牽連的洞

單補nil判斷只會得到空戰場，因為`BattleUnitsFor`對nil回空切片。追下去有三件事要一起處理。

| 洞 | 現象 | 影響範圍 |
|---|---|---|
| 新遊戲沒有原池 | `BattleFleetsFor`讀nil | 沒存檔的新遊戲第一場戰鬥崩潰 |
| 新編艦隊到達時池索引是-1 | `PlaceFleetArrival`安靜回false，兩軍分隊全在(0,0) | 存過檔的新遊戲、讀檔局新編的艦隊 |
| 新編、沒行軍的艦隊從未配置 | 同上，座標是零值 | 同上 |

另有一個獨立的舊洞，也是這條路徑測出來的：`disbandFleet`只改typed表，原始表那一列留著。電腦在下一次同步前於同一格重編艦隊時，那一列不算fresh，新艦隊的`Scout`零值被當成池#0，`SyncFleets`回「艦艇池索引0的擁有者／艦種不符現況」。這會讓存檔失敗，也會讓`startBattle`不開戰直接返回。

## 原版怎麼配置

原版建艦隊走`297B:03EB(槽, 1)`（[212](212-fleet-composition-template.md)），第二個參數非0就呼叫`297B:08D8(艦隊, 星域, 星域)`（[230](230-composition-change-refresh.md)）。新編艦隊在建立當下就配置，來源與目的都是自己的星域。

入口表只有兩筆`from`等於自己的列，都在首都：星域0的(336,280)、星域31的(1696,1740)。原版傾印與它逐值相符。

| 收據 | SHA-256 | 艦隊表位 | 星域 | 中心 | 航向 | 說明 |
|---|---|---|---|---:|---:|---|
| `fcmd/fcmd-base-fleet0.bin` | `3a077e0847f611d3db1cfd18a88c6bfb768fc4e24e7822e40e5251b6708704d7` | 0 | 0 | 336,280 | 34 | 玩家新編，未行軍 |
| `t260_post.bin` | `4b4e6083be68a2ad4b86f9d93685642d99d7a1b33ad022f2c382ca45b03bb1da` | 28 | 31 | 1696,1740 | 162 | 電腦新編，入口本身 |
| 同上 | 同上 | 31 | 31 | 1728,1708 | 160 | 電腦新編，環狀第1格 |
| 同上 | 同上 | 23 | 31 | 1728,1740 | 161 | 電腦新編，環狀第2格 |

環狀搜尋依序試(0,-32)、(32,-32)、(32,0)。第28支佔入口，第31支取第1格，第23支取第2格，所以配置順序是28、31、23，與表位順序不同。傾印在`~/cht/dosgolem-workplace-archive/logh3/`，未進版控。

## `297B:0ABF`選中心的順序

從`t260_post.bin`的線性0x2A26F起反組譯，objdump在Docker內執行。原版CS:IP：

| CS:IP | 指令 | 語意 |
|---|---|---|
| `297B:0ACB`、`0AD2` | `cmp` `[bp+1A]`／`[bp+1C]`與0xFF | 任一個是FF就跳到`0B6D` |
| `297B:0AE7` | `lcall 113A:0D2C` | 用來源與目的查入口表 |
| `297B:0AFF`、`0B0D` | `mov es:[bx+5C]`、`[bx+5E]` | 查到：欄×16、列×20寫中心，跳到`0B6D` |
| `297B:0B26` | `lcall 2AF5:060D` | 查不到：取該陣營經濟力最高的行星序號 |
| `297B:0B5C` | `lcall 0DEA:1759` | 行星座標寫中心，內部再走一次`08D8(FF, FF)` |
| `297B:0B73` | `lcall 297B:0BD6` | 碰撞環狀避讓 |

`0DEA:1759`（線性0xF5F9）拿序號乘0x52加上行星表基底直接讀，沒有判斷0xFF。`2AF5:060D`找不到行星時回0xFF，原版就讀到表外。remake不重現這個越界，改取該星域第一個入口。**strong inference**：行星分支與越界都只有靜態證據，沒有動態收據。

## 證據等級

| 結論 | 等級 |
|---|---|
| 崩潰是`g.slot`為nil造成，路徑如上 | proven |
| 首都新編艦隊的中心與航向 | proven，兩份原版傾印 |
| 同一星域多支新編艦隊的配置順序不是表位順序 | strong inference，由`t260_post`的環狀位置推得 |
| 非首都新編艦隊用待命行星座標 | strong inference，靜態 |
| 沒有待命行星時的原版行為 | strong inference，靜態；remake刻意不重現 |
