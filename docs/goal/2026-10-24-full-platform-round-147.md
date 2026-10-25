# 第147輪：三平台完整版重新打包

使用者要求將目前修正重新打包為三平台完整版。沿v.1.5.2-20261024相同產品來源，包含145前線及人事修正與150預設關閉的非戰艦旗艦。既有Windows替換ZIP、影片及公開版本保留；完整版只留本機。

## 目前狀態表

| 工作 | 狀態 | 驗證 |
|---|---|---|
| 凍結及共用素材 | 通過 | 1593項凍結來源、615項非vendor與146同來源；三平台598項共同素材及資料一致 |
| Windows amd64 | 完整ZIP通過 | 603項檔案；Wine直接入口及兩啟動器、正常96步、19張畫面、六表讀回及重啟 |
| Linux amd64 | 完整ZIP通過 | 603項檔案；直接入口及兩啟動器、正常96步、19張畫面、六表讀回及重啟 |
| macOS通用 | 完整ZIP及靜態檢查通過 | 605項檔案；Intel／Apple Silicon、最低11.0、SDK15.5、僅系統依賴，ARM3261頁簽章全同；未真機驗收 |
| 交付核對 | 已產生索引及校驗清單 | ZIP逐檔、權限、UTF-8、版本、素材、資料與存檔位置通過；原修正包與影片保持 |

## 入口

- [完整版建置器](../../tools/scripts/full_platform_repack.py)沿[既有跨平台封包器](../../tools/scripts/v140_platform_delivery.py)及[macOS建置器](../../tools/scripts/local_test_macos_build.sh)，不改產品程式。
- [直接入口與啟動器抽驗](../../tools/scripts/full_package_smoke.py)、[獨立交付核對器](../../tools/scripts/full_platform_receipt.py)，正常鍵鼠沿[146錄影器](../../tools/scripts/garrison_promo_gui.py)。[來源及交付索引](../internals/data/v152-full-delivery.json)已產生。
- 產品及測試依[146](2026-10-24-garrison-flagship-promo-round-146.md)與[150](../spec/150-non-battleship-flagship.md)。本輪只驗封包，沒有新增DOS原版或平台真機聲明。
- 工作區`workplace/package92-v.1.5.2-20261024-release147/`及`workplace/macos-release147/`；正式三ZIP位於`dist-all/v.1.5.2-20261024/full/`。

開工Issue快照102張，#48 OPEN，282條verify零失效。工作樹與使用者文字檔保留，不寫Issue、不發布Release。

## 本機交付

| 平台 | 完整ZIP | 大小 |
|---|---|---|
| Windows amd64 | [下載](../../dist-all/v.1.5.2-20261024/full/LOGH3-v.1.5.2-20261024-windows-amd64.zip) | 417534032 bytes |
| Linux amd64 | [下載](../../dist-all/v.1.5.2-20261024/full/LOGH3-v.1.5.2-20261024-linux-amd64.zip) | 417671677 bytes |
| macOS Intel／Apple Silicon | [通用包](../../dist-all/v.1.5.2-20261024/full/LOGH3-v.1.5.2-20261024-macos-universal.zip) | 424656460 bytes |

[完整版SHA-256](../../dist-all/v.1.5.2-20261024/full/SHA256SUMS.txt)及[建置／來源／驗收manifest](../../dist-all/v.1.5.2-20261024/full/BUILD.json)保存身份。各包均含原版、HD、FM／GM／MT-32／HD曲庫、音效及音色庫，解壓即用，不需另套修正包。存檔沿原使用者目錄，包內沒有玩家偏好。完整版只留本機。

## 驗收與重建

Windows與Linux從實際ZIP解壓，直接執行及兩啟動器共6條診斷入口不傳orig／data／art／music，從`/tmp`驗證自動尋找包內資料。HD、GM及MT-32曲庫可用，非戰艦旗艦預設關閉，兩啟動器AI等級及強化HD配樂選擇正確。這些是診斷入口，不當成正常鍵鼠證據。

另各以正常主選單及真實鍵鼠走完96步：駐留、改選提督、當地領船、E／T拒絕、C旗艦編成、實際抵達、六表存讀及退出重啟，共38張完整視窗。Linux C旗艦及Windows新設定頁目視通過。正常鍵鼠器沿146的`/orig`唯讀掛載，核對器確認它與ZIP內original每個檔案相同；data及art取自各實際解壓包。兩側安裝檔案在驗收後均與manifest相同，保存寫到獨立測試目錄。

產品沒有修改，615項產品來源與146最後標準19套件15 PASS／4無測試一致，本輪不重跑相同產品測試。Windows新建EXE與146已驗替換包逐位元組相同。Linux／Windows用`logh3-ebiten:2`、Go1.24.13，macOS用既有`wizardry7-osxcross:go1.24.13-python3.13.15-15.5-r1`；Windows驗證用`eob-remake-wine-verify:ubuntu-noble-20261008-r2`，正常X11用既有HD預覽映像。沒有新工具映像。

```sh
tools/gui.sh xrun python3 tools/scripts/full_platform_repack.py prepare
tools/gui.sh xrun python3 tools/scripts/full_platform_repack.py windows
tools/gui.sh xrun python3 tools/scripts/full_platform_repack.py linux
bash tools/macos.sh macos-release147 bash /src/tools/scripts/local_test_macos_build.sh \
  v.1.5.2-20261024 workplace/package92-v.1.5.2-20261024-release147/source
tools/gui.sh xrun python3 tools/scripts/full_platform_repack.py macos
tools/gui.sh xrun python3 tools/scripts/full_platform_repack.py extract
# 沿索引所列命令完成兩平台診斷與正常鍵鼠驗證後，產生交付收據。
tools/gui.sh xrun python3 tools/scripts/full_platform_receipt.py
# 不改寫交付，重驗既有收據。
tools/gui.sh xrun python3 tools/scripts/full_platform_receipt.py --check
```

建置要求146候選存在且新full目錄尚未建立，拒絕覆寫交付。原根校驗清單先保存到工作區`previous-SHA256SUMS.txt`，本輪只追加完整版清單；舊替換ZIP、BUILD.json、影片及promo證據雜湊全部保持。沒有建立額外引擎包、來源包或重錄影片。

Python語法、差異、擁有權及容器清理檢查通過。沒有root-owned檔或誤建`.md`目錄，一次性容器均已退出自刪。Windows為Wine、macOS未真機、音訊未作人耳確認，不新增DOS parity聲明。未commit、push、tag或發布。

2026-10-24第148輪清理可重建工作副本，實際ZIP及本輪驗收證據保持；重跑`--check`前須先依[清理及還原對照](2026-10-24-workspace-cleanup-round-148.md)還原所需安裝副本。這不改變本輪原驗收結果或其雜湊。
