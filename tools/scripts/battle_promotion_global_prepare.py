"""第86輪修正STELLAR觀測基底，舊來源與收據保持。"""
from pathlib import Path
import hashlib

p = Path('workplace/promotion79-r85-recalc.go')
b = p.read_bytes()
assert hashlib.sha256(b).hexdigest() == '5054248450c84bc7b10d4f6467e538c02323d7ac83f6c77c26d8665c44d2bd72'
s = b.decode()
assert s.count('0xA7BC') == 1
s = s.replace('0xA7BC', '0xA79C')
s = s.replace('"stellar":hex.EncodeToString', '''"planet":hex.EncodeToString(append(o.Bytes(far(0x54B3,0xCFB8),64),o.Bytes(far(o.Word(far(0x54B3,0x5338)),o.Word(far(0x54B3,0x5336))),81*82)...)),
  "paths":hex.EncodeToString(o.Bytes(far(o.Word(far(0x54B3,0xD064)),o.Word(far(0x54B3,0xD062))),6144)),
  "stellar":hex.EncodeToString''')
s = s.replace('logh3.battle-promotion-recalc-native.v1', 'logh3.battle-promotion-global-native.v1')
s = s.replace('bash tools/battle_promotion_recalc_oracle.sh <a|b>', 'bash tools/battle_promotion_global_oracle.sh <a|b>')
dest = Path('workplace/promotion79-r86-global.go')
with dest.open('x') as f:
    f.write(s)
print(dest, hashlib.sha256(dest.read_bytes()).hexdigest())
