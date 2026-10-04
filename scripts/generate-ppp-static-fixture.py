#!/usr/bin/env python3
"""Create the public known-truth PPP product from the approved SP3 source.

The generator selects the approved product's seventh epoch, repeats its eight
position records unchanged across all 13 regular epochs, and replaces satellite
clock values with explicit zeroes. It refuses a source hash, epoch count, cadence,
or satellite-layout change. The separate observation recipe in ppp_test.go uses
the core's documented RTKLIB geodist first-order Sagnac equation at known ECEF
receiver coordinates; no Sidereon solve output enters this fixture.
"""
import hashlib
import json
import math
from datetime import datetime, timedelta
from pathlib import Path

SOURCE = Path("testdata/trimmed.sp3")
OUTPUT = Path("testdata/ppp-static-known-truth.sp3")
MANIFEST = Path("testdata/ppp-static-known-truth.json")
SOURCE_SHA256 = "fd0aa5e7047b67b41a567e1d8f9307b3d28af448dd5cf288098fc1300358c6ea"
RECEIVER_M = (4_500_000.0, 500_000.0, 4_500_000.0)
SATELLITES = ("G08", "G10", "G16", "G18", "G20", "G21")
OMEGA_RAD_S = 7.2921151467e-5
C_M_S = 299_792_458.0


def main():
    source = SOURCE.read_bytes()
    source_hash = hashlib.sha256(source).hexdigest()
    if source_hash != SOURCE_SHA256:
        raise SystemExit(f"approved SP3 source hash mismatch: {source_hash}")
    lines = source.decode("ascii").splitlines()
    epochs = []
    target = None
    for line in lines:
        if line.startswith("* "):
            epochs.append(line)
            if len(epochs) == 7:
                target = {}
        elif line.startswith("P") and target is not None and len(epochs) == 7:
            target[line[1:4]] = line
    if len(epochs) != 13 or target is None:
        raise SystemExit(f"expected 13 epochs, found {len(epochs)}")
    for before, after in zip(epochs, epochs[1:]):
        def parse_epoch(line):
            fields = line.split()
            base = datetime(int(fields[1]), int(fields[2]), int(fields[3]), int(fields[4]), int(fields[5]))
            return base + timedelta(seconds=float(fields[6]))
        if (parse_epoch(after) - parse_epoch(before)).total_seconds() != 900.0:
            raise SystemExit("source SP3 epoch cadence is not 900 seconds")
    if len(target) != 8:
        raise SystemExit(f"expected eight position records at source epoch 6, got {len(target)}")

    out = []
    in_epochs = 0
    for line in lines:
        if line.startswith("* "):
            in_epochs += 1
            out.append(line)
        elif line.startswith("P"):
            token = line[1:4]
            source_line = target.get(token)
            if source_line is None:
                raise SystemExit(f"satellite {token} missing from source reference epoch")
            # Keep the source's exact coordinate fields; set a stated zero clock.
            out.append(source_line[:46] + f"{0.0:14.6f}" + source_line[60:])
        elif line.startswith("V"):
            # Remove velocity rows so repeated positions describe static states.
            continue
        else:
            out.append(line)
    if in_epochs != 13:
        raise SystemExit(f"internal epoch count changed: {in_epochs}")
    fixture = ("\n".join(out) + "\n").encode("ascii")
    OUTPUT.write_bytes(fixture)
    positions = {}
    for satellite in SATELLITES:
        line = target[satellite]
        xyz_m = [float(line[4 + i * 14:18 + i * 14]) * 1000.0 for i in range(3)]
        dx, dy, dz = (xyz_m[i] - RECEIVER_M[i] for i in range(3))
        euclidean = math.sqrt(dx * dx + dy * dy + dz * dz)
        sagnac = OMEGA_RAD_S * (xyz_m[0] * RECEIVER_M[1] - xyz_m[1] * RECEIVER_M[0]) / C_M_S
        positions[satellite] = {
            "position_ecef_m": xyz_m,
            "geometric_range_m": euclidean,
            "first_order_sagnac_m": sagnac,
            "code_phase_range_m": euclidean + sagnac,
        }
    manifest = {
        "generator": "generate_ppp_static_fixture.py",
        "source_path": str(SOURCE),
        "source_sha256": source_hash,
        "fixture_path": str(OUTPUT),
        "fixture_sha256": hashlib.sha256(fixture).hexdigest(),
        "epoch_count": 13,
        "cadence_seconds": 900,
        "static_reference_epoch_index_zero_based": 6,
        "satellite_clock_seconds": 0.0,
        "receiver_ecef_m": list(RECEIVER_M),
        "measurement_model": "Euclidean range + omega*(sat_x*receiver_y - sat_y*receiver_x)/c; code=phase; troposphere and other corrections disabled",
        "independent_formula_reference": "RTKLIB 2.4.3 b34, git 75a2e56275485b21a67bd35bc94bbeb8936e1a74, src/rtkcmn.c geodist (first-order Sagnac term)",
        "constants": {"omega_rad_s": OMEGA_RAD_S, "c_m_s": C_M_S},
        "satellites": positions,
    }
    MANIFEST.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
    print(json.dumps({"fixture": str(OUTPUT), "sha256": manifest["fixture_sha256"], "manifest": str(MANIFEST)}, sort_keys=True))

if __name__ == "__main__":
    main()
