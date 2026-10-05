# 086：正常標題玩家流程

第97輪，GitHub #68。父提交 `481221ad222b02f8c88b2f1f1644fb3a8c6797af`，驗證時工作樹含本輪未提交修改。
正常入口與全部來源身份見[完整來源](../internals/data/title-player.json)，不將本機檔名當作唯一索引。
原版版本、輸入、SHA-256及IDA定位見[431](../re/431-title-player.md)；契約為[112](../spec/112-title-player.md)。

## 原版與正常Game

原版dosgolem `f7523f9ac7ed6f17bc4e37686b10748c5e9bf89c`／乾淨，Go image `golang:1.24-bookworm`。
末格、標題等待及無輸入逾時各以相同快照／相同旗標雙跑，screen.bin、pal.bin及旗標逐byte相同。
左鍵主選單沿第96輪雙跑；右鍵／普通A／BIOS A樣本與本輪按鈕0輪詢靜態證據合併核實。
來源快照沿正常GIN3.COM入口產生，無RAM注入，不需RND seed；不是注入局面的診斷。

正常 `bootstrap.Load → LoadIntro → game.New` 完整421格、末格等待、標題、拒絕輸入、逾時重播、第二趟標題、左鍵主選單及正常138開局。
第0格、末格、標題、拒絕輸入、重播、第二趟標題、主選單、正常開局與60 Hz基準共29檔，b／c兩個獨立批次逐檔相同。
期滿前1 ns仍留原狀態；期滿才轉場，大步進保留跨狀態餘量。既有左鍵／右鍵／鍵盤片頭跳過、缺標題及損壞資料測試通過。

| 比較 | 色號像素／差異 | RGB像素／差異 | 最大通道差／bbox |
|---|---:|---:|---|
| 原版末格350頁 → Game末格第65列 | 224000／0 | 224000／0 | 0／null |
| 原版標題350頁 → Game標題第65列 | 224000／0 | 224000／0 | 0／null |
| 兩張頁面正規化後完整480畫布 | 每張307200／0 | 每張307200／0 | 0／null |
| 每張上下65列補邊 | 83200／0 | 83200／0 | 0／null |

比較等級為 `same-state`，限定原版靜態末格與標題的350頁；原版mode10h／remake640×480。
正規化只依020核准畫布決定置於第65列，不裁掉任何頁內差異。上下補邊亦全部核對，不用遮罩。
完整原版PNG、Game PNG、正規化PNG、差分及區域統計在完整來源索引中保存SHA-256。
末格色號SHA-256 `b9b8f179788ec26d9812d5acd63cdd89c294ab59ccb6b1e0a9f0d73340ba0db9`；
標題色號SHA-256 `addde5e0929eb5f99a5430a714b19041a2cbe5530a4720a3641abebbfa3bf8ae`。
原版與Game的末格／標題16色均逐byte相同，TITSPT尾段色盤漏接已修。

## 正式60 Hz視窗

兩條正常 `cmd/gin3`，無 `-script`、direct-entry、TPS調整、局面或時間注入。
使用 `logh3-ebiten:2`、Xvfb、Mesa及隱藏系統游標。真實X11滑鼠位於 `(600,340)`。

| 流程 | 實際結果 |
|---|---|
| 自然標題→右鍵→Enter→左鍵 | 前三張仍是標題，左鍵回主選單；5200次Update結束於StateMainMenu |
| 自然標題→無輸入等待 | 5480次Update結束於StateIntro第2格，完整307200色號符合獨立60 Hz Game計畫 |

五張真實X11全圖為兩張標題、右鍵後、Enter後及主選單。
每張640×480的307200像素與正式Game計畫全部相同，共1536000像素；沒有排除區。
兩條正式GUI末端dump各307200色號也與獨立Game相同。
重播末端PNG為正式renderer輸出；該動態第2格未另做X11快照，不升格為逐幀原版同步。
已目視X11標題及主選單，文字、位置、色彩及游標符合本輪核對範圍。

重生入口：

```sh
bash tools/gui.sh xrun bash -c 'LOGH3_TITLE_RECEIPTS=/src/workplace/<新玩家目錄> go test ./internal/bootstrap -run TestTitlePlayerPath -count=1 -v'
# 先在同一image建正式cmd/gin3與既有tools/scripts/fleet_player_gui_capture.go；
# 本輪固定二進位位置由title_player_gui.py引用，來源與SHA見完整來源索引。
LOGH3_CPUS=2 bash tools/gui.sh xrun python3 tools/scripts/title_player_gui.py \
  --case left --out /src/workplace/<新GUI目錄> --plan /src/workplace/<新玩家目錄>
# 無輸入重播把--case換成replay，輸出仍用另一新目錄。
```

## 配樂與標準測試

正常Game→正式app播放器記錄器：動畫到標題不中斷OPEN_M，逾時先Stop再從頭要求OPEN_M，左鍵主選單停止。
這驗正式配樂要求，沒有冒充實體裝置、音色或人耳驗收；#69仍OPEN。

```sh
bash tools/gui.sh xrun bash -c 'go test ./internal/... ./cmd/... -count=1'
```

18套件：13 PASS／5無測試；Game175.569秒、gin3 88.746秒。
log SHA-256 `0a17f16cd7adba7cd83aa1e71226f23c29d6dbf5491aca851549249ece0d582a`。
第95輪1889來源與第96輪1701來源核對未變。GUI與標準測試後只修正兩处程式註解，功能程式碼未變。

末格5.396420663秒、標題15.442619818秒採PIT規格近似；不宣稱硬體wall-clock或逐週期一致。
原版素材仍限玩家自備／private保存；不加入一般發行包。

## 工具與收尾

IDA初次COM EA假設錯誤與後續ASCII locale讀中文JSON失敗均屬匯出腳本問題。
改檔案位移映射與明示UTF-8後，同image／同原始bytes乾淨重跑成功；保留失敗log與版本，不當產品缺陷。
分級匯出681條原指令、九項proven，其餘unknown均有警示，不改原始函式名。
完整來源核對、輸出1000:1000、工作樹與專案容器清理於收尾核對。
112限定CONFORMED，#68可依正常玩家回歸關閉；其他七項閘門及原Goal保持。
