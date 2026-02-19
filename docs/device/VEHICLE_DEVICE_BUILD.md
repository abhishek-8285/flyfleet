# Vehicle Device Build and Test

## Module

- path: `apps/mobile/vehicle-device/android-device`
- build system: Gradle Kotlin/JVM module for runtime core logic tests.

## Local Test

Preferred:

```bash
bash scripts/run_vehicle_device_tests.sh
```

Prerequisite:

- Gradle `7+` (or module-local `gradlew`) is required.

Fallback:

```bash
gradle -p apps/mobile/vehicle-device/android-device --no-daemon test
```

## CI Test

The workflow `.github/workflows/vehicle-device-foundation.yml` runs:

- foundation contract checker
- Gradle unit tests for the vehicle-device module

## Why JVM First

Current runtime core is platform-agnostic Kotlin (`state machine`, `resolver`, `contracts`).
Running JVM tests early provides deterministic anti-regression checks before Android framework wiring expands.
