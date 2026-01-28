# Hook Libraries

This directory contains compiled hook libraries for process interception.

## Structure
- `bin/` - Compiled hook libraries by platform
- `src/` - Hook library source code (C)

## Implementation Status
Hook libraries will be implemented in Phase 6.

Supported platforms:
- Linux (amd64, arm64) via LD_PRELOAD
- macOS (amd64, arm64) via DYLD_INSERT_LIBRARIES