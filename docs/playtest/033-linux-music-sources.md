# 033 — Linux 音樂工具固定來源集合驗收

日期：2026-10-01；工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
規格：[073](../spec/073-linux-music-source-closure.md)；證據：[研究 361](../re/361-linux-music-source-closure.md)。

已驗：198 個實際二進位套件對應的 157 個同版本來源集合完整取得；
513 個描述與來源檔、748,421,917 bytes 全數符合固定 SHA-256／大小。
無網路容器全集合重驗成功，全部描述重新解析後來源清單完全相同。
五組錯身份／重複欄位／缺雜湊／路徑越界／來源缺項／碰撞／損壞與缺檔快取邊界測試通過。
本輪只改建置器與文件，沒有變更播放器、遊戲或合成器；不重複宣稱新的原版對拍。
標準 `tools/gui.sh test ./internal/... ./cmd/...` 全部通過，部分套件使用 Go 快取，
cmd/gin3 約 24.24 秒。

## 固定收據與本機產物

- [固定鎖](../internals/data/linux-music-source-lock.json)：
  `f102cbe2028118a0b3e560a2666bf0d733913dcc95f41bd0d86f40c9d7dd7bac`。
- [驗收 JSON](../internals/data/linux-music-sources.json)：
  `737aedbd205f9b0aeea78d39b361c64d8e87e96f957425f5be5b0c92a62b48e8`。
- 本機 tar：`workplace/music69-linux/source-delivery/linux-music-debian-sources.tar`，
  734 項、752,414,720 bytes，雜湊
  `c0bf4070d492d41dcff9cf883c56c3aba90e5117abb84feea3bb2241ba1e3f74`。

tar 已重讀逐項驗全部來源、完整著作權／授權及附加 manifest／建置來源，
沒有解包執行、連結或多餘 entry。封裝前也重驗原 Linux 工具包全部檔案。
快取與 tar 只留本機，Git 保存固定鎖、準備器、測試及收據。

## 重生入口

```sh
# 建置時明示開放網路；只依已提交固定鎖下載。
LOGH3_NET=1 tools/py.sh /work/tools/scripts/music_linux_sources.py --download
# 預設 --network none：全檔 SHA-256／大小及描述集合重驗。
tools/py.sh /work/tools/scripts/music_linux_sources.py
tools/py.sh /work/tools/scripts/music_linux_sources_test.py
# 全部來源及原工具包重驗後封裝；拒絕覆蓋已存在 tar。
tools/py.sh /work/tools/scripts/music_linux_sources.py --bundle
```

前置是[驗證 032](032-linux-music-tools.md) 的實際 BUILD.json 與未變動工具包。
--lock-from-probe 僅供初次描述探針審查後建鎖；日常建置不需探針或重建鎖，
也不可用該旗標繞過錯版本／雜湊拒絕。需要更新版本時先重新審查身份與證據。

## 尚未完成

本機來源集合可供正式封裝接線，但不是 dist-all 正式版號交付。
未逐套重建來源，未驗 Debian 簽章信任鏈；三平台來源／建置方式及完整發行義務仍需
隨正式封包核對。Windows／macOS 音樂工具封裝、實際聆聽與循環接點仍缺，#69 保持開啟。
