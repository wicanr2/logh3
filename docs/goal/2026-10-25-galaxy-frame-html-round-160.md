# 第160輪：銀河串接、伯倫希爾下移與休伯利安外形

使用者要求伯倫希爾再往下；指出159休伯利安與原著有差距，並要求上下兩艘船的背景連接楊威利與萊因哈特海報的銀河。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 伯倫希爾 | 提案完成 | 原圖與朝右配置保留，群組座標相較159實際下移24像素 |
| 休伯利安 | 依使用者兩圖重繪，待審閱 | 方形艦首、上下兩組3×4圓孔、細長艦身及144M；替換尖叉與大型紅色結構 |
| 銀河背景 | 提案完成 | 原人物海報上方星空區，以相同星點尺度連續鏡像鋪設；上下與四角共用，漸層融合兩側 |
| HTML | 瀏覽器抽驗通過，待使用者檢視 | workplace/poster160/proposal.html，12,721,766 bytes；只修改提案 |

入口：[目前脈絡](../../CONTEXT.md)、[159提案](2026-10-25-single-flagships-html-round-159.md)。來源及限制隨本輪輸出記錄，正式遊戲、舊稿與交付保持。

## 檢視入口

- [HTML提案](../../workplace/poster160/proposal.html)、[構圖截圖](../../workplace/poster160/proposal-frame.png)、[完整頁面](../../workplace/poster160/proposal-page.png)、[手機截圖](../../workplace/poster160/proposal-mobile.png)。
- [來源](../../workplace/poster160/SOURCE.json)、[瀏覽器收據](../../workplace/poster160/BROWSER.json)、[稽核收據](../../workplace/poster160/AUDIT.json)。三者綁HTML SHA-256 `e039b59fe2a9c8f77419c4308042ff3d1fe21289c7baf1a149e2942661099a05`。
- [使用者整體示意](../../workplace/poster160/layout-reference.png)、[休伯利安艦首參考](../../workplace/poster160/hyperion-bow-reference.png)、[全艦參考](../../workplace/poster160/hyperion-body-reference.png)。

## 休伯利安素材與提示

先以[壽屋KP771產品資料](https://www.kotobukiya.co.jp/product/detail/p4934054063635/)核對作品分類，該頁圖片位址明列`KP771_hyperion_ishiguroban`。官方縮圖下載回403，沒有冒稱已保存原圖；使用者隨後提供的兩張圖是本輪主要外形依據。159的尖叉艦首及巨大紅色結構不是可沿用的原著證據。

使用內建imagegen，沒有使用API／CLI。先依兩張使用者圖重繪，再調整鏡頭，使細長艦身與兩組圓孔能在上框辨認。

- [初稿透明圖](../../workplace/poster160/hyperion-classic.png)與[完整初稿提示](../../workplace/poster160/hyperion-prompt.txt)。
- [採用的側面取景圖](../../workplace/poster160/hyperion-side.png)，2146×733 RGBA，SHA-256 `568093a586c348431fa86618bd7653365f04e1186800ed0da3887049867a1d93`；[完整視角調整提示](../../workplace/poster160/hyperion-side-prompt.txt)。

內建工具原輸出分別為`exec-4bec5a6b-6796-4b44-928e-6d8cd96ad288.png`及`exec-f2478a58-b5bb-472a-8f21-9f93a137aab3.png`，位於`/home/anr2/.codex/generated_images/01a16b43-4ac7-74b3-95be-186e30695ea3/`，已原樣複製到本輪工作區。保留原始透明層，素材屬參考重構，不宣稱動畫逐形完全一致或DOS證據，限本機提案。

## 銀河與版面

[帝國人物原圖](../../workplace/poster160/empire-portrait.png)上方836×282無人物星空作為紋理，以人物圖同一縮放率顯示為320×107.9426像素。SVG鏡像鋪設讓接縫連續；上下及四角漸層使用同一全域紋理，不再分成兩塊深色實心區域。兩側人物PNG、名言與姓名保持，不另重繪角色或改寫原PNG。

下方伯倫希爾原圖、大小、朝右方向及剛體轉角保持，只將預設bottomY改24；與159瀏覽器群組座標相比實際增加24像素。休伯利安寬度預設65%，讓兩組艦首圓孔同時可辨識，仍可調60–140%。沿用88%／100%遊戲大小比較，此配置仍未核准接正式版。武器起點跟隨新的艦首及上下取景，本體固定。

## 工具與重現

[生成器](../../tools/scripts/poster_galaxy_html.py)讀159單檔提案，只替換新船體、背景與預設位置；[瀏覽器驗證](../../tools/scripts/poster_galaxy_browser.mjs)沿[既有有界Chromium入口](../../tools/scripts/poster_closeup_capture.sh)。

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_TIMEOUT=5m LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/poster_galaxy_html.py
LOGH3_POSTER_BATCH=poster160 LOGH3_POSTER_BROWSER=poster_galaxy_browser.mjs bash tools/scripts/poster_closeup_capture.sh
```

Chromium驗上下及四角同一銀河來源、下移設定、單艦與方向、完整船體群組不隨武器相位移動、五語圖片內文字、大小比較、漸層、武器及實際設定下載。桌面1600／手機390無水平溢出，零JavaScript例外與外網請求；構圖、兩組圓孔與銀河星點尺度已目視。

稽核另比159的中央背景、兩側人物與伯倫希爾內嵌素材完全相同，休伯利安確實替換；來源及收據雜湊、Python／shell語法、1000:1000擁有權及整專案root-owned／誤建.md目錄檢查通過。正式poster.go、poster_decoration.go與159提案保持，未跑未改動的全產品測試或打包。名言尖端舊版紙本逐字／頁碼核對仍未完成。未commit、push或發布；一次性Docker容器退出自刪，其他專案資源保持。
