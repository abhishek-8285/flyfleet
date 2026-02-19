#!/usr/bin/env bash
set -euo pipefail

MODULE_DIR="apps/mobile/vehicle-device/android-device"

if [[ -x "${MODULE_DIR}/gradlew" ]]; then
  echo "vehicle_device_test_runner=wrapper"
  "${MODULE_DIR}/gradlew" -p "${MODULE_DIR}" --no-daemon test
  exit 0
fi

if command -v gradle >/dev/null 2>&1; then
  gradle_version="$(gradle -v 2>/dev/null | awk '/^Gradle / {print $2; exit}')"
  gradle_major="${gradle_version%%.*}"
  if [[ -z "${gradle_version}" || -z "${gradle_major}" || "${gradle_major}" -lt 7 ]]; then
    echo "vehicle_device_test_runner=system_gradle_incompatible"
    echo "error=gradle_7_or_newer_required_for_vehicle_device_module"
    echo "hint=use_ci_workflow_or_install_modern_gradle_or_add_wrapper"
    exit 2
  fi
  echo "vehicle_device_test_runner=system_gradle"
  gradle -p "${MODULE_DIR}" --no-daemon test
  exit 0
fi

echo "vehicle_device_test_runner=unavailable"
echo "error=neither_gradle_wrapper_nor_gradle_found"
exit 2
