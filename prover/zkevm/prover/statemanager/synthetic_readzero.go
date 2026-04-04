package statemanager

import (
	"encoding/binary"
	"log"
	"slices"
	"strings"

	execstatemanager "github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
)

// syntheticEntry represents a storage entry identified in the SCP that needs a
// synthetic ReadZero trace injected into the Shomei data. This handles the edge
// case where the VALUE_ZEROIZATION constraint incorrectly checks valueCurr
// instead of valueNext, preventing the insertion filter from being set to 0.
type syntheticEntry struct {
	address    types.EthAddress
	storageKey types.FullBytes32
	block      int // 1-based block/batch number
}

// injectSyntheticReadZeroTraces detects insertion segments with transient
// 0→0 storage writes and injects synthetic ReadZero traces so the hub→summary
// lookup has matching entries without needing to change the constraints.
func (sm *StateManager) injectSyntheticReadZeroTraces(run *wizard.ProverRuntime, shomeiTraces *[][]execstatemanager.DecodedTrace) {
	if sm.StateSummary.ArithmetizationLink == nil {
		return
	}
	scp := &sm.StateSummary.ArithmetizationLink.Scp

	entries := detectTransientInsertionEntries(run, scp)
	if len(entries) == 0 {
		return
	}

	// Group entries by (block, address) so we create one empty trie per account
	type addrBlock struct {
		address types.EthAddress
		block   int
	}
	grouped := map[addrBlock][]types.FullBytes32{}
	for _, e := range entries {
		key := addrBlock{address: e.address, block: e.block}
		grouped[key] = append(grouped[key], e.storageKey)
	}

	// For each (address, block), generate ReadZero proofs and inject them
	for ab, keys := range grouped {
		// The traces array is 0-indexed but block numbers are 1-based
		blockIdx := ab.block - 1
		if blockIdx < 0 || blockIdx >= len(*shomeiTraces) {
			continue
		}

		// Create an empty storage trie for this account and generate the proofs
		emptyTrie := execstatemanager.NewStorageTrie(ab.address)
		var syntheticTraces []execstatemanager.DecodedTrace
		for _, storageKey := range keys {
			readZeroTrace := emptyTrie.ReadZeroAndProve(storageKey)
			syntheticTraces = append(syntheticTraces, execstatemanager.DecodedTrace{
				Location:   ab.address.Hex(),
				Type:       execstatemanager.READ_ZERO_TRACE_CODE,
				Underlying: readZeroTrace,
			})
		}

		// Find the insertion point: right after the world-state trace for this
		// account, before any existing storage traces for the same account.
		blockTraces := (*shomeiTraces)[blockIdx]
		insertIdx := findStorageInsertionPoint(blockTraces, ab.address)
		if insertIdx < 0 {
			// Account not found in traces; this shouldn't happen for a valid
			// insertion segment, but skip gracefully.
			continue
		}

		log.Printf("synthetic_readzero: injecting %d synthetic ReadZero traces for address=%s block=%d at trace index %d",
			len(syntheticTraces), ab.address.Hex(), ab.block, insertIdx)

		// Insert synthetic traces at the found position
		(*shomeiTraces)[blockIdx] = slices.Insert(blockTraces, insertIdx, syntheticTraces...)
	}
}

// findStorageInsertionPoint finds the index in blockTraces where synthetic
// storage traces should be injected for the given account.
//
// For newly created accounts (insertions), Shomei emits storage traces BEFORE
// the WS insertion trace. We must inject the ReadZero BEFORE the first existing
// storage trace so it is processed first on the empty trie, preserving root
// chaining in the state summary.
//
// Falls back to the position after the WS trace if no storage traces exist.
// Returns -1 if the account is not found.
func findStorageInsertionPoint(blockTraces []execstatemanager.DecodedTrace, targetAddr types.EthAddress) int {
	targetHex := targetAddr.Hex()
	// Look for the first storage trace for this address. Storage traces have
	// Location = address hex (e.g. "0xabc..."), unlike WS traces ("0x").
	for i, trace := range blockTraces {
		if trace.Location == execstatemanager.WS_LOCATION {
			continue
		}
		if strings.EqualFold(trace.Location, targetHex) {
			return i
		}
	}
	// Fallback: no storage traces found, insert after the WS trace.
	for i, trace := range blockTraces {
		if trace.Location != execstatemanager.WS_LOCATION {
			continue
		}
		addr, err := trace.GetRelatedAccount()
		if err != nil {
			continue
		}
		if addr == targetAddr {
			return i + 1
		}
	}
	return -1
}

// detectTransientInsertionEntries scans the SCP to find insertion segments where
// valueNext=0 on the last KOC row but valueCurr≠0 on some row in the segment.
// These are entries that the broken VALUE_ZEROIZATION constraint prevents from
// being filtered out via filterAccountInsert.
func detectTransientInsertionEntries(run *wizard.ProverRuntime, scp *statesummary.HubColumnSet) []syntheticEntry {
	var entries []syntheticEntry
	lastSegmentStart := 0

	for index := 0; index < scp.PeekAtStorage.Size(); index++ {
		isStorage := scp.PeekAtStorage.GetColAssignmentAt(run, index)
		if !isStorage.IsOne() {
			continue
		}

		firstKOC := scp.FirstKOCBlock.GetColAssignmentAt(run, index)
		if firstKOC.IsOne() {
			lastSegmentStart = index
		}

		lastKOC := scp.LastKOCBlock.GetColAssignmentAt(run, index)
		if !lastKOC.IsOne() {
			continue
		}

		// Must be an insertion segment: ExistsFirstInBlock=0, ExistsFinalInBlock=1
		existsStart := scp.ExistsFirstInBlock.GetColAssignmentAt(run, lastSegmentStart)
		if !existsStart.IsZero() {
			continue
		}
		existsEnd := scp.ExistsFinalInBlock.GetColAssignmentAt(run, index)
		if !existsEnd.IsOne() {
			continue
		}

		// ValueNext must be 0 on last KOC row (all limbs)
		if !allLimbsZero(run, scp.ValueHINext[:], index) || !allLimbsZero(run, scp.ValueLONext[:], index) {
			continue
		}

		// Check if valueCurr OR valueNext is non-zero on ANY row in the segment.
		// This matches the overly-strict check in the broken assignment that
		// prevents filterAccountInsert from being set to 0.
		hasNonZero := false
		for j := lastSegmentStart; j <= index; j++ {
			for i := range common.NbLimbU128 {
				vchi := scp.ValueHICurr[i].GetColAssignmentAt(run, j)
				vclo := scp.ValueLOCurr[i].GetColAssignmentAt(run, j)
				vnhi := scp.ValueHINext[i].GetColAssignmentAt(run, j)
				vnlo := scp.ValueLONext[i].GetColAssignmentAt(run, j)
				if !vchi.IsZero() || !vclo.IsZero() || !vnhi.IsZero() || !vnlo.IsZero() {
					hasNonZero = true
					break
				}
			}
			if hasNonZero {
				break
			}
		}
		if !hasNonZero {
			// All values are zero throughout the segment; the existing filter
			// handles this case correctly. No synthetic trace needed.
			continue
		}

		// This is a problematic entry: extract address, key, block
		addr := extractEthAddressFromScp(run, scp, index)
		key := extractStorageKeyFromScp(run, scp, index)
		block := extractBlockNumberFromScp(run, scp, index)

		log.Printf("synthetic_readzero: detected transient insertion entry at SCP index %d: address=%s key=%s block=%d",
			index, addr.Hex(), key.Hex(), block)

		entries = append(entries, syntheticEntry{
			address:    addr,
			storageKey: key,
			block:      block,
		})
	}
	return entries
}

// allLimbsZero checks if all limb columns are zero at the given index.
func allLimbsZero(run *wizard.ProverRuntime, cols []ifaces.Column, index int) bool {
	for _, col := range cols {
		v := col.GetColAssignmentAt(run, index)
		if !v.IsZero() {
			return false
		}
	}
	return true
}

// extractEthAddressFromScp reconstructs a 20-byte Ethereum address from the
// SCP's AddressHI (2 limbs × 16 bits) and AddressLO (8 limbs × 16 bits).
func extractEthAddressFromScp(run *wizard.ProverRuntime, scp *statesummary.HubColumnSet, index int) types.EthAddress {
	var addr [20]byte
	offset := 0
	// AddressHI: NbLimbU32=2 limbs of 16 bits each = 4 bytes
	for i := range common.NbLimbU32 {
		v := scp.AddressHI[i].GetColAssignmentAt(run, index)
		b := v.Bytes()
		// Field element is 4 bytes big-endian; the 16-bit limb is in the low 2 bytes
		addr[offset] = b[2]
		addr[offset+1] = b[3]
		offset += 2
	}
	// AddressLO: NbLimbU128=8 limbs of 16 bits each = 16 bytes
	for i := range common.NbLimbU128 {
		v := scp.AddressLO[i].GetColAssignmentAt(run, index)
		b := v.Bytes()
		addr[offset] = b[2]
		addr[offset+1] = b[3]
		offset += 2
	}
	return types.EthAddress(addr)
}

// extractStorageKeyFromScp reconstructs a 32-byte storage key from the SCP's
// KeyHI and KeyLO columns.
func extractStorageKeyFromScp(run *wizard.ProverRuntime, scp *statesummary.HubColumnSet, index int) types.FullBytes32 {
	var key [32]byte
	offset := 0
	for i := range common.NbLimbU128 {
		v := scp.KeyHI[i].GetColAssignmentAt(run, index)
		b := v.Bytes()
		key[offset] = b[2]
		key[offset+1] = b[3]
		offset += 2
	}
	for i := range common.NbLimbU128 {
		v := scp.KeyLO[i].GetColAssignmentAt(run, index)
		b := v.Bytes()
		key[offset] = b[2]
		key[offset+1] = b[3]
		offset += 2
	}
	return types.FullBytes32(key)
}

// extractBlockNumberFromScp extracts the block number from the SCP's
// BlockNumber columns.
func extractBlockNumberFromScp(run *wizard.ProverRuntime, scp *statesummary.HubColumnSet, index int) int {
	var blockBytes [8]byte
	for i := range common.NbLimbU64 {
		v := scp.BlockNumber[i].GetColAssignmentAt(run, index)
		b := v.Bytes()
		blockBytes[i*2] = b[2]
		blockBytes[i*2+1] = b[3]
	}
	return utils.ToInt(binary.BigEndian.Uint64(blockBytes[:]))
}
