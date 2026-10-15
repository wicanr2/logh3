# 規格

第141輪：[148設定左緣補充修正](148-clear-settings-font.md)，保留陰影並補上可見金屬外緣；重新打包的版號決定及進度見[狀態表141](../goal/2026-10-15-settings-repackage-round-141.md)。

第140輪：[148系統設定專用清晰字型](148-clear-settings-font.md)，以原生小字與高解析遮罩修正縮小字模的斷筆及陰影重疊。[驗收129](../playtest/129-clear-settings-font.md)／[狀態表140](../goal/2026-10-15-settings-font-round-140.md)。

第139輪：[147緊湊系統設定與AI選擇](147-compact-system-settings.md)，使用者指定縮小文字、恢復顯示敵軍入口及加入現有AI難度。顯示敵軍預設關閉；三平台v.1.4.0候選重新打包，進度集中於[狀態表139](../goal/2026-10-15-release-and-promo-round-139.md)。

第136輪：[145戰術保存成功返回](145-battle-save-return.md)、[驗收125](../playtest/125-battle-save-return.md)與[研究449](../re/449-original-battle-save-load.md)。原版冷啟動可直接續戰已證實；正式remake只修保存成功返回，讀回重建及DOS整段同態保持未完成。[目前狀態表](../goal/2026-10-14-original-battle-load-round-136.md)

第135輪：[登陸分段118限定CONFORMED](118-landing-scene.md)與[驗收124](../playtest/124-engagement-integration.md)，459完整幀及目前CPU政策下S3正常登陸／反擊／占領、兩種顯示、音訊、F7與公開六表通過。[目前狀態表](../goal/2026-10-14-engagement-integration-round-135.md)保留DOS整回合同態及戰場續讀限制。

第134輪：[要塞砲呈現144限定CONFORMED](144-fortress-gun-presentation.md)與[驗收123](../playtest/123-fortress-presentation.md)，423完整幀、正常GUI、音訊、F7與六表存讀通過。[目前狀態表](../goal/2026-10-14-fortress-presentation-round-134.md)分開要塞完成與續讀／整段同態限制。

第133輪歷史：[計畫133](../goal/2026-10-14-fortress-gun-round-133.md)沿[028要塞砲](028-fortress-gun-fire.md)完成原版三案雙跑、規則回歸及正常伊謝爾倫開火；[研究446](../re/446-fortress-gun-scene.md)與[驗證122](../playtest/122-fortress-gun-trigger.md)保存證據。當輪尚未接的正式要塞砲演出，已由上方144與驗收123完成；完整056保持DRAFT。

第132輪成果：[118交戰分段](118-landing-scene.md)與[114音色選擇](114-midi-interface.md)，狀態見[計畫132](../goal/2026-10-14-landing-soundfont-round-132.md)及[驗證121](../playtest/121-landing-and-soundfont.md)。使用者已驗收現行音樂並關閉#69，人工檢視一致並關閉#16；原版欄名未知及未驗真機仍保留證據限制。

第131輪呈現及輸入修正：[141系統視窗](141-system-dialog-cursor.md)、[142啟動進度](142-startup-progress.md)、[143四字型與HD音色](143-font-and-hd-music.md)。機器結果限定於[驗證120](../playtest/120-system-dialog-polish.md)的remake範圍。

[117：正常玩家公開退出](117-player-public-quit.md)，沿 #90 補正式 GUI 接線。

`AGENTS.md` 的流程：

```
RE 證據 → DRAFT spec → 證據審查 → READY → implementation → 同狀態驗證 → CONFORMED
```

只有標 `READY` 的規格可以動 production path；標 `CONFORMED` 表示實作完成
而且有同狀態收據撐著。每一份規格的頂端都要寫現在是哪一個狀態。

`docs/internals/` 是逆向出來的事實，`docs/re/` 是怎麼查到的，
這裡是**要照著做什麼**——把事實翻成 remake 的行為與驗收條件。

| # | 規格 | 狀態 |
|---|---|---|
| [`001`](001-region-summary-panel.md) | 星域摘要面板（左下角） | CONFORMED |
| [`002`](002-capital-panel.md) | 右側面板（日期、錢、自己的首都） | CONFORMED |
| [`003`](003-budget-panel.md) | 預算計劃／徵稅計劃的數字 | CONFORMED |
| [`004`](004-production-panel.md) | 生產計劃的數字 | CONFORMED |
| [`005`](005-save-slot-loading.md) | 讀一格存檔 | CONFORMED |
| [`006`](006-unrest-at-load.md) | 載入時重算未安定度 | CONFORMED |
| [`007`](007-load-game-path.md) | 「取進度」那條玩家路徑 | CONFORMED |
| [`008`](008-save-slot-list.md) | 存檔畫面那 11 列 | CONFORMED |
| [`009`](009-save-slot-writing.md) | 存檔寫出 | CONFORMED |
| [`010`](010-save-game-screen.md) | 「存進度」畫面 | CONFORMED |
| [`011`](011-opening-report.md) | 疊在戰略底板上那幾張裡跟局面走的內容 | CONFORMED |
| [`012`](012-admiral-detail.md) | 提督詳細資料 | CONFORMED |
| [`013`](013-fleet-organisation.md) | 艦隊編成 | CONFORMED |
| [`015`](015-fleet-composition.md) | 艦隊組成 | CONFORMED |
| [`016`](016-versus-mode.md) | 對戰模式 | CONFORMED |
| [`017`](017-computer-ai.md) | 電腦 AI：原版行為與強化版 | 兩層都有了 |
| [`018`](018-battle-resolution.md) | 戰鬥結算（射程、攻擊名單、損失分攤）| 部分 `READY`（逐節標示）|
| [`019`](019-planet-assault.md) | 攻打行星 | READY |
| [`020`](020-intro-animation.md) | 開頭動畫 | CONFORMED（dosgolem 播放順序、尾段與時程抽樣對拍） |
| [`021`](021-battle-data-screen.md) | 戰術選單「資料」 | CONFORMED |
| [`022`](022-fleet-detail.md) | 艦隊詳細資料 | CONFORMED |
| [`023`](023-battle-command-menus.md) | 戰場的命令選單 | CONFORMED |
| [`024`](024-battle-move-order.md) | 戰場的「移動」命令（分隊層） | READY／已實作 |
| [`025`](025-battle-formation-order.md) | 戰場的「編隊」命令 | CONFORMED |
| [`026`](026-battle-supply-repair.md) | 戰場的「補給」與「修理」 | READY |
| [`027`](027-fleet-move-and-battle-order.md) | 艦隊選單的「移動」與「戰鬥」 | READY |
| [`028`](028-fortress-gun-fire.md) | 要塞砲開火 | READY |
| [`029`](029-strategic-ai-turn.md) | 原版的戰略 AI 回合 | READY |
| [`030`](030-blend-and-tax-rate.md) | `Blend` 與稅率重算 | CONFORMED |
| [`031`](031-fleet-refit.md) | AI 回合的艦隊整補 | CONFORMED |
| [`032`](032-region-posture.md) | 星域態勢重算（`+0x13` 與 `+0x24`） | CONFORMED |
| [`033`](033-region-ownership-chain.md) | 回合結束的星域歸屬鏈 | CONFORMED |
| [`034`](034-region-flip-report.md) | 星域易主的戰略報告 | CONFORMED |
| [`035`](035-month-end-gate.md) | 月底的閘門與每回合的重算 | READY |
| [`036`](036-ship-delivery.md) | 月底交付的船艦放到哪裡 | READY |
| [`037`](037-unit-price-from-staff.md) | 船艦單價由軍事專任參謀決定 | CONFORMED |
| [`038`](038-monthly-growth-and-tax-rate.md) | 月底重算成長率與稅率 | CONFORMED |
| [`039`](039-region-supply-pool.md) | 星域的飛彈／艦載機存量每回合回充 | DRAFT |
| [`040`](040-region-morale-drift.md) | 民心每回合往擁有者的目標值走一格 | CONFORMED |
| [`041`](041-rebellion.md) | 叛亂：判定、效果與接進回合的位置 | CONFORMED |
| [`042`](042-region-unrest-recompute.md) | 星域未安定度重算與呼叫點 | CONFORMED |
| [`043`](043-ai-capital-reach.md) | 戰略 AI 的首都可達表 | CONFORMED |
| [`044`](044-ending-image-position.md) | 400 列結局 VO5 的垂直位置 | CONFORMED |
| [`045`](045-ship-delivery-screen-details.md) | 月底船艦接收報告的底板細節 | CONFORMED |
| [`046`](046-battle-grid-residue.md) | 戰場小地圖下方的格表殘留 | READY |
| [`046`](046-rebellion-report.md) | 逐格叛亂的戰略報告 | CONFORMED |
| [`047`](047-occupation-report.md) | 單方面占據的戰略報告 | CONFORMED |
| [`048`](048-admiral-fate-report.md) | 失蹤提督生死判定與報告 | READY／已實作 |
| [`049`](049-phezzan-event-report.md) | 費沙事件雙頁報告 | CONFORMED |
| [`050`](050-victory-defeat-report.md) | 勝負報告與結局轉場 | CONFORMED |
| [`051`](051-strategic-fleet-command.md) | 戰略地圖上的艦隊命令（艦隊視窗 → 作戰 → 五項 → 確認、變更、副官、編組） | CONFORMED |
| [`052`](052-fleet-list-tabs.md) | 右下角清單的四個分頁與捲動 | CONFORMED（敵艦隊頁有列的局面用改記憶體造，`docs/re/325`） |
| [`053`](053-region-info-window.md) | 左鍵點星圖：星系情報視窗與換選星域 | CONFORMED |
| [`054`](054-fleet-arrival.md) | 艦隊走到之後：回合裡的一格、到達收尾、占據旗標與戰略打行星 | 走一格 CONFORMED／打行星 DRAFT |
| [`055`](055-list-screen-rows.md) | 一覽表四張的列 | CONFORMED（分頁 DRAFT） |
| [`056`](056-engagement-scene.md) | 交戰演出（打行星那一種）與戰鬥曲的出口 | DRAFT |
| [`057`](057-battle-view-scroll.md) | 戰場視野的滑鼠捲動（按住右鍵） | CONFORMED |
| [`058`](058-strategic-planet-assault.md) | 戰略打行星（占據那一格每回合打行星、反擊、士氣、換手） | CONFORMED（規則層）；訊息框 DRAFT |
| [`059`](059-turn-transition-windows.md) | 回合裡的過場視窗與報告框（出發前往／已經到達、被擋、經費不足、問話、處理中面板、攻占訊息框、士氣報告、已全軍覆沒） | CONFORMED；「已經恢復」DRAFT |
| [`060`](060-strategic-attrition-annihilation.md) | 戰略階段的耗損與艦隊全滅（士氣低落停在別人星域每回合扣兵、第 0 組 0 艘就全滅、提督失蹤） | CONFORMED（全滅）；「沒全滅」那一支 DRAFT |
| [`061`](061-passive-battle-computer.md) | 戰場上的電腦方不造成傷害（使用者裁決，刻意偏離原版；`originalBattleAI` 關掉） | CONFORMED（remake 規則） |
| [`062`](062-battle-turn-eight-steps.md) | 戰場一個回合的結算：原版的八步與第 8 步的五個子步驟（全滅、星域換手、索敵燃料、提督升級、結束判定） | DRAFT |
| [`063`](063-region-resupply.md) | 自家星域的艦艇補充（逐艦種補到滿編、修護艦重置、補完印「士氣值已經恢復」） | READY |
| [`064`](064-staff-appointment.md) | 參謀任命（情報收集與徵稅計劃的「參謀」鈕：清單、詳細資料、決定／中止、換內政參謀會重算稅收） | CONFORMED |
