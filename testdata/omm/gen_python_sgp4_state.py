"""Generate the independent 25544 OMM SGP4 state reference.

Run with python-sgp4 2.22 and its accelerated extension. The output records the
reference distribution version, source fixture digest, exact binary64 hex
values, and decimal values suitable for the Go test's JSON decoder.
"""

from __future__ import annotations

import hashlib
import json
import platform
from pathlib import Path

import sgp4
from sgp4 import omm
from sgp4.api import Satrec, accelerated


HERE = Path(__file__).resolve().parent
INPUT = HERE / "25544.json"
OUTPUT = HERE / "python_sgp4_25544_state.json"
EXPECTED_VERSION = "2.22"
TIME_OFFSETS_MIN = (0.0, 10.0)


def main() -> None:
    if sgp4.__version__ != EXPECTED_VERSION or not accelerated:
        raise SystemExit("requires accelerated python-sgp4 2.22")

    source = INPUT.read_bytes()
    document = json.loads(source)
    if not isinstance(document, list) or len(document) != 1 or not isinstance(document[0], dict):
        raise SystemExit("25544.json must contain exactly one OMM object")

    satellite = Satrec()
    omm.initialize(satellite, document[0])
    states = []
    for minutes in TIME_OFFSETS_MIN:
        error, position, velocity = satellite.sgp4_tsince(minutes)
        if error != 0:
            raise SystemExit(f"python-sgp4 refused {minutes} minutes since epoch: error={error}")
        position = tuple(float(value) for value in position)
        velocity = tuple(float(value) for value in velocity)
        states.append(
            {
                "minutes_since_epoch": minutes,
                "error": error,
                "position_km": position,
                "position_km_hex": [value.hex() for value in position],
                "velocity_km_s": velocity,
                "velocity_km_s_hex": [value.hex() for value in velocity],
            }
        )

    result = {
        "generator": "python-sgp4 2.22 accelerated Satrec.sgp4_tsince",
        "python_version": platform.python_version(),
        "sgp4_version": sgp4.__version__,
        "input_fixture": INPUT.name,
        "input_fixture_sha256": hashlib.sha256(source).hexdigest(),
        "states": states,
    }
    OUTPUT.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
