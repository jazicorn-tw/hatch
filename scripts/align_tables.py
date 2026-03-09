#!/usr/bin/env python3
"""Align markdown table columns to satisfy MD060 'aligned' style."""

import re
import sys

ESCAPED_PIPE = "\x00"


def parse_row(line):
    line = line.replace("\\|", ESCAPED_PIPE).strip()
    if line.startswith("|"):
        line = line[1:]
    if line.endswith("|"):
        line = line[:-1]
    return [c.strip().replace(ESCAPED_PIPE, "\\|") for c in line.split("|")]


def is_separator_row(cells):
    return all(re.match(r"^:?-+:?$", c) for c in cells)


def align_table(table_lines):
    rows = [parse_row(line) for line in table_lines]
    num_cols = max(len(r) for r in rows)

    for row in rows:
        while len(row) < num_cols:
            row.append("")

    col_widths = [3] * num_cols
    for row in rows:
        if not is_separator_row(row):
            for j, cell in enumerate(row):
                col_widths[j] = max(col_widths[j], len(cell))

    result = []
    for row in rows:
        if is_separator_row(row):
            formatted = ["-" * col_widths[j] for j in range(num_cols)]
        else:
            formatted = [row[j].ljust(col_widths[j]) for j in range(num_cols)]
        result.append("| " + " | ".join(formatted) + " |")

    return result


def process_file(path):
    with open(path) as f:
        lines = f.read().splitlines()

    out = []
    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()
        if stripped.startswith("|") and stripped.endswith("|") and "|" in stripped[1:-1]:
            table = []
            while i < len(lines):
                s = lines[i].strip()
                if s.startswith("|") and s.endswith("|"):
                    table.append(lines[i])
                    i += 1
                else:
                    break
            out.extend(align_table(table))
        else:
            out.append(line)
            i += 1

    with open(path, "w") as f:
        f.write("\n".join(out) + "\n")

    print(f"aligned: {path}")


for path in sys.argv[1:]:
    process_file(path)
