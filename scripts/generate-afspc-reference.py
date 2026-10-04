#!/usr/bin/env python3
"""Independent AFSPC-mode Vallado propagation at Skyfield-derived UTC splits.

Run from the repository root with skyfield==1.54 and sgp4==2.22 installed.
This invokes python-sgp4's Vallado propagation directly. It parses the TLE,
initializes AFSPC mode explicitly with Satrec.sgp4init(..., 'a', ...), then
restores the exact split epoch before propagation. It does not call the Go or
Sidereon propagation implementation.
"""
import hashlib
import json
import math
from pathlib import Path

import skyfield
from skyfield.api import load
from skyfield.constants import DAY_S
import sgp4
from sgp4.api import Satrec, WGS72

TLE = Path("testdata/iss.tle")
EXPECTED_SHA256 = "b7866a81fefe47fa04a271b86cc129064748b36d906a335061ca1b5cbc9dc10b"
EXPECTED_SKYFIELD = "1.54"
EXPECTED_SGP4 = "2.22"

def main():
    if skyfield.__version__ != EXPECTED_SKYFIELD or sgp4.__version__ != EXPECTED_SGP4:
        raise SystemExit(f"expected skyfield {EXPECTED_SKYFIELD} and sgp4 {EXPECTED_SGP4}; got {skyfield.__version__} and {sgp4.__version__}")
    data = TLE.read_bytes()
    digest = hashlib.sha256(data).hexdigest()
    if digest != EXPECTED_SHA256:
        raise SystemExit(f"unexpected TLE SHA-256 {digest}")
    lines = [line.decode("ascii").rstrip("\r\n") for line in data.splitlines() if line.strip()]
    if len(lines) != 2:
        raise SystemExit(f"expected two TLE lines, got {len(lines)}")
    # Satrec.twoline2rv uses the extension's exact whole/fraction epoch split.
    # Reinitialize those parsed elements with the published AFSPC mode. The
    # C++ sgp4init API takes one epoch scalar, so form it from the split epoch
    # in days since 1949-12-31, then restore both native split fields exactly.
    parsed = Satrec.twoline2rv(lines[0], lines[1], WGS72)
    if parsed.error:
        raise SystemExit(f"TLE parse error {parsed.error}")
    sat = Satrec()
    epoch_days = (parsed.jdsatepoch - 2433281.5) + parsed.jdsatepochF
    sat.sgp4init(
        WGS72, "a", parsed.satnum, epoch_days, parsed.bstar, parsed.ndot,
        parsed.nddot, parsed.ecco, parsed.argpo, parsed.inclo, parsed.mo,
        parsed.no_kozai, parsed.nodeo,
    )
    sat.jdsatepoch = parsed.jdsatepoch
    sat.jdsatepochF = parsed.jdsatepochF
    if sat.operationmode != "a":
        raise SystemExit(f"AFSPC opsmode was not retained: {sat.operationmode!r}")
    ts = load.timescale(builtin=True)
    records = []
    for minute in (0, 1):
        t = ts.utc(2018, 7, 3, 12, minute, 0)
        jd = float(t.whole)
        fraction = float(t.tai_fraction - t._leap_seconds() / DAY_S)
        # Satrec.sgp4 uses the native split JD fields; the C++ extension applies
        # Vallado propagation after the explicit AFSPC reinitialization above.
        error, position, velocity = sat.sgp4(jd, fraction)
        if error != 0 or not all(math.isfinite(value) for value in (*position, *velocity)):
            raise SystemExit(f"Vallado reference failed at 12:{minute:02d}: error={error}, r={position}, v={velocity}")
        records.append({
            "utc": f"2018-07-03T12:{minute:02d}:00Z",
            "jd_hex": jd.hex(),
            "fraction_hex": fraction.hex(),
            "error": int(error),
            "position_km": list(position),
            "position_km_hex": [float(value).hex() for value in position],
            "velocity_km_s": list(velocity),
            "velocity_km_s_hex": [float(value).hex() for value in velocity],
        })
    print(json.dumps({
        "generator": "python-sgp4 Vallado propagation + Skyfield UTC split",
        "skyfield_version": skyfield.__version__,
        "sgp4_version": sgp4.__version__,
        "gravity_model": "WGS-72",
        "opsmode": "a (AFSPC)",
        "tle_path": str(TLE),
        "tle_sha256": digest,
        "records": records,
    }, indent=2, sort_keys=True))

if __name__ == "__main__":
    main()
