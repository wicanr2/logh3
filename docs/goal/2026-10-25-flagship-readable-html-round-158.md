# 第158輪：主旗艦可辨識、仰視與上向隊列、漸層及尖端舊譯查證

使用者要求伯倫希爾、休伯利安的主要船體都能看清楚。已澄清：下方帝國艦首朝畫面上方；上方同盟維持朝左，鏡頭由下仰視，要看得到Hyperion。上下與左右海報四處銜接增加漸層；名言及姓名仍在人物圖內，名言需查尖端當年譯文。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 主旗艦呈現 | HTML已完成 | 前景兩艘船體完整可見，預設顯示倍率100%，瀏覽器驗證完整落在上下範圍內 |
| 上方同盟 | HTML已完成 | 休伯利安朝左、艦底仰視；新視角依既有HD圖生成，不冒稱原版逐形證據 |
| 下方帝國 | HTML已完成 | 伯倫希爾艦首朝上；後景艦列同向，按景深排列，本體固定 |
| 銜接 | HTML已完成 | 四處漸層遮罩使用與上下飾框相同的底色，末端沒有異色色帶；人物名言及姓名在上層 |
| 尖端名言 | 二手來源已核對，紙本未核對 | 萊因哈特採標示尖端舊版的轉錄用語；楊威利採同站語錄的短句節錄，未取得舊版書頁與確定頁碼 |
| HTML | 瀏覽器抽驗通過，待使用者審閱 | 自包含14,458,666 bytes；五語文字、控制項、設定下載、桌面與手機無溢出、零外網請求 |
| 遊戲大小取捨 | 僅供比較的提案 | 預設88%，4:3比例；可切回100%。尚未核准將此配置接入正式遊戲 |

入口：[目前脈絡](../../CONTEXT.md)、[157提案](2026-10-25-flagship-depth-html-round-157.md)。本輪仍沿既有HTML／Chromium工具鏈；新視角原圖、完整提示及來源記錄保存在workplace/poster158，不作DOS或動畫逐形精確聲明。

## 檢視與重現

- [HTML提案](../../workplace/poster158/proposal.html)、[構圖截圖](../../workplace/poster158/proposal-frame.png)、[完整頁面](../../workplace/poster158/proposal-page.png)、[手機截圖](../../workplace/poster158/proposal-mobile.png)。
- [來源](../../workplace/poster158/SOURCE.json)、[瀏覽器驗證](../../workplace/poster158/BROWSER.json)、[收尾稽核](../../workplace/poster158/AUDIT.json)。三者綁定HTML SHA-256 `25de4f97b3d4ec7d6ae38bffd8a1f5ef626b054f25fdb5ad13428250fc65f301`。
- [生成器](../../tools/scripts/poster_readable_html.py)、[SVG繪製片段](../../tools/scripts/poster_readable_js.txt)、[瀏覽器檢查](../../tools/scripts/poster_readable_browser.mjs)，沿用[157模板](../../tools/scripts/poster_flagship_rows.html)及[有界擷取入口](../../tools/scripts/poster_closeup_capture.sh)。

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_TIMEOUT=5m LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/poster_readable_html.py
LOGH3_POSTER_BATCH=poster158 LOGH3_POSTER_BROWSER=poster_readable_browser.mjs bash tools/scripts/poster_closeup_capture.sh
```

原62／93像素上下空間難以同時容納大型、完整的直向船體，因此提案增加遊戲顯示大小比較。88%時中央仍完整顯示原畫面，兩側人物填滿增加的空間。100%可檢視原尺寸。這是可丟棄原型，不視為使用者已同意改正式配置。

## 新視角素材

使用內建imagegen，參考157的45B1／46BD圖，製作透明底提案圖並重新排列成圖集。提案保留原始輸出，只將兩艘各自裁出，沒有把舊平視圖旋轉後宣稱是仰視圖。

- [第一張原圖](../../workplace/poster158/heroes-raw.png)與[完整生成提示](../../workplace/poster158/hero-prompt.txt)。
- [分離排版圖集](../../workplace/poster158/heroes-layout.png)與[完整編輯提示](../../workplace/poster158/layout-prompt.txt)。
- [朝上伯倫希爾](../../workplace/poster158/brunhild-up.png)，458×798；[仰視休伯利安](../../workplace/poster158/hyperion-low.png)，1110×481。

素材雜湊在SOURCE.json。新視角是藝術重構，不能證明動畫正典、DOS或原畫輪廓完全一致；限本機提案，沒有加入公開發行包。

## 名言查證與限制

| 人物 | 提案文字 | 來源及證據範圍 |
|---|---|---|
| 萊因哈特 | 嗯！星星好美啊！ | [《黎明篇》第一章〈永恆的夜〉轉錄](https://www.hkweb.org/logh/story/1-1.htm)。[小說入口](https://www.hkweb.org/logh/story/story.htm)明列尖端二十冊版、譯者及清大BBS轉錄背景，屬二手文字證據 |
| 楊威利 | 我想最好的還是和平吧。 | [《楊語錄》黎明篇第五章〈伊謝爾倫攻略戰〉](https://hkweb.org/ginei/yang/)，從長句節錄此子句；該語錄頁未獨立提供版權頁與譯者核對，不能只憑同站就認定版次 |

目前未取得尖端當年紙本頁面或確定頁碼，不能宣稱已由出版社原書逐字證實。HTML明列來源與限制；簡體為對應轉換，日／英／韓為本專案對應譯文，不宣稱尖端外語版。正式名言定稿仍需舊版書頁交叉核對。

## 抽樣與收尾

Chromium檢查上下各六艘、主旗艦整體邊界、上左下上、後景景深次序、四角漸層開關、本體固定、五語文字位於人物圖片內、4:3大小切換、武器開關及實際JSON下載。桌面1600及手機390無水平溢出，零JavaScript例外及外網請求；完整構圖已目視。

Python／shell語法、來源雜湊、擁有權及整個專案root-owned／誤建.md目錄檢查通過。正式poster.go、poster_decoration.go與157提案保持。未改正式版、README、封包或Issue；未commit、push或發布。既有Docker映像沿用，一次性容器退出自刪。
