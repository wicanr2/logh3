# 136：海報戰機與多旗艦循環裝飾

使用者指定上方兩軍戰機交戰、下方多艘名將旗艦，採輕量循環動畫。正式契約為[156](../spec/156-poster-combat-decoration.md)，唯一狀態表為[154](../goal/2026-10-25-poster-header-footer-round-154.md)。

本輪使用既有HD圖集的16種帝國與12種同盟旗艦外形。上方雙方各四架戰機穿梭，下方每側八艘、兩列相向，每12秒換批。戰機為生成小圖，來源、完整提示及SHA-256保存於`assets/hd/poster/fighters.SOURCE.json`，不冒稱原版影格。

裝飾只畫在實際背景完整黑列，當前HD星圖為上方0–40、下方418–479邏輯像素；主選單不畫。海報關閉則隨之關閉，沿原有動畫設定切換動態或靜態，不增加偏好或存檔格式。工具列、說明與游標仍在上層。

## 驗證與交付入口

三平台採同一凍結來源，由`tools/scripts/poster_decoration_release.py`建立`v.1.6.2-20261025`本機完整版。獨立核對器為`tools/scripts/poster_decoration_receipt.py`，正式收據保存於[poster-decoration154.json](../internals/data/poster-decoration154.json)。舊v1.6.1不覆寫。

正常操作腳本為`tools/scripts/poster_decoration_gui.py`。從主選單進入，輸入防拷答案，經系統設定開啟三組海報，測五語、F1／F7、工具列、靜態動畫與六表存讀；不使用狀態注入、固定seed或診斷入口。

`tools/scripts/poster_decoration_preview.py`僅從通過的正式AppImage收據複製[README原圖與來源](../images/readme-v162-poster.json)，並將18秒無聲正常視窗預覽放在同版`promo/`，保存FFprobe、來源與SHA-256。

## 已驗結果

同一1620項凍結來源含642項非vendor，618項共用檔三平台一致。獨立核對器逐包驗證精確檔案集合、CRC、SHA-256、版本、架構、UTF-8、權限及素材來源。Linux實際解出744項AppDir，使用正式AppImage自行找素材與預設XDG存檔，未指定orig／data／art／save。

| 平台 | 檔案 | bytes |
|---|---|---:|
| Linux x86_64 | `LOGH3-v.1.6.2-20261025-linux-x86_64.AppImage` | 452856312 |
| Windows amd64 | `LOGH3-v.1.6.2-20261025-windows-amd64.zip` | 452991276 |
| macOS Intel／Apple Silicon | `LOGH3-v.1.6.2-20261025-macos-universal.zip` | 462850221 |

三包位於`dist-all/v.1.6.2-20261025/full/`，同版根目錄及full內含SHA256SUMS。Linux啟動方式：

```sh
chmod +x LOGH3-v.1.6.2-20261025-linux-x86_64.AppImage
./LOGH3-v.1.6.2-20261025-linux-x86_64.AppImage
```

缺FUSE時加`--appimage-extract-and-run`。存檔位置及其他主機相依沿[135](135-v161-full-appimage.md)契約。

標準20套件16 PASS／4無測試。正式AppImage帝國、同盟與實際Windows ZIP的Wine各23張，共69張正常畫面；三組海報、五語、工具列、F1／F7、動畫設定、主選單不畫、六表存讀與重啟通過。原版／HD六組共3254400個非留白像素零差，四組側欄比較合計2764800像素零差，靜態模式全圖兩次擷取一致。同盟跨平台六表與既有基準相同。

循環各21、22、27種相位且至少兩批，背景上傳均20至20。容器Draw中位依序20.37、16.15、17.60ms；並行工作負載不同，不據此宣稱加速或固定FPS。Windows不錄影片，仍觀測相位及狀態；無錄影路徑至少18秒、最長90秒，保留相同相位及換批門檻。三份實際腳本SHA-256與快照記在收據，不將環境缺少FFmpeg或並行錄影取樣不足記成遊戲缺陷。

預覽為18秒、1280×720、15fps、270格、790160 bytes的實際視窗H.264，刻意無音軌。黑段0.5秒與凍結2秒門檻均無事件，來源及檢測記在同版promo。macOS最低11.0／SDK15.5，雙架構及ARM4092頁簽章通過，未真機。

## 驗證限制

循環影片來自實際視窗，僅作功能預覽，沒有另播預製影片或新增音效。Xvfb使用軟體繪圖，CPU耗時與相位只作容器抽樣，不宣稱實機FPS或硬體輸入延遲。Linux採AppImage的extract-and-run，Windows採Wine；macOS核對雙架構、相依與ARM簽章，未作真機驗收。新增裝飾為remake呈現，不增加DOS parity聲明。v1.6.0完整推廣片保留。
