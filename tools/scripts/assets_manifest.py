# assets/original/ 的清單：每一類的來源檔、產生命令、每個檔的 SHA-256。
# → assets/original/README.md（說明＋分類表）、assets/original/MANIFEST.tsv（逐檔雜湊）
#
# 新增一類素材時在 KINDS 補一列；這份清單是「怎麼重做」的唯一記錄。
import hashlib, os

ROOT = '/work/assets/original'
KINDS = [
    # 目錄, 原版檔, 產生命令, 內容, 規格文件
    ('face', 'FACE.AMG', 'amg_export.py', '人物頭像 83 張 80×100（第 82 張是空白格）', 'docs/internals/09-face-amg.md'),
    ('sprite', 'SPRITE.SPR', 'spr_export.py 16 SPRITE.SPR', '圖塊 75 張 16×16（艦隊組成格盤）', 'docs/internals/10-sprite-spr.md'),
    ('frame', 'FRAME.BLK', 'spr_export.py 20 FRAME.BLK', '視窗外框圖塊 33 張 16×20', 'docs/internals/47-window-frame-tiles.md'),
    ('block_3', 'BLOCK_3.BLK', 'spr_export.py 20 BLOCK_3.BLK', '戰場圖塊 128 張 16×20', 'docs/internals/35-block3-blk.md'),
    ('3dblock', '3DBLOCK.SPR', 'spr_sheet.py 3DBLOCK.SPR 48 64 4 4 row', '等距方塊 48×64（放大 4 倍的單張）', 'docs/internals/21-3dblock-spr.md'),
    ('comflag', 'COM_FLAG.SPR', 'comflag_export.py', '指揮旗標籤五張 ＋ 總表', 'docs/internals/01-assets-inventory.md'),
    ('strmap', 'STRMAP.BIT', 'strmap_export.py', '戰略星圖底圖', 'docs/internals/12-strmap-bit.md'),
    ('vo5', 'END*.VO5', 'vo5_export.py', '結局圖 50 張', 'docs/internals/18-vo5.md'),
    ('comani', 'COMANI.BMG', 'bmg_export.py', '交戰演出動畫 254 段 1,438 格，每張 256×102 是畫完那一格之後的視窗；XOR 開頭的段疊在同單位 D0 第 0 格上', 'docs/internals/20-comani-bmg.md'),
    ('font', 'GRAPH.IMG', 'font_extract.py；font_head.py', '倚天中文字型總表與半形字頭', 'docs/internals/06-font-graph-img.md'),
    ('music/mid', '*.MDS', 'mds_export.py', '22 首樂曲轉標準 MIDI', 'docs/internals/17-mds.md'),
    ('music/vgm', '*_F.MDS', 'tools/music_fm_capture.sh', 'ＦＭ版 11 首在 dosgolem 裡跑出來的 OPL 暫存器記錄（.meta 是擷取參數）', 'docs/internals/17-mds.md'),
    ('music/ogg', '*.MDS', 'tools/music_to_ogg.sh', '`_F` 由 VGM 合成（OPL 模擬）、`_M` 由 MIDI 以 fluidsynth ＋ FluidR3_GM 合成；`SOURCES.tsv` 是 `_F` 的來源雜湊', 'docs/internals/17-mds.md'),
]

rows, counts = [], {}
for d, *_ in KINDS:
    base = f'{ROOT}/{d}'
    for dp, _, fns in sorted(os.walk(base)):
        for fn in sorted(fns):
            p = f'{dp}/{fn}'
            h = hashlib.sha256(open(p, 'rb').read()).hexdigest()
            rows.append((os.path.relpath(p, ROOT), os.path.getsize(p), h))
            counts[d] = counts.get(d, 0) + 1

known = {r[0] for r in rows}
stray = []
for dp, _, fns in os.walk(ROOT):
    for fn in fns:
        rel = os.path.relpath(f'{dp}/{fn}', ROOT)
        if rel not in known and rel not in ('README.md', 'MANIFEST.tsv'):
            stray.append(rel)
if stray:
    raise SystemExit('有不在 KINDS 裡的檔（先補分類）：' + ', '.join(stray[:10]))

with open(f'{ROOT}/MANIFEST.tsv', 'w') as fp:
    fp.write('# 路徑\tbytes\tSHA-256\n')
    for r in rows:
        fp.write('%s\t%d\t%s\n' % r)

md = ['# 原版素材（保存用）', '',
      '從《銀河英雄傳說III SP》DOS 版（亨和企業．微波軟體，1994）原版檔案解出來的素材，',
      '進版控是為了保存遊戲資產、給日後重新繪圖當參考。規則見 `AGENTS.md`「Docker-only 與',
      '原版素材邊界」：只放匯出腳本產生的可檢視格式，不手改；**remake 的發行包不含這個目錄**；',
      'repo 轉公開前要先處理這個目錄（它在 git 歷史裡）。', '',
      '重做：自備原版放 `org_game/extracted/`，`tools/py.sh tools/scripts/<命令>` 多半寫到',
      '`workplace/assets/`，再複製過來（`bmg_export.py` 直接寫這裡）；然後 `tools/py.sh tools/scripts/assets_manifest.py`',
      '重產這份說明與 `MANIFEST.tsv`（逐檔 SHA-256）。音樂那兩支是 shell 包裝器，要 dosgolem。', '',
      '| 目錄 | 檔數 | 原版檔 | 產生命令 | 內容 | 格式文件 |', '|---|---:|---|---|---|---|']
for d, src, cmd, what, doc in KINDS:
    md.append(f'| `{d}/` | {counts.get(d, 0)} | `{src}` | `{cmd}` | {what} | [`{os.path.basename(doc)}`](../../{doc}) |')
md += ['', '還沒匯出成圖的（解碼在 `internal/asset` 或只有格式文件）：`ANI_PNL0`–`2.A98`、',
       '`BLOCK_0`–`2.BLK`、`MAP*.CMP`、`IBM16.350`（開頭動畫）、`SPTIT.350`。', '']
open(f'{ROOT}/README.md', 'w').write('\n'.join(md))
print('寫出 README.md 與 MANIFEST.tsv：%d 個檔' % len(rows))
