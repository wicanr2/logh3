# 第134輪：完成要塞砲演出驗收

使用者要求完成要塞演出驗收，並沿用提交、推送與README更新要求。
目前權威為#48、現行程式及[原版觸發122](../playtest/122-fortress-gun-trigger.md)。

| 項目 | 開工狀態 | 本輪完成閘門 |
|---|---|---|
| 要塞規則／正常開火 | 已通過 | 保持命中、戰損、分數及seed |
| 原版逐分隊樣本 | 五類1956檔雙跑相同 | 423完整幀全圖／八區／RGB零差，完成 |
| 正式要塞演出 | [144限定CONFORMED](../spec/144-fortress-gun-presentation.md) | 先DRAFT核證、再READY接線及正式驗收，完成 |
| 正常操作／保存 | 正常S1伊謝爾倫、HD／原版顯示、動畫關閉、音訊、F7及六表公開存讀通過 | 要塞子項完成；同戰場續讀仍未完成 |

參考入口為[COMANI](../internals/20-comani-bmg.md#要塞砲演出)、[研究446](../re/446-fortress-gun-scene.md)與[來源索引](../internals/data/fortress-scene.json)。流水帳追加唯一[WORKLOG](../../WORKLOG.md)。
完整登陸／占領或DOS整場同態不由本輪要塞演出抽樣自動升格；#48是否可關依所有剩餘判準核對。

本輪探針：[原版分支入口](../../tools/fortress_presentation_oracle.sh)、[原版分支程式](../../tools/scripts/fortress_presentation_native.go)、[typed輸入生成](../../tools/scripts/fortress_frame_prepare.py)、[DRAFT完整幀診斷](../../tools/scripts/fortress_frame_probe.go)。原版分支明示注入，與正常玩家驗收分開。

最終成果：[研究447](../re/447-fortress-presentation.md)、[驗收123](../playtest/123-fortress-presentation.md)、[凍結索引](../internals/data/fortress-presentation.json)、[核對器](../../tools/scripts/fortress_presentation_receipt.py)與[正常GUI驅動器](../../tools/scripts/fortress_presentation_gui.py)。標準19套件15PASS／4無測試；#48保留完整登陸正常鏈、續讀與DOS整段同態，不由要塞子項代替。
