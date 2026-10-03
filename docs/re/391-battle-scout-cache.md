# 391 未出擊偵察隊的圖示快取

沿[#89](https://github.com/wicanr2/logh3/issues/89)及[第53輪](../goal/2026-10-03-battle-scout-cache-round-53.md)。
父提交517fecc21f8ad0868e26e49448e7edea70e3ee25，第52輪JSON保持不變。

## 原始證據

原版輸入與工具沿用[研究390](390-battle-hover-marks.md)的固定bundle與dosgolem f7523f9。
正常首戰完整RAM為`workplace/oracle/normal89-r51-base-a/entry.bin`，SHA-256
`ef34310fe3e2c133522ed1612f16faeaff8f14e94ae9e0b6ea965e9749424a97`。
其獨立b側相同；完整旗標、版本與輸入在第51輪收據，這輪不更動執行器或重寫原收據。

這輪回查兩份既有IDA Pro 9.4資料庫匯出。它們的自動函式名只供導覽。
26E5:00AC共113條、25DB:03E9共96條原始指令，全部逐byte與上述RAM相同。
IDA位址基準10200、原CS:IP與dosgolem線性位址保持分欄。

| 匯出 | SHA-256 | 載入COM SHA-256 |
|---|---|---|
| `re79-marker-r33-reviewed-26E5.json` | `67c9e061d877dcbc91d59500450d7f651e0521aa023b1e072e29e0839e29bd2a` | `8c5364070af105b4c49c5dd539d3f9cfcecaf72a9098ea97534d3be14afdeeb2` |
| `re79-second-r36-25DB.json` | `b2abee692687909ff67958d75cc8bf8a7241f363767e439f6293cd57b912153b` | `d1004845670387a15986f585658051965a2cb899f1f499d062b7b7ba8bd34b7a` |

## 已證實的資料流

- 26E5:016F從FLEET+38讀偵察隊池，0173乘78；沒有以可點選單位清單排除它。
- 0188比較FLEET+3A與FF。未出擊時018F令force為0，01B6仍呼叫25DB:03E9。
  01C4再呼叫1727:02E2刷新標記，沿082既有契約。
- 25DB:0408檢查force；非強制且SHIP+17不為3則進0466。
  0491讀SHIP+04艦種，艦種7於0498以`26c6472f09`寫SHIP+2F為9。
- 正常S3七筆偵察隊皆未出擊、艦種7、Detail0，原版快取均9。
  remake底稿仍FF；七筆正是第52輪SHIP276 bytes中的七個+2F差異。

| 艦隊索引 | 偵察池 | 原版+2F | remake修正前+2F |
|---|---|---|---|
| 12 | 81 | 9 | FF |
| 16 | 103 | 9 | FF |
| 17 | 143 | 9 | FF |
| 28 | 110 | 9 | FF |
| 29 | 120 | 9 | FF |
| 30 | 127 | 9 | FF |
| 31 | 136 | 9 | FF |

根因是Game.refreshBattleMarkers只遍歷battleUnits，未出擊偵察隊不在該清單。
修正只補相同時點的原池快取更新，使用既有BattleCachedSprite與ScoutUnit。
不把偵察隊加入renderer或命中清單，不改+30、far指標、RNG或玩法。
完整版面與完整存檔仍不能稱same-state；剩餘一像素及其他欄位保持待驗。

## 限定驗收

[結構化收據](../internals/data/battle-scout-cache.json)SHA-256
`1203dc687f6f2481ac53acfaa3a2da8866d8d82638ebeda6a9dd452d875bf3d9`。
第51／52輪原JSON保持，既有209條原指令自動合併十項附加分級，原名／原位址不覆寫。

- 正常Game55檔a／b相同；相較第52輪只改七個+2F bytes，WORK／FLEET／STELLAR／PLANET及seed不變。
- 兩AI正常入口55筆快取與原版相同；七筆偵察池維持未出擊，48可點選單位及完整SHIP存讀保持。
- 五張完整畫面與第52輪相同。全圖差1／0／0／0／0，五張小地圖各0／16384。
  原圖及差分沿用逐byte及雜湊重核的第52輪影像，沒有裁切或新增遮罩。
- 137限定PASS零skip、18標準產品套件及正常cmd/gin3 X11兩倍率通過。
  完整SHIP差276減269；WORK14／FLEET27及far指標／SHIP+30保留。

重播使用既有工作目錄與包裝器，輸出已存在會拒絕覆寫：

```sh
bash tools/gui.sh xrun python3 tools/scripts/battle_normal_probe_seed.py --prefix round53-normal
bash tools/gui.sh xrun go run -overlay workplace/round53-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round53-pass-a -input-display
bash tools/gui.sh xrun go run -overlay workplace/round53-normal-overlay.json tools/scripts/battle_normal_probe.go -out workplace/round53-pass-b -input-display
bash tools/gui.sh xrun env LOGH3_TURN_GUI_PREFIX=round53-gui LOGH3_TURN_PLAYER_DIR=round53-pass-a bash tools/turn_counter_gui.sh
bash tools/gui.sh xrun python3 tools/scripts/battle_scout_cache_receipt.py
```

收據內保存全部測試log雜湊；限定137事件含子案例，不稱137個獨立玩家流程。
未出擊／回收後Detail非零的完整標記刷新鏈未新增正常原版驗收，沿既有082限制。
原版雙跑為前輪收據重核，不宣稱本輪重新執行原版。完整088仍DRAFT，#89／#79保持OPEN。
