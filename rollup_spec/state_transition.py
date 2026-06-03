from dataclasses import dataclass, field
from typing import List, Optional, Sequence

from ethereum.crypto.hash import Hash32
from ethereum.state import Account, Address
from ethereum_types.bytes import Bytes32

from .block import StatelessInput
from .fork import Log


@dataclass
class ExecutionWitness:
    """
    Decoded execution witness for one payload (part of the parsed
    `StatelessInput`).

    `headers` are RLP ancestor headers ordered by block number, ending at the
    payload parent. `keys` is not in the SSZ wire format (JSON/debug only),
    defaulted empty.

    The pool must cover every account/slot read — both what block execution
    touches (served by the underlying engine) and the Linea-extra reads
    (L1->L2 rolling-hash slots at the boundary roots, FTX-sender accounts for
    §6.5 'Invalid').
    """
    state: List[bytes]
    codes: List[bytes]
    headers: List[bytes]
    keys: List[bytes] = field(default_factory=list)


@dataclass
class StatelessExecutionResult:
    """
    Result of executing one stateless input (see `execute_stateless_input`).

    Mirrors the proof output an underlying engine exposes; the Linea layer
    consumes these without re-running execution.
    """
    pre_state_root: Hash32     # parent (pre-execution) state root of the block
    post_state_root: Hash32    # post-execution state root (== executionPayload.stateRoot)
    block_logs: List[Log]      # ordered logs emitted by the block, for L2->L1 message extraction


def execute_stateless_input(stateless_input: StatelessInput) -> StatelessExecutionResult:
    """
    Validate and execute one block from its parsed `StatelessInput`.

    This is the spec's boundary to the underlying stateless block-execution
    engine (for example a Zesu-style guest, imported as a dependency or
    reimplemented). The spec does not model its internals; an underlying
    implementation is expected to perform, from the parsed `StatelessInput`
    alone:

      - witness header-chain validation and parent-header anchoring (the last
        witness header hash must equal `executionPayload.parentHash`);
      - full Engine-API payload validation (block hash, versioned hashes,
        `parentBeaconBlockRoot`, base-fee correctness, gas / blob-gas, timestamp
        vs parent, state root, receipts root, logs bloom, EIP-7928 block access
        list);
      - the EVM state transition itself.

    It returns the boundary state roots and the block logs the Linea layer builds
    on. It does NOT enforce Linea policy (e.g. empty `executionRequests`) or
    conflation-level invariants — those live in `run_l2_execution_guest`.
    """
    raise NotImplementedError(
        "stateless block execution is provided by the underlying engine "
        "(e.g. Zesu); the spec models only the Linea-specific logic on top"
    )


@dataclass
class L2State:
    """
    Read-only, witness-backed view of L2 state at a state root.

    Exposes the EVM state-database reads the Linea layer needs on top of
    execution: the L1->L2 bridge rolling hash (`storage`) and forced-tx sender
    accounts (`account`). Reads are served by the underlying engine's witness
    database (inclusion proofs against `state_root` over the witness node pool);
    the spec does not model the trie walk.
    """
    state_root: Hash32
    witnesses: Sequence["ExecutionWitness"]

    def account(self, address: Address) -> Optional[Account]:
        """Account at `address`, or None if absent (proven by the witness)."""
        raise NotImplementedError(
            "served by the underlying engine's witness-backed state database"
        )

    def storage(self, address: Address, slot: Bytes32) -> Bytes32:
        """Storage value at (`address`, `slot`), or zero if unset."""
        raise NotImplementedError(
            "served by the underlying engine's witness-backed state database"
        )
