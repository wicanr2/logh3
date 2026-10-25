# 第163輪：楊威利切線與三組海報銜接

使用者指出162右側楊威利海報仍有明顯切線，並要求確認其他兩組。162的內側28像素漸層不足以消除不同星空之間的視覺邊界，本輪改用原PNG天然留邊向外延伸，並把三組接入同一HTML。

## 目前狀態表

| 項目 | 狀態 | 依據與下一步 |
|---|---|---|
| 原因 | 已定位 | 原裁切線內羽化仍在兩張星空交界變色；原圖其實尚有可用的裁切範圍外星空 |
| 銜接 | HTML完成並目視 | 原圖越過舊裁切位置，在天然留邊內向外淡出；星空延伸為上下獨立層，人物原側欄維持不透明 |
| 三組 | 全部抽驗通過 | v1沿原圖，v2/v3沿正式narrow PNG；組別切換、88%／100%六組樣本與左右交界均已擷取 |
| 範圍 | HTML提案 | 不重繪人物，不改正式程式與交付 |
| 核對 | 通過 | 三組中央各363268像素零差；v1臉部與腳部四區零差，素材及艦艇幾何保持 |

入口：[目前脈絡](../../CONTEXT.md)、[162提案](2026-10-25-poster-seam-html-round-162.md)、[三組規格](../spec/152-poster-sets-and-quotes.md)。

## 檢視入口

- [HTML提案](../../workplace/poster163/proposal.html)，新增v1／v2／v3選擇；[三組對照圖](../../workplace/poster163/sets-compare.png)、[本機比較頁](../../workplace/poster163/sets-compare.html)。
- 完整構圖：[v1](../../workplace/poster163/v1.png)、[v2](../../workplace/poster163/v2.png)、[v3](../../workplace/poster163/v3.png)。
- 右側交界：[楊威利](../../workplace/poster163/v1-right-seam.png)、[同盟將領](../../workplace/poster163/v2-right-seam.png)、[親近者](../../workplace/poster163/v3-right-seam.png)。
- 左側交界：[萊因哈特](../../workplace/poster163/v1-left-seam.png)、[帝國將領](../../workplace/poster163/v2-left-seam.png)、[親近者](../../workplace/poster163/v3-left-seam.png)。
- 下方：[v1](../../workplace/poster163/v1-footer.png)、[v2](../../workplace/poster163/v2-footer.png)、[v3](../../workplace/poster163/v3-footer.png)；100%遊戲大小另存各vN-original-size.png。
- [SOURCE](../../workplace/poster163/SOURCE.json)、[BROWSER](../../workplace/poster163/BROWSER.json)、[AUDIT](../../workplace/poster163/AUDIT.json)。HTML為38,666,416 bytes，SHA-256 `ac843a911fce876e67292940285d22f19c988e092465fe09ab0746f6bf34c202`。

## 呈現方式與來源

162在舊裁切線內混合不同星空，仍會顯出視覺邊界。本輪讓原PNG的自然留邊越過原裁切線，再向外淡出。88%配置左側天然留邊45.056像素、右側51.2像素，實際延伸44.056／50.2像素，在原圖結束前留1像素並達到透明。100%配置延伸69.4／79像素。原側欄及人物不淡出，並非用更寬漸層去遮臉或腳。

沿既有中央畫面繪製路徑，星空底框先畫，兩側原圖留邊在上下獨立層延伸，再畫艦艇。主畫面旁的海報仍用原裁切區；上下延伸區與原側欄重疊1像素，避免裁切反鋸齒留下細線。不使用反射、拉伸、補畫人物或生成新PNG。

v1原圖直接沿162內嵌資料，v2／v3取`assets/hd/poster/v2-empire-narrow.png`、`v2-alliance-narrow.png`、`v3-empire-narrow.png`、`v3-alliance-narrow.png`，皆為836×1882；原檔雜湊記SOURCE。依152規格，只有v1顯示五語名言與署名，其他組不疊v1姓名。

## 工具與驗證

[生成器](../../tools/scripts/poster_sets_blend_html.py)、[三組繪製片段](../../tools/scripts/poster_sets_blend_js.txt)、[瀏覽器驗證](../../tools/scripts/poster_sets_blend_browser.mjs)，沿[既有有界擷取入口](../../tools/scripts/poster_closeup_capture.sh)。

```sh
LOGH3_GUI_IMAGE=logh3-hd-preview:20261011-r1 LOGH3_XVFB_SCRIPT=/src/tools/with-xvfb.sh LOGH3_TIMEOUT=5m LOGH3_CPUS=2 tools/gui.sh xrun python3 tools/scripts/poster_sets_blend_html.py
LOGH3_POSTER_BATCH=poster163 LOGH3_POSTER_BROWSER=poster_sets_blend_browser.mjs bash tools/scripts/poster_closeup_capture.sh
```

Chromium逐組檢查88%／100%兩種大小，原圖融合都發生在舊裁切線外，且沒有超出原圖的天然留邊；人物側欄不淡出。六張完整構圖、六張左右交界及三張下方近照已擷取，三組比較圖已目視。原有艦首艦尾安全、本體固定、v1五語文字、控制下載與桌面1600／手機390無水平溢出抽驗通過，零JavaScript例外與外網請求。

AUDIT確認三組中央各363268像素與162零差，v1兩人臉部、腳部四個抽樣區域零差，所有既有內嵌素材及艦艇幾何保持。四張群像直接使用原PNG，沒有修改原檔。Python／shell語法、來源及收據雜湊、1000:1000與整專案root-owned／誤建.md目錄檢查通過。

中途嘗試給中央遊戲加SVG裁切，Chromium造成718像素取樣差異；[裁切開／關診斷](../../workplace/poster163/DIAGNOSTIC-CLIP.json)證明單獨移除裁切就與162零差。最終移除新增遊戲裁切，改星空獨立延伸層，重跑恢復零差。另一次生成器殘留舊名稱檢查導致未重生HTML，瀏覽器正確拒絕舊分層；修正生成器後同容器命令乾淨重跑。這些屬原型工具與瀏覽器取樣問題，不列為遊戲缺陷。

仍為本機HTML原型，正式程式、README、舊稿及交付保持；未commit、push、打包或發布。一次性Docker容器退出自刪，沒有新建映像或清理其他專案資源。
