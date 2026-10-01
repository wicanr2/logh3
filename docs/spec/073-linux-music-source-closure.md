# 073 — Linux 音樂工具固定來源集合

狀態：**CONFORMED**（2026-10-01，限固定來源集合取得、離線重驗與本機封裝）。
工作項目：[#69](https://github.com/wicanr2/logh3/issues/69)。
證據審查：既有 BUILD.json 的 198 個二進位套件對應 157 組來源套件／版本，
同版本 `.dsc` 已全部取得；描述列出 356 個去重來源檔，747,995,610 bytes。
格式依 [Debian Policy](https://www.debian.org/doc/debian-policy/ch-controlfields.html#debian-source-control-files-dsc)。

1. 唯一輸入為 072 的實際 BUILD.json 套件身份。保留 epoch 與完整來源版本；
   只有 Debian 檔名去除 epoch。不能以新版替代缺少的舊版本。
2. Git 保存固定鎖定資料：BUILD.json 雜湊、全部二進位至來源對照、157 份描述的
   URL／大小／SHA-256，及每份描述 Checksums-Sha256 所列的全部來源與 Debian 修補檔。
3. 描述只接受一份 UTF-8 控制段落；可以帶 PGP 明文簽章包裝。
   欄位不分大小寫、拒絕重複欄位；Source、Version 必須等於鎖定身份。
   來源檔名只接受普通 basename，拒絕路徑、碰撞、錯誤大小或非 SHA-256。
   保存簽章但本階段沒有驗證 Debian 簽章信任鏈，不宣稱完成簽章驗證。
4. 日常準備只讀 Git 的固定鎖定資料。預設離線驗證快取；明示 --download
   才使用 Docker 網路下載。下載串流有單檔／總量上限，SHA-256 與大小通過後
   才原子提交；快取損壞、缺件、來源對照不符直接拒絕，不能偷偷重建鎖。
5. 全集合離線重驗，重新解析全部描述並核對來源檔集合完整。
   保存集合 manifest／SHA-256 與所有檔案，不解開未知 archive，不執行來源腳本。
   輸入、快取與輸出拒絕連結，UID／GID 必須為目前使用者。
6. 來源 archive 留本機 workplace，正式交付依 dist-all 契約接續。
   本規格只完成第三方 Debian 來源取得／重驗；不外推為全部 GPL／LGPL
   發行義務、三平台封包或人耳驗收完成，也不改音源或玩家規則。

驗收：157 身份／198 對照／513 描述與來源檔完整，所有 SHA-256 與大小符合固定鎖；
無網路乾淨重驗通過，損壞快取、缺檔、描述錯身份／重複欄位／路徑越界均拒絕。

實際結果：513 檔、748,421,917 bytes 全數驗證；五組邊界測試通過。
本機來源 tar 734 項、752,414,720 bytes，重讀每項大小／雜湊符合輸入；
身份、命令、固定鎖及 tar 雜湊見[驗證 033](../playtest/033-linux-music-sources.md)。
