# 第143輪：提督編輯器完整版、影片與README截圖

使用者要求重新打包完整版、更新推廣影片及README截圖。沿[149](../spec/149-admiral-editor.md)已驗功能，版本為`v.1.5.0-20261019`。保留舊版，完整版及含原版音畫影片留本機，不建立遠端tag或Release。

## 目前狀態表

| 工作 | 狀態 | 驗收 |
|---|---|---|
| 三平台封包 | 七包已建、平台檢查通過 | 同一1590項凍結來源，612項非vendor與目前產品相同；Linux11入口、Windows8條Wine入口、macOS雙架構／11.0／SDK15.5及ARM3253頁簽章通過 |
| 正常GUI與截圖 | 正式ZIP兩平台通過，README19圖已匯出 | Linux及Windows Wine均兩次正常片頭至S1，十三能力／四特殊、取消／套用、F7、重啟與讀檔通過；各11張完整960×720畫面 |
| 推廣影片 | 本機交付及聲畫／目視通過 | 37段101.016667秒、1080p60；新三武器交火27.9667秒，明列沿用後合計51.2833秒；36剪接點單曲連續、AAC SNR33.96dB、零黑段，來源畫格與凍結區間核對通過 |
| 既有Issue | 102張、#48 OPEN | 282條verify無失效；不擴張原版對拍、戰術讀回或平台真機聲明 |

入口：[驗收131](../playtest/131-admiral-editor-delivery.md)。封包沿[Windows凍結建置器](../../tools/scripts/windows_hotfix_package.py)、[跨平台封包器](../../tools/scripts/v140_platform_delivery.py)及[macOS包裝器](../../tools/macos.sh)，錄影沿[正常X11錄影器](../../tools/scripts/ai_gameplay_record.py)及[提督編輯GUI](../../tools/scripts/admiral_editor_gui.py)。

正式包副本、README實錄與交付收據由[本輪核對器](../../tools/scripts/admiral_delivery.py)管理，影片由[製片器](../../tools/scripts/promo_v150.py)沿既有編碼與單曲混音流程建立。

原`psychicwar-osxcross:go1.24.13-15.5-r2`已不存在。沿既有`wizardry7-osxcross:go1.24.13-python3.13.15-15.5-r1`，固定映像`sha256:1aa0f82b2c880d33751bf26259a6453d8c0009fd7f13e593db92897710aa9370`，保留Go1.24.13、SDK15.5、雙架構與最低11.0契約；建置後仍核對Mach-O及ARM簽章全部頁。未重建工具映像。

本輪工作根為`workplace/admiral143/`，實際Linux ZIP解壓於`package/`，`editor-linux/`保存正常GUI與兩段錄影。Windows八入口在`workplace/music69-windows/windows143-delivery/`，正常GUI在`windows143-editor-r2/`。前次`windows143-editor/`因原Wine映像缺xdotool而失敗，保留不作通過收據；改用現有`eob-remake-wine-verify:ubuntu-noble-20261008-r2`，固定`sha256:c047a6a259a1fc9a90babf7c937310fa9df4b3465bf75f3fcaad1bc640099c2f`，同一ZIP與操作重跑通過，沒有安裝或借用主機套件。

最初`gameplay-f0/`與`gameplay-f1/`只到第3回合，保留為短錄影。已查明既有帝國艦隊戰從第4回合、同盟戰鬥機從第3回合開始，故在執行前固定沿原第5回合終點，新來源為`gameplay-f0-final/`及`gameplay-f1-final/`，輸入為`game-f0-final.json`及`game-f1-final.json`。沒有換seed或注入局面。同盟第3回合的多場完整演出用時較長，正常目標輸入期間另存只讀`DIAG.png`，未改遊戲或輸入程式。

兩陣營已完成預定錄影。分鏡初次只使用新素材時，三武器齊全但零跳格完整場次僅27.9667秒，未達既有50秒實戰篇幅。維持完整場次／零跳格判準，不放寬；從v1.4已驗分鏡補足，片中以「前版實戰回顧」及版本明列，來源索引分開新舊。沿用場次只讀原有凍結審計，不重新寫入歷史來源目錄。新README19圖全部取自v1.5正式包；沒有沿用前版README圖。

已完成的重建順序為`promo_v150.py plan`、`hero_story_promo.py prepare --plan /src/workplace/admiral143/PROMO-PLAN.json --out /src/workplace/admiral143/movie`、`promo_v150.py render`。再執行`hero_story_promo.py audit`、`promo_v150.py source_audit`與`promo_v150.py audit`，目視CONTACT及原尺寸編輯器、要塞、片尾後，以`publish --visual-reviewed`交付。README的19張正式包原尺寸圖由`admiral_delivery.py images`匯出並`--check`重核，整體交付由`receipt`建立[唯一索引](../internals/data/v150-delivery.json)。上述Python命令均由既有`logh3-hd-preview:20261011-r1`與`tools/gui.sh`的Docker入口執行。既有交付拒絕覆寫，重跑用獨立輸出及凍結身份。
