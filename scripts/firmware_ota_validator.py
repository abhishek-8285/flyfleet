#!/usr/bin/env python3
"""Validates OTA manifest and artifact integrity for CI safety gates."""

from __future__ import annotations

import hashlib
import json
import base64
import subprocess
import sys
import tempfile
from pathlib import Path

SCHEMA_PATH = Path("docs/firmware/ota-manifest.schema.json")
TRUSTED_KEYS_PATH = Path("firmware/keys/trusted_keys.json")

REQUIRED_TOP_LEVEL = [
    "manifest_version",
    "release_id",
    "firmware_version",
    "hardware_compatibility",
    "min_bootloader_version",
    "artifact",
    "signature",
    "released_at",
]


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        while True:
            chunk = f.read(8192)
            if not chunk:
                break
            h.update(chunk)
    return h.hexdigest()


def canonical_signed_payload(manifest: dict) -> bytes:
    payload = {
        "artifact": {
            "sha256": manifest["artifact"]["sha256"],
            "size_bytes": manifest["artifact"]["size_bytes"],
            "uri": manifest["artifact"]["uri"],
        },
        "firmware_version": manifest["firmware_version"],
        "hardware_compatibility": manifest["hardware_compatibility"],
        "min_bootloader_version": manifest["min_bootloader_version"],
        "release_id": manifest["release_id"],
    }
    return json.dumps(payload, separators=(",", ":"), sort_keys=True).encode("utf-8")


def load_trusted_key(key_id: str) -> tuple[str, Path]:
    if not TRUSTED_KEYS_PATH.exists():
        raise ValueError(f"missing_trusted_keys={TRUSTED_KEYS_PATH}")
    trusted = json.loads(TRUSTED_KEYS_PATH.read_text(encoding="utf-8"))
    keys = trusted.get("keys")
    if not isinstance(keys, dict):
        raise ValueError(f"invalid_trusted_keys_format={TRUSTED_KEYS_PATH}")
    entry = keys.get(key_id)
    if not isinstance(entry, dict):
        raise ValueError(f"unknown_key_id={key_id}")
    algorithm = entry.get("algorithm")
    key_path_raw = entry.get("public_key_path")
    if not isinstance(algorithm, str) or not isinstance(key_path_raw, str):
        raise ValueError(f"invalid_key_entry={key_id}")
    key_path = Path(key_path_raw)
    if not key_path.exists():
        raise ValueError(f"missing_public_key={key_path}")
    return algorithm, key_path


def verify_ed25519_signature(public_key_path: Path, payload: bytes, signature: bytes) -> tuple[bool, str]:
    with tempfile.TemporaryDirectory(prefix="ota-signature-verify-") as tmpdir:
        payload_file = Path(tmpdir) / "payload.bin"
        sig_file = Path(tmpdir) / "signature.bin"
        payload_file.write_bytes(payload)
        sig_file.write_bytes(signature)
        run = subprocess.run(
            [
                "openssl",
                "pkeyutl",
                "-verify",
                "-pubin",
                "-inkey",
                str(public_key_path),
                "-rawin",
                "-in",
                str(payload_file),
                "-sigfile",
                str(sig_file),
            ],
            capture_output=True,
            text=True,
        )
        if run.returncode != 0:
            stderr = (run.stderr or "").strip()
            if stderr:
                return False, f"signature_verification_failed={stderr}"
            return False, "signature_verification_failed"
    return True, "cryptographic_signature_verification=PASS"


def verify_manifest_signature(manifest: dict) -> tuple[bool, str]:
    signature = manifest["signature"]
    key_id = signature.get("key_id")
    algorithm = signature.get("algorithm")
    signature_b64 = signature.get("value")

    if not isinstance(key_id, str) or not key_id:
        return False, "signature_key_id_missing"
    if not isinstance(algorithm, str) or not algorithm:
        return False, "signature_algorithm_missing"
    if not isinstance(signature_b64, str) or not signature_b64:
        return False, "signature_value_missing"

    try:
        signature_bytes = base64.b64decode(signature_b64, validate=True)
    except Exception:
        return False, "signature_value_not_base64"
    if not signature_bytes:
        return False, "signature_value_empty"

    try:
        trusted_algorithm, public_key_path = load_trusted_key(key_id)
    except ValueError as exc:
        return False, str(exc)

    if trusted_algorithm != algorithm:
        return False, f"signature_algorithm_mismatch key={trusted_algorithm} manifest={algorithm}"

    payload = canonical_signed_payload(manifest)
    if algorithm == "ed25519":
        return verify_ed25519_signature(public_key_path, payload, signature_bytes)
    return False, f"unsupported_signature_algorithm={algorithm}"


def validate_manifest(manifest_path: Path) -> int:
    if not SCHEMA_PATH.exists():
        print(f"missing_schema={SCHEMA_PATH}")
        return 1

    if not manifest_path.exists():
        print(f"missing_manifest={manifest_path}")
        return 1

    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))

    for k in REQUIRED_TOP_LEVEL:
        if k not in manifest:
            print(f"missing_field={k}")
            return 1

    artifact = manifest["artifact"]
    artifact_uri = artifact.get("uri", "")
    if not artifact_uri:
        print("missing_artifact_uri")
        return 1

    artifact_path = (manifest_path.parent / artifact_uri).resolve()
    if not artifact_path.exists():
        print(f"missing_artifact_file={artifact_path}")
        return 1

    expected_hash = artifact.get("sha256")
    actual_hash = sha256_file(artifact_path)
    if expected_hash != actual_hash:
        print(f"artifact_hash_mismatch expected={expected_hash} actual={actual_hash}")
        return 1

    expected_size = artifact.get("size_bytes")
    actual_size = artifact_path.stat().st_size
    if expected_size != actual_size:
        print(f"artifact_size_mismatch expected={expected_size} actual={actual_size}")
        return 1

    ok, message = verify_manifest_signature(manifest)
    if not ok:
        print(message)
        return 1

    print(f"manifest={manifest_path}")
    print(f"artifact={artifact_path.name}")
    print(f"artifact_sha256={actual_hash}")
    print(message)
    print("ota_manifest_validation=PASS")
    return 0


def main() -> int:
    manifest_arg = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("firmware/releases/sample/ota-manifest.json")
    return validate_manifest(manifest_arg)


if __name__ == "__main__":
    sys.exit(main())
