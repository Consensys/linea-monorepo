#!/usr/bin/env python3
import json
import sys

FIELDS_IN = [
    "h0h1_be_input", "h2h3_be_input", "h4h5_be_input", "h6h7_be_input",
    "m0m1_be", "m2m3_be", "m4m5_be", "m6m7_be", "m8m9_be", "m10m11_be",
    "m12m13_be", "m14m15_be", "t0t1_be",
]
FIELDS_OUT = ["h0h1_be", "h2h3_be", "h4h5_be", "h6h7_be"]


def be(value, length):
    return int(value).to_bytes(length, "big").hex()


with open(sys.argv[1]) as f:
    for n, line in enumerate(f, 1):
        line = line.strip()
        if not line or line.startswith(";;"):
            continue
        case = json.loads(line)["F"]
        data = be(case["r"][0], 4)
        data += "".join(be(case[k][0], 16) for k in FIELDS_IN)
        data += be(case["f"][0], 1)
        data += "".join(be(case[k][0], 16) for k in FIELDS_OUT)
        print(f"{n}:0x{data}")
