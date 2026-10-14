# 要塞砲演出驗收

2026-10-14，第134輪。要塞砲已依[144限定CONFORMED](../spec/144-fortress-gun-presentation.md)正式接線。
原版證據與失敗分類見[研究447](../re/447-fortress-presentation.md)，唯一來源索引為[fortress-presentation.json](../internals/data/fortress-presentation.json)。

## 已通過

| 驗收 | 結果 |
|---|---|
| 原版固定重播 | 命中、空射、動畫關閉、伊謝爾倫及帝國目標／換側，1956檔雙跑相同 |
| 正式完整幀 | 423張640×480，129945600像素，全圖／八區／RGB零差，差分圖齊全 |
| 發射通知 | 352×100指定區域含原游標零差；正常GUI通知及射程橢圓等待通過 |
| 規則與逐擊資料 | 148有效分隊，命中10支／10637艘，副作用／雙計分器／seed不變；逐擊完整度及量表分母符合原版 |
| 正常玩家路徑 | 片頭→S1同盟→星域8撤退→伊謝爾倫行星選單→瞄準→實行→逐分隊演出→第2回合 |
| 兩種顯示 | 正常HD／普通速度及原版顯示／動畫關閉；F7在演出中往返不重設流程 |
| 正常戰損 | 兩支友軍分隊損失2678／2254艘，剩322／746艘；合計4932艘、245分 |
| 音訊 | 正式PulseAudio輸出，44100 Hz／16-bit／雙聲道、非靜音；cue前後及實際播放完成閘門通過 |
| 公開存讀 | 兩套GUI各由戰術保存、公開讀檔、再保存；六表全部逐位元組相同 |
| 標準測試 | 19套件，15 PASS／4無測試；新原版通知、清圖、單次點擊、空射、保存拒絕、音效及HD結果回歸通過 |

正常GUI沒有狀態、seed、腳本或被動CPU覆寫。試射朝最近同盟艦隊19，原版候選本就含雙方分隊，不把本次稱為命中敵軍。
正式執行檔為`workplace/round134/gin3-verified`，SHA-256 `4e30d368b1be3b5bb703340539bfe260764ae2b223ff73f95b08c0f6b8e4702d`。
正常圖、完整輸入、播放幀、cue、音訊與六表雜湊全部收錄來源索引。

## 收據位置與重跑

| 內容 | 本機路徑 |
|---|---|
| 最終HD／音訊／F7／存讀 | `workplace/round134/gui-hd-accepted/` |
| 最終原版顯示／動畫關閉／存讀 | `workplace/round134/gui-off-accepted/` |
| 兩支結果畫面 | 上述目錄的`unit-7-result.png`及`unit-8-result.png` |
| typed來源與完整比較 | `workplace/round134/{hit,off,iserlohn,empire}-typed-final.json`及`*-production-final/` |
| 新原版分支 | `workplace/oracle/fortress48-r134-r3-{iserlohn,empire}-{a,b}/` |
| 標準結果 | `workplace/round134/standard-final.log` |

工具均拒絕覆寫收據；重跑改用獨立輸出目錄或工作副本，保留既有來源。

```sh
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_presentation_oracle.sh iserlohn a
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_presentation_oracle.sh iserlohn b
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_presentation_oracle.sh empire a
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/fortress_presentation_oracle.sh empire b
tools/gui.sh run tools/scripts/fortress_frame_probe.go \
  workplace/round134/hit-typed-final.json workplace/round134/new-hit-comparison production
tools/gui.sh test ./cmd/... ./internal/...
LOGH3_EXTRA_RO="$PWD" tools/py.sh tools/scripts/fortress_presentation_receipt.py
```

[typed生成器](../../tools/scripts/fortress_frame_prepare.py)依呼叫事件、原始欄位及清圖順序建立輸入；原版像素只供比較。
正常GUI重跑範例：

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_CPUS=2 LOGH3_TIMEOUT=10m \
tools/gui.sh xrun python3 tools/scripts/fortress_presentation_gui.py \
  --binary /src/workplace/round134/gin3-verified \
  --out /src/workplace/round134/new-gui-hd --music /src/workplace/round132/music
```

原版顯示／動畫關閉改用`--graphics original --off`，不加`--music`。

## 完成範圍

同狀態比較限定完整typed renderer幀與通知指定區域，正常GUI為獨立remake驗收，沒有稱作原版／remake整場同狀態。
原版／HD結果截圖已目視確認分隊數字、下方方塊、完整文字安全區及回合續行。
X11擷取在既有結果等待內取穩定畫面，沒有改動畫時鐘或跳格。

六表保存通過，不等於同戰場續讀。額外續戰探針回星域8，未回伊謝爾倫；088原契約已明示這個限制。本輪保留問題，不宣稱續讀完成。
完整登陸／占領正常GUI、DOS整段同狀態、Windows／macOS真機與實體音訊人耳仍未新增驗收。
因此要塞砲子項完成，#48整張仍保留其他未完成項。本輪沒有新發行包。

來源索引4885557 bytes，SHA-256 `5180e1968e7a1579157f87c11d1ec11586639aa787367cd5c69309fda18bbf12`。
