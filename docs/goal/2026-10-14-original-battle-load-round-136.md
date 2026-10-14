# 第136輪：原版戰術保存與續讀語意

日期：2026-10-14。起點`e402794d`，工作樹乾淨，已推送main。
權威清單為[GitHub #48](https://github.com/wicanr2/logh3/issues/48)。[118](../spec/118-landing-scene.md)及[144](../spec/144-fortress-gun-presentation.md)的限定完成保持，正常整合證據見[驗收124](../playtest/124-engagement-integration.md)。

| 項目 | 開工狀態 | 本輪最小問題 |
|---|---|---|
| remake演出 | 正常S3兩種顯示各24段、293格完整 | 不重開已完成項 |
| 戰術保存續讀 | 原版oracle未知，remake目前回戰略 | 原版保存後返回哪裡，讀回是否沿WORK+02／03／06續戰 |
| DOS整段同態 | 未完成 | 先建立可比的原版控制邊界，不猜補資料 |

本輪只先取原版契約證據。使用專用dosgolem `f7523f9`、原版唯讀目錄及獨立Scratch寫入層，記錄完整公開輸入、檔案讀寫、WORK、回合／星域與畫面。
原版自然S3父狀態續行，不注入位置、艘數、歸屬或seed；保存和讀回透過原始程式的公開選單。
目前兩份較早的正常首戰state已不存在，不能用舊文件當作本輪已驗；專用副本的S3載入state仍在，須核來源後重生。
依[005保存資料](../spec/005-save-slot-loading.md)、[007載入入口](../spec/007-load-game-path.md)與[088控制範圍](../spec/088-normal-battle-player-entry.md)判定缺口。
發現差異後先補證據、規格與READY審查，再改正式玩家路徑；本輪不先改保存格式或LoadSlot。
調查過程追加[WORKLOG](../../WORKLOG.md)，現況及證據入口同步[CONTEXT](../../CONTEXT.md)。

原版工具：[公開存讀入口](../../tools/battle_load_oracle.sh)、[容器執行](../../tools/battle_load_native_run.sh)、[原版觀察器](../../tools/scripts/battle_load_native.go)。

窄實作契約：[145戰術保存返回](../spec/145-battle-save-return.md)。

正常玩家驗收：[戰術保存X11](../../tools/scripts/battle_save_gui.py)，涵蓋HD／原版顯示的保存成功返回與下一回合。

唯一來源索引：[battle-save-return.json](../internals/data/battle-save-return.json)，由[核對工具](../../tools/scripts/battle_save_receipt.py)產生與重核。

## 本輪結果

145經DRAFT、原版C／D審查READY、實作及正常GUI後限定CONFORMED。保存成功自動回原戰場，失敗不離開；兩種顯示各公開保存並續下一回合，8圖與12份六表落地，狀態及seed不變，修正前負向對照明確失敗。標準19套件15 PASS、4無測試。
原版C／D37對、E／F45對全部相同；S3帝國冷啟動直接續星域26第2回合，並可實行至第3回合。鏡頭與圖示快取實際差分、RNG不在WORK內均保存於[研究449](../re/449-original-battle-save-load.md)。

| 項目 | 收尾狀態 | 下一閘門 |
|---|---|---|
| 戰術保存成功返回 | remake已完成，145限定CONFORMED | 不重開 |
| 原版戰術冷啟動續讀 | S3帝國已證實 | 同盟／雙人與鏡頭／快取／計分續接 |
| 正式戰場讀回重建 | 未完成 | 證據與READY後實作，不重置保存回合／部署 |
| DOS整段演出同態 | 未完成 | 保留#48，不由本輪原版重播或GUI代替 |
