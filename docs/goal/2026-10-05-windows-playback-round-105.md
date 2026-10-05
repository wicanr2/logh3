# 第105輪：Windows 正常曲庫播放

父提交 `62e925a8ee081623144a17317f85fca62579b25b`。前輪已推進Windows GUI。
遠端89張worklist、6 OPEN／83 CLOSED，269條核對沒有失效斷言。
本輪沿已開啟的 #69，不新增平行工作清單。

## 工作

- 沿用既有Wine及虛擬音訊image，以隔離Unix socket連接，無網路、無實體裝置。
- 使用第103輪來源重建的Windows玩家程式及已核Windows OGG曲庫。
- 從正式入口播片頭，再以真實輸入開新局，驗主選單停止及平時配樂。
- 保存WASAPI經Wine輸出的實際PCM、端點身份、原圖、輸入時間戳及完整命令。
- 原版及資料唯讀，不改正式選曲、MIDI I/F、戰術AI或既有取捨。

## 證據限制

命中OGG正式音訊及Wine正常玩家試玩路由，載入
`knowledge-base/local/remake-audio-ogg-default.md`與
`knowledge-base/sources/claude/retro-cht/retro-game-playtest.md`。
Docker-only及文件分工仍依專案AGENTS。

虛擬音訊錄得PCM不等於實體卡、人耳或循環接點驗收。正式版號／dist-all、macOS及
其他玩法閘門仍缺。前輪1.5倍實體像素差異保持完整記錄，本輪不改顯示規則。

## 接續範圍

連續播放核對失敗後命中平台契約停止線，補讀`local/retro-platform-spec-first.md`。
在同一Issue內核固定PCM／Windows音訊後端／正常GUI的對照，保留失敗而不放寬門檻。
既有Wine缺ALSA的Pulse外掛時，延伸父映像並鎖新增依賴；不改產品播放器或音樂合成。
完成範圍以[093觀測](../playtest/093-windows-music-playback.md)及其來源收據為準。
