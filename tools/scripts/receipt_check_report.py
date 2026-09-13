"""把 receipt-check.tsv 的每個名稱最後一列整理成 JSON 報表。"""

import datetime
import json
import pathlib
import sys


def main() -> None:
    if len(sys.argv) != 4:
        raise SystemExit(
            "用法：receipt_check_report.py <輸入.tsv> <輸出.json> <dosgolem-head>"
        )

    source = pathlib.Path(sys.argv[1])
    output = pathlib.Path(sys.argv[2])
    seen: dict[str, dict[str, str | int]] = {}
    for line_no, line in enumerate(source.read_text(encoding="utf-8").splitlines(), 1):
        fields = line.split("\t")
        if len(fields) != 3:
            raise SystemExit(f"{source}:{line_no}: 應有三欄，實際為 {len(fields)} 欄")
        name, screen_diff, palette_diff = fields
        try:
            screen = int(screen_diff)
            palette = int(palette_diff)
        except ValueError as error:
            raise SystemExit(f"{source}:{line_no}: 差異數不是整數") from error
        seen[name] = {
            "name": name,
            "screen_diff": screen,
            "pal_diff": palette,
        }

    rows = sorted(seen.values(), key=lambda row: str(row["name"]))
    mismatched = [
        str(row["name"])
        for row in rows
        if row["screen_diff"] != 0 or row["pal_diff"] != 0
    ]
    report = {
        "generated": datetime.datetime.now().astimezone().isoformat(timespec="seconds"),
        "golem_head": sys.argv[3],
        "checked": len(rows),
        "mismatched": mismatched,
        "results": rows,
    }
    output.write_text(
        json.dumps(report, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(
        f"\n比對 {len(rows)} 份：{len(rows) - len(mismatched)} 份逐位元組一致、"
        f"{len(mismatched)} 份對不上"
    )


if __name__ == "__main__":
    main()
