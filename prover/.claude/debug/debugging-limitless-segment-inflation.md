# Debugging: Limitless Prover Segment Inflation

**Branch**: `prover/eip7702-setup`
**Trace**: `10-19.conflated.beta-v5.0-rc3-93d7863.25.12.0-linea4.1.lt.gz` (19 txs, 6.6M gas)
**Symptom**: ~156 segments (78 LPP + 78 GL) for a tiny trace with <2% module utilization.

---

## Root Cause

`manual_shift.go:Assign` always produces a `*Regular` smart vector (flat array). When
`AssignColumn` ran without `DisableAssignmentSizeReduction` (the default), the optimizer
`TryReduceSizeRight` stripped the trailing zeros introduced by the shift operation and
re-encoded the result as a right-padded `PaddedCircularWindow` with a near-full window
(e.g. `131072 - offset = 131070`).

`RecordAssignmentStats` then called `CoWindowRange` on that window, got `density = 131070`,
and set `NbActiveRows = 131070` — even though only ~100 rows had real data.

Segment count = `DivCeil(131070, segmentSize)` → **32 segments** per module instead of 1.

### Offending modules (from `run_limitless.log`)
| Module       | Segments | QBM driving it   |
|--------------|----------|------------------|
| HUB          | 32       | rom QBM          |
| BN-EC-OPS    | 32       | ecdata QBM       |
| KECCAK       | 4        | keccak.$ret QBM  |
| ECDSA        | 4        | ecrecover QBM    |

### Debug stats that revealed the bug (for `Module_489_keccak.$ret`)
```
NbColumns:81           NbPragmaLeftPadded:68   NbPragmaRightPadded:0
NbAssignedLeftPadded:68  NbAssignedRightPadded:10  NbAssignedFullColumn:2
NbActiveRows:131070
LastRightPadded:MANUALLY_SHIFTED_keccak.limb/10971_COL
```

Key observations:
- `NbPragmaRightPadded:0` but `NbAssignedRightPadded:10` → the 10 MANUALLY_SHIFTED columns
  had no right-padding pragma but were being stored as right-padded by `TryReduceSizeRight`.
- The `2^n - epsilon` pattern (131072 - 2 = 131070) = `size - min_shift_offset`.

### Why `TryReduceSizeRight` misclassifies the shifted column

For `m.Offset > 0`, `Assign` produces:
```
[0 ... 0 | real data (H rows) | 0 ... 0]
           ↑                    ↑
    data in middle          m.Offset trailing zeros
```

`TryReduceSizeRight` scans only from the RIGHT. It strips the `m.Offset` trailing zeros and
declares the window as `[0, size - m.Offset)`. The `131070` leading zeros are not stripped.

---

## Fix

**File**: `prover/protocol/dedicated/manual_shift.go:99`

```go
// Before:
run.AssignColumn(m.Natural.ID, res)

// After:
run.AssignColumn(m.Natural.ID, res, wizard.DisableAssignmentSizeReduction)
```

### Why it works

With `DisableAssignmentSizeReduction`, the `*Regular` is stored as-is. `CoWindowRange` on a
`*Regular` returns `(0, fullSize)` (the `default` branch in `smartvectors.go:527`), so both
`isRightPadded` and `isLeftPadded` are true. `RecordAssignmentStats` classifies this as
`NbAssignedFullColumn` and **skips it** with `continue` (line 1117), meaning the shifted
columns do NOT contribute to `NbActiveRows`. Only the 68 real left-padded `keccak.*` columns
contribute their actual density (~H rows) → `NbActiveRows ≈ H` → 1 segment.

---

## How to add the debug logging

In `module_discovery_standard.go`, inside `SegmentBoundaries` after calling
`RecordAssignmentStats`:

```go
stats = mod.RecordAssignmentStats(run)
logrus.Infof("module-segmentation-stats: qbm=%v stats=%++v", mod.ModuleName, stats)
```

The `QueryBasedAssignmentStatsRecord` struct (line 107) prints all the fields you need.

---

## Key files

| File | Role |
|------|------|
| `protocol/dedicated/manual_shift.go` | Bug fix location |
| `protocol/wizard/prover.go:453` | `AssignColumn` + `TryReduceSizeRight` logic |
| `maths/common/smartvectors/smartvectors.go:313` | `TryReduceSizeRight` implementation |
| `maths/common/smartvectors/smartvectors.go:516` | `CoWindowRange` — `default` returns `(0, Len())` |
| `protocol/distributed/module_discovery_standard.go:1003` | `RecordAssignmentStats` |
| `protocol/distributed/module_discovery_standard.go:898` | `SegmentBoundaries` |
| `zkevm/limitless.go:96` | `DiscoveryAdvices` with BaseSize / Cluster / Regexp entries |

---

## Other bugs found (not fixed)

1. **Cache ignores `segmentSize`** (`module_discovery_standard.go:901`): cache key is
   `unsafe.Pointer(run)` only. If two columns in the same QBM have different `segmentSize`,
   the second call returns the first's cached result — could cause a `DivExact` panic.

2. **`foundAny` never set in `CoWindowRange`** (`smartvectors.go:545`): in the
   `*PaddedCircularWindow` branch, `foundAny = true` is placed AFTER `continue`, so it is
   never executed. Affects multi-vector calls only (single-vector calls in
   `RecordAssignmentStats` are unaffected).