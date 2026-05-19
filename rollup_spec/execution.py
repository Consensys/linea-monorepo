from dataclasses import dataclass, field
from typing import List, Sequence

from ethereum.crypto.hash import Hash32, keccak256
from ethereum.forks.osaka.blocks import Block as EthereumBlock, Header
from ethereum.forks.osaka.transactions import recover_sender
from ethereum.state import Address
from ethereum_types.bytes import Bytes32
from ethereum_types.numeric import U64

from .block import (
    AccountProof,
    ChainConfig,
    RollupBlock,
    StorageProof,
    block_hash,
    decode_block_rlp,
    decode_signed_transaction_rlp,
    parse_block_transaction_rlps,
    validate_forced_transactions,
)
from .state_transition import ExecutionWitness, state_transition_modified

BRIDGE_L2L1_MESSAGE_SENT_TOPIC_0 = Hash32(
    bytes.fromhex("e856c2b8bd4eb0027ce32eeaf595c21b0b6b4644b326e5b7bd80a1cf8db72e6c"),
)


@dataclass
class ExecutionProofPublicInput:
    """
    The 14-field execution-proof public input tuple from Readme.md section 2.1.
    """
    parent_block_hash: Hash32
    end_block_hash: Hash32
    end_block_number: U64
    l2_l1_messages_hash: Hash32
    parent_l1_l2_bridge_rolling_hash: Hash32
    parent_l1_l2_bridge_rolling_hash_message_number: U64
    end_l1_l2_bridge_rolling_hash: Hash32
    end_l1_l2_bridge_rolling_hash_message_number: U64
    dynamic_chain_config_hash: Hash32
    parent_ftx_rolling_hash: Hash32
    end_ftx_rolling_hash: Hash32
    last_processed_ftx_number: U64
    filtered_addresses_hash: Hash32
    tx_froms_hash: Hash32


@dataclass
class L1L2BridgeStorageWitness:
    """
    Storage witness for the L2MessageService L1->L2 rolling-hash slots at one
    state root.

    The account proof authenticates the L2MessageService account against the
    state root. The two storage proofs authenticate the rolling hash and its
    message number against the account's `storageRoot`. All three are
    Ethereum MPT proofs (Type-1 state, `eth_getProof` shape).
    """
    rolling_hash: Hash32
    rolling_hash_message_number: U64
    account_proof: AccountProof
    rolling_hash_proof: StorageProof
    rolling_hash_message_number_proof: StorageProof

    def check_inclusion(self, state_root: Hash32, l2_message_service_address: Address) -> bool:
        if len(state_root) != 32:
            return False
        if not self.account_proof.check_shape():
            return False
        if self.account_proof.address != l2_message_service_address:
            return False
        return (
            self.rolling_hash_proof.check_shape() and
            self.rolling_hash_message_number_proof.check_shape()
        )


@dataclass
class L1L2BridgeStateTransitionWitness:
    """
    Old and new storage witnesses for the L1->L2 rolling-hash accumulator.
    """
    parent: L1L2BridgeStorageWitness
    end: L1L2BridgeStorageWitness


@dataclass
class ExecutionProofPrivateInput:
    """
    Logical execution-proof request. `blocks` carries canonical block RLP bytes
    plus per-block FTX metadata. `execution_witnesses` carries the corresponding
    Besu `debug_executionWitness` payload for each block. The first witness must
    contain the parent header whose hash is `blocks[0].header.parent_hash`.
    """
    l1_l2_bridge_state: L1L2BridgeStateTransitionWitness
    parent_ftx_rolling_hash: Hash32
    parent_last_processed_ftx_number: U64
    blocks: List[RollupBlock]
    execution_witnesses: List[ExecutionWitness]
    chain_config: ChainConfig


@dataclass
class ExecutionProof:
    """
    Reference wrapper for an execution proof plus the hash preimages consumed by
    the blob guest. `proof` stands in for the execution STARK bytes that the
    blob guest recursively verifies.
    """
    public_inputs: ExecutionProofPublicInput
    start_block_number: U64
    end_block_number: U64
    proof: bytes = b""
    l2_l1_messages: List[Hash32] = field(default_factory=list)
    tx_froms: List[Address] = field(default_factory=list)
    filtered_addresses: List[Address] = field(default_factory=list)


def check_execution_proof(execution_input: ExecutionProofPrivateInput) -> ExecutionProof:
    """
    Execution proof: validates the EVM state transition for a contiguous block
    range and emits the 14-field execution PI.
    """
    if len(execution_input.blocks) == 0:
        raise Exception("execution proof must cover at least one block")
    decoded_blocks = [
        decode_block_rlp(rollup_block.block_rlp)
        for rollup_block in execution_input.blocks
    ]
    if len(execution_input.execution_witnesses) != len(execution_input.blocks):
        raise Exception("execution witness count must match block count")

    parent_header = parent_header_from_execution_witness(
        decoded_blocks[0],
        execution_input.execution_witnesses[0],
    )
    parent_block_hash = block_hash(parent_header)
    start_block_number = parent_header.number + 1
    base_fee = execution_input.chain_config.base_fee
    current_parent_header = parent_header
    parent_l1_l2_bridge_state = execution_input.l1_l2_bridge_state.parent
    if not parent_l1_l2_bridge_state.check_inclusion(
        parent_header.state_root,
        execution_input.chain_config.l2_message_service_address,
    ):
        raise Exception("invalid parent L1-to-L2 rolling-hash storage proof")
    current_ftx_rolling_hash = execution_input.parent_ftx_rolling_hash
    current_last_processed_ftx_number = execution_input.parent_last_processed_ftx_number
    l2_l1_message_hashes: List[Hash32] = []
    tx_froms: List[Address] = []
    filtered_addresses: List[Address] = []

    if decoded_blocks[0].header.number != start_block_number:
        raise Exception("execution proof block range does not start after parent block")

    for rollup_block, execution_witness, block in zip(
        execution_input.blocks,
        execution_input.execution_witnesses,
        decoded_blocks,
    ):
        if block.header.base_fee_per_gas != base_fee:
            raise Exception("baseFee must be constant across an execution proof")
        if block.header.coinbase != execution_input.chain_config.coinbase:
            raise Exception("block coinbase does not match chain configuration")

        for tx_rlp in parse_block_transaction_rlps(rollup_block.block_rlp):
            tx_froms.append(
                recover_sender(
                    execution_input.chain_config.chain_id,
                    decode_signed_transaction_rlp(tx_rlp),
                ),
            )

        block_filtered_addresses, current_ftx_rolling_hash, current_last_processed_ftx_number = (
            validate_forced_transactions(
                current_ftx_rolling_hash,
                current_last_processed_ftx_number,
                execution_input.chain_config,
                base_fee,
                current_parent_header.number,
                current_parent_header.state_root,
                rollup_block,
            )
        )
        filtered_addresses.extend(block_filtered_addresses)

        block_output = state_transition_modified(
            execution_input.chain_config.chain_id,
            current_parent_header,
            execution_witness,
            block,
            rollup_block.block_rlp,
        )
        current_parent_header = block.header
        current_block = block

        for log in block_output.block_logs:
            if log.address != execution_input.chain_config.l2_message_service_address:
                continue
            if log.topics[0] == BRIDGE_L2L1_MESSAGE_SENT_TOPIC_0:
                l2_l1_message_hashes.append(Hash32(log.topics[3]))

    end_l1_l2_bridge_state = execution_input.l1_l2_bridge_state.end
    if not end_l1_l2_bridge_state.check_inclusion(
        current_block.header.state_root,
        execution_input.chain_config.l2_message_service_address,
    ):
        raise Exception("invalid end L1-to-L2 rolling-hash storage proof")
    if end_l1_l2_bridge_state.rolling_hash_message_number < (
        parent_l1_l2_bridge_state.rolling_hash_message_number
    ):
        raise Exception("L1-to-L2 rolling-hash message number cannot decrease")

    public_inputs = ExecutionProofPublicInput(
        parent_block_hash=parent_block_hash,
        end_block_hash=block_hash(current_block.header),
        end_block_number=current_block.header.number,
        l2_l1_messages_hash=hash_hash_list(l2_l1_message_hashes),
        parent_l1_l2_bridge_rolling_hash=parent_l1_l2_bridge_state.rolling_hash,
        parent_l1_l2_bridge_rolling_hash_message_number=(
            parent_l1_l2_bridge_state.rolling_hash_message_number
        ),
        end_l1_l2_bridge_rolling_hash=end_l1_l2_bridge_state.rolling_hash,
        end_l1_l2_bridge_rolling_hash_message_number=(
            end_l1_l2_bridge_state.rolling_hash_message_number
        ),
        dynamic_chain_config_hash=execution_input.chain_config.hash(),
        parent_ftx_rolling_hash=execution_input.parent_ftx_rolling_hash,
        end_ftx_rolling_hash=current_ftx_rolling_hash,
        last_processed_ftx_number=current_last_processed_ftx_number,
        filtered_addresses_hash=hash_address_list(filtered_addresses),
        tx_froms_hash=hash_address_list(tx_froms),
    )

    return ExecutionProof(
        public_inputs=public_inputs,
        start_block_number=start_block_number,
        end_block_number=current_block.header.number,
        l2_l1_messages=l2_l1_message_hashes,
        tx_froms=tx_froms,
        filtered_addresses=filtered_addresses,
    )


def parent_header_from_execution_witness(block: EthereumBlock, execution_witness: ExecutionWitness) -> Header:
    """
    Extract the parent header for the first execution block from the witness.

    The protocol witness supplies recent headers through
    `debug_executionWitness.headers`; the parent header is the one whose block
    hash equals the first block's `parent_hash`.
    """
    parent_headers = [
        header
        for header in execution_witness.headers
        if block_hash(header) == block.header.parent_hash
    ]
    if len(parent_headers) != 1:
        raise Exception("execution witness must contain exactly one parent header for the first block")
    return parent_headers[0]


def hash_hash_list(values: Sequence[Hash32]) -> Hash32:
    return keccak256(b"".join(bytes(value) for value in values))


def hash_address_list(values: Sequence[Address]) -> Hash32:
    return keccak256(b"".join(bytes(value) for value in values))


def uint256_topic_to_int(topic: Bytes32) -> int:
    return int.from_bytes(bytes(topic), "big")
