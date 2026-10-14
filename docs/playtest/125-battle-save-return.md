# 125：戰術保存成功返回戰場

2026-10-14，來源父提交`e402794d`加本輪工作樹。[145](../spec/145-battle-save-return.md)的返回範圍；[原版研究449](../re/449-original-battle-save-load.md)與[來源索引](../internals/data/battle-save-return.json)區分原版續戰證據及正式remake實作。

索引57840 bytes，SHA-256 `63cd2991064314bbae236cbcd8df33fc3f39a3da1a51a9fc77ae7e2d0e6ec2fc`，綁194項本機輸出及11項現行來源。

## 原版證據

dosgolem專用副本`f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`，無狀態／seed注入，唯讀原版與獨立Scratch。C／D各37檔、E／F各45檔雙跑相同，共82對。公開保存第0格自動回原戰場第2回合；冷啟動取進度第0格直接續同戰場，再公開實行到第3回合。這是原版自己的可重現流程，不是跨引擎整段same-state。

```sh
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/battle_load_oracle.sh e
DOSGOLEM_CPUS=2 DOSGOLEM_MEM=3g bash tools/battle_load_oracle.sh f
```

工具拒絕覆寫既有目錄。歷史C／D觀察器由各輸出目錄的`observer-source.go`保存，SHA-256及完整輸入在唯一索引。

## 正式remake

正式片頭→取進度→138→S3→帝國→公開實行→星域26第1回合，保留目前CPU政策與自然seed。HD保存第0格後直接回第1回合，再公開實行到第2回合；F7切原版顯示，保存第1格後回第2回合，再實行到第3回合。兩模式各保存／返回／下一回合與保存前留圖，共8張完整960×720視窗，六表各6份落地。
公開輸入由[GUI工具](../../tools/scripts/battle_save_gui.py)送X11，實際游標及按下／放開由正式frame trace回授，無-script、局面、seed或被動CPU覆寫。

```sh
tools/gui.sh build -o workplace/round136/gin3 ./cmd/gin3
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 \
LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_CPUS=2 LOGH3_MEM=3g \
tools/gui.sh xrun python3 /src/tools/scripts/battle_save_gui.py \
  --binary /src/workplace/round136/gin3 --out /src/workplace/round136/gui-save-c
```

目標回歸驗證返回不改分隊、艦隊、行星、鏡頭、回合、命令、兩種分數及seed；保存失敗保持視窗。只移除本次成功返回分支的負向對照，明確以`StateSaveGame <nil>`失敗。該回歸使用明標診斷fixture，正常入口由上方GUI補足。
標準19產品套件15 PASS、4無測試。程式修正後沒有改動畫、傷害規則、六表格式或讀回重建。

```sh
tools/go.sh test ./internal/game -run TestBattleSaveSuccessReturnsWithoutAdvancing -v
tools/gui.sh test ./internal/... ./cmd/...
LOGH3_EXTRA_RO=$PWD tools/py.sh tools/scripts/battle_save_receipt.py
```

首次GUI視窗未移至螢幕內，座標被虛擬桌面邊界截斷。第二次流程通過，但下一回合截圖可能早於視窗呈現，正式取樣加等待後以相同產品重跑。這些是驗證器問題；沒有改產品時鐘或讓畫面規則配合截圖。

## 完成邊界

145僅保存成功返回已驗。remake同戰場讀回重建仍待實作，不能由六表存在或原版可續戰推成已完成。完整056／062與DOS整段同狀態保持原狀，#48保持OPEN。既有118／144成果不重開。
本輪GUI只驗保存與返回，音訊沿既有124限定結果；ALSA空輸出不列作新音訊驗收。沒有新發行包、tag或Release。
