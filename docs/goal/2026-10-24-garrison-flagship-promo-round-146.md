# 第146輪：駐地出擊實錄與非戰艦旗艦

使用者要求新版推廣影片包含換提督、從駐地選船及出發，另新增預設關閉的非戰艦旗艦選項。
使用者隨後明確指定轉移旗艦可選存活艦艇，但E／T仍不得當旗艦。沿[150](../spec/150-non-battleship-flagship.md)實作。

## 目前狀態表

| 工作 | 狀態 | 證據與限制 |
|---|---|---|
| 使用者規則 | 150限定CONFORMED | 編成與轉移交易共用艦種政策，E／T始終排除 |
| 設定及偏好 | 標準19套件15 PASS／4無測試 | 預設關閉、保存重啟、未知偏好鍵及拒絕不改寫通過 |
| 駐地操作實錄 | Linux與Windows Wine正常路徑通過 | 各96步、19個檢查點，共38張；未作Windows真機驗收 |
| 推廣影片 | 87秒、1080p60、8段已完成 | 配樂連續、7剪接點及無黑段；字幕邊界及目視通過，未作人耳驗收 |
| Windows交付 | v.1.5.2-20261024本機替換包已核對 | 1593項凍結來源、615項現行產品一致；PE、版本、manifest及SHA-256通過 |

本輪開工Issue快照102張、#48 OPEN，282條verify零失效。前輪工作樹與使用者文字檔保留，
不寫GitHub Issue，不移動已發布tag／Release。影片含原版及衍生音畫，沿既有本機交付邊界。

## 入口

- [驗收132](../playtest/132-garrison-flagship-promo.md)保存正常操作、驗證層級及交付限制。
- [輸入模型](../../tools/scripts/garrison_promo_plan.go)、[真實視窗操作](../../tools/scripts/garrison_promo_gui.py)、[剪輯及聲畫檢查](../../tools/scripts/garrison_promo_movie.py)、[獨立核對](../../tools/scripts/garrison_promo_receipt.py)。
- [來源索引](../internals/data/garrison-flagship146.json)由獨立核對器產生，綁定兩平台GUI、保存、偏好、影片與來源。
- [87秒影片](../../dist-all/v.1.5.2-20261024/promo/LOGH3-v.1.5.2-20261024-garrison-promo.mp4)及[Windows替換包](../../dist-all/v.1.5.2-20261024/release/LOGH3-v.1.5.2-20261024-windows-amd64-patch.zip)留在本機。只替換既有v1.5的`gin3.exe`。

本輪收尾Issue快照仍為102張、#48 OPEN，282條verify零失效。格式及差異檢查通過；整工作根未見root-owned檔案或誤建的`.md`目錄，輸出為1000:1000。一次性容器均退出自刪，沒有建新映像、commit、push或發布。
