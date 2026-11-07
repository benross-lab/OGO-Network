#!/usr/bin/env python3
"""P1 24-hour continuous observation of the Milestone C wired chain.

Polls all 4 validators every 60s and appends one line per poll to
p1-24h.log (same directory as this script):
  <UTC timestamp> height=<min>-<max> hash_agree=<bool> heights=[...] hashes=[...]

Every hour it also appends a summary line with the block rate over
the last hour and the number of missed polls. Run detached:
  nohup python3 scripts/p1_observe.py > p1-24h.out 2>&1 &
The observer never writes to the chain and never restarts nodes.
"""
import json
import time
import urllib.request
from datetime import datetime, timezone

NODES = {
    "val1": "http://127.0.0.1:27657",
    "val2": "http://127.0.0.1:27667",
    "val3": "http://127.0.0.1:27677",
    "val4": "http://127.0.0.1:27687",
}
INTERVAL_S = 60
DURATION_S = 24 * 3600
LOG = "p1-24h.log"


def status(url):
    with urllib.request.urlopen(url + "/status", timeout=10) as r:
        return json.load(r)["result"]["sync_info"]


def main():
    start = time.time()
    start_height = None
    last_hour_mark = start
    hour_start_height = None
    missed = 0
    with open(LOG, "a") as log:
        log.write(
            f"# P1 started {datetime.now(timezone.utc).isoformat()} "
            f"on the post-upgrade v2 chain\n"
        )
        log.flush()
        while time.time() - start < DURATION_S:
            now = time.time()
            ts = datetime.now(timezone.utc).isoformat()
            heights = {}
            hashes = {}
            for name, url in NODES.items():
                try:
                    info = status(url)
                    heights[name] = int(info["latest_block_height"])
                    hashes[name] = info["latest_block_hash"]
                except Exception as e:  # noqa: BLE001 - record and continue
                    heights[name] = None
                    hashes[name] = f"ERR:{e}"
                    missed += 1
            if start_height is None:
                start_height = max(h for h in heights.values() if h is not None)
                hour_start_height = start_height
            hs = [h for h in heights.values() if h is not None]
            agree = len(set(hashes.values())) == 1 and len(hs) == 4
            log.write(
                f"{ts} height={min(hs)}-{max(hs)} hash_agree={agree} "
                f"heights={heights} hashes={hashes}\n"
            )
            if now - last_hour_mark >= 3600:
                elapsed_h = (now - start) / 3600
                rate = (max(hs) - hour_start_height) / ((now - last_hour_mark) / 3600)
                log.write(
                    f"# HOUR {elapsed_h:.1f}h: blocks/hour={rate:.1f} "
                    f"(~{3600 / rate:.2f}s/block) missed_polls={missed} "
                    f"height={max(hs)}\n"
                )
                last_hour_mark = now
                hour_start_height = max(hs)
            log.flush()
            time.sleep(INTERVAL_S)
        log.write(f"# P1 complete {datetime.now(timezone.utc).isoformat()}\n")


if __name__ == "__main__":
    main()
