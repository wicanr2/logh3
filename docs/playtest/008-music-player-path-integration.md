# 配樂正常玩家路徑整合收據

日期：2026-09-13  
remake Git HEAD：`4133e5c2680b52fad9bbe0e5d8b4a9c6b1038f29`  
GUI image：`logh3-ebiten:2`，manifest SHA-256
`6ccf04cfd896019d83b92641b9c072f65e1403cc4ef41e0afdb09b5ee1bab8f8`  
Ebitengine：`v2.8.8`

## 自動化範圍

```sh
tools/gui.sh test -count=1 ./cmd/gin3 ./internal/audio
tools/gui.sh xrun ./tools/gui_input_check.sh
```

`TestPlayerPathMusicSwitches` 直接使用正式 `Game` 狀態與 `app.syncMusic`，播放器端換成
只記錄狀態、不開音效裝置的實作。輸入路徑為主選單 → 防拷 → 遊戲本體 → 關兩張
月報 → 系統指令 → 設定；實際切換序列是：

```text
BGM_0F → BGM_1F → <stop> → BGM_1F
```

其中改戰鬥曲不會試聽，符合原版設定畫面行為。正式產品尚無獨立的交戰演出狀態，
所以目前沒有可到達、會把 `BattleTune` 送進播放器的玩家路徑；曲目映射由
`TestBattleTune` 與 `TestEveryTuneIsReachable` 客觀驗證，不把尚不存在的畫面冒稱已整合。

## 曲庫與錯誤處理

本機玩家自備曲庫 `workplace/assets/music/ogg/` 有 22/22 個名稱；
`TestAvailableMusicLibraryDecodes` 逐首走與正式播放器相同的 Vorbis 解碼及 44.1 kHz
閘門，全數通過。代表檔案：

| 曲名 | SHA-256 |
|---|---|
| `BGM_0F.ogg` | `0113b315b86b125f3d10050ce670f6c9f8f74d55d66989f84bccdf833873cdc5` |
| `BGM_1F.ogg` | `7f710e859abaf1d898abf5f9003224a2ab483759a9317ee6424d7998e11f2745` |
| `BGM_6F.ogg` | `04e7642666d38f4e3fd5e2527403177b22ec9c3e44e53b97ec88c9447748037e` |
| `OPEN_M.ogg` | `eb5b3d2dcdef82da001319eb6a1c54d63bcee317d02d3bcd9524abf63051895d` |

`TestMusicPlayerMissingAndInvalidFile` 證明缺檔不啟動播放器且不回錯，損壞 OGG 則回
包含曲名的明確解碼錯誤。真實 X11 程式另以 `-music` 指向該曲庫；隔離環境沒有
`/dev/snd` 時印出「沒有音效裝置」，仍留在 `StateMainMenu` 且正常倒圖退出。

## 裁決

- 已客觀驗證曲目選擇、設定即時切換、停止、恢復、同曲不重啟、曲庫解碼、缺檔、
  壞檔與無裝置降級。
- 原版或衍生 OGG 仍只在 `workplace/`，沒有加入 Git 或公開交付。
- 本收據不宣稱音量、音色或人耳聽感通過，也不宣稱尚未實作的交戰演出已接上戰鬥曲。
