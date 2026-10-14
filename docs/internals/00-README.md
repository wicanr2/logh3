# 原版內部構造參考

[交戰整合驗收124](../playtest/124-engagement-integration.md)、[研究448](../re/448-engagement-integration.md)及[凍結來源](data/engagement-integration.json)保存118限定CONFORMED、459完整幀、目前CPU政策的正常S3登陸占領與公開六表。[核對器](../../tools/scripts/engagement_integration_receipt.py)檢查原版792雙跑檔及兩種顯示各24段／293格。原版整回合同態及戰場續讀仍未升格。

[要塞演出驗收123](../playtest/123-fortress-presentation.md)、[研究447](../re/447-fortress-presentation.md)及[凍結來源](data/fortress-presentation.json)保存144正式逐分隊演出、423完整幀、正常GUI與六表存讀。[核對器](../../tools/scripts/fortress_presentation_receipt.py)重核原版1956檔及所有來源；戰場續讀與整段DOS同態未升格。

[要塞砲正常觸發122](../playtest/122-fortress-gun-trigger.md)、[研究446](../re/446-fortress-gun-scene.md)及[唯一來源索引](data/fortress-scene.json)保存原版三案雙跑、IDA778指令、逐分隊圖序列及正常S1伊謝爾倫開火；[核對器](../../tools/scripts/fortress_scene_receipt.py)可重核。COMANI已證實事實見[20](20-comani-bmg.md#要塞砲演出)，正式要塞砲已由上方123接續驗收。

[第132輪音色與登陸分段](../playtest/121-landing-and-soundfont.md)及[來源索引](data/landing-soundfont.json)區分正式音色驗收、原版完整幀診斷與尚未完成的#48整合。音色入口見[17](17-mds.md)，核對器為[landing_soundfont_receipt.py](../../tools/scripts/landing_soundfont_receipt.py)。

[介面操作第128輪](../goal/2026-10-12-launch-hover-round-128.md)、[功能列與反光138](../spec/138-top-toolbar-and-hover.md)及[正常介面試樣器](../../tools/scripts/interface_hover_preview.go)保存本次啟動確認、F1／F7／F8配置及調暗試樣。

[原配置HD交付111](../playtest/111-production-hd.md)及[唯一來源索引](data/production-hd.json)保存正式A介面、地圖、動畫、正常S1實錄與三平台交付；[最終核對器](../../tools/scripts/hd_finalize.py)逐檔驗證凍結來源、封包與收據。

[HD共用艦型匯入器](../../tools/scripts/hd_ship_import.py)只複製生成原PNG，保存2×2圖集與原D0像素雜湊；正式載入器取格，不以影像腳本修圖。

[HD動畫批次診斷](../../tools/scripts/hd_animation_check.go)枚舉實際COMANI段、解覆寫／XOR與片頭播放序列，原色重播及HD抽樣分開記錄。

[動畫艦殼與光效133](../spec/133-hd-animation-surfaces.md)共用原解碼幀，保留片頭／交火幀序、鏡頭及事件；完整快照不能抹掉外側HD文字。

[策略面材與高解析符號132](../spec/132-hd-strategy-and-symbols.md)保留32遮罩、當幀色號、75張真實SPRITE索引、情報與游標裁切；策略前景以原PNG定位重畫，不能驅動規則或點選。

[正式HD第125輪](../goal/2026-10-11-production-hd-round-125.md)沿已確認的A原配置接正式遊戲；[繪製紀錄規格130](../spec/130-hd-drawing-journal.md)定義原色號保持、高解析文字、面框及疊窗順序。

[HD字型圖集產生器](../../tools/scripts/hd_font_atlas.py)在既有HD工具映像排版Noto Serif CJK TC；輸出保存來源字型、授權及生成器雜湊，缺字沿原字模回退。

[HD地圖材質131](../spec/131-hd-map-materials.md)沿原格號、世界座標、星域遮罩及小地圖bitmap；共享生成材質不取代地形或規則資料。

[32張HD地圖診斷](../../tools/scripts/hd_map_check.go)從正式Battle消費端檢查64個鏡頭、53實際圖塊及天體四格；正常玩家錄影另外驗收。

[HD介面與S1影片第123輪](../goal/2026-10-11-hd-interface-and-s1-promo-round-123.md)保存本輪狀態；[三劇本目錄129](../spec/129-bundled-scenario-catalog.md)、[正常X11實錄](../../tools/scripts/s1_gameplay_record.py)與[有界錄影顯示](../../tools/s1_record_display.sh)為重現入口。

[下一階段HD提案](../goal/2026-10-10-v1-and-hd-plan-round-122.md)列實際素材範圍、畫布方向與分階段驗收，未確認部分不進正式程式。

[v1引擎封包器](../../tools/scripts/v1_release.py)、[啟動器抽驗](../../tools/scripts/v1_release_smoke.py)與[推廣片製作](../../tools/scripts/v1_promo.py)保存正式交付的重現入口；[來源索引](data/v1-release.json)集中輸入、輸出及驗收身份。

[HD畫布試樣產生器](../../tools/scripts/v1_hd_preview.py)只產生可丟棄HTML；[集中交付核對器](../../tools/scripts/v1_finalize.py)核六份ZIP、正常視窗、影音及來源，產生最終SHA清單。

[封包美術說明更正](../../tools/scripts/v1_package_notes.py)保留本次尚未發布的舊ZIP，從相同程式及素材重封，逐檔驗證新版說明、manifest與平台權限。

[按鈕修正128](../spec/128-visible-button-activation.md)與[正式交付109](../playtest/109-v1-release.md)保存第二畫面確認、月初報告接線及v.1.0.0封包驗收。

[原素材HD取樣](../../tools/scripts/hd_refinement_probe.go)驗各類首／中／尾PNG的邏輯像素中心與色盤，與正常玩家路徑收據分開標示。

[完整人物核對器](../../tools/scripts/hd_complete_check.go)以實際FACE.AMG驗82個人物及第82號單色空格，逐項核對生成PNG、原圖與提示詞。

[HD邊緣細化](../../internal/presentation/refinement.go)與[中心／細線／動態色盤回歸](../../internal/presentation/refinement_test.go)保存123的固定取樣契約。

[全畫面HD規格123](../spec/123-hd-complete-presentation.md)區分逐張人物修復與地圖、艦艇、動畫的呈現層邊緣細化，沿原索引與動態色盤。

HD 素材匯入使用[原始輸出匯入器](../../tools/scripts/hd_art_import.py)保存 image_gen PNG 與提示詞；[人物核對表](../../tools/scripts/hd_art_review.go)只產生原圖與 HD 並排的診斷圖，不修改素材。

HD 美術沿使用者核准的原畫風試樣：[規格122](../spec/122-hd-artwork.md)定義可選高解析度圖片、原索引、遮擋及輸入契約。定版圖保存在 `assets/hd/face/face-00.png`；完整提示詞及來源見 WORKLOG 的 HD 試樣記錄。這是美術修復方向，不宣稱原版像素對拍。

[HD素材入口](../../assets/hd/README.md)、[來源與雜湊](../../assets/hd/MANIFEST.json)及[正常操作與試玩108](../playtest/108-hd-artwork.md)保存實際覆蓋與交付。HD驗收沿[正式操作模型](../../tools/scripts/hd_art_model.go)、[真實X11輸入](../../tools/scripts/hd_art_gui.py)與[全視窗擷取](../../tools/scripts/hd_art_capture.go)；歷史原畫面收據保持。

[HD交付核對器](../../tools/scripts/hd_art_finalize.py)集中封包、56張真實視窗、來源與前版保持；[HD來源索引](data/hd-artwork.json)保存首批限定完成與整套美術未完成的區別。

[第121輪HD計畫](../goal/2026-10-06-hd-artwork-round-121.md)沿#97記錄第一批與下一批界線。

[HD合成器](../../internal/presentation/artwork.go)沿[原圖繪製](../../internal/asset/draw.go)與[可見區域](../../internal/gfx/artwork.go)保留原索引及遮擋；[合成回歸](../../internal/presentation/artwork_test.go)和[滑鼠映射回歸](../../cmd/gin3/artwork_test.go)釘裁切、同色覆蓋與邏輯座標。

第120輪：[規格121](../spec/121-player-input-hotspots.md)修正會計／訂造／建造畫面熱區與可見按鈕相差40像素、三種下拉選單游標停住。[研究441](../re/441-player-input-hotspot-space.md)保留歷史勘誤，[正常操作107](../playtest/107-player-input-hotspots.md)與[來源索引](data/player-input-hotspots.json)集中驗收與新封包。
[本輪計畫](../goal/2026-10-06-player-input-round-120.md)沿使用者實際阻塞點，不重新開啟已完成的戰場對拍。
[正式Game預期畫面](../../tools/scripts/player_input_model.go)、[真實X11輸入](../../tools/scripts/player_input_gui.py)、[視窗擷取與差分](../../tools/scripts/player_input_capture.go)沿正常入口核全圖；[大視窗容器入口](../../tools/player_input_display.sh)讓2倍畫面完整留在虛擬螢幕內。
[交付身份核對器](../../tools/scripts/player_input_receipt.py)核三平台新包、48張正式包全圖及歷史提交中的凍結來源，不覆寫前兩版。

第119輪：[macOS封包器](../../tools/scripts/local_test_macos_package.py)核對通用Mach-O、Apple Silicon程式頁簽章、資產及ZIP權限。
[交付核對與身份保存器](../../tools/scripts/macos_promo_finalize.py)集中三平台、72秒影片與抽驗，保存[本輪來源](data/macos-retro-promo.json)，前版凍結來源保持。
[影片素材與分鏡](../../tools/scripts/retro_promo_prepare.py)沿既有真實收據，不新增戰場差異驗證。
[本作影片主題](../../tools/scripts/retro_promo_theme.sh)保存來源色彩；[素材與合成](../../tools/scripts/retro_promo_render.sh)分開產生版面及72秒影片。

第118輪沿 [#92](https://github.com/wicanr2/logh3/issues/92)：[本機完整版計畫](../goal/2026-10-06-local-test-package-round-118.md)。
[封包建置器](../../tools/scripts/local_test_package.py)使用既有GUI Docker，產生完整Windows／Linux測試包、授權、來源與逐檔雜湊。
[封包抽驗](../../tools/scripts/local_test_smoke.py)沿公開CLI輸入核實實際封包、兩種戰略AI、存讀並保存推廣片來源畫面。
[120](../spec/120-active-tactical-ai.md)依使用者明示裁決啟用既有戰術AI，取代正常單人被動政策；歷史診斷仍可明示選用。
[正常戰術抽驗](../../tools/scripts/active_tactical_player.go)由公開S3一路實行，保存兩模式CPU下令、位移、傷害與六表，不注入局面或乱數。
[活動政策回歸](../../internal/game/activebattle_test.go)核對253艘反擊保持與旗艦戰鬥命令接線。
[影片合成器](../../tools/scripts/local_test_promo.sh)沿既有FFmpeg／中文字型映像，合成60秒實際遊戲畫面靜態展示及真實遊戲配樂，保存抽樣幀與音畫檢查。
[Windows封包抽驗](../../tools/scripts/local_test_windows_smoke.py)解壓實際ZIP，核對版本、全部檔案與Wine正常入口；不冒稱實機。
[交付核對器](../../tools/scripts/local_test_finalize.py)補完整Go授權、核目前來源與實際ZIP，集中影片、抽驗及最終SHA-256。
[440](../re/440-normal-first-turn-and-active-ai.md)、[105](../playtest/105-local-full-test-package.md)及[唯一來源索引](data/local-test-package.json)保存既有117原版首回合及118交付。
[身份保存器](../../tools/scripts/local_test_receipt.py)核最終交付SHA-256與既有原版49檔，保存實作、限制與工具來源，不重跑原版。
交付入口限本機 `dist-all/v.0.1.1-20261006/`；雙玩家命令與戰場差異停止新增驗證，既有證據保持。

第109輪沿 [#48](https://github.com/wicanr2/logh3/issues/48)：[登陸完整單幀計畫](../goal/2026-10-06-landing-frame-round-109.md)。
[IDA caller匯出器](../../tools/scripts/landing_frame_ida.py)只開新暫存資料庫，核文字與行星下方的原始位址及指令。
[typed輸入探針](../../tools/scripts/landing_frame_prepare.py)與[單幀合成探針](../../tools/scripts/landing_frame_probe.go)沿118 DRAFT核全圖，不接正式Game。
[分支產生器](../../tools/scripts/landing_frame_branch_prepare.py)、[觀察器](../../tools/scripts/landing_frame_native.go)、[原版雙跑](../../tools/landing_frame_branch_oracle.sh)與[容器入口](../../tools/landing_frame_branch_native_run.sh)補先反擊、帝國及雙名冊的明示診斷。
[r2觀察器](../../tools/scripts/landing_frame_native_r2.go)補第二攻方的行星目標；r1只消費第一支的收據保留。
[研究437](../re/437-landing-frame.md)、[驗證097](../playtest/097-landing-frame.md)與[完整來源](data/landing-frame.json)保存十八張全圖及新分支；118保持DRAFT。
[凍結核對器](../../tools/scripts/landing_frame_receipt.py)重核原版雙跑、十八張完整圖、色盤、原始指令及歷史118身份。

第108輪沿 [#48](https://github.com/wicanr2/logh3/issues/48)：[登陸攻擊計畫](../goal/2026-10-05-landing-scene-round-108.md)。
[原版入口](../../tools/landing_scene_oracle.sh)、[容器內入口](../../tools/landing_scene_native_run.sh)與[完整流程觀察器](../../tools/scripts/landing_scene_native.go)使用專用dosgolem，明示注入，拒絕覆寫舊收據。
[IDA 有界匯出器](../../tools/scripts/landing_scene_ida.py)只開新暫存資料庫，保留原始位址與指令，語意另審。
[研究436](../re/436-landing-scene.md)與[118草稿](../spec/118-landing-scene.md)區分登陸、反擊、占領結果，不把注入診斷當正常玩家驗收。
[驗證096](../playtest/096-landing-scene.md)及[核對器](../../tools/scripts/landing_scene_receipt.py)核五案原版雙跑、每格畫面與[完整來源](data/landing-scene.json)。

第107輪沿 [#79](https://github.com/wicanr2/logh3/issues/79)：[正常占領玩家計畫](../goal/2026-10-05-battle-occupation-player-round-107.md)。
[出貨劇本操作探針](../../tools/scripts/battle_occupation_player_probe.py)只產生測試overlay，固定首戰種子一次，不注入局面或改正式規則。
[獨立重播與正常GUI入口](../../tools/scripts/battle_occupation_player_replay.py)沿正式bootstrap、60 Hz與公開存讀，觀測層不改局面。
[驗證095](../playtest/095-battle-occupation-player.md)限定第⑥步攻打行星與8-2星域換手，不補110勝後占領。
[本輪核對器](../../tools/scripts/battle_occupation_player_receipt.py)驗正常公開存讀、全圖與完整身份，產生[來源索引](data/battle-occupation-player.json)。

第106輪沿 [#69](https://github.com/wicanr2/logh3/issues/69)：[Windows 正常存讀計畫](../goal/2026-10-05-windows-save-load-round-106.md)。
[唯讀 Windows 入口](../../tools/windows_save_load.sh)及[正常輸入／Linux 對照器](../../tools/scripts/windows_save_load.py)驗公開保存、退出及新程序讀回，不注入局面。
[117退出接線規格](../spec/117-player-public-quit.md)沿研究341原版契約；[正式 app 回歸](../../cmd/gin3/quit_test.go)覆蓋取消及確認。
[正常保存核對器](../../tools/scripts/windows_save_load_receipt.py)凍結完整來源與歷史身份，產生[存讀來源索引](data/windows-save-load.json)；[驗證094](../playtest/094-windows-save-load.md)記錄本輪範圍。

第100輪沿[#56](https://github.com/wicanr2/logh3/issues/56)：[戰後清單與接線計畫](../goal/2026-10-05-strategic-guard-lifecycle-round-100.md)。
[研究434](../re/434-strategic-guard-lifecycle.md)核正常戰後續行與下一星域載圖的清單保留；[115限定CONFORMED](../spec/115-strategic-assault-guard.md)定義正式原池查詢與錯誤停止。
[原池探針](../../tools/strategic_guard_raw_probe.sh)與[比較器](../../tools/scripts/strategic_guard_raw_probe.go)在容器副本核共用1512寫回；不直接改正式Game。
[驗證088](../playtest/088-strategic-guard.md)限定正式規則／Game五案原池與保存；[完整現行來源](data/strategic-guard.json)由[核對器](../../tools/scripts/strategic_guard_review.py)保存，舊索引按歷史提交核實。
[清單觀察器產生器](../../tools/scripts/strategic_guard_lifecycle_prepare.py)續跑固定8-1保存，記錄退場、下一星域載圖與筆數寫端；不注入CS/IP。
[原版雙跑入口](../../tools/strategic_guard_lifecycle_oracle.sh)與[容器內入口](../../tools/strategic_guard_lifecycle_native_run.sh)保留全部局面注入與工具版本。

第99輪沿[#56](https://github.com/wicanr2/logh3/issues/56)：[戰略攻擊邊界計畫](../goal/2026-10-05-strategic-assault-edges-round-99.md)。
[戰略入口探針產生器](../../tools/scripts/strategic_assault_edge_prepare.py)沿固定完整原版保存，記錄明示注入及查詢結果。
[原版雙跑入口](../../tools/strategic_assault_edge_oracle.sh)與[容器內入口](../../tools/strategic_assault_edge_native_run.sh)只作#56窄幅診斷，不替代正常玩家路徑。
[實際地圖掃描器](../../tools/scripts/strategic_assault_terrain_scan.go)列出行星周圍的不可通行候選；結果是remake地形線索，另須原版返回值確認。
[正式規則比較器](../../tools/scripts/strategic_assault_edge_model.go)直接呼叫StrategicAssault，照實保存有／無地形的typed差異；執行成功不代表差異為零。
[守軍修法探針](../../tools/strategic_assault_edge_probe.sh)在容器/tmp副本加入已證實查詢，原版曲線與正式/src/internal保持；只用來審查DRAFT。
[研究433](../re/433-strategic-assault-edges.md)及[115草稿](../spec/115-strategic-assault-guard.md)記錄戰略漏算守軍與正式來源缺口，凍結資料在[完整來源](data/strategic-assault-edges.json)。
[來源核對器](../../tools/scripts/strategic_assault_edge_review.py)分別核正式反例及副本零差，不將探針成功混作正式修正完成。

第98輪沿[#66](https://github.com/wicanr2/logh3/issues/66)：[設定子選單計畫](../goal/2026-10-05-settings-interface-round-98.md)。
[IDA選單匯出器](../../tools/scripts/settings_interface_ida.py)讀取當輪唯讀原生記憶體複本，保留CS:IP、檔案位移、IDA EA與原始指令；未審查語意標unknown。
[MIDI介面操作探針](../../tools/scripts/midi_interface_prototype.html)僅供取捨，未進正式路徑。
[研究432](../re/432-settings-interface.md)核MIDI控制流並勘誤BUG!出貨可達性；[規格113](../spec/113-settings-bug-unavailable.md)限定正常不可選熱區。
[正常設定回歸](../../internal/bootstrap/settings_interface_test.go)使用正式bootstrap及公開輸入，並核全圖及保存不變。
[正常設定GUI](../../tools/scripts/settings_interface_gui.py)以正式cmd/gin3及X11真實鍵鼠核對兩張全圖。
[設定全圖比較器](../../tools/scripts/settings_interface_compare.go)從同一規格取得矩形，保存原圖、remake、差分及逐區統計。
[驗證087](../playtest/087-settings-interface.md)限定出貨BUG!正常路徑；[MIDI草稿114](../spec/114-midi-interface.md)保持決策閘門。
[設定收據核對器](../../tools/scripts/settings_interface_review.py)生成並核對[完整來源](data/settings-interface.json)，不改寫舊凍結索引。


第91輪沿#79：[原池回寫計畫](../goal/2026-10-05-battle-finish-raw-round-91.md)、
[逐指令寫入觀察器產生器](../../tools/scripts/battle_finish_raw_prepare.py)、
[固定原版雙跑](../../tools/battle_finish_raw_oracle.sh)與
[隔離容器入口](../../tools/battle_finish_raw_native_run.sh)。
沿110 DRAFT核定位目標與戰損副作用，不改正式程式或既有凍結收據。
[完整原池暫存模型](../../tools/scripts/battle_finish_raw_model.go)、
[固定比較器產生器](../../tools/scripts/battle_finish_raw_model_prepare.py)與
[容器模型入口](../../tools/battle_finish_raw_probe.sh)比較完整六表及清單。
[原池來源核對器](../../tools/scripts/battle_finish_raw_review.py)生成
[完整原池來源](data/battle-finish-raw.json)，證據入口為
[研究425](../re/425-battle-finish-raw.md)與[驗證080](../playtest/080-battle-finish-raw.md)。
[占領報告暫停點產生器](../../tools/scripts/battle_finish_cursor_prepare.py)、
[固定callback雙跑](../../tools/battle_finish_cursor_oracle.sh)與
[callback容器入口](../../tools/battle_finish_cursor_native_run.sh)核對報告前後的資料提交順序。
[callback核對器](../../tools/scripts/battle_finish_cursor_review.py)產生
[續接來源與原指令](data/battle-finish-cursor.json)，由110規格及研究425回查。

第90輪沿#79：[占領傷害與守軍計畫](../goal/2026-10-05-battle-finish-attack-round-90.md)，
[IDA輸入產生器](../../tools/scripts/battle_finish_attack_ida_prepare.py)、
[容器暫存模型](../../tools/battle_finish_attack_probe.sh)與
[獨立比較探針](../../tools/scripts/battle_finish_attack_probe_test.go)。
這些是110 DRAFT診斷，沒有修改正式rules或Game。
[敗方守軍探針](../../tools/scripts/battle_finish_guard_prepare.py)、
[原版雙跑](../../tools/battle_finish_guard_oracle.sh)與
[容器入口](../../tools/battle_finish_guard_native_run.sh)核對查詢筆數及守軍艘數。
[核心證據核對器](../../tools/scripts/battle_finish_attack_review.py)產生
[占領核心與守軍來源](data/battle-finish-attack.json)；入口為
[研究424](../re/424-battle-finish-attack.md)及[驗證079](../playtest/079-battle-finish-attack.md)。

第89輪沿#79：[結束判定與勝方占領計畫](../goal/2026-10-05-battle-finish-round-89.md)，
[固定IDA輸入產生器](../../tools/scripts/battle_finish_ida_prepare.py)。
[原版觀測器產生器](../../tools/scripts/battle_finish_native_prepare.py)、
[雙跑入口](../../tools/battle_finish_oracle.sh)與[容器內入口](../../tools/battle_finish_native_run.sh)
固定8-5注入與完整返回，診斷收據不代替正常玩家勝負。
[獨立核對器](../../tools/scripts/battle_finish_review.py)保存
[結束判定事實與來源](data/battle-finish.json)，入口為
[研究423](../re/423-battle-finish.md)、[草稿110](../spec/110-battle-finish.md)及[驗證078](../playtest/078-battle-finish.md)。

第87輪沿#79：[升級資料與保存計畫](../goal/2026-10-05-battle-promotion-save-round-87.md)。
[路徑快取診斷產生器](../../tools/scripts/battle_promotion_cache_prepare.py)、
[原版雙跑](../../tools/battle_promotion_cache_oracle.sh)及[容器入口](../../tools/battle_promotion_cache_native_run.sh)區分星域換手與路徑重建時點。
[快取凍結器](../../tools/scripts/battle_promotion_cache_receipt.py)產生[快取來源](data/battle-promotion-cache.json)，
[研究421](../re/421-battle-promotion-save.md)記錄快取時點與現行保存覆寫邊界。

第86輪沿#79：[戰場升級全局回寫計畫](../goal/2026-10-04-battle-promotion-global-round-86.md)。
108保持DRAFT，全局完整表已核；剩typed窄回寫契約與正式玩家接線。
[全局組合探針](../../tools/scripts/battle_promotion_global_probe.go)以第85輪已核的提督／艦隊資料與明示父RAM輸入驗算。
[修正觀測器產生器](../../tools/scripts/battle_promotion_global_prepare.py)、
[雙跑入口](../../tools/battle_promotion_global_oracle.sh)與[容器內入口](../../tools/battle_promotion_global_native_run.sh)修正STELLAR基底並補完整行星及路徑輸入。
[全局來源凍結器](../../tools/scripts/battle_promotion_global_receipt.py)產生[完整全局收據](data/battle-promotion-global.json)。
[研究420](../re/420-battle-promotion-global.md)及[驗證075](../playtest/075-battle-promotion-global.md)保存53次全局回寫與STELLAR基底勘誤。

第85輪沿#79：[戰場升級重算與接線計畫](../goal/2026-10-04-battle-promotion-wiring-round-85.md)。
沿108草稿補297B:05B2／0E8B原版完整資料鏈，達READY後才接正式升級。
[原版重算觀測器產生器](../../tools/scripts/battle_promotion_recalc_prepare.py)、
[容器內入口](../../tools/battle_promotion_recalc_native_run.sh)與
[雙跑入口](../../tools/battle_promotion_recalc_oracle.sh)保存66案及逐次五表快照。
[RAM切段](../../tools/scripts/battle_promotion_recalc_ida_prepare.py)與
[原始IDA查詢](../../tools/battle_promotion_recalc_ida.sh)固定六支helper邊界。
[獨立重算模型](../../tools/scripts/battle_promotion_recalc_receipt.py)與
[第一版來源收據](data/battle-promotion-recalc.json)保留三表、亂數及費用核對；全局費用以第二版取代。
[分級索引產生器](../../tools/scripts/battle_promotion_recalc_annotations.py)、
[定位索引](data/battle-promotion-recalc-annotations.json)、
[自動合併匯出器](../../tools/scripts/battle_promotion_recalc_review_ida.py)與
[分級查詢入口](../../tools/battle_promotion_recalc_review_ida.sh)保留原始bytes及unknown警示。
[分級來源核對器](../../tools/scripts/battle_promotion_recalc_review_receipt.py)、
[分級收據](data/battle-promotion-recalc-review.json)、
[研究419](../re/419-battle-promotion-recalc.md)與
[對拍074](../playtest/074-battle-promotion-recalc.md)回查53次完整重算。
全局稅收／星域及typed回寫、正式UI／保存仍缺，108 DRAFT。
[第二版獨立模型](../../tools/scripts/battle_promotion_recalc_receipt_v2.py)與
[第二版來源](data/battle-promotion-recalc-v2.json)補核全局移動費及其67條既有原指令。
[第二版語意產生器](../../tools/scripts/battle_promotion_recalc_annotations_v2.py)、
[第二版定位索引](data/battle-promotion-recalc-annotations-v2.json)、
[第二版IDA匯出器](../../tools/scripts/battle_promotion_recalc_review_ida_v2.py)與
[第二版查詢入口](../../tools/battle_promotion_recalc_review_ida_v2.sh)取代第一版漏追callee的語意。
[第二版審查凍結器](../../tools/scripts/battle_promotion_recalc_review_receipt_v2.py)及
[第二版分級收據](data/battle-promotion-recalc-review-v2.json)保留808條原指令與42項proven。

第84輪沿#79：[戰場升級計畫](../goal/2026-10-04-battle-promotion-round-84.md)、
[108草稿](../spec/108-battle-promotion.md)與
[IDA輸入產生器](../../tools/scripts/battle_promotion_prepare.py)。
本輪先核8-4成長、艦隊重算及通知順序；證據未足前不改正式路徑。
[原指令匯出](../../tools/scripts/battle_promotion_ida.py)、
[完整保存觀測器產生器](../../tools/scripts/battle_promotion_native_prepare.py)、
[隔離模組](../../tools/battle_promotion_native_run.sh)與
[固定版本雙跑](../../tools/battle_promotion_oracle.sh)為重生入口。
[成長獨立驗算與凍結器](../../tools/scripts/battle_promotion_receipt.py)
核完整提督表，保留原版完整FLEET／SHIP差分。
[原版來源與成長模型](data/battle-promotion.json)、
[分級定位索引](data/battle-promotion-annotations.json)及
[自動合併定位索引的IDA匯出器](../../tools/scripts/battle_promotion_review_ida.py)
保留原始bytes與未知警示，不取代原始名稱。
[分級來源核對器](../../tools/scripts/battle_promotion_review_receipt.py)及
[分級原指令與軍階文字](data/battle-promotion-review.json)回查兩陣營12級文字。
[研究418](../re/418-battle-promotion.md)與[原版診斷073](../playtest/073-battle-promotion-oracle.md)
保存60案原版雙跑及成長模型；艦隊／SHIP速度重算仍待補，108 DRAFT。

第83輪沿#79：[索敵燃料與回收計畫](../goal/2026-10-04-scout-fuel-round-83.md)。
限定062子步8-3，先核固定原版的派出／冷卻邊界、報告等待與原池回寫，再接正式回合。
[原指令匯出器](../../tools/scripts/scout_fuel_ida.py)保留0491／05A9與艦隊名格式的定位及交叉參照。
[常式觀測器](../../tools/scripts/scout_fuel_probe.go)保存派出／冷卻邊界、完整艦隊／索敵記錄與報告當下狀態。
該冷狀態探針遇到報告繪圖未返回，正式補證改用
[完整保存觀測器產生器](../../tools/scripts/scout_fuel_native_prepare.py)與
[隔離模組執行器](../../tools/scout_fuel_native_run.sh)沿固定machine／DOS保存自然續行。
[固定版本雙跑入口](../../tools/scout_fuel_oracle.sh)拒絕髒執行器及覆寫收據。
[來源凍結器](../../tools/scripts/scout_fuel_receipt.py)、[原版來源](data/scout-fuel.json)、
[107](../spec/107-scout-fuel.md)與[研究417](../re/417-scout-fuel.md)保存完整燃料及報告證據。
[正式索敵燃料](../../internal/game/scoutfuel.go)在8-2返回後續接，報告後才回收並保留原池。
[原版與玩家回歸](../../internal/game/scoutfuel_test.go)分開控制案例、報告矩形與正常操作鏈。
[對拍072](../playtest/072-scout-fuel.md)、[產品來源凍結器](../../tools/scripts/scout_fuel_game_receipt.py)
及[產品來源](data/scout-fuel-game.json)保存兩報告矩形、正常一次實行與診斷存讀的各自範圍。

第82輪沿#79：[戰術星域歸屬計畫](../goal/2026-10-04-battle-ownership-round-82.md)，
限定子步8-2，先核原版重算、座位報告及費沙登記，再接正式回合。
[IDA匯出器](../../tools/scripts/battle_ownership_ida.py)保留函式、原始bytes及交叉參照。
[常式觀測器](../../tools/scripts/battle_ownership_probe.go)保留明示注入、完整星域記錄與報告呼叫。
[自然入口重播](../../tools/battle_ownership_oracle.sh)沿既有第五回合收據續行，與常式診斷分開。
[原版凍結器](../../tools/scripts/battle_ownership_receipt.py)、[規格106](../spec/106-battle-region-ownership.md)
及[原版來源v2](data/battle-ownership-v2.json)保存156案、14組完整收據與分級原指令。
[原版來源v1](data/battle-ownership.json)保留被推翻的無頭像推論；更正見研究416。
[保存解碼器產生器](../../tools/scripts/battle_ownership_state_prepare.py)逐字抽固定dosgolem型別，
比較所有保存欄位，保留gob map列舉順序造成的原檔雜湊差異。
[對拍071](../playtest/071-battle-ownership.md)、
[產品凍結器](../../tools/scripts/battle_ownership_game_receipt.py)及
[產品來源](data/battle-ownership-game.json)保存156案規則、六張報告矩形、
兩陣營公開保存選單與正常首戰GUI兩倍率的各自驗證範圍。
[正式8-2實作](../../internal/game/battleownership.go)依106串接座位報告與費沙登記。

第73輪沿#48／#79：[交戰開場與裁切計畫](../goal/2026-10-04-engagement-opening-round-73.md)。
從固定原版RAM核對開場計時與合法編成容量，不由模擬CPU步數猜補固定延遲。
[開場IDA匯出器](../../tools/scripts/engagement_opening_ida.py)保留原始bytes與位址空間。
[完整畫面來源凍結器](../../tools/scripts/fleet_frame_receipt.py)及[typed來源](data/fleet-frame.json)
保存傷害前後兩幀的文字、量表、圖段、投影與完整原生畫面。
[099規格](../spec/099-fleet-frame.md)、[研究410](../re/410-fleet-frame.md)與
[對拍062](../playtest/062-fleet-frame.md)限定無時間的完整單幀合成。

第72輪沿#48／#79：[投影與交戰接線計畫](../goal/2026-10-04-fleet-projection-round-72.md)，
從321F:000A／0223／0513閉合typed投影、裁切域與等待輸入，往完整Game接線前進。
[097規格](../spec/097-fleet-projection.md)、[研究409](../re/409-fleet-projection.md)、
[對拍061](../playtest/061-fleet-projection.md)與[投影收據](data/fleet-projection.json)保存30案及原指令。
[原版重播](../../tools/fleet_projection_oracle.sh)、[觀測器](../../tools/scripts/fleet_projection_native.go)、
[IDA匯出](../../tools/scripts/fleet_projection_ida.py)、[凍結器](../../tools/scripts/fleet_projection_receipt.py)為重生入口。
[場景來源凍結器](../../tools/scripts/fleet_scene_plan_receipt.py)及[完整下方收據](data/fleet-scene-plan.json)
把typed投影接到095，完整092、等待起點與Game仍待驗。
[098等待規格](../spec/098-engagement-wait.md)、[等待重播](../../tools/fleet_wait_oracle.sh)、
[等待觀測器](../../tools/scripts/fleet_wait_native.go)、[等待凍結器](../../tools/scripts/fleet_wait_receipt.py)及
[11案等待收據](data/fleet-wait.json)補結果等待與左右輸入排隊；起始tick、完整Game仍待驗。

第71輪沿#48／#79：[完整交戰呈現計畫](../goal/2026-10-04-fleet-presentation-round-71.md)，
閉合文字、色盤及等待輸入，往完整scene與Game接線前進。
[呈現原指令匯出](../../tools/scripts/fleet_presentation_ida.py)保留原名／bytes與三種定位。
[原版文字重播](../../tools/fleet_presentation_oracle.sh)、[觀測器](../../tools/scripts/fleet_presentation_native.go)及
[凍結器](../../tools/scripts/fleet_presentation_receipt.py)保存92案雙跑、691條原指令與完整注入。
[096規格](../spec/096-fleet-presentation.md)、[研究408](../re/408-fleet-presentation.md)、
[對拍060](../playtest/060-fleet-presentation.md)、[格式資料](data/fleet-presentation-spec.json)與
[凍結收據](data/fleet-presentation.json)為同一入口。正式文字已實作，Game交戰色盤已接；完整092仍DRAFT。

第70輪沿#48／#79：[交戰下方場景計畫](../goal/2026-10-04-fleet-blocks-round-70.md)，
從321F:06FC閉合透明堆疊、裁切與背景；0B43無buffer搬移。095已READY，完整092未接。
[下方場景IDA匯出](../../tools/scripts/fleet_blocks_ida.py)保留原名、定位與bytes，未知不改名。
[完整下方重播](../../tools/fleet_blocks_oracle.sh)與[逐層觀測器](../../tools/scripts/fleet_blocks_native.go)
保存傷害前後的背景、方塊、標籤、呈現、連線與箭頭邊界及可回載完整state。
[暫存器觀測器](../../tools/scripts/fleet_blocks_registers.go)、
[裁切診斷](../../tools/scripts/fleet_blocks_edges.go)與[可丟棄合成探針](../../tools/scripts/fleet_blocks_probe.py)
分開正式原版續接與直接入口注入；[凍結器](../../tools/scripts/fleet_blocks_receipt.py)保存原版期望值。
[095規格](../spec/095-fleet-blocks.md)、[研究407](../re/407-fleet-blocks.md)、
[對拍059](../playtest/059-fleet-blocks.md)與[凍結收據](data/fleet-blocks.json)是同一輪的入口。

第69輪沿#48／#79：[艦隊完整場景計畫](../goal/2026-10-04-fleet-scene-round-69.md)，
追1DD6:05D7的文字／量表／方塊／圖視窗／戰損／等待／返回，再接093到正式Game。
[場景IDA匯出](../../tools/scripts/fleet_scene_ida.py)保存原始定位、bytes與未審查語意。
[完整交戰重播](../../tools/fleet_scene_oracle.sh)與[原版觀測器](../../tools/scripts/fleet_scene_native.go)
由已提交命令的完整state續跑到返回戰場，不新增局面注入或玩家輸入。
[場景收據凍結器](../../tools/scripts/fleet_scene_receipt.py)核四案雙跑、原始bytes及祖先收據；
死亡反例另外明列目標1艘注入，不與前三案的無新增注入混稱。
[研究406](../re/406-fleet-scene.md)、[原版重播058](../playtest/058-fleet-scene.md)與
[凍結場景收據](data/fleet-scene.json)保存完整呼叫鏈及死亡後停止；事實21／38回填現行用途。

第68輪沿#48／#79：[交戰量表接線計畫](../goal/2026-10-04-engagement-gauges-round-68.md)與
[量表IDA匯出](../../tools/scripts/engagement_gauges_ida.py)，閉合共用量表後接正式演出。
[094子規格](../spec/094-engagement-gauges.md)、[研究405](../re/405-engagement-gauges.md)與
[驗證057](../playtest/057-engagement-gauges.md)保存範圍、原版證據及正式renderer抽樣。
[原版入口](../../tools/engagement_gauges_oracle.sh)沿用專用dosgolem；
[原版探針](../../tools/scripts/engagement_gauges_native.go)、
[凍結器](../../tools/scripts/engagement_gauges_receipt.py)與[結構化收據](data/engagement-gauges.json)
保留兩側量表32案的原始像素。量表事實回填[37](37-ani-pnl-a98.md)。
第67輪沿#79／#48：[正式交戰核心計畫](../goal/2026-10-04-fleet-combat-core-round-67.md)與
[093子規格](../spec/093-fleet-combat-core.md)。單擊與排序先READY再實作；
[純核心收據](data/fleet-combat-core.json)及[凍結器](../../tools/scripts/fleet_combat_core_receipt.py)
保存原版直接資料對拍，正式Game／UI／保存未接，093 READY、完整092仍DRAFT。

第66輪沿#79／#48：[多目標與多攻擊者計畫](../goal/2026-10-04-combat-sequence-round-66.md)。
先閉合重疊座標曲線與原版命令序列，避免以單擊probe代替正式交戰鏈；092仍DRAFT。
[序列重播入口](../../tools/combat_sequence_oracle.sh)與
[固定命令／係數計畫](../../tools/scripts/combat_sequence_plan.py)保留完整注入與點擊。
[係數域可丟棄probe](../../tools/scripts/combat_sequence_probe.go)沿第64輪模型；零距離另外保存原版殘留stack，不接production。
[機動力原始指令匯出](../../tools/scripts/combat_sequence_ida.py)核對既有機動力公式與排序加值。
[序列收據凍結器](../../tools/scripts/combat_sequence_receipt.py)核原版雙跑、來源指紋、祖先注入與模型，拒絕覆寫。
[研究404](../re/404-combat-sequence.md)、[原版診斷056](../playtest/056-combat-sequence.md)與
[完整序列收據](data/combat-sequence.json)保存零距離未初始化指標、已核准差異、多分隊排序及累積戰損。

第65輪沿#79／#48：[武器候選計畫](../goal/2026-10-04-weapon-candidates-round-65.md)與
[候選owner原始指令匯出](../../tools/scripts/weapon_candidates_ida.py)。
固定第62輪父狀態，先閉合可用戰鬥機及候選條件；092仍DRAFT。
[候選重播入口](../../tools/weapon_candidates_oracle.sh)與
[固定注入及選單計畫](../../tools/scripts/weapon_candidates_plan.py)保留完整輸入。
[可丟棄候選探針](../../tools/scripts/weapon_candidates_probe.go)核候選、射線及合法戰鬥機資料更新，不接正式Game。
[收據凍結器](../../tools/scripts/weapon_candidates_receipt.py)核原版雙跑、來源雜湊及IDA原始bytes，拒絕覆寫。
[研究403](../re/403-weapon-candidates.md)、[原版診斷055](../playtest/055-weapon-candidates.md)及
[候選固定收據](data/weapon-candidates.json)保存48案候選、合法戰鬥機命令及返回戰場。

第64輪沿#79／#48：[交戰傷害計畫](../goal/2026-10-03-fleet-combat-damage-round-64.md)。
由第63輪固定提交狀態追0C54及203A係數／資料更新，092保持DRAFT直到證據審查。
[第64輪IDA匯出](../../tools/scripts/fleet_combat_damage_ida.py)保留固定原版bytes與三種位址空間。
[傷害owner原版重播](../../tools/fleet_combat_damage_oracle.sh)及
[固定重播計畫](../../tools/scripts/fleet_combat_damage_plan.py)隔離提交、係數、扣艘數與計分。
[可丟棄戰損計算探針](../../tools/scripts/fleet_combat_damage_probe.go)沿既有定點及扣艘數規則，
核單擊係數、庫存、戰損、分數與功績；092仍DRAFT。
[傷害收據凍結器](../../tools/scripts/fleet_combat_damage_receipt.py)保存所有注入、原始定位與來源雜湊，拒絕覆寫。
[研究402](../re/402-fleet-combat-damage.md)、[原版診斷054](../playtest/054-fleet-combat-damage.md)及
[傷害固定收據](data/fleet-combat-damage.json)分開原版雙跑與未接線的正式流程。

第63輪沿#79／#48：[目標與武器提交計畫](../goal/2026-10-03-fleet-combat-command-round-63.md)、
[原版命令操作重播](../../tools/fleet_combat_command_oracle.sh)與
[固定輸入計畫](../../tools/scripts/fleet_combat_command_plan.py)。
沿092 DRAFT，先核目標命中與選擇提交；不覆寫第62輪收據或把診斷當正常玩家驗收。
[人類命令IDA匯出](../../tools/scripts/fleet_combat_command_ida.py)分開原始bytes與待審查語意。
[研究401](../re/401-fleet-combat-command.md)、[原版重播053](../playtest/053-fleet-combat-command.md)與
[固定提交／戰損收據](data/fleet-combat-command.json)保存16組雙跑、提交邊界及兩武器交戰末端。
[收據驗證工具](../../tools/scripts/fleet_combat_command_receipt.py)核固定父快照、303份來源及1506條IDA bytes。
命令子選單見[事實44](44-battle-command-menus.md)，交戰排序與損失owner見[事實38](38-battle-units.md)。

第62輪沿#79／#48：[艦隊戰鬥入口計畫](../goal/2026-10-03-fleet-combat-entry-round-62.md)與
[第四階段原始指令匯出](../../tools/scripts/fleet_combat_phase_ida.py)。
先核參戰門檻與玩家命令，不以行星對艦隊近似代替原版艦隊戰鬥。
[第四階段自然入口重播](../../tools/fleet_combat_entry_oracle.sh)保存可回載入口及完整原版資料。
[第四階段受控門檻計畫](../../tools/scripts/fleet_combat_entry_plan.py)明列唯一入口、所有注入與待驗排序。
[研究400](../re/400-fleet-combat-entry.md)、[092草稿](../spec/092-fleet-combat-entry.md)與
[對拍052](../playtest/052-fleet-combat-entry.md)分開入口雙跑及完整階段閘門。
[入口收據凍結器](../../tools/scripts/fleet_combat_entry_receipt.py)保存完整注入、原指令及未知項，拒絕覆寫。
[艦隊戰鬥入口完整收據](data/fleet-combat-entry.json)只驗原版重播，未作remake same-state聲明。

第61輪沿#79：[正常首次索敵計畫](../goal/2026-10-03-normal-spotting-round-61.md)與
[正常原版索敵邊界重播](../../tools/normal_spotting_oracle.sh)。
保持完整資料與畫面差異，不以無報告或局部情報一致升格完整087。
[正常bootstrap驗證](../../internal/game/normalspotting_player_test.go)與
[只讀續接觀測](../../internal/game/normalspotting_test.go)補情報與完整保存路徑。
[正常索敵收據凍結器](../../tools/scripts/normal_spotting_receipt.py)核雙跑、完整差異、標準回歸及來源身份。
[研究399](../re/399-normal-spotting.md)、[對拍051](../playtest/051-normal-spotting.md)與
[正常索敵固定收據](data/normal-spotting.json)保存限定情報、無報告續接及未同態的完整差異。

第60輪沿#59：[可見艦隊數工作快照](../goal/2026-10-03-region-visible-fleets-round-60.md)。
追查四格導出資料的寫入與正式玩家入口，完成聲明依當輪證據。
[原版可見艦隊數重播](../../tools/region_visible_fleets_oracle.sh)固定正常回合與明示索敵邊界；
[第60輪IDA匯出](../../tools/scripts/region_visible_fleets_ida.py)保留寫入者及玩家入口原始定位。
[091更新契約](../spec/091-region-visible-fleets.md)與[正式命令接線](../../internal/game/regionvisible.go)
限定可見艦隊數導出，不改索敵規則。
[可見數原版規則對拍](../../internal/rules/region_visible_test.go)與
[正常回合／保存接線驗證](../../internal/game/regionvisible_test.go)分開記錄受控與正常路徑。
[可見數正常GUI](../../tools/region_visible_fleets_gui.sh)以真實鍵鼠建隊、實行及開一覽表，驗1×／2×全圖。
[可見數收據凍結器](../../tools/scripts/region_visible_fleets_receipt.py)保存原版雙跑、來源雜湊與附加分級語意。
[研究398](../re/398-region-visible-fleets.md)、[對拍050](../playtest/050-region-visible-fleets.md)及
[可見數完整收據](data/region-visible-fleets.json)保存寫入生命週期與無戰鬥清理證據。

第59輪沿#59：[一覽表分頁邊界工作快照](../goal/2026-10-03-list-tab-edges-round-59.md)與
[固定原版邊界重播](../../tools/list_tab_edge_oracle.sh)，四格顯示與情報寫入生命週期分開驗收。
[第59輪IDA匯出](../../tools/scripts/list_tab_edge_ida.py)保留原始指令與三種位址空間，語意逐項審查。
[分頁邊界對拍](../../internal/game/listtabedges_test.go)從正常Game開清單，明示四格／旗標／士氣注入並比較完整畫面。
[090顯示契約](../spec/090-list-tab-edge-display.md)只核已證實的讀法，不以注入通過宣稱情報更新完成。
[四格讀取邊界](../../internal/rules/region_list_test.go)驗FF、254、0及無效座位／短記錄，不改原始資料。
[分頁正常GUI](../../tools/list_tab_edge_gui.sh)用真實鍵鼠從同盟主選單開星系清單，驗一倍／兩倍完整畫面。
[分頁凍結工具](../../tools/scripts/list_tab_edge_receipt.py)保存原版雙跑、正式全圖、存檔、GUI及原始指令，拒絕覆寫。
[研究397](../re/397-list-tab-edges.md)、[對拍049](../playtest/049-list-tab-edges.md)及
[分頁完整收據](data/list-tab-edges.json)保存十張零差及未驗情報寫入的界線。

第58輪沿#60：[左鍵保持工作快照](../goal/2026-10-03-left-hold-round-58.md)、
[研究396](../re/396-left-button-repeat.md)、[089契約](../spec/089-left-button-repeat.md)與
[原始指令匯出](../../tools/scripts/left_hold_ida.py)、[固定原版保持探針](../../tools/left_hold_oracle.sh)、
[速度操作探針](../../tools/scripts/left_hold_prototype.html)、[補證凍結工具](../../tools/scripts/left_hold_receipt.py)與
[完整保持補證](data/left-button-repeat.json)。補證凍結於裁決前；現行契約見089。

第58輪已確認350／100ms清單速度，限定CONFORMED子契約與[正式左鍵保持](../../internal/game/lefthold.go)接線依089。
[保持回歸](../../internal/game/lefthold_test.go)核時間邊界、列選擇、上下界、模態及保持取消。
[正常首戰保持全圖](../../internal/game/lefthold_normal_test.go)沿主選單138／S3同盟到資料窗，保持到26及0，不遮罩。
[真實GUI保持](../../tools/left_hold_gui.sh)用正常鍵盤與滑鼠按住五秒到上界，驗一倍／兩倍與全圖。
[GUI就緒觀測](../../tools/scripts/left_hold_gui_ready.py)唯讀比對首戰輸入標籤，再送正常滑鼠操作。
[月初提督清單保持補驗](../../tools/monthly_admiral_hold_oracle.sh)從正常建造計劃取得獨立父狀態與雙跑。
[月初框線游標清理](../../tools/scripts/monthly_admiral_cursor.py)由雙跑證據剔除25個舊量測混入的游標像素，產生器同時接線。
[保持成果凍結工具](../../tools/scripts/left_hold_game_receipt.py)、[正式保持收據](data/left-button-repeat-game.json)與
[對拍048](../playtest/048-left-button-repeat.md)保存明示時間差異、正常GUI、全圖及保存範圍。
[新原版保持結果比較](../../internal/game/lefthold_oracle_test.go)以明示等價輸入映射核右下清單／一覽表／稅率／百分比全圖。

《銀河英雄傳說III SP》PC DOS 中文版（1994，亨和企業．微波軟體發行）的
結構化逆向資料。這裡放**目前為止查證出來的事實本身**；查證過程與逐次
調查紀錄在 `docs/re/`，執行器規格在 dosgolem 的 `docs/spec/`。

這個專案除了 remake 之外，還有一個目的是保存這款中文版的技術細節——
1994 年台灣中文化作品的封裝方式、文字編碼取捨、資料格式與製作團隊，
在原始碼與開發資料都不存在的情況下，只剩執行檔本身可以問。所以逆向出來
的東西一律整理成可讀的 markdown，不停留在對話紀錄或一次性的輸出檔裡。

- [索敵顯示最小契約研究381](../re/381-battle-spotting-display.md)
- [索敵顯示完整原版資料](data/battle-spotting-display.json)
- [索敵顯示分級定位與固定61案計畫](data/battle-spotting-display-semantics.json)
- [索敵顯示收據工具](../../tools/scripts/battle_spotting_display_evidence.py)


第57輪入口：[資料視窗捲動工作快照](../goal/2026-10-03-battle-data-scroll-round-57.md)。沿#67，以自然S3同盟局面補超過十四列的完整畫面收據。

[資料視窗依賴原指令匯出](../../tools/scripts/battle_data_scroll_ida.py)保留原始定位及bytes，未審查項保持unknown。
[資料視窗四十列原版重播](../../tools/battle_data_scroll_oracle.sh)固定正常同盟路徑與seed，不注入局面。
[四十列及摘要／視野框研究395](../re/395-battle-data-scroll.md)保存本輪呈現根因與原始定位。
[正常入口全畫面回歸](../../internal/game/battledata_wholeframe_test.go)與[正式GUI鍵鼠驗證](../../tools/battle_data_scroll_gui.sh)覆蓋三個偏移及兩種倍率。
[收據產生器](../../tools/scripts/battle_data_scroll_receipt.py)、[完整收據](data/battle-data-scroll.json)與[對拍047](../playtest/047-battle-data-scroll.md)保存原圖、差分、逐區統計及驗證限制。

第56輪入口：[工作快照](../goal/2026-10-03-battle-object-keys-round-56.md)、
[研究394](../re/394-battle-object-keys.md)、
[IDA匯出](../../tools/scripts/battle_object_keys_ida.py)、[原版鍵反例](../../tools/battle_object_keys_oracle.sh)、
[收據產生器](../../tools/scripts/battle_object_keys_receipt.py)與[結構化證據](data/battle-object-keys.json)。限定顯示物件指標與排序鍵，不改完整入口門檻。


## 這裡的東西怎麼寫

- [第五十輪移動後戰術座標](../goal/2026-10-03-fleet-arrival-placement-round-50.md)
- [移動後座標原版擷取器](../../tools/fleet_arrival_oracle.sh)
- [移動後配置IDA定位工具](../../tools/scripts/fleet_arrival_ida.py)
- [移動後原池配置規則](../../internal/rules/fleetarrival.go)
- [移動過場後的正式配置接線](../../internal/game/fleetarrival.go)
- [正常S3配置與保存驗證](../../internal/game/fleetarrival_test.go)

- [第四十四輪正常玩家戰場入口配對](../goal/2026-10-03-battle-normal-entry-round-44.md)
- [正常原版冷啟動擷取器](../../tools/battle_normal_entry_oracle.sh)
- [正式資源正常玩家探測](../../tools/scripts/battle_normal_probe.go)
- [僅固定種子的測試期Go覆蓋](../../tools/scripts/battle_normal_probe_seed.py)
- [正常入口完整資料](data/battle-normal-entry.json)
- [正常入口雙跑與差異核對](../../tools/scripts/battle_normal_entry_receipt.py)
- [正常首戰入口研究382](../re/382-normal-battle-player-entry.md)
- [正常首戰入口草稿088](../spec/088-normal-battle-player-entry.md)
- [第四十五輪正常AI一次性閘門](../goal/2026-10-03-normal-ai-gate-round-45.md)

第三階段可見性／快取與正式續接：[第四十一輪](../goal/2026-10-03-battle-spotting-integration-round-41.md)，工作項目 #79。
- [第四十二輪正式索敵接線](../goal/2026-10-03-battle-spotting-integration-round-42.md)
- [第四十三輪來源標頭與逐格移鏡](../goal/2026-10-03-battle-discovery-scroll-round-43.md)
- [來源標頭／逐格移鏡固定計畫](data/battle-discovery-scroll-plan.json)
- [來源標頭／逐格移鏡完整原版及正式Game收據](data/battle-discovery-scroll.json)
- [原版擷取器](../../tools/battle_discovery_scroll_oracle.sh)
- [計畫凍結與雙跑／全畫面核對工具](../../tools/scripts/battle_discovery_scroll.py)

第三階段發現報告與正式續接：[第四十輪](../goal/2026-10-02-battle-discovery-round-40.md)，工作項目 #79。

第三階段索敵與戰鬥前續接：[第三十九輪](../goal/2026-10-02-battle-spotting-round-39.md)，工作項目 #79。
固定原版雙跑入口：[battle_spotting_oracle.sh](../../tools/battle_spotting_oracle.sh)，
邊界與父狀態身份：[battle_spotting_plan.py](../../tools/scripts/battle_spotting_plan.py)。
完整前後紀錄與分級索引：[battle-spotting-phase.json](data/battle-spotting-phase.json)，
重生工具：[battle_spotting_evidence.py](../../tools/scripts/battle_spotting_evidence.py)。
有限狀態探測：[battle_spotting_probe.go](../../tools/scripts/battle_spotting_probe.go)，不接正式玩家路徑。
第三階段規格：[087](../spec/087-battle-spotting-phase.md)，研究：[379](../re/379-battle-spotting-phase.md)。
原版重現範圍：[046](../playtest/046-battle-spotting-phase.md)。
六案現在位置／目的地分離與半徑邊界：[battle-spotting-coordinates.json](data/battle-spotting-coordinates.json)。

第二戰場殘留列來源：[第三十八輪](../goal/2026-10-02-battle-residue-source-round-38.md)，工作項目 #79。
小地圖bitmap契約：[086](../spec/086-battle-minimap-bitmap.md)，[研究378](../re/378-minimap-bitmap-bottom.md)。
分級來源及45份核對：[battle-minimap-bitmap.json](data/battle-minimap-bitmap.json)，
重生入口：[battle_minimap_evidence.py](../../tools/scripts/battle_minimap_evidence.py)。

第二戰場剩餘呈現來源：[第三十七輪](../goal/2026-10-02-battle-render-order-round-37.md)，工作項目 #79。
分隊繪製次序契約：[085](../spec/085-battle-unit-draw-order.md)。
調查紀錄：[研究377](../re/377-battle-render-order.md)。
分隊鍵分級索引：[battle-draw-order.json](data/battle-draw-order.json)，重生入口[原版核對工具](../../tools/scripts/battle_order_evidence.py)。

戰場傷害與方位規則接續範圍：[第二十四輪](../goal/2026-10-02-battle-rules-round-24.md)，工作項目 #53。

三項戰場規則勘誤：[規格 078](../spec/078-battle-rule-corrections.md)、
[研究 366](../re/366-battle-rule-corrections.md)、[對拍 038](../playtest/038-battle-rule-corrections.md)。
解出的事實：[欄位摘要](data/battle-rule-corrections.json)、
[IDA 自動合併分級索引](data/battle-rule-semantics.json)。

八步戰鬥的補給／修理階段接續範圍：[第二十五輪](../goal/2026-10-02-battle-preparation-round-25.md)，工作項目 #79。
本輪原版事實：[65 — 自動補給與修理](65-battle-preparation.md)、
[研究367](../re/367-battle-preparation.md)、[規格079](../spec/079-battle-preparation.md)、
[驗證039](../playtest/039-battle-preparation.md)。
欄位與分級索引：[battle-preparation.json](data/battle-preparation.json)、
[battle-preparation-semantics.json](data/battle-preparation-semantics.json)。
跨艦隊排序接續：[第二十六輪](../goal/2026-10-02-battle-preparation-round-26.md)、
[研究368](../re/368-battle-preparation-sort.md)。
原版排序欄位：[battle-preparation-sort.json](data/battle-preparation-sort.json)，
審查索引：[battle-preparation-sort-semantics.json](data/battle-preparation-sort-semantics.json)；
重驗入口：`tools/scripts/battle_preparation_sort_receipt.py`，方法與限制見研究368。
驗證範圍：[040 — 自動候選排序](../playtest/040-battle-preparation-sort.md)。
來源生命週期接續：[第二十七輪](../goal/2026-10-02-battle-preparation-round-27.md)、
[研究369](../re/369-battle-preparation-lifecycle.md)。
來源及結果驗證：[041](../playtest/041-battle-preparation-lifecycle.md)，
解出欄位：[battle-preparation-lifecycle.json](data/battle-preparation-lifecycle.json)，
審查索引：[battle-preparation-lifecycle-semantics.json](data/battle-preparation-lifecycle-semantics.json)。
重播／匯出工具的入口與限制見研究369，完整第一階段仍由規格079管理。
來源重設及正式接線接續：[第二十八輪](../goal/2026-10-02-battle-preparation-round-28.md)、
[研究370](../re/370-battle-preparation-reset.md)。
修理目標身分的[分級索引](data/battle-preparation-target-semantics.json)由同一IDA匯出工具合併。
正式第一階段的資源、面板、等待、正常入口與存讀驗收見[對拍042](../playtest/042-battle-preparation-first-phase.md)。
第二階段移動接續：[第二十九輪](../goal/2026-10-02-battle-movement-round-29.md)、
[研究371](../re/371-battle-movement-phase.md)、[草稿080](../spec/080-battle-movement-phase.md)、
[分級索引與局面摘要](data/battle-movement-phase.json)；原版重播入口與限制由研究371管理。
單步／避障接續：[第三十輪](../goal/2026-10-02-battle-movement-round-30.md)、
[研究372](../re/372-battle-step-and-collision.md)。
艦隊單步與碰撞的限定正式契約：[規格081](../spec/081-battle-fleet-step.md)、
[分級原版102案](data/battle-step.json)；重生入口為`tools/scripts/battle_step_receipts.py`。
限定同狀態資料與正常保存鏈驗收：[驗證043](../playtest/043-battle-fleet-step.md)。

音色庫來源與完整授權索引：[`player-music-library.json`](data/player-music-library.json)，
用途與原版素材邊界見 [`17-mds.md`](17-mds.md)。
原生 GM 移植與真實轉檔驗收索引：[`native-gm-import.json`](data/native-gm-import.json)，
方法及限制見[驗證 030](../playtest/030-native-gm-import.md)。
原生 FM 的固定版本／曲目／封裝索引：[`fm-import-lock.json`](data/fm-import-lock.json)，
原版觸發與重生限制見[規格 071](../spec/071-native-fm-import.md)、[研究 359](../re/359-native-fm-import.md)。
完整原生曲庫匯入與三路徑重跑收據：[`native-fm-import.json`](data/native-fm-import.json)，
正式解碼、重生命令與平台限制見[驗證 031](../playtest/031-native-fm-import.md)。
Linux 隨包工具、共享依賴與授權索引：[`linux-music-tools.json`](data/linux-music-tools.json)，
原生啟動、乾淨容器抽驗與交付限制見[驗證 032](../playtest/032-linux-music-tools.md)。
Linux 工具固定來源集合：[`linux-music-source-lock.json`](data/linux-music-source-lock.json)，
完整取得／離線重驗與本機封裝收據：[`linux-music-sources.json`](data/linux-music-sources.json)，
命令與正式交付限制見[驗證 033](../playtest/033-linux-music-sources.md)。
Windows FM 反例與完整 GM 探針收據：[`windows-music-probes.json`](data/windows-music-probes.json)，
重生命令及限制見[研究 362](../re/362-windows-music-tools.md)；
跨平台 FM 無損雜湊契約見[規格 074](../spec/074-cross-platform-fm-pcm.md)。
其十一首正式原生匯入與三路重跑收據由 `tools/scripts/music_windows_fm_receipt.py`
產生 [`windows-fm-pcm.json`](data/windows-fm-pcm.json)，未完成批次不得產生通過收據。
Windows 固定建置來源鎖：[`windows-music-source-lock.json`](data/windows-music-source-lock.json)，
十四組來源與兩組額外來源完整核對，方法與正式交付限制見[驗證 035](../playtest/035-windows-gm-sources-unicode.md)。
Windows GM 的 Unicode 命令列與載入失敗拒絕契約：[規格 075](../spec/075-windows-gm-unicode.md)。
Windows FM／GM 同包工具與預設完整入口：[規格 076](../spec/076-windows-combined-music-tools.md)，
候選依賴與驗證過程見[研究 364](../re/364-windows-combined-music-tools.md)。
同包二十二首、Unicode OPL 與中文 ZIP 搬移收據：[`windows-combined-music-tools.json`](data/windows-combined-music-tools.json)，
由 `tools/scripts/music_windows_bundle_receipt.py` 重驗產生；完整命令、反例與限制見
[驗證 036](../playtest/036-windows-combined-music-tools.md)。
音樂工具的本作品授權與可重建來源契約：[規格 077](../spec/077-music-source-delivery.md)，
本輪範圍見[第二十三輪](../goal/2026-10-02-music-source-delivery-round-23.md)。
來源重建正式 FM 抽驗入口：`tools/scripts/music_source_import_check.py`，方法見[驗證 037](../playtest/037-music-source-delivery.md)。
完整來源、授權、重建程式與首批反例的唯一收據：[`music-source-delivery.json`](data/music-source-delivery.json)，
由 `tools/scripts/music_source_receipt.py` 重驗產生；命令與限制見
[驗證 037](../playtest/037-music-source-delivery.md)，調查過程見[研究 365](../re/365-music-source-delivery.md)。
其固定來源工具與原生曲庫驗收收據：[`windows-gm-source-tools.json`](data/windows-gm-source-tools.json)，
由 `tools/scripts/music_windows_gm_source_receipt.py` 重驗產生；過程與勘誤見[研究 363](../re/363-windows-music-source-build.md)。
remake 增強策略的正式入口驗證索引：[`enhanced-ai-player-path.json`](data/enhanced-ai-player-path.json)，
範圍及原版差異見[規格 068](../spec/068-enhanced-ai-player-path.md)。

**每一項都要標推論等級**：`confirmed`（有實測或反組譯位元組證據）、
`強證據`、`假說`、`未知`。推測不得因為「看起來很合理」就升格，
升格要補證據並註明來源。

**什麼是穩定的、什麼不是**——這決定資料該怎麼記：

| 穩定 | 不穩定 |
|---|---|
| 結構欄位的**位移**（`rec[+0x0B]`）| **段值**（`6B93:0068`）|
| 資料段內的**變數位移**（`ds:94E0`）| 執行期的**線性位址** |
| **位元組簽章**（一段沒有絕對位址的機器碼）| 步數（`#25757331`）|
| 演算法、編碼、檔案格式 | 暫存器快照 |

段值取決於程式被載到哪裡，而那取決於 DOS 的記憶體佈局。實例：修好
dosgolem 的 EXEC 記憶體回收之後，`GIN3PS.EXE` 的 PSP 從 `0x3922` 降到
`0x0339`，所有模組跟著搬家，`docs/re/05` 之前記下的段值全部失效
（結構位移與演算法則完全不受影響）。

**所以定位程式碼一律用位元組簽章。** dosgolem 的 probe 可以一次驗一批：

```
tools/go.sh run ./cmd/probe -exe /orig/GIN3.COM -root /orig ... \
  -find "26 8A 47 0B 30 E4 26 8A 4F 0C 30 ED F7 E1, 83 3E E6 94 FF 75 1A"
```

## 產出格式

**每一項解出來的東西都要有機器可讀版本。** markdown 給人看，
`data/*.json` 給後續的工具與 remake 用；兩邊由同一份證據產生，不得只更新
一邊。素材（字型、圖庫、地圖）解出來之後匯出成 **PNG** 放
`workplace/assets/`——那是原版素材，**不進版控**，但要讓下一輪能直接拿來用。

挑簽章時避開含絕對位址的位元組（`9A xx xx ss ss` 的 far call 目標、
`A1 xx xx` 之類指到別段的存取），那些會隨載入位址改變。

## 工具

| 包裝器 | 用途 |
|---|---|
| `tools/py.sh` | Python 分析（`python:3.12-slim`）。原版素材唯讀掛在 `/orig`，倚天字型之類的外部參照用 `LOGH3_EXTRA_RO` 掛在 `/ref`。|
| `tools/ida.sh` | IDA Pro 9.4 headless。`analyze <exe>` 建 `.i64`，`script <exe> <py>` 跑 IDAPython。**exit code 不能當證據，只看輸出檔**；腳本一律寫檔，一支腳本一次處理多個目標（對同一個 `.i64` 連續開關會弄壞它）。|

`GIN3PS.EXE` 是六層封裝（GLUE／TINYPROG／PKLITE 系，見 `02`），IDA 只看得到最外層的解壓殼——本體要在執行期逆向
（dosgolem 的 `-peek`／`-find`／trace／檔案存取記錄）。`OPEN.EXE` 與
`ENDING.EXE` 沒有封裝，適合 IDA。

## 目錄

| 檔 | 內容 |
|---|---|
| [`01-assets-inventory.md`](01-assets-inventory.md) | 原版素材清單 |
| [`02-executables-and-credits.md`](02-executables-and-credits.md) | 執行檔、封裝器與製作團隊 |
| [`03-startup-and-memory.md`](03-startup-and-memory.md) | 啟動鏈、記憶體佈局與堆積配置器 |
| [`04-text-windows-and-big5.md`](04-text-windows-and-big5.md) | 文字視窗、儲存格格式與中文編碼 |
| [`05-palette-and-fade.md`](05-palette-and-fade.md) | 調色盤與淡入淡出 |
| [`06-font-graph-img.md`](06-font-graph-img.md) | 中文字型：`GRAPH.IMG` |
| [`07-window-frame.md`](07-window-frame.md) | 視窗外框 |
| [`08-admiral-dt.md`](08-admiral-dt.md) | ADMIRAL.DT[89AX] — 提督能力表 |
| [`09-face-amg.md`](09-face-amg.md) | FACE.AMG — 人物頭像（.AMG 圖庫格式） |
| [`10-sprite-spr.md`](10-sprite-spr.md) | SPRITE.SPR — 艦隊編號圖塊 |
| [`11-copy-protection-screen.md`](11-copy-protection-screen.md) | 防拷問答畫面 |
| [`12-strmap-bit.md`](12-strmap-bit.md) | STRMAP.BIT — 星域地圖底圖 |
| [`13-strmap-rgn.md`](13-strmap-rgn.md) | STRMAP.RGN — 星域區域遮罩 |
| [`14-screen-strings.md`](14-screen-strings.md) | 畫面文字表與熱區 |
| [`15-keireki-dat.md`](15-keireki-dat.md) | KEIREKI.DAT — 人物傳 |
| [`16-ending-exe.md`](16-ending-exe.md) | ENDING.EXE — 結局文字與職員表 |
| [`17-mds.md`](17-mds.md) | `.MDS` — 樂曲 |
| [`18-vo5.md`](18-vo5.md) | `.VO5` — 結局圖 |
| [`19-ibm16-350.md`](19-ibm16-350.md) | `IBM16.350` ＋ `IBMFLS16.350` — 開頭動畫 |
| [`20-comani-bmg.md`](20-comani-bmg.md) | `COMANI.BMG` — 戰鬥動畫封存檔 |
| [`21-3dblock-spr.md`](21-3dblock-spr.md) | `3DBLOCK.SPR` — 等距方塊 |
| [`22-stellar-dt.md`](22-stellar-dt.md) | `STELLAR.DT[89AX]` — 星域資料 |
| [`23-hotspot-lists.md`](23-hotspot-lists.md) | 熱區清單與按鈕標籤 |
| [`24-window-list-and-right-button.md`](24-window-list-and-right-button.md) | 視窗串列、命中測試與右鍵 |
| [`25-text-rendering.md`](25-text-rendering.md) | 文字繪製：粗體、陰影、底線與網點 |
| [`26-star-region-graph.md`](26-star-region-graph.md) | 星域之間的相鄰關係與路徑表 |
| [`27-treasury-globals.md`](27-treasury-globals.md) | 財政的全域變數 |
| [`28-screen-specs.md`](28-screen-specs.md) | 畫面規格 JSON：`data/screen-*.json` |
| [`29-planets-and-economy.md`](29-planets-and-economy.md) | `PLADAT.DAT` 與經濟力 |
| [`30-save-format.md`](30-save-format.md) | 存檔：`*.DT[89A]` 就是執行期的表 |
| [`31-text-plane.md`](31-text-plane.md) | 全螢幕字元平面（`ds:D81E`）|
| [`32-ship-build-costs.md`](32-ship-build-costs.md) | 船艦建造：單價（兩個陣營不同）、一單位幾艘與捨去規則 |
| [`data/`](data/) | 機器可讀版本（JSON），給後續工具與 remake 用 |

- [`34-battle-screen.md`](34-battle-screen.md)：戰術（戰鬥）畫面的版面、命令選單與怎麼走到。
- [`35-block3-blk.md`](35-block3-blk.md)：戰場的圖塊集（128 塊 16×20）。
- [`36-map-cmp.md`](36-map-cmp.md)：每個星域的戰場地圖（128×102 格，兩條規則的編碼）。
- [`37-ani-pnl-a98.md`](37-ani-pnl-a98.md)：戰鬥畫面的底圖（`.AMG` 編碼、畫在 y=40）。
- [`38-battle-units.md`](38-battle-units.md)：戰場上的單位是「分隊」，不是艦隊。
- [`39-battle-entry-points.md`](39-battle-entry-points.md)：戰場入口表。
- [`40-ship-kind-table.md`](40-ship-kind-table.md)：艦種屬性表。
- [`data/live-fleet-save.json`](data/live-fleet-save.json)：現況艦隊存讀、速度寫入的分級語意與原版／GUI 雜湊；入口為[規格 067](../spec/067-live-fleet-save.md)。
- [`41-sin-table.md`](41-sin-table.md)：原版的正弦表。
- [`42-screen-tile-table.md`](42-screen-tile-table.md)：畫面的圖塊表 `ds:D81E`。
- [`43-battle-terrain-and-rally.md`](43-battle-terrain-and-rally.md)：戰場地形的通行規則與集結點。
- [`44-battle-command-menus.md`](44-battle-command-menus.md)：戰場的命令選單（五張表、可選旗標、版面公式）。
- [`45-battle-unit-panel.md`](45-battle-unit-panel.md)：戰場右側的分隊面板（八列、值的來源、三個顏色規則）。
- [`46-battle-formation-chooser.md`](46-battle-formation-chooser.md)：戰場「編隊」的選格子畫面。
- [`47-window-frame-tiles.md`](47-window-frame-tiles.md)：視窗外框的圖塊（`FRAME.BLK`）。
- [`48-message-inventory.md`](48-message-inventory.md)：訊息框說了什麼。
- [`49-text-engine.md`](49-text-engine.md)：印字進視窗。
- [`50-fortress-gun.md`](50-fortress-gun.md)：要塞砲的射程、瞄準與開火。
- [`51-borland-long-helpers.md`](51-borland-long-helpers.md)：Datalight C 的 32 位元輔助函式（`3CC9` 那一段，檔名沿用舊名）。
- [`52-plot-primitives-and-canvas.md`](52-plot-primitives-and-canvas.md)：畫點的三支常式與那 40 列的畫布偏移。
- [`53-strategic-planet-assault.md`](53-strategic-planet-assault.md)：戰略回合裡的打行星（`20B4:1512` → `20B4:1662`）。
- [`54-strategic-ai-decision.md`](54-strategic-ai-decision.md)：戰略 AI 決定一支艦隊要不要動。
- [`56-fleet-composition-panel.md`](56-fleet-composition-panel.md)：艦隊組成畫面的工作區（編隊那一行的字母與格盤的圖塊）。
- [`57-phezzan-event.md`](57-phezzan-event.md)：費沙事件——第一次易主時觸發、經濟力流向對手首都、邊境叛離、人類座位受罰。
- [`58-strategic-fleet-window.md`](58-strategic-fleet-window.md)：戰略地圖上的艦隊視窗與命令（清單點列 → 作戰 → 目的地 → 確認）。
- [`59-list-screens.md`](59-list-screens.md)：一覽表四張（星系／艦隊／提督／敵將資料）：點列開什麼、游標反白、位移與分頁的變數。
- [`60-mouse-roles.md`](60-mouse-roles.md)：滑鼠在每張畫面的作用總表（左鍵、右鍵、按住、游標經過），逐格掛證據；原版有、remake 沒有的列在最後。
- [`61-turn-transition-windows.md`](61-turn-transition-windows.md)：回合處理中的畫面（「戰略模式處理中」、清單清空、地圖不看座位）與每支艦隊走一格的過場視窗（出發前往、被擋、經費不足、問話）、等待怎麼結束；報告框（拿下行星、士氣、已全軍覆沒）。
- [`62-strategic-attrition.md`](62-strategic-attrition.md)：戰略階段士氣低落停在別人星域的艦隊每回合扣兵、第 0 組 0 艘就全滅（報告、艦艇作廢、提督失蹤、星域與陣營艦隊數減一），沒全滅就移掉 0 艘的組。
- [`63-info-screen-tabs.md`](63-info-screen-tabs.md)：情報收集的三個分頁（治安／勢力／敵艦）與「確認」。
- [`64-ship-pool.md`](64-ship-pool.md)：艦艇池一筆 78 bytes 的欄位、配置與作廢（分隊就是它）。
- [完整交錯移動審查](../re/373-battle-interleaved-movement.md)：[第三十一輪](../goal/2026-10-02-battle-movement-round-31.md)依既有三邊界及102案核對完整第二階段。
- [交錯移動驗證044](../playtest/044-battle-interleaved-movement.md)：完整7艦隊／55分隊資料、正式Game轉接、正常命令中間幀及存讀；整階段畫面仍待對拍。
- [回合後距離／座標語意索引](data/post-turn-runtime-fields.json)：研究 354 的原始定位與證據分級，由 IDA 匯出自動合併；[全表收據摘要](data/post-turn-runtime-receipts.json) 的驗收入口是[對拍 027](../playtest/027-post-turn-runtime-fields.md)。
- [第三十二輪：索敵回收與移動畫面](../goal/2026-10-02-battle-movement-round-32.md)：接續正常回收、原池重用與產品畫面收據。
- [移動畫面收據工具](../../tools/scripts/battle_movement_render_receipt.py)：依371固定收據及正式Game.Frame記錄全畫面差分，不排除未對齊區域。
- [移動畫面與正常回收驗證045](../playtest/045-battle-movement-render.md)：隱藏圖示、點選及游標修正，原池再放出與原生差分。
- [第三十三輪：小地圖標記生命週期](../goal/2026-10-02-battle-marker-round-33.md)：接續正常初始化、更新及正式呈現。
- [小地圖標記原版探針](../../tools/scripts/battle_marker_probe.go)：固定原版RAM與seed，驗開關及座標快照；只限受控常式。
- [正常標記原版重播](../../tools/battle_marker_oracle.sh)：固定S3檢查點、輸入與seed，確認佈陣及顯示開關的玩家呼叫鏈。
- [標記生命週期研究374](../re/374-battle-marker-lifecycle.md)、[規格082](../spec/082-battle-minimap-markers.md)：正常佈陣、顯示開關與已畫快照的呈現契約。
- [標記收據產生器](../../tools/scripts/battle_marker_receipts.py)、[標記資料](data/battle-markers.json)及[語意索引](data/battle-marker-semantics.json)：96案、正常55筆與原始指令分級。
- [第三十四輪：階段面板與移動中間幀](../goal/2026-10-02-battle-phase-panel-round-34.md)、[階段面板規格083](../spec/083-battle-phase-panel.md)：接續#79固定第6回合逐圈畫面與階段文字。
- [逐圈原版重播](../../tools/battle_phase_oracle.sh)及[固定計數計畫](../../tools/scripts/battle_phase_plan.py)：唯讀snapshot、35個完整RAM／planar PNG／色號。
- [階段文字研究375](../re/375-battle-phase-panel.md)、[語意索引](data/battle-phase-semantics.json)、[逐圈全紀錄](data/battle-phase-frames.json)：呼叫順序、三字串與35邊界雙跑；由[證據工具](../../tools/scripts/battle_phase_evidence.py)重生。
- [階段面板正式畫面收據](../../tools/scripts/battle_phase_render_receipt.py)：35邊界全畫面及逐區有效RGB差分，不裁切或遮罩。
- [第三十五輪：第二戰場及正常資源結果](../goal/2026-10-02-battle-second-sample-round-35.md)：接續#79現行READY範圍，抽正式載入路徑及另一戰場。
- [隨附存檔正常玩家收據](../../tools/scripts/battle_samples_receipt.py)：三份存檔進場、實行、面板還原與完整存讀；[驗證045](../playtest/045-battle-movement-render.md)限定為remake正常路徑，不宣稱原版第二戰場對拍。
- [第三十六輪：原版第二戰場同狀態](../goal/2026-10-02-battle-second-oracle-round-36.md)：接續#79原版S1正常過場與固定戰場邊界，不重做三份remake玩家抽樣。
- [S1第二戰場原版重播](../../tools/battle_second_oracle.sh)與[固定計畫／匯出](../../tools/scripts/battle_second_plan.py)：43個邊界、18艦隊／135分隊，入口seed與完整旗標固定。
- [S1第二戰場正式Game驗證](../../internal/game/battlesecond_test.go)與[逐圈紀錄](data/battle-second-frames.json)：唯一入口注入，後續全部使用正式移動引擎。
- [S1第二戰場畫面差分](../../tools/scripts/battle_second_render_receipt.py)：原生全畫面與逐區色彩統計，保留修正前後差分，不遮罩。
- [第二戰場研究376](../re/376-second-battle-render.md)與[方向圖示規格084](../spec/084-battle-fleet-directions.md)：方向位置、條件及圖塊來源；同順位重疊與203列差分另待補證。
- [方向圖示分級索引](data/battle-directions.json)由[證據工具](../../tools/scripts/battle_second_render_evidence.py)核原始指令、八向位移及素材bytes。

第40輪索敵呈現收據入口：[固定七分支計畫](data/battle-discovery-plan.json)、[原版收據包裝器](../../tools/battle_discovery_oracle.sh)、[計畫旗標工具](../../tools/scripts/battle_discovery_plan.py)。
第40輪呈現事實：[分級指令索引](data/battle-discovery-semantics.json)、[完整原版證據](data/battle-discovery-evidence.json)、[證據重生工具](../../tools/scripts/battle_discovery_evidence.py)。
第40輪研究入口：[380發現標頭／暫時標亮](../re/380-battle-discovery-presentation.md)，目前事實同步於[戰場畫面](34-battle-screen.md)及[戰場分隊](38-battle-units.md)。

第42輪正式索敵：[Game](../../internal/game/battlespotting.go)、[087](../spec/087-battle-spotting-phase.md)。

第42輪可比UI邊界收據：[battle-spotting-game.json](data/battle-spotting-game.json)，由[索敵收據工具phase-render](../../tools/scripts/battle_spotting_display_evidence.py)重生。
- [第45輪AI閘門證據](../re/383-normal-ai-once-gate.md)：[完整收據](data/normal-ai-once-gate.json)，原版雙跑由[工具](../../tools/normal_ai_gate_oracle.sh)重生，[IDA匯出](../../tools/scripts/normal_ai_gate_ida.py)與[收據產生器](../../tools/scripts/normal_ai_gate_receipt.py)保存原始定位。
- [第46輪正常讀檔座位／日期](../goal/2026-10-03-load-seat-date-round-46.md)，接續#89與088未完成範圍。
- [讀檔陣營研究384](../re/384-load-faction-selection.md)、[完整座位收據](data/load-seat-selection.json)：[原版雙跑](../../tools/load_seat_oracle.sh)、[IDA匯出](../../tools/scripts/load_seat_ida.py)、[可丟棄選單](../../tools/scripts/load_seat_probe.go)與[收據重生](../../tools/scripts/load_seat_receipt.py)。正式入口見[Game讀檔選單](../../internal/game/loadfaction.go)及[正常玩家探針](../../tools/scripts/load_seat_player.go)，驗收範圍由088限定。
- [讀檔陣營正常回歸](../../internal/game/loadfaction_test.go)與[真實X11輸入抽驗](../../tools/load_seat_gui.sh)：依088驗正常主選單、三選擇、右鍵反白與放大命中。
- [第47輪日期進位](../goal/2026-10-03-turn-counter-round-47.md)：接續#89，核對0374:03D9在所有戰場與模態尾段之後進位；watch03DD為執行後IP，前輪座位子契約保持。
- [回合計數器原版重播](../../tools/turn_counter_oracle.sh)及[IDA原始指令匯出](../../tools/scripts/turn_counter_ida.py)：固定無戰鬥進位與正常S3戰場保持，輸入、工具版本與限制見第47輪。
- [回合計數器研究385](../re/385-turn-counter-boundary.md)：主迴圈原始定位、watch執行後IP勘誤及正常首戰日期修正依據。
- [正常回合與兩AI計數驗證](../../internal/game/turncounter_test.go)：正式玩家完整回合／WORK保存、兩種AI與兩陣營的建隊／命令時點；全戰場與模態尾段抽樣另由既有regionturn／victoryreport測試驗收。
- [回合修正真實X11輸入](../../tools/turn_counter_gui.sh)：正式cmd/gin3正常主選單至S3首戰，兩種視窗尺寸；不使用script或direct-entry。
- [計數器收據工具](../../tools/scripts/turn_counter_receipt.py)及[完整收據](data/turn-counter-boundary.json)：原版雙跑／IDA／19份正式資料／全圖差分、GUI與來源SHA，研究385及088限定驗收。
- [第48輪活動座位](../goal/2026-10-03-active-seat-round-48.md)：接續#89，查正常讀檔、實行與雙人交接的WORK控制生命週期。
- [活動座位原版重播](../../tools/active_seat_oracle.sh)：固定正常無戰鬥回合、同盟S3處理與雙人首次交接，全部RAM與watch保留，語意未證實欄位不補值。
- [活動座位IDA匯出](../../tools/scripts/active_seat_ida.py)：04F8四個座位函式的原始名稱、範圍、bytes與附加分級，不將自動名稱當證據。
- [活動座位研究386](../re/386-active-seat-lifecycle.md)與[事實27](27-treasury-globals.md)：兩層座位迴圈、思考／處理／命令的原始控制值與同盟435M實際停點，088限定驗收。
- [正式活動控制更新](../../internal/game/activeseat.go)：依088在載入／交接、思考／AI、處理與首戰更新WORK+09／+0A，不改固定玩家視角。
- [活動控制正常驗證](../../internal/game/activeseat_test.go)：原版捕獲值、正常回合／雙人交接／正式保存、兩AI實際命令與缺WORK／非法上下文邊界。

- [活動座位正常GUI](../../tools/active_seat_gui.sh)：正式三選擇、雙人同盟交接與兩倍視窗，使用產品scale及PID選視窗。
- [活動座位收據生成](../../tools/scripts/active_seat_receipt.py)與[結構化收據](data/active-seat.json)：原版控制邊界、正式保存與完整全圖比較。
- [第49輪月底座位閘門](../goal/2026-10-03-month-seat-gate-round-49.md)與[原版重播](../../tools/month_seat_gate_oracle.sh)：接續#89，隔離S3載入的WORK+E0／+E1旗標與錯誤重複結算。
- [月底閘門IDA匯出](../../tools/scripts/month_seat_gate_ida.py)：0654:0034的原始範圍、兩條閘門及共同清旗標尾巴。
- [月底閘門完整收據](data/month-seat-gate.json)由[驗證工具](../../tools/scripts/month_seat_gate_receipt.py)重生，保存原版自然與受控雙跑、IDA原始bytes及未解差異。
- [月底旗標研究387](../re/387-month-seat-gate.md)：重複結算根因、原版受控變因及088子契約的證據審查。
- [正式月底旗標](../../internal/game/monthseat.go)：單人raw WORK閘門、座位清除與回合進位重設，雙人仍依既有016範圍。
- [月底旗標正常驗證](../../internal/game/monthseat_test.go)：原版正常旗標、兩AI／兩陣營的正式存讀、旗標非零與跨月一次結算。
- [到達配置研究388](../re/388-fleet-arrival-placement.md)、[完整收據](data/fleet-arrival-placement.json)及[產生器](../../tools/scripts/fleet_arrival_receipt.py)：九艦隊／七十筆池的到達配置、完整表差分與正常首回合無報告的原版雙跑。
- [第51輪正常首戰輸入顯示](../goal/2026-10-03-battle-input-display-round-51.md)：接續#89，核對游標／右側摘要與選取、sprite快取，完整088仍DRAFT。
- [正常首戰純移動原版重播](../../tools/battle_input_display_oracle.sh)：固定435M輸入邊界，分隊、空白及面板位置的單次純移動與完整RAM／畫面。
- [純移動摘要研究389](../re/389-battle-input-display.md)與[IDA匯出](../../tools/scripts/battle_input_display_ida.py)：摘要快取、退出與兩個SPRITE 10物件的原始定位；088限定正常Input範圍。
- [純移動完整收據產生器](../../tools/scripts/battle_input_display_receipt.py)：保存五組獨立原版雙跑、正常入口全圖及未解像素，不改第50輪收據。
- [純移動結構化收據](data/battle-input-display.json)：五個IDA函式535條原指令與22項分級定位、55份正常入口雙跑及兩尺寸X11；全圖2／1／1／0／0 pixels差保留。
- [第52輪正常首戰摘要標記](../goal/2026-10-03-battle-hover-marks-round-52.md)：沿#89追剩餘小地圖與位置反白像素，既有第51輪收據保持。
- [摘要標記IDA探針](../../tools/scripts/battle_hover_marks_ida.py)：由固定原版RAM切段保存原始指令，未審語意保持unknown。
- [摘要小地圖原版重播](../../tools/battle_hover_marks_oracle.sh)：空白Input移回分隊及明示集結點注入，保存新RAM／畫面與完整旗標。
- [摘要小地圖研究390](../re/390-battle-hover-marks.md)：旗艦航線、一般分隊集結點與剩餘方向物件覆蓋的分級證據。
- [摘要小地圖收據產生器](../../tools/scripts/battle_hover_marks_receipt.py)與[結構化收據](data/battle-hover-marks.json)：八組原版雙跑、正常入口五個完整畫面與剩餘差異，原JSON保持。
- [第53輪隱藏偵察隊快取](../goal/2026-10-03-battle-scout-cache-round-53.md)：沿#89核對完整戰場池的圖示刷新與存讀，不重開既有畫面收據。
- [隱藏偵察隊研究391](../re/391-battle-scout-cache.md)：既有IDA資料庫的209條原始bytes與原版七筆快取；規格088限定接線。
- [偵察隊快取收據產生器](../../tools/scripts/battle_scout_cache_receipt.py)與[結構化收據](data/battle-scout-cache.json)：核對固定原版雙跑、正常Game雙跑、七個byte與未改畫面，保留原始位址與分級。
- [第54輪戰場圖示順位](../goal/2026-10-03-battle-sprite-priority-round-54.md)與[IDA探針](../../tools/scripts/battle_sprite_priority_ida.py)：沿#89核+30初始化及顯示消費端，未審查語意保持unknown。
- [圖示順位原版重播](../../tools/battle_sprite_priority_oracle.sh)：固定正常S3狀態、輸入及seed，追完整分隊與偵察池的建立／刷新，輸出拒絕覆寫。
- [圖示順位研究392](../re/392-battle-sprite-priority.md)：55池初始化、原指令及顯示消費端，分開正常收據與未重畫的診斷探針。
- [圖示順位收據產生器](../../tools/scripts/battle_sprite_priority_receipt.py)與[結構化收據](data/battle-sprite-priority.json)：原版55筆物件、完整指令索引、正常資料／保存／畫面及比較器回歸，前輪JSON保持。
- [第55輪首戰WORK控制與經費保存](../goal/2026-10-03-battle-work-control-round-55.md)：沿#89追剩餘控制與經費的原版寫入生命週期，既有收據不重生。
- [首戰WORK原版重播](../../tools/battle_work_control_oracle.sh)：固定S3父狀態與seed，監看控制、兩座位鏡頭與兩方移動經費；完整旗標與原始輸出分批保存。
- [首戰WORK的IDA匯出](../../tools/scripts/battle_work_control_ida.py)：原始定位、指令與附加語意分欄，未審查的掃描範圍保持unknown。
- [首戰WORK研究393](../re/393-battle-work-control.md)：控制與經費的寫入生命週期、座位鏡頭及未驗續戰的限制。
- [正式戰場WORK同步](../../internal/game/battlework.go)：088限定控制、鏡頭及經費更新；原版存檔格式保持。
- [WORK邊界回歸](../../internal/game/battlework_test.go)：兩座位鏡頭、Input保存邊界、狀態2／5及signed經費；完整正常WORK對拍沿既有fleetarrival_test。
- [WORK完整收據產生器](../../tools/scripts/battle_work_control_receipt.py)與[結構化收據](data/battle-work-control.json)：原版雙跑／寫入監看、785條IDA原始指令、完整230-byte保存／重讀及未變畫面；前輪JSON保持。

- [第74輪交戰命令資料鏈](../goal/2026-10-04-combat-command-data-round-74.md)與[100候選／原池命令規格](../spec/100-fleet-combat-candidates.md)：沿研究401／403，逐次候選與原存檔欄位；完整092仍DRAFT。
- [候選與命令驗證索引](data/fleet-command-data.json)由[第74輪凍結器](../../tools/scripts/fleet_command_data_receipt.py)重建；直接核原版逐查詢，不讀舊probe期望。
- [候選與命令接線研究411](../re/411-fleet-command-data.md)及[限定驗證063](../playtest/063-fleet-command-data.md)：111次原版查詢、命令原池存讀、正常bootstrap唯讀域；100不取代完整092。

- [第75輪正式交戰命令](../goal/2026-10-04-combat-player-command-round-75.md)：接續100的候選與原池資料，補原版自動／防禦與逐分隊續接，完整092仍DRAFT。
- [第75輪自動命令IDA匯出](../../tools/scripts/combat_player_command_ida.py)：保留原始bytes與分欄位址，未審語意unknown；一次性資料庫留workplace/ida。
- [第75輪固定自動分支重播](../../tools/combat_player_command_oracle.sh)與[完整注入計畫](../../tools/scripts/combat_player_command_plan.py)：固定父state，按預設RND條件驗門檻，不試種子或覆寫收據。
- [101自動命令規格](../spec/101-fleet-auto-command.md)、[完整來源](data/fleet-auto-command.json)與[凍結器](../../tools/scripts/combat_auto_receipt.py)：31組雙跑及原版預估／選擇暫存，缺beam的原版錯誤明列。
- [自動命令研究412](../re/412-fleet-auto-command.md)、[驗證064](../playtest/064-fleet-auto-command.md)、[Game索引](data/fleet-auto-game.json)與[凍結器](../../tools/scripts/fleet_auto_game_receipt.py)：31案115筆原版預估、21案有效原池存讀及診斷身分限制；完整092仍DRAFT。

- [第76輪正式玩家鏈](../goal/2026-10-04-fleet-player-chain-round-76.md)與[名冊原版重播](../../tools/fleet_player_chain_oracle.sh)、[完整注入計畫](../../tools/scripts/fleet_player_chain_plan.py)：接既有命令、核心與畫面；固定名冊有限排列補順位，拒絕覆寫歷史收據。
- [正式生命週期102](../spec/102-fleet-player-chain.md)、[研究413](../re/413-fleet-player-chain.md)、[原版來源](data/fleet-player-chain-native.json)與[凍結器](../../tools/scripts/fleet_player_chain_receipt.py)：名冊順位、UI來源及命令／核心／演出續接的最小充分契約。
- [交戰UI格式](data/fleet-command-ui.json)、[動態選單與標頭](../../internal/screen/fleet_command.go)：從凍結原版格式產生，沿既有命令框幾何與文字緩衝區。
- [正式玩家鏈驗證065](../playtest/065-fleet-player-chain.md)、[完整Game索引](data/fleet-player-chain-game.json)與[正常bootstrap重播](../../tools/scripts/fleet_player_chain_replay.go)：四種原版戰損、四圖零差、兩陣營正常公開輸入及保存；GUI鍵鼠、命令亮標與音效仍缺，102保持READY。

- [第77輪交戰命令顯示](../goal/2026-10-04-fleet-command-display-round-77.md)：沿#48／#79核對等待目標的射程、亮標、敵方資訊與游標，不覆寫既有收據。
- [交戰命令顯示IDA匯出](../../tools/scripts/fleet_command_display_ida.py)：第77輪固定RAM的原始bytes及分欄位址，附加語意預設unknown。
- [命令顯示原版重播](../../tools/fleet_command_display_oracle.sh)與[注入計畫](../../tools/scripts/fleet_command_display_plan.py)：固定池102情報1／2／3與滑鼠輸入，等待／摘要／彈窗的新收據分開保存。
- [命令顯示研究414](../re/414-fleet-command-display.md)與[顯示契約103](../spec/103-fleet-command-display.md)：射程與情報等級的資料流及受控驗收範圍。
- [命令顯示凍結器](../../tools/scripts/fleet_command_display_receipt.py)：固定原版雙跑、IDA bytes及格式的來源雜湊，拒絕覆寫。
- [敵方摘要原格式](data/fleet-command-display-spec.json)：固定DS4534／4541／4528／455E／456F，包含情報遮蔽及距離的四個字面零。
- [原版快照核對](../../tools/scripts/fleet_command_state_check.go)：第77輪逐項比較完整機器欄位與DOS段，辨別gob map序列化順序的差異。
- [命令顯示原版收據](data/fleet-command-display-native.json)：七組新原版雙跑及完整快照、固定旗標、情報注入、IDA來源的索引。
- [命令顯示對拍](../../internal/game/fleetcommanddisplay_test.go)：原池還原、完整renderer、公開滑鼠輸入及無遮罩的全圖／區域差分。
- [命令顯示驗證066](../playtest/066-fleet-command-display.md)、[正式顯示索引](data/fleet-command-display-game.json)與[Game凍結器](../../tools/scripts/fleet_command_display_game_receipt.py)：七圖各保留一個未核准像素、兩陣營正常命令畫面與六檔保存；103仍READY。
- [命令顯示正常重播](../../tools/scripts/fleet_command_display_replay.go)：正式bootstrap與公開輸入，逐張核對等待／彈窗與存檔，不注入位置或命令。

- [第78輪交戰音效](../goal/2026-10-04-fleet-sound-round-78.md)與[研究415](../re/415-fleet-sound.md)：自訂SVF標頭、音效編號及最小樣本消費鏈，硬體時序採規格停止線。
- [音效來源IDA匯出](../../tools/scripts/fleet_sound_ida.py)：固定原版RAM的原始bytes、三種位址基準及unknown語意，不修改正式資料庫。
- [SVF來源凍結器](../../tools/scripts/fleet_sound_source.py)、[完整來源](data/fleet-sound-source.json)及[兩側音效表](data/fleet-sound-spec.json)：28段payload、播放長度與硬體規格近似時長，附原始定位及雜湊。
- [音效契約104](../spec/104-fleet-sound.md)與[SVF解析器](../../internal/asset/svf.go)：原版樣本到OGG與一次性COMANI音效事件，正常播放與驗收各保留範圍限制。
- [一次性Game音效](../../internal/game/soundeffects.go)及[音效曲庫](../../internal/audio/effects.go)：與配樂共用素材目錄，保留首格前的近似播放完成閘門。
- [音效OGG轉檔器](../../tools/scripts/fleet_sound_convert.go)與[正式一次性播放器](../../cmd/gin3/effects.go)：沿既有音訊Docker工具與唯一Context，分開控制配樂與音效。
- [音效原版序列回歸](../../internal/game/soundeffects_test.go)：四案音效、三速度／動畫關閉、等待不重播與原版戰損／seed；正常玩家音效沿既有fleetplayerchain_test補驗。
- [正式音效解碼驗證](../../cmd/gin3/effects_test.go)：28段OGG雜湊、有效結束位置與完整PCM串流，不能替代裝置或人耳驗收。
- [音效驗證067](../playtest/067-fleet-sound.md)、[Game完整索引](data/fleet-sound-game.json)及[Game凍結器](../../tools/scripts/fleet_sound_game_receipt.py)：正常兩陣營、正式app事件及六份保存，分開硬體近似與未驗GUI／聆聽。
- [音效轉檔入口](../../tools/sound_to_ogg.sh)：沿既有GUI／音訊Docker映像，產生本機28段OGG並拒絕覆寫。
- [第79輪真實GUI交戰](../goal/2026-10-04-fleet-player-gui-round-79.md)、[只讀GUI overlay](../../tools/scripts/fleet_player_gui_prepare.py)及[公開輸入計畫](../../tools/scripts/fleet_player_gui_plan.go)：固定RND、觀測正常階段，不以Game呼叫代替真實鍵鼠。
- [真實GUI驅動器](../../tools/scripts/fleet_player_gui_driver.py)：以X11鍵鼠與狀態等待，核實際視窗、色號與診斷保存；不偽造原版或聲音裝置證據。
- [實際X11擷取](../../tools/scripts/fleet_player_gui_capture.go)：查核window及visual，保留色彩遮罩及像素排列，整數最近鄰正規化後保存完整差分。
- [常駐XTEST輸入](../../tools/scripts/fleet_player_gui_input.go)：使用既有X11相依套件送實際按下／放開，觀測器核送達；不呼叫Game輸入API，完整保持動作另記。
- [正常GUI交戰入口](../../tools/fleet_player_gui.sh)及[容器內Xvfb生命週期](../../tools/fleet_player_gui_xrun.sh)：沿既有映像，二倍視窗另開1600×1200顯示並有界清理，不裁掉960高度的下方。
- [實際原生全圖核對](../../tools/scripts/fleet_player_gui_native_check.go)：逐點核所有X11像素，一倍307200、二倍1228800；不只抽每個放大區塊的首點。
- [GUI驗證068](../playtest/068-fleet-player-gui.md)、[GUI凍結器](../../tools/scripts/fleet_player_gui_receipt.py)與[完整GUI索引](data/fleet-player-gui.json)：兩陣營、一倍／兩倍正常X11、Game計畫及六份診斷保存，與原版及人耳門檻分開。
- [第80輪正式播放](../goal/2026-10-04-fleet-playback-round-80.md)與[GUI播放映像來源](../../tools/Dockerfile.gui-playback)：固定PulseAudio虛擬輸出，補正式後端PCM與正常GUI音效；真機／人耳另驗。
- [正式播放驗證入口](../../tools/playback.sh)、[虛擬輸出與擷取](../../tools/scripts/playback_sink.py)及[ALSA裝置探針](../../tools/scripts/alsa_device_probe.go)：無/dev/snd時核真正default PCM能否開啟，不生成合成代用品，子程序有界清理。
- [真實播放曲庫組合](../../tools/scripts/playback_library.py)及[monitor完整樣本核對](../../tools/scripts/playback_stats.py)：保持22首／28段來源雜湊，非靜音、峰值及NaN全樣本核對；人耳另驗。
- [Linux實際裝置判定](../../cmd/gin3/audio_device_linux.go)及[實際ALSA測試](../../cmd/gin3/audio_device_linux_test.go)：依104補PulseAudio可用端點，其他平台沿[原本判定](../../cmd/gin3/audio_device_other.go)。
- [正常GUI播放觀測](../../tools/scripts/fleet_playback_prepare.py)：沿第79輪真實X11，開正式音效等待，只記實際播放器生命週期及輸出連線，舊工具不改。
- [播放來源凍結器](../../tools/scripts/fleet_playback_receipt.py)與[完整播放索引](data/fleet-playback.json)：綁固定映像、正式PCM、正常兩陣營及未解音效收尾，實體音效卡與人耳仍待驗。
- [正式播放驗證069](../playtest/069-fleet-playback.md)：Linux裝置誤判修正、正常X11混音輸出與四次音效收尾觀測；不升格真機、聆聽或完整原版同態。
- [第81輪音效尾端](../goal/2026-10-04-sound-tail-round-81.md)與[正式PCM局部探針](../../tools/scripts/sound_tail_probe_prepare.py)：由真正ALSA輸出核28段與舊替換方式，只讀Oto buffer，不合成測試音源。
- [音效全PCM核對](../../tools/scripts/sound_tail_compare.py)及[正常關配樂GUI](../../tools/scripts/sound_tail_gui_prepare.py)：由正式熱區送公開輸入，沿真正monitor比完整音效，不用Position估計尾端。
- [插入靜音分類](../../tools/scripts/sound_tail_gaps.py)保留全檔失敗，再核整段尾端；[常駐XTEST放開](../../tools/scripts/sound_tail_input.go)由正式輸入邊緣觸發，不改遊戲時鐘。
- [105正式音效完成回報](../spec/105-sound-completion.md)：104的完成等待接到實際後端，PCM與buffer共同核截短。
- [第81輪播放索引生成](../../tools/scripts/sound_tail_receipt.py)：固定修正前後完整PCM、buffer、正常GUI與保存來源。
- [驗證070正式音效完整尾端](../playtest/070-sound-tail.md)：正常兩陣營關配樂，核完整來源與真正後端完成。
- [正式音效尾端索引](data/sound-tail.json)：521份來源、修正前後PCM及正常玩家完整輸入，綁父提交1993f1a。

- [戰術星域歸屬研究416](../re/416-battle-ownership.md)與[106子規格](../spec/106-battle-region-ownership.md)：三方行星數、通知等待、歸屬寫入及費沙登記時點。
- [正式8-2規則及原版回歸](../../internal/game/battleownership_test.go)：156案完整記錄及無遮罩報告矩形；完整062仍DRAFT。
- [正常首戰Game基準](../../tools/scripts/battle_ownership_player_entry.go)：同版本bootstrap及公開輸入，供既有turn_counter_gui.sh一倍／兩倍回歸；不宣稱正常占領報告。

- [108正式升級資料生成](../../tools/scripts/battle_promotion_spec.py)、[唯一軍階與門檻表](data/admiral-promotion-spec.json)、[原池升級規則](../../internal/rules/promotion.go)與[Game接線](../../internal/game/promotion.go)：固定證據輸入、原子回寫五表、跨換手路徑快取及逐則等待。
- [升級規則對拍](../../internal/rules/promotion_test.go)、[Game報告與保存對拍](../../internal/game/promotion_test.go)及[正常玩家重播](../../tools/scripts/battle_promotion_player.go)：固定原版逐次五表與亂數；正常公開入口另驗續接與存進度。
- [升級正式來源凍結器](../../tools/scripts/battle_promotion_game_receipt.py)、[完整產品來源](data/battle-promotion-game.json)與[驗證076](../playtest/076-battle-promotion-game.md)：控制案原版對拍、正常玩家續接、首戰GUI及尚未驗收的輸入鏈分開記錄。
- [第88輪清空分隊輸入鏈](../goal/2026-10-05-battle-cleanup-round-88.md)與[8-1前原版保存重生](../../tools/battle_cleanup_entry_oracle.sh)：固定第五回合自然到185C:001D，獨立保存，不覆寫既有8-2父狀態。
- [清空分隊觀測器生成](../../tools/scripts/battle_cleanup_prepare.py)、[原版雙跑](../../tools/battle_cleanup_oracle.sh)與[原版模組執行器](../../tools/battle_cleanup_native_run.sh)：七艦隊、五種死亡位置，完整原池及原始呼叫順序，不改正式程式。
## 第88輪清空分隊證據

- [109清理契約](../spec/109-battle-empty-groups.md)、[研究422](../re/422-battle-empty-groups.md)、[077驗證](../playtest/077-battle-empty-groups.md)。
- [完整原版來源與指令分級](data/battle-cleanup.json)、[獨立核對腳本](../../tools/scripts/battle_cleanup_review.py)。
- [正式Game清理回歸](../../internal/game/battlecleanup_test.go)。
- [原池清理規則](../../internal/rules/battlecleanup.go)與[Game同步](../../internal/game/battlecleanup.go)。
- [純規則原版回歸](../../internal/rules/battlecleanup_test.go)。
- [清理至升級鏈產生器](../../tools/scripts/battle_cleanup_chain_prepare.py)與[固定原版雙跑](../../tools/battle_cleanup_chain_oracle.sh)。
- [產品來源凍結器](../../tools/scripts/battle_cleanup_game_receipt.py)與[完整產品收據](data/battle-cleanup-game.json)。

## 第92輪入口

- [正式戰場結束實作](../goal/2026-10-05-battle-finish-implementation-round-92.md)
- [正式勝負與占領規則](../../internal/rules/battlefinish.go)
- [固定原版回歸](../../internal/rules/battlefinish_test.go)
- [戰場結束 Game 接線](../../internal/game/battlefinish.go)
- [報告原始格式擷取](../../tools/scripts/battle_finish_formats.py)
- [Game報告與保存回歸](../../internal/game/battlefinish_test.go)
- [報告讀取端研究426](../re/426-battle-finish-game.md)與[正式驗證081](../playtest/081-battle-finish-game.md)
- [正式收據凍結器](../../tools/scripts/battle_finish_game_receipt.py)與[完整來源](data/battle-finish-game.json)

## 第93輪正常玩家探索

- [計畫93](../goal/2026-10-05-battle-finish-player-round-93.md)與[自然勝負探針生成器](../../tools/scripts/battle_finish_player_probe.py)。
- [旗艦行動碼研究427](../re/427-fleet-order-flagship-action.md)、[原版準備器](../../tools/scripts/fleet_order_action_prepare.py)與[原版雙跑](../../tools/fleet_order_action_oracle.sh)。
- [行動碼來源收據](data/fleet-order-action.json)與[重播／凍結器](../../tools/scripts/fleet_order_action_receipt.py)。
- [公開命令與保存驗證082](../playtest/082-fleet-order-flagship-action.md)。
- [正常首戰旗艦GUI驗證器](../../tools/scripts/fleet_order_action_gui.py)。

## 第94輪正常勝負

- [計畫94](../goal/2026-10-05-battle-finish-natural-round-94.md)與[公開玩家勝負探針](../../tools/scripts/battle_finish_victory_probe.py)。
- [正常勝負重播與收據工具](../../tools/scripts/battle_finish_victory_receipt.py)。
- [正常勝負研究428](../re/428-battle-finish-natural.md)、[玩家驗證083](../playtest/083-battle-finish-natural.md)與[完整來源](data/battle-finish-natural.json)。

## 第95輪正常結算

- [計畫95](../goal/2026-10-05-battle-lifecycle-player-round-95.md)與[正常結算追蹤工具](../../tools/scripts/battle_lifecycle_player_receipt.py)。
- [全滅與公開存讀研究429](../re/429-battle-lifecycle-player.md)、[玩家驗證084](../playtest/084-battle-lifecycle-player.md)與[完整來源](data/battle-lifecycle-player.json)。

## 第96輪動畫後標題

- [計畫96](../goal/2026-10-05-title-player-round-96.md)，沿GitHub #68／#70收斂標題解碼及正常入口。
- [標題圖規格111](../spec/111-title-image.md)、[研究430](../re/430-title-image.md)及[保存素材收據](data/remaining-assets.json)。
- [五類素材匯出器](../../tools/scripts/remaining_assets_export.go)，共用正式解碼器並核對PNG重讀的尺寸、色號及palette。
- [保存素材驗證085](../playtest/085-original-asset-exports.md)與[PNG完整性回歸](../../internal/asset/export_catalog_test.go)。
- [標題與匯出收據核對器](../../tools/scripts/title_asset_receipt.py)與[完整驗證來源](data/title-asset-receipts.json)。

## 第97輪標題玩家流程

- [計畫97](../goal/2026-10-05-title-player-round-97.md)與[原始載入器IDA匯出](../../tools/scripts/title_loader_ida.py)，沿#68量末端等待再接正式玩家路徑。
- [等待與重播研究431](../re/431-title-player.md)、[玩家規格112](../spec/112-title-player.md)與[玩家驗收086](../playtest/086-title-player.md)。
- [正式標題流程](../../internal/game/title.go)、[正常bootstrap驗收](../../internal/bootstrap/title_player_test.go)與[app配樂續接回歸](../../cmd/gin3/title_player_test.go)。
- [正常60 Hz GUI流程](../../tools/scripts/title_player_gui.py)，沿用[完整X11擷取器](../../tools/scripts/fleet_player_gui_capture.go)核圖。
- [原版全圖與區域比較器](../../tools/scripts/title_player_compare.go)保存未縮放頁面、補邊及差分。
- [分級定位索引](data/title-player-annotations.json)合併進原始IDA匯出，保留原始bytes與未知警示。
- [完整來源核對器](../../tools/scripts/title_player_receipt.py)與[完整玩家來源](data/title-player.json)。

## 第101輪正常索敵流程

- [計畫101](../goal/2026-10-05-scout-player-round-101.md)：沿#79補公開索敵、完整回合、失聯、冷卻及同池再派出。
- [只讀觀測器生成](../../tools/scripts/scout_player_prepare.py)與[公開輸入探針](../../tools/scripts/scout_player_probe.go)。
- [正常60 Hz真實GUI驗證](../../tools/scripts/scout_player_gui.py)沿用正式視窗、X11事件與完整畫面擷取。
- [報告游標研究435](../re/435-scout-player-report-cursor.md)、[玩家驗證089](../playtest/089-scout-player.md)及[完整來源](data/scout-player.json)。
- [差分與來源核對器](../../tools/scripts/scout_player_receipt.py)保存修正前後全圖與報告矩形，不隱藏背景差異。

## 第102輪正式60 Hz戰鬥長流程

- [計畫102](../goal/2026-10-05-battle-lifecycle-60hz-round-102.md)與[目前程式重播工具](../../tools/scripts/battle_lifecycle_60hz.py)，沿#79補正常全滅、敗退及公開存讀的正式更新頻率。
- [完整收據核對器](../../tools/scripts/battle_lifecycle_60hz_receipt.py)核對獨立雙跑、實際每更新TPS、原生畫面及六檔公開存讀。
- [玩家驗證090](../playtest/090-battle-lifecycle-60hz.md)及[完整來源](data/battle-lifecycle-60hz.json)保存範圍與限制。

## 第103輪目前來源交付準備

- [計畫103](../goal/2026-10-05-current-source-round-103.md)、[來源規格116](../spec/116-current-source.md)及[乾淨來源封裝器](../../tools/scripts/remake_source_kit.py)，沿#69補目前玩家及音樂Go入口的離線重建。
- [離線重建器](../../tools/scripts/remake_source_rebuild.py)核對完整二進位、真正dosgolem身份及正常片頭。
- [來源執行抽驗](../../tools/scripts/remake_source_runtime.py)補正常動畫至標題及重建程式的正式FM匯入。
- [玩家及音樂來源驗證091](../playtest/091-current-source.md)、[完整來源](data/current-source.json)及[來源核對器](../../tools/scripts/remake_source_receipt.py)。

## 第104輪 Windows 玩家入口

- [計畫104](../goal/2026-10-05-windows-player-round-104.md)，沿 #69 抽測來源重建的 Windows 正常 GUI。
- [Wine 玩家包裝器](../../tools/player_wine.sh)沿用既有 Wine image、唯讀原版與資料及有界 Xvfb。
- [玩家抽驗工具](../../tools/scripts/windows_player_check.py)保存預設片頭、真實 X11 輸入及同來源 Linux 完整畫面對照。
- [Windows玩家驗證092](../playtest/092-windows-player.md)分開原生畫布、一倍視窗及1.5倍量化差異。
- [完整像素核對](../../tools/scripts/windows_player_pixels.py)、[來源凍結器](../../tools/scripts/windows_player_receipt.py)與[唯一來源收據](data/windows-player.json)保存通過範圍及完整反例。

## 第105輪 Windows 正常播放

- [計畫105](../goal/2026-10-05-windows-playback-round-105.md)，沿 #69 驗正式Windows播放器輸出。
- [Wine音訊包裝器](../../tools/player_wine_audio.sh)及[虛擬輸出擷取](../../tools/scripts/wine_playback_sink.py)沿既有image連接隔離Unix socket。
- [Windows正常配樂玩家路徑](../../tools/scripts/windows_music_player.py)驗預設片頭、主選單停止、平時曲切換及關閉。
- [正式Vorbis解碼對照](../../tools/scripts/windows_music_decode.go)保存signed16來源及相同的float轉換。
- [原首段核對](../../tools/scripts/windows_music_compare.py)保持3秒連續相同及停止門檻；[完整固定曲段觀測](../../tools/scripts/windows_music_analysis.py)另記缺口，不取代首段FAIL。
- [直接PCM對照](../../tools/scripts/windows_music_pulse_control.py)、[Windows Oto探針](../../tools/scripts/windows_music_oto_probe.go)、[Oto容器控制器](../../tools/scripts/windows_music_oto_control.py)與[全來源核對](../../tools/scripts/windows_music_control_compare.py)只作環境診斷。
- [音訊映像來源](../../tools/Dockerfile.windows-audio)、[50套件固定鎖](../../tools/windows-audio-packages.lock)及[建置入口](../../tools/build_windows_audio.sh)延伸既有Wine父映像；[隔離driver設定](../../tools/scripts/wine_audio_driver.py)不改遊戲MIDI I/F。
- [093觀測](../playtest/093-windows-music-playback.md)、[來源核對器](../../tools/scripts/windows_music_playback_receipt.py)及[本輪來源](data/windows-music-playback.json)保存e首段失敗、f限定內段與正常換曲／停止。
- 093的WASAPI名稱屬鎖版Oto優先分支強推論，實際API選擇未直接追蹤；勘誤及現行範圍見[CONTEXT](../../CONTEXT.md)。Windows PCM／Wine ALSA／Pulse端點及波形有實測。

## 第110輪登陸分段接線

- [計畫110](../goal/2026-10-06-landing-caller-round-110.md)：沿#48補齊反擊及跨攻方獨立快照，依118閘門接正式演出。
- [各層快照準備器](../../tools/scripts/landing_caller_prepare.py)與[完整幀探針](../../tools/scripts/landing_caller_probe.go)依原版寫入者及逐格返回建立獨立文字、量表、圖段與下方場景。
- [計時與格提交IDA匯出](../../tools/scripts/landing_caller_ida.py)、[研究438](../re/438-landing-caller.md)及[驗證098](../playtest/098-landing-caller.md)保留首批差分與原指令勘誤。
- [原版文字規格生成](../../tools/scripts/landing_caller_formats.py)與[正式完整幀renderer](../../internal/screen/planet_frame.go)使用各層獨立typed快照。
- [候選renderer驗證](../../tools/scripts/landing_caller_candidate.go)直接載入獨立候選規格，尚未接Game。
- [本輪唯一來源](data/landing-caller.json)與[來源核對器](../../tools/scripts/landing_caller_receipt.py)記錄候選、原始RAM、首批差分及父提交歷史驗證。
- [#91鎖定向量診斷](../../tools/scripts/planet_eligibility_probe.go)只比較固定原版父RAM與目前Targets，未重跑0091或修改規則。

## 第111輪行星鎖定

- [計畫111](../goal/2026-10-06-planet-eligibility-round-111.md)與[規格119](../spec/119-planet-assault-eligibility.md)，沿#91重跑原版獨立鎖定條件。
- [原版矩陣入口](../../tools/planet_eligibility_oracle.sh)、[隔離模組入口](../../tools/planet_eligibility_native_run.sh)及[原版CPU注入探針](../../tools/scripts/planet_eligibility_native.go)保存完整注入，不修改dosgolem或原始遊戲檔。
- [原版向量與結算生成器](../../tools/scripts/planet_eligibility_receipt.py)及[結構化原版事實](data/planet-eligibility.json)從新執行的原版RAM生成預期值。
- [原版矩陣及只受反擊回歸](../../internal/rules/planet_eligibility_test.go)核對目標、位元、分隊、行星及最終亂數。
- [研究439](../re/439-planet-eligibility.md)及[驗證099](../playtest/099-planet-eligibility.md)記錄本輪分級與回歸範圍。
- [Game政策與動畫模式回歸](../../internal/game/planet_eligibility_test.go)驗雙陣營被動CPU正對照及普通／快速／最快消費。

## 第112輪正常勝後占領

- [計畫112](../goal/2026-10-06-winner-occupation-player-round-112.md)與[正常輸入探針](../../tools/scripts/battle_winner_player_probe.py)沿#79及110 READY補勝方占領的玩家流程。
- 第112輪正常勝利的獨立bootstrap及60 Hz公開存讀入口：[重播工具](../../tools/scripts/battle_winner_player_replay.py)。占領驗收仍未完成。
- [驗證100](../playtest/100-battle-winner-player-search.md)、[完整來源與探索分類](data/battle-winner-player-search.json)及[來源核對器](../../tools/scripts/battle_winner_player_receipt.py)保留晚回合零占領與正式視窗回歸。

## 第113輪旗艦接近

- [計畫113](../goal/2026-10-06-flagship-approach-round-113.md)及[正常輸入探針](../../tools/scripts/battle_flagship_approach_probe.py)核對公開行進目標與旗艦優先策略，不更改產品規則。
- [公開編組重播](../../tools/scripts/battle_flagship_composition_replay.py)補獨立bootstrap、正式60 Hz視窗與公開存讀。
- [驗證101](../playtest/101-battle-flagship-approach.md)、[來源索引](data/battle-flagship-approach.json)及[核對器](../../tools/scripts/battle_flagship_approach_receipt.py)保存早期勝利仍零占領的核心結果、完整失敗與編組證據。

## 第114輪正常增援

- [計畫114](../goal/2026-10-06-battle-reinforcement-round-114.md)及[公開增援探針](../../tools/scripts/battle_reinforcement_probe.py)沿#79驗正常移動、占領命令與勝後占領事件。
- [獨立重播](../../tools/scripts/battle_reinforcement_replay.py)驗正式入口、60 Hz視窗與公開保存讀回。
- [來源核對器](../../tools/scripts/battle_reinforcement_receipt.py)核對原始收據、占領等待及存讀狀態。
- [唯一來源索引](data/battle-reinforcement.json)凍結正常雙人增援、勝後占領及公開存讀的輸入與輸出，保留未完成GUI批次及原始標籤勘誤。
- [驗證102](../playtest/102-battle-reinforcement.md)保存正式資源缺項、失敗分類及實際勝後占領範圍。

## 第115輪自然功績接續

- [計畫115](../goal/2026-10-06-battle-natural-promotion-round-115.md)沿#79補正常升級、軍階UI與存讀，108限定CONFORMED。
- [正常接續探針](../../tools/scripts/battle_natural_promotion_probe.py)從第114輪公開存檔走正式載入、艦隊命令及戰術實行，不注入功績。
- [驗證103](../playtest/103-battle-natural-promotion.md)保存實際選單門檻、未形成戰場與自然升級接續的證據限制。
- [正式入口與GUI重播](../../tools/scripts/battle_natural_promotion_replay.py)核對自然升級報告、軍階視窗及公開存讀。
- [來源核對器](../../tools/scripts/battle_natural_promotion_receipt.py)凍結本輪輸入、全圖及六檔存讀，保存於[唯一來源索引](data/battle-natural-promotion.json)。

## 第116輪自然空分隊清理

- [計畫116](../goal/2026-10-06-battle-natural-cleanup-round-116.md)沿#79及109 READY補正常一般分隊死亡、引用搬移與存讀。
- [公開選敵探針](../../tools/scripts/battle_natural_cleanup_probe.py)只改測試玩家的公開選敵，唯讀觀測清理前後，不注入死亡。
- [驗證104](../playtest/104-battle-natural-cleanup.md)保存自然清理、正常GUI與原版整回合同態的範圍區分。
- [正式重播](../../tools/scripts/battle_natural_cleanup_replay.py)及[唯讀觀測片段](../../tools/scripts/battle_natural_cleanup_observer.go)驗正式bootstrap與60 Hz視窗的清理前後完整表。
- [來源核對器](../../tools/scripts/battle_natural_cleanup_receipt.py)沿109已證實byte契約核對中間槽搬移、下一回合及完整六檔，保存於[唯一來源](data/battle-natural-cleanup.json)。
- [109限定CONFORMED](../spec/109-battle-empty-groups.md)沿原版35案與本輪正常Game／GUI／存讀鏈；助手取圖時點另由公開右鍵後完整圖核實，062保持DRAFT。

## 第117輪正常首戰整回合

- [計畫117](../goal/2026-10-06-battle-normal-turn-round-117.md)沿#79，由第55輪正常原版完整保存自然續跑，不改CPU入口。
- 現行入口與原版保存身份見[研究393](../re/393-battle-work-control.md)；完整088與062保持DRAFT。
- [正常原版續行入口](../../tools/battle_normal_turn_oracle.sh)、[容器內入口](../../tools/battle_normal_turn_native_run.sh)與[只讀步驟觀測器](../../tools/scripts/battle_normal_turn_native.go)固定父保存與首戰seed，僅以公開滑鼠續跑。
- [正常Game觀測覆蓋](../../tools/scripts/battle_normal_turn_prepare.py)與[公開玩家探針](../../tools/scripts/battle_normal_turn_probe.go)從bootstrap、主選單及S3實際輸入續跑，完整保留每步原表與分隊。
# 第119輪交付入口

- [計畫119](../goal/2026-10-06-macos-retro-promo-round-119.md)及[macOS與新版推廣片106](../playtest/106-macos-and-retro-promo.md)。
- [macOS隔離入口](../../tools/macos.sh)與[兩架構建置器](../../tools/scripts/local_test_macos_build.sh)沿既有Go1.24.13映像。

## 第123輪三劇本、實錄影片與HD介面

- [計畫123](../goal/2026-10-11-hd-interface-and-s1-promo-round-123.md)列操作介面及動畫HD的分批工作與樣式閘門。
- [129劇本目錄契約](../spec/129-bundled-scenario-catalog.md)、[封包三劇本抽驗](../../tools/scripts/bundled_scenarios_smoke.py)與[驗收110](../playtest/110-hd-interface-s1-promo.md)保存個人索引空尾格的修正。
- [正常X11實錄器](../../tools/scripts/s1_gameplay_record.py)、[有界顯示](../../tools/s1_record_display.sh)、[雙旗故事製片](../../tools/scripts/s1_story_promo.py)與[交付核對器](../../tools/scripts/s1_finalize.py)保存來源、剪接及影音檢查。
- [素材盤點器](../../tools/scripts/hd_interface_inventory.py)、[可操作介面模板](../../tools/scripts/hd_interface_preview.html)與[瀏覽器抽驗](../../tools/scripts/hd_interface_browser_check.py)保持與正式遊戲獨立。
- [瀏覽器增補映像](../../tools/Dockerfile.hd-preview)及[套件下載核對器](../../tools/scripts/hd_preview_deps.py)延伸既有影片工具；精確套件與SHA-256收據由映像保存。
- [唯一來源索引](data/s1-gameplay-promo.json)記錄本次修正版、正常實錄、素材盤點與試樣的驗收範圍。

## 第124輪原介面與地圖HD

- [計畫124](../goal/2026-10-11-original-layout-hd-round-124.md)記錄A風格、原介面配置及戰略／戰術／小地圖HD範圍，取代前輪尚待選型的現行說法。
- [原配置試樣產生器](../../tools/scripts/hd_original_layout_preview.py)只寫入workplace，沿現行JSON、星域遮罩及正常S1截圖製作；不接正式Game或存檔。
- [原配置瀏覽器抽驗](../../tools/scripts/hd_original_layout_check.py)核對疊窗與六鈕矩形、32星域選取、鍵盤返回、文字格及4:3比例，另產生目視截圖。

## 第126輪原版介面與敵軍AI展示

- [計畫126](../goal/2026-10-11-original-ui-ai-promo-round-126.md)是本輪唯一狀態表，保留既有交付並記錄外觀修正及影片要求。
- [規格134](../spec/134-enemy-visibility-display.md)定義預設關閉、遊戲中切換的敵軍全顯示，以及顯示與規則的資料邊界。
- [規格135](../spec/135-original-style-hd-ui.md)沿使用者確認的原配色、原框線及原字模輪廓強化方向。
- [四款陣營HTML概念產生器](../../tools/scripts/faction_hd_concepts.py)保留各兩款比較。第2版依使用者選擇，帝國B改亮藍／黃金，同盟A改淡黃／淡赤紅；[瀏覽器核對](../../tools/scripts/faction_hd_concept_check.py)檢查四畫面、原色切換、試點框與對齊線。本機核准概念入口為`workplace/hd126/FACTION-HD-CONCEPTS.html`，正式接線的證據見113。
- [驗證112](../playtest/112-original-ui-concepts-and-enemy-display.md)與[來源索引](data/original-ui-concepts.json)保存68張原框／原字模正常視窗、敵軍顯示測試、HTML第2版及待完成的AI錄影範圍。
- [正式陣營主題資料](../../assets/hd/interface-themes.json)沿使用者核准的[帝國B原PNG](../../assets/hd/interface-empire-approved.png)與最新[同盟B原PNG](../../assets/hd/interface-alliance-approved.png)，配合135追加的戰略地圖主題契約；只供顯示，不改規則或存檔。
- [正式陣營外觀核對器](../../tools/scripts/faction_hd_check.py)比較核准HTML的四畫面與正式呈現，保存完整差圖與已分類差異，並產生兩陣營的離線實際畫面頁。
- [正式主題驗證113](../playtest/113-faction-themes.md)及[凍結來源](data/faction-themes.json)保存改選前同盟A的歷史收據；最新同盟B見114，影片與封包見115。本機實際畫面頁為`workplace/hd126/FACTION-HD-GAME.html`。
- [AI影片模型](../../tools/scripts/ai_promo_model.go)從公開S3載入兩種模式，保存相同初始六表與正常逐回合圖，供實際X11錄影核對；不能單獨代替實錄或DOS原版證據。
- [AI與S1正常實錄](../../tools/scripts/ai_gameplay_record.py)沿正式遊戲、X11鍵鼠及有時鐘虛擬音訊，同次錄製遊戲畫面、配樂與音效；可用模型逐張核完整視窗。

## decomp 與編譯器歸屬

- [`66-fleet-attack-203a-0424.md`](66-fleet-attack-203a-0424.md)：戰略艦隊攻擊力 `203A:0424` 的精確語意與 16 位元溢位區。
- [研究442](../re/442-compiler-is-datalight-c.md)：`GIN3PS.EXE` 的 C runtime 是 Datalight C 3.10 Large，不是 Borland。
- [研究443](../re/443-fleet-attack-differential-test.md)：`203A:0424` 的差異測試收據。
- [規格124](../spec/124-decomp-reference-oracle.md)：C 參考實作作為規則 oracle 的驗收契約，C 程式在 [`decomp/`](../../decomp/README.md)。
- [規格125](../spec/125-repair-ship-repairable.md)：修護艦可被其他修護艦修理（可選規則，預設關閉）。

- [戰略AI推廣片正常新局模型](../../tools/scripts/strategy_ai_promo_model.go)：正常實行及存讀換陣營檢視電腦部署，供第126輪實錄核對。

- [陣營正式HTML瀏覽器核對](../../tools/scripts/faction_hd_game_browser_check.py)：最新帝國B／同盟B的離線圖片、切換與手機版面。

- [雙旗HD故事及實戰比較影片](../../tools/scripts/faction_story_promo.py)：有界逐幀片頭、正式封包錄音與來源影格核對，僅本機交付。

- [114 同盟赤焰戰線](../playtest/114-alliance-fireline.md)：最新同盟B與戰略星圖，68個正常視窗及[來源索引](data/alliance-fireline-theme.json)。

- [114來源凍結工具](../../tools/scripts/faction_fireline_finalize.py)：核對同盟B的正常收據並保存來源副本。

- [115交付核對工具](../../tools/scripts/faction_delivery_finalize.py)：核對最新陣營素材、封包、正常實錄及故事影片，將本機交付集中於`dist-all/v.1.1.2-20261011/`。

- [115雙陣營HD與AI實錄交付](../playtest/115-faction-hd-ai-promo.md)及[來源索引](data/faction-hd-ai-delivery.json)：最新同盟B、戰場F7、兩種AI比較、118秒HD故事影片與新版封包。

- [第127輪目前狀態](../goal/2026-10-11-map-icons-cadence-round-127.md)：戰鬥節奏、原版圖示／地圖比較與即時切換。
- [規格136](../spec/136-live-graphics-and-cadence.md)：F6即時顯示切換、視覺快取及交火時鐘，依116限定CONFORMED。

- [規格137](../spec/137-original-symbols-and-clear-map.md)：已確認戰術圖示與待選星圖前景保護。
- 第三稿本機比較頁：`workplace/hd127/CLARITY-COMPARISON.html`。新繪星球／徽章及完整提示、雜湊在`workplace/hd127/map-sprites/`，原版地名與航路保留，使用者已明確核准第三稿。
- [正常切換與節奏核對器](../../tools/scripts/live_graphics_audit.py)：從正式視窗的唯讀顯示收據、X11事件及同次錄音驗F6、畫面復原與原生來源格間隔。
- [驗證116](../playtest/116-live-graphics-review.md)與[來源索引](data/live-graphics-review.json)：正常S1六次F6、單調時鐘修復、全部星圖地名保護及第三稿限定驗收。

第127輪核准第三稿：[星圖素材](../../assets/hd/strategy/MANIFEST.json)、[提示](../../assets/hd/strategy/PROMPTS.json)、[正式畫面逐像素核對工具](../../tools/scripts/approved_map_check.go)。驗收契約見[規格137](../spec/137-original-symbols-and-clear-map.md)。

第127輪正式交付：[來源索引](data/live-graphics-review.json)、[交付核對工具](../../tools/scripts/live_graphics_finalize.py)，沿[驗收116](../playtest/116-live-graphics-review.md)。
## 第128輪驗證工具

第129輪接續入口：[唯一狀態表](../goal/2026-10-12-battle-playback-round-129.md)、[規格139](../spec/139-battle-playback-and-settings.md)。敵軍顯示改為設定選單，另核對正常模式戰鬥播放。

[驗證118](../playtest/118-battle-playback-and-settings.md)保存正常S1／S3播放、設定返回修正、工具時序與交付限制。

[正常播放唯讀觀測](../../internal/game/playbackobservation.go)經既有[正常錄影工具](../../tools/scripts/ai_gameplay_record.py)保存段名、格號及時點，不注入局面或控制遊戲時鐘。

[正常動畫順序核對器](../../tools/scripts/battle_playback_audit.py)檢查實際Draw觀測的逐格順序、音效閘門及最後一格後才顯示結果。

[本輪交付核對器](../../tools/scripts/battle_delivery_finalize.py)在正常播放、設定與封包全通過後產生[唯一來源索引](data/battle-playback-settings.json)。

[正常艦隊計畫準備](../../tools/scripts/battle_fleet_prepare.py)沿既有公開玩家鏈，保持現行活動AI並停用舊種子注入；觀測overlay不更動戰鬥規則或時間。

[正常艦隊錄影](../../tools/scripts/battle_fleet_record.py)沿第79輪鍵鼠與鏡頭回授工具，追加正式聲音、逐格觀測及F7顯示切換；診斷保存另行標示。

[頂端功能列正常視窗驗收](../../tools/scripts/toolbar_gui_check.py)沿片頭、主選單、財政、訂造、建造及戰略畫面，以X11鍵鼠輸入檢查規格138。

[功能列與反光來源索引](data/top-toolbar-hover.json)只在本次驗收全通過後記錄正式八包交付，前版影片身份沿驗證116。

[功能列交付核對器](../../tools/scripts/toolbar_delivery_finalize.py)核對八封包、兩陣營正式啟動器、反光亮度、標準測試與來源身份。

[頂端功能列驗收117](../playtest/117-top-toolbar-hover.md)保存已核准A配色、正常啟動器流程、反光修正及前版影片邊界。

## 第130輪名將戰術與發行

第131輪目前入口：[狀態表131](../goal/2026-10-13-system-dialog-and-polish-round-131.md)、[規格141系統視窗](../spec/141-system-dialog-cursor.md)、[驗證120](../playtest/120-system-dialog-polish.md)及[交付索引](data/system-dialog-polish.json)。保存游標與取消修正、四字型、原曲HD音色、啟動進度及v.1.3.0三平台交付。

[規格142啟動進度](../spec/142-startup-progress.md)限定真實載入階段、小視窗與錯誤顯示，保持正常遊戲入口及原版時鐘。

[規格143字型與HD音樂](../spec/143-font-and-hd-music.md)記錄原版點陣／強化／宋體／黑體、原曲升級音色及獨立顯示選項。[載入字集工具](../../tools/scripts/startup_font.py)產生可散布啟動字集，[正常系統視窗驗收](../../tools/scripts/system_dialog_gui_check.py)追蹤載入、取消、存讀及退出。

[HD音樂MIDI匯出](../../tools/scripts/hd_music_export.py)沿既有原曲匯出器，[HD音訊核對](../../tools/scripts/hd_music_check.py)保存OGG格式、音量、循環邊界、來源及權利分類。[啟動字集來源](data/startup-font.json)保存內嵌授權字集的來源與雜湊。

[HD音色合成](../../tools/scripts/hd_music_render.sh)在既有音訊容器執行，輸入原曲MIDI與固定音色庫，輸出44.1kHz雙聲道OGG。

[第131輪最終核對器](../../tools/scripts/system_dialog_finalize.py)驗1553凍結來源、八封包、正常GUI及音樂；[通用發行核對器](../../tools/scripts/release_publish_check.py)以實際下載核對新Release身份，沿[引擎抽驗器](../../tools/scripts/v1_release_smoke.py)檢查外部原版匯入。既有v.1.2.0專用核對器保留歷史契約。

[核准字型試樣](../images/font-v130-preview.html)與[決定索引](../images/font-v130-preview.json)保存使用者選擇；[新版首頁截圖收據](../images/readme-v130.json)保存載入視窗及兩陣營正常設定畫面，沒有裁切或縮放。

[目前狀態表130](../goal/2026-10-12-hero-ai-release-round-130.md)與[規格140](../spec/140-hero-tactics-and-release.md)記錄已核准1.5倍陣營加速、名將包圍、1988動畫來源及新版交付。

[可撤回戰術試樣產生器](../../tools/scripts/hero_tactics_preview.py)只產生規則示意HTML，不接正式遊戲；本機入口為`workplace/hero130/HERO-TACTICS-PREVIEW.html`。

[舊版動畫來源保存器](../../tools/scripts/hero_promo_sources.py)保存官方石黑版人物圖與年代／權利頁，僅供本機試剪，不把角色圖稱為1988正片影格。

[指定動畫下載器](../../tools/scripts/hero_anime_download.py)只下載指定影片可公開讀取的串流，不帶登入憑證；來源原生尺寸與SHA-256保存在本機收據。

[名將整隊戰術](../../internal/ai/hero.go)與[存活帶隊名將接線](../../internal/game/herocommand.go)沿規格140，移動加成只作用於每回合新額度。

[正常名將玩家模型](../../tools/scripts/hero_gameplay_model.go)從主選單讀S1／S3，保存正式輸入與實際艦隊命令，供X11重播；不使用狀態或種子注入。

[雙旗名將故事製片工具](../../tools/scripts/hero_story_promo.py)保留正常遊玩與動畫原速、遊戲聲音、來源雜湊及字幕矩形；指定動畫的原生尺寸另列，不稱原生HD。

[正常名將實錄核對器](../../tools/scripts/hero_gameplay_audit.py)逐回合核對正式封包、公開玩家輸入模型、兩側同初始狀態及F7實際切換。

[名將故事剪輯計畫](../../tools/scripts/hero_promo_plan.py)從完整實戰場次產生可重現時碼，保留普通速度及遊戲聲音。

[名將交付核對器](../../tools/scripts/hero_delivery_finalize.py)只在封包、正常實錄與新片驗收通過後建立[來源索引](data/hero-ai-release-promo.json)。

[驗證119](../playtest/119-hero-ai-release-promo.md)保存1.5倍新額度、兩位名將正常玩家比較、普通動畫、新片來源及三平台交付限制。

[影片低變動回查](../../tools/scripts/hero_promo_review.py)逐區間核對來源與成片，不把固定背景、選單停留或音效等待直接當成異常凍結。

[GitHub發行下載核對](../../tools/scripts/hero_release_publish_check.py)先核草稿五資產的實際下載雜湊，正式發布後再核遠端tag、說明及完成訊號；完整遊戲與影片不列入遠端資產。

[README實錄截圖工具](../../tools/scripts/readme_release_shots.py)由本版正常玩家實錄保存八張完整視窗，原圖／HD切換兩張核對同戰術狀態；[來源收據](../images/readme-v120.json)連結正式程式、時碼與雜湊。
