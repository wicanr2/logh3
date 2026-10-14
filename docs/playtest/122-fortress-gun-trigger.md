# 要塞砲正常操作與原版觸發驗證

2026-10-14，第133輪。HD／decomp已先以`b36b950f`提交並推送。
本輪完成要塞砲觸發及原版演出取證，沒有實作新的要塞砲演出。
唯一來源索引為[fortress-scene.json](../internals/data/fortress-scene.json)，SHA-256：
`22160d0c3c08d8c72fba9208328de2e2da11818d305955f67a117b1a137c04b7`。

## 正常玩家路徑

| 條件 | 結果 |
|---|---|
| 入口 | 正式片頭→取進度→S1亞姆力札→同盟；防拷138 |
| CPU | `-ai -1`，目前原版規則政策；沒有被動覆寫 |
| 狀態／seed／腳本旗標注入 | 全部沒有 |
| 先前戰場 | 星域8正常選撤退，續到星域13伊謝爾倫 |
| 要塞 | 真實行星33、同盟、防衛力上限12800、座標936,1172 |
| 操作 | 平移→行星選單「要塞砲」→橢圓瞄準→確認→戰術「實行」 |
| 結果 | 第1回合開火，續到第2回合；GunState由3轉2，GunAim161 |
| 戰損 | 2支分隊，2678及2254艘，共4932艘；逐支記133及112分 |
| 測試射擊 | 朝最近的同盟艦隊19方向，會傷及友軍；未宣稱命中敵軍 |
| 時間 | 約27.2秒 |

測試器只增加只讀行星／選單／戰損觀測，輸入皆為真實X11滑鼠與鍵盤。
原版候選含雙方分隊，本次不改友軍傷害政策。
畫面已檢視要塞選單、橢圓與回合續行。這證明正常操作可開火及結算，不能證明要塞砲動畫已接。

| 本機收據 | SHA-256 |
|---|---|
| `workplace/round133/gin3` | `f57ee5bbd0347a572cbf57f7e35129a44290864c3c95ef1d6ec37dda7e4c4c96` |
| `gui-b/fortress-entry.png` | `ec00c440059520fa938332d77e314b6a8b73e28e9d63923f184fb44404e5930e` |
| `gui-b/fortress-menu-45.png` | `4008b84c61bae72b603678d4a64b54cc7f4370df65172ef34dcafac6bf516a39` |
| `gui-b/aim-ellipse-48.png` | `b94a19dce03d346363a337e8acd0e9952f57bec2782b85821e5e4149fe584374` |
| `gui-b/after-turn-1.png` | `4cf5a07c2b94065d145c53af2577d8a24e9b57c56baa70dea3f635ccb3ce8d46` |
| `gui-b/final.png` | `f8e0c02378593a314a28c079149e6caac4c15e574928e8c6f5803cc4d4ded3e5` |

上述`gui-b/`皆位於`workplace/round133/`。完整點擊、觀測、trace及圖片雜湊亦存來源索引。
本輪沒有公開SaveSlot／重讀抽樣，也沒有實體音訊驗收。

## 原版與規則

[研究446](../re/446-fortress-gun-scene.md)保存明示注入範圍、原始CS:IP與IDA交叉核對。
三案雙跑共828對RAM／色號／色盤／PNG相同，包含138完整命中幀、5空射幀及52動畫關閉幀。
命中10支、擊沉10637艘；動畫開關不改戰損、seed及分數；空射不開交戰面板。
IDA的778條指令經63個MZ relocation正規化後，全與原版RAM相同。
這是原版可重現性與資料鏈驗證，沒有做remake／原版逐像素比較。

`TestFortressGunSceneOracle`把原版池中148支有效分隊全部搬入正式規則，不用原版命中標記篩選。
射程、排序、戰損、修理上限、飛彈、艦載機、雙計分器與末端seed逐一相同。
測試綁定同次開火入口及返回RAM雜湊，原版收據缺失便skip。
驗證範圍是本次輸入域的規則層，不升格為畫面或正常整段同狀態。

標準`tools/gui.sh test ./cmd/... ./internal/...`完成19套件，15 PASS／4無測試。
本輪Go檔的`gofmt -l`無輸出，來源索引完整重核通過。

## 重跑入口

原版輸出拒絕覆寫。既有六份收據在`workplace/oracle/fortress48-r133-r3-{hit,empty,off}-{a,b}/`。
要重生時先保留舊收據並使用獨立工作副本。父狀態位於專用dosgolem的`workplace/fcmd/eng-t5.state`。

```sh
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_scene_oracle.sh hit a
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_scene_oracle.sh hit b
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_scene_oracle.sh empty a
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_scene_oracle.sh empty b
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_scene_oracle.sh off a
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_scene_oracle.sh off b
tools/gui.sh test ./internal/rules -run TestFortressGunSceneOracle -count=1 -v
tools/gui.sh test ./cmd/... ./internal/...
LOGH3_EXTRA_RO="$PWD" tools/py.sh tools/scripts/fortress_scene_receipt.py
```

GUI輸出同樣拒絕覆寫，重跑改用新的`--out`，原始命令為：

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_CPUS=2 LOGH3_TIMEOUT=10m \
tools/gui.sh xrun python3 tools/scripts/fortress_player_gui.py \
  --binary /src/workplace/round133/gin3 --out /src/workplace/round133/gui-b
```

## 未完成範圍

- 正式要塞砲的逐分隊文字、PFAM與目標中彈、下方方塊及結果等待尚未接線。
- 目前CPU政策下完整登陸／占領正常GUI、公開存讀及DOS整段同狀態仍待驗證。
- Windows／macOS真機、實體音訊不由本輪Linux收據升格。

#48保持OPEN。觸發方法、原版動態樣本及規則層回歸已完成，不再列為缺樣本。
