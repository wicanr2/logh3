# .AMG 解碼器能不能套到其他圖庫？判準一樣：每張都要剛好用完。
import sys, struct, os
sys.path.insert(0, '/work/tools/scripts')
from amg_decode import decode
for name in ('COMANI.BMG', 'SPRITE.SPR', '3DBLOCK.SPR', 'COM_FLAG.SPR',
             'FRAME.BLK', 'BLOCK_3.BLK'):
    p = '/orig/' + name
    if not os.path.exists(p):
        print(name, '不存在'); continue
    f = open(p, 'rb').read()
    first = struct.unpack_from('<I', f, 2)[0]
    print(f'\n{name}  {len(f)} bytes  前 16: ' + ' '.join(f'{b:02X}' for b in f[:16]))
    if first % 6 or first == 0 or first > len(f):
        print(f'  第一筆位移 {first} → 不是 .AMG 那種 6-byte 索引表')
        continue
    n = first // 6
    idx = [struct.unpack_from('<HI', f, i*6) for i in range(n)]
    ok = sum(1 for s, o in idx if o + s <= len(f))
    print(f'  索引 {n} 筆，落在檔內 {ok}')
    good = 0; sizes = set()
    for s, o in idx[:n]:
        try:
            W, H, px, used = decode(f[o:o+s])
            if used == s:
                good += 1
            sizes.add((W, H))
        except Exception:
            pass
    print(f'  剛好用完 {good}/{n}  尺寸 {sorted(sizes)[:5]}')
