#!/usr/bin/env bash
# 印出一個函式的 IDA 反組譯，呼叫目標的 IDA 名稱後面附 seg:off 與分類。
#
#   decomp/tools/asm.sh <SEG:OFF>        例：decomp/tools/asm.sh 203A:0424
# 需要 workplace/ida/GIN3PS_UNPACKED.EXE.asm 與 decomp/ledger/functions.tsv。
# 只讀，在 Docker 內執行。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
key="${1:?用法: asm.sh <SEG:OFF>}"
test -f "$ROOT/workplace/ida/GIN3PS_UNPACKED.EXE.asm"
docker run --rm --network none --memory 256m --cpus 1 --pids-limit 32 \
  --log-opt max-size=10m --log-opt max-file=3 -u "$(id -u):$(id -g)" \
  -v "$ROOT/decomp/ledger:/l:ro" -v "$ROOT/workplace/ida:/i:ro" golang:1.24-bookworm python3 -I -c '
import re, sys
key = sys.argv[1].upper()
rows = [l.rstrip("\n").split("\t") for l in open("/l/functions.tsv")][1:]
by_name = {r[3]: r for r in rows}
me = next((r for r in rows if r[2] == key), None)
if me is None:
    sys.exit("找不到 " + key)
name = me[3]
out, on = [], False
for line in open("/i/GIN3PS_UNPACKED.EXE.asm", encoding="utf-8", errors="replace"):
    if re.match(r"^%s\s+proc\b" % re.escape(name), line):
        on = True
    if on:
        def tag(m):
            r = by_name.get(m.group(0))
            return "%s{%s %s}" % (m.group(0), r[2], r[8]) if r and m.group(0) != name else m.group(0)
        out.append(re.sub(r"\b(?:sub|nullsub|loc)_[0-9A-F]+\b", lambda m: tag(m) if m.group(0).startswith(("sub_", "nullsub_")) else m.group(0), line.rstrip("\n")))
    if on and re.match(r"^%s\s+endp\b" % re.escape(name), line):
        break
print("; %s  %s  size=%s  class=%s  far=%s  callers=%s" % (me[2], name, me[4], me[8], me[5], me[6]))
print("\n".join(out))
' "$key"
