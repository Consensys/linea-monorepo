from dataclasses import dataclass, field
from ethereum_types.numeric import U64
from ethereum.crypto.hash import Hash32, keccak256
from ethereum_types.bytes import Bytes32, Bytes48
from ethereum.state import Address
from typing import List, Optional, Sequence, Tuple

from ethereum.crypto.kzg import (
    KZGCommitment,
    kzg_commitment_to_versioned_hash,
    verify_kzg_proof,
)
from .execution import (
    ExecutionProof,
    ExecutionProofPublicInput,
    hash_address_list,
    hash_hash_list,
)

L2_L1_TREE_DEPTH = 5
ZERO_HASH32 = Hash32(b"\x00" * 32)

@dataclass
class TruncatedEthereumBlock:
    """
    TruncatedEthereumBlock is the truncated content of an Ethereum block as it
    appears in the DA blob payload (§3.2). The blob-proof guest cross-checks
    `froms` against each execution proof's `txFromsHash` (§2.2 step 5) and
    `block_hash` against the execution proofs' `endBlockHash` at boundaries
    (§2.2 step 7); no separate full-block comparison is required.
    """
    timestamp: U64
    block_hash: Hash32
    prev_randao: Bytes32
    transactions: List[bytes]
    froms: List[Address]

@dataclass
class ShnarfWitness:
    """
    ShnarfWitness is the preimage of a shnarf (§3.1):
    `Hash(parentShnarf, lastBlockHash, blobHash)`.
    """
    parent_shnarf: Hash32
    last_block_hash: Hash32
    blob_hash: Hash32

    def hash(self) -> Hash32:
        return keccak256(self.parent_shnarf + self.last_block_hash + self.blob_hash)

@dataclass
class BlobWitness:
    """
    Per-blob witness for the blob proof.

    Fields:
      - `blob_bytes`: the EIP-4844 compressed payload as submitted on L1.
      - `block_number_range`: `(startBlockNumber, endBlockNumber)` of the
        L2 blocks contained in this blob.
      - `block_rlps`: the canonical full block RLPs (same shape the
        execution proof receives — header + tx list [+ withdrawals], EIP-2718
        typed transactions in full signed form), one per block in
        `block_number_range`. Truncation per §3.2 happens *inside* the
        guest from these full RLPs; there is no separately-witnessed
        truncated form.
      - `blob_kzg_commitment` / `blob_kzg_proof`: KZG witness that binds
        `blob_bytes` to the on-chain versioned `blobHash`.
      - `expected_blob_hash`: optional cross-check against the L1-anchored
        versioned hash.

    The proof statement (§2.2 step 2) is compression of the RLP-encoded
    truncated form:
      `lz4_compress(rlp_encode_truncated_blocks(truncate(block_rlps))) == blob_bytes`.
    The truncated blocks themselves are an internal-only intermediate;
    their downstream consumers (block-hash boundary checks in §2.2 step 6,
    sender-list cross-checks in step 4) take the computed truncated form,
    not a witnessed copy.
    """
    blob_bytes: bytes
    block_number_range: tuple[int, int]
    block_rlps: List[bytes]
    blob_kzg_commitment: KZGCommitment
    blob_kzg_proof: Bytes48
    expected_blob_hash: Optional[Hash32] = None

    def blob_hash(self) -> Hash32:
        """
        blob_hash returns the versioned hash of the blob KZG commitment.
        """
        return Hash32(kzg_commitment_to_versioned_hash(self.blob_kzg_commitment))

    def is_authenticated_blob_bytes(self) -> Tuple[bool, Hash32] :
        """
        Check the KZG proof binds `blob_bytes` to the declared `blob_hash`.

        Returns (is_authenticated, blob_hash).
        """
        blob_hash = self.blob_hash()
        if self.expected_blob_hash is not None and self.expected_blob_hash != blob_hash:
            return False, blob_hash
        blob_kzg_x = Bytes32(keccak256(self.blob_bytes + blob_hash))
        blob_poly = parse_as_bls12_381_fr_vec(self.blob_bytes)
        blob_kzg_y = eval_lagrange_bls12_381(blob_poly, blob_kzg_x)
        return verify_kzg_proof(self.blob_kzg_commitment, blob_kzg_x, blob_kzg_y, self.blob_kzg_proof), blob_hash

    def verify_compression(self) -> List["TruncatedEthereumBlock"]:
        """
        Apply the canonical DA truncation (§3.2) to each `block_rlps[i]`
        internally, RLP-encode the truncated form, LZ4-compress it, and
        assert the result equals `blob_bytes` (§2.2 step 2). Returns the
        computed truncated blocks for downstream steps (block-hash
        boundary alignment, sender-list cross-check).

        The compression-equality check binds the full-RLP witness to the
        KZG-authenticated blob bytes — no decompression of `blob_bytes`
        happens inside the proof.
        """
        truncated = [truncate_block_rlp(rlp_bytes) for rlp_bytes in self.block_rlps]
        serialized = rlp_encode_truncated_blocks(truncated)
        if compress_lz4(serialized) != self.blob_bytes:
            raise Exception("blob compression mismatch: truncated form does not compress to blobContent")
        return truncated


@dataclass
class AggregatedPublicInput:
    """
    The 14-field blob-proof / aggregation-proof public input tuple from
    Readme.md section 2.4.
    """
    end_block_number: U64
    end_block_timestamp: U64
    l2_l1_bridge_transaction_tree: Hash32
    parent_l1_l2_bridge_rolling_hash: Hash32
    parent_l1_l2_bridge_rolling_hash_message_number: U64
    end_l1_l2_bridge_rolling_hash: Hash32
    end_l1_l2_bridge_rolling_hash_message_number: U64
    dynamic_chain_config_hash: Hash32
    parent_ftx_rolling_hash: Hash32
    end_ftx_rolling_hash: Hash32
    last_processed_ftx_number: U64
    filtered_addresses_hash: Hash32
    parent_shnarf: Hash32
    end_shnarf: Hash32


@dataclass
class BlobProofPrivateInput:
    """
    Logical blob-proof request. One blob proof folds K >= 1 DA blobs and N >= 1
    execution proofs tiling the combined block range.
    """
    parent_shnarf: Hash32
    blobs: List[BlobWitness]
    execution_proofs: List[ExecutionProof]


@dataclass
class BlobProof:
    """
    Reference wrapper for a blob proof plus the root/address preimages consumed
    by the aggregation proof. `proof` stands in for recursive STARK bytes.
    """
    public_inputs: AggregatedPublicInput
    start_block_number: U64
    end_block_number: U64
    proof: bytes = b""
    l2_l1_roots: List[Hash32] = field(default_factory=list)
    filtered_addresses: List[Address] = field(default_factory=list)


def check_blob_proof(blob_input: BlobProofPrivateInput) -> BlobProof:
    """
    Blob proof: verifies KZG/decompression for K blobs, recursively verifies
    the N execution proofs, checks continuity, and emits the 13-field PI.
    """
    if len(blob_input.blobs) == 0:
        raise Exception("blob proof must cover at least one blob")
    if len(blob_input.execution_proofs) == 0:
        raise Exception("blob proof must consume at least one execution proof")

    current_shnarf = blob_input.parent_shnarf
    truncated_blocks = []
    expected_blob_start: Optional[int] = None

    for blob in blob_input.blobs:
        # KZG binding: blob_bytes matches the declared blobHash.
        blob_auth, blob_hash = blob.is_authenticated_blob_bytes()
        if not blob_auth:
            raise Exception("invalid KZG proof")

        # Compression equality (§2.2 step 2): the witnessed truncated blocks
        # serialize+compress to blob_bytes. Downstream cross-checks anchor
        # their content to the execution proofs.
        blob_blocks = blob.verify_compression()
        start_block_number, end_block_number = blob.block_number_range
        expected_block_count = end_block_number + 1 - start_block_number
        if len(blob_blocks) != expected_block_count:
            raise Exception("blob block range is inconsistent with witnessed truncated blocks")
        if len(blob_blocks) == 0:
            raise Exception("blob proof cannot include an empty blob block range")
        if expected_blob_start is not None and start_block_number != expected_blob_start:
            raise Exception("blob ranges must be contiguous")

        truncated_blocks.extend(blob_blocks)
        current_shnarf = ShnarfWitness(
            current_shnarf,
            blob_blocks[-1].block_hash,
            blob_hash,
        ).hash()
        expected_blob_start = end_block_number + 1

    blob_start_block_number = blob_input.blobs[0].block_number_range[0]
    blob_end_block_number = blob_input.blobs[-1].block_number_range[1]
    verify_execution_proof_tiling(blob_input.execution_proofs, blob_start_block_number, blob_end_block_number)

    concatenated_froms: List[Address] = []
    concatenated_l2_l1_messages: List[Hash32] = []
    concatenated_filtered_addresses: List[Address] = []
    truncated_froms: List[Address] = []
    truncated_block_hashes = [block.block_hash for block in truncated_blocks]

    for proof in blob_input.execution_proofs:
        verify_execution_proof(proof)
        concatenated_froms.extend(proof.tx_froms)
        concatenated_l2_l1_messages.extend(proof.l2_l1_messages)
        concatenated_filtered_addresses.extend(proof.filtered_addresses)

    for block in truncated_blocks:
        truncated_froms.extend(block.froms)

    if concatenated_froms != truncated_froms:
        raise Exception("execution-proof txFroms do not match blob blockData.froms")

    first_proof = blob_input.execution_proofs[0]
    last_proof = blob_input.execution_proofs[-1]
    for proof in blob_input.execution_proofs:
        boundary_index = int(proof.end_block_number) - blob_start_block_number
        if boundary_index < 0 or boundary_index >= len(truncated_block_hashes):
            raise Exception("execution proof boundary falls outside the blob block range")
        if proof.public_inputs.end_block_hash != truncated_block_hashes[boundary_index]:
            raise Exception("execution proof end block hash does not match blob data at its boundary")

    for left, right in zip(blob_input.execution_proofs, blob_input.execution_proofs[1:]):
        assert_execution_continuity(left.public_inputs, right.public_inputs)

    l2_l1_roots, l2_l1_bridge_transaction_tree = build_l2_messages_tree(
        concatenated_l2_l1_messages,
    )
    public_inputs = AggregatedPublicInput(
        end_block_number=last_proof.public_inputs.end_block_number,
        end_block_timestamp=last_proof.public_inputs.end_block_timestamp,
        l2_l1_bridge_transaction_tree=l2_l1_bridge_transaction_tree,
        parent_l1_l2_bridge_rolling_hash=first_proof.public_inputs.parent_l1_l2_bridge_rolling_hash,
        parent_l1_l2_bridge_rolling_hash_message_number=(
            first_proof.public_inputs.parent_l1_l2_bridge_rolling_hash_message_number
        ),
        end_l1_l2_bridge_rolling_hash=last_proof.public_inputs.end_l1_l2_bridge_rolling_hash,
        end_l1_l2_bridge_rolling_hash_message_number=(
            last_proof.public_inputs.end_l1_l2_bridge_rolling_hash_message_number
        ),
        dynamic_chain_config_hash=first_proof.public_inputs.dynamic_chain_config_hash,
        parent_ftx_rolling_hash=first_proof.public_inputs.parent_ftx_rolling_hash,
        end_ftx_rolling_hash=last_proof.public_inputs.end_ftx_rolling_hash,
        last_processed_ftx_number=last_proof.public_inputs.last_processed_ftx_number,
        filtered_addresses_hash=hash_address_list(concatenated_filtered_addresses),
        parent_shnarf=blob_input.parent_shnarf,
        end_shnarf=current_shnarf,
    )

    return BlobProof(
        public_inputs=public_inputs,
        start_block_number=blob_start_block_number,
        end_block_number=blob_end_block_number,
        l2_l1_roots=l2_l1_roots,
        filtered_addresses=concatenated_filtered_addresses,
    )


def verify_execution_proof(proof: ExecutionProof) -> None:
    """
    Placeholder for recursive proof verification plus explicit checks for the
    hash preimages that the blob proof consumes.
    """
    if proof.public_inputs.end_block_number != proof.end_block_number:
        raise Exception("execution proof range metadata does not match public inputs")
    if hash_hash_list(proof.l2_l1_messages) != proof.public_inputs.l2_l1_messages_hash:
        raise Exception("invalid L2-to-L1 message-list preimage")
    if hash_address_list(proof.tx_froms) != proof.public_inputs.tx_froms_hash:
        raise Exception("invalid txFromsHash preimage")
    if hash_address_list(proof.filtered_addresses) != proof.public_inputs.filtered_addresses_hash:
        raise Exception("invalid execution filteredAddressesHash preimage")


def verify_execution_proof_tiling(
    execution_proofs: Sequence[ExecutionProof],
    start_block_number: int,
    end_block_number: int,
) -> None:
    expected_start = start_block_number
    for proof in execution_proofs:
        if proof.start_block_number != expected_start:
            raise Exception("execution proofs do not tile the blob block range")
        expected_start = proof.end_block_number + 1
    if expected_start != end_block_number + 1:
        raise Exception("execution proofs do not cover the full blob block range")


def assert_execution_continuity(left: ExecutionProofPublicInput, right: ExecutionProofPublicInput) -> None:
    if left.end_block_hash != right.parent_block_hash:
        raise Exception("execution block-hash continuity failed")
    if left.end_l1_l2_bridge_rolling_hash != right.parent_l1_l2_bridge_rolling_hash:
        raise Exception("execution L1-to-L2 rolling-hash continuity failed")
    if left.end_l1_l2_bridge_rolling_hash_message_number != right.parent_l1_l2_bridge_rolling_hash_message_number:
        raise Exception("execution L1-to-L2 rolling-hash-number continuity failed")
    if left.dynamic_chain_config_hash != right.dynamic_chain_config_hash:
        raise Exception("execution dynamic chain configuration continuity failed")
    if left.end_ftx_rolling_hash != right.parent_ftx_rolling_hash:
        raise Exception("execution FTX rolling-hash continuity failed")


def build_l2_messages_tree(msgs: Sequence[Hash32]) -> Tuple[List[Hash32], Hash32]:
    """
    Build L2-to-L1 message trees exactly as specified:
    - Pad the ordered message-hash list with zero Hash32 values until its
      length is a multiple of 32.
    - Split the padded list into consecutive 32-leaf chunks.
    - Merkle-hash each chunk as a complete depth-5 binary tree with keccak.
    - Flat-hash the ordered roots with keccak256(root_1 || ... || root_n).

    The returned root list is the private preimage used by aggregation and L1
    calldata; the returned hash is the public `L2L1BridgeTransactionTree`.
    """
    roots = build_l2_message_roots(msgs)
    return roots, hash_hash_list(roots)


def build_l2_message_roots(msgs: Sequence[Hash32]) -> List[Hash32]:
    leaves_per_tree = 1 << L2_L1_TREE_DEPTH
    padded_msgs = list(msgs)
    padding = (-len(padded_msgs)) % leaves_per_tree
    padded_msgs.extend(ZERO_HASH32 for _ in range(padding))

    roots = []
    for start in range(0, len(padded_msgs), leaves_per_tree):
        roots.append(
            merkle_root_fixed_depth(
                padded_msgs[start:start + leaves_per_tree],
                L2_L1_TREE_DEPTH,
            ),
        )
    return roots


def merkle_root_fixed_depth(leaves: Sequence[Hash32], depth: int) -> Hash32:
    leaf_count = 1 << depth
    if len(leaves) > leaf_count:
        raise Exception("too many leaves for fixed-depth tree")

    layer = [bytes(leaf) for leaf in leaves]
    layer.extend(bytes(ZERO_HASH32) for _ in range(leaf_count - len(layer)))
    while len(layer) > 1:
        layer = [
            keccak256(layer[i] + layer[i + 1])
            for i in range(0, len(layer), 2)
        ]
    return Hash32(layer[0])


def truncate_block_rlp(block_rlp: bytes) -> "TruncatedEthereumBlock":
    """
    Decode the canonical full block RLP (header + tx list [+ withdrawals]
    with EIP-2718 typed transactions in full signed form) and apply the
    §3.2 DA truncation rule, returning a `TruncatedEthereumBlock`:

      - `timestamp`, `prev_randao`, `block_hash` extracted from the header
        (block_hash = keccak256(header_rlp));
      - per-transaction signature-stripped bytes (drop `(v, r, s)`) with
        the recovered `from` recorded explicitly;
      - `froms` parallel-list of sender addresses recovered via
        `recover_sender(signedTxRlp, chainID)`.

    The Python reference defers the actual decode/truncate to the
    production guest.
    """
    raise NotImplementedError("truncate_block_rlp")


def rlp_encode_truncated_blocks(blocks: Sequence["TruncatedEthereumBlock"]) -> bytes:
    """
    Canonical RLP serialization of the per-blob truncated-block list
    (§3.2). Both the sequencer (when producing the blob) and the
    blob-proof guest (when verifying compression equality) must use the
    same routine. The resulting byte string is what the LZ4 compressor
    operates on.
    """
    raise NotImplementedError("rlp_encode_truncated_blocks")


def compress_lz4(data: bytes) -> bytes:
    """
    LZ4-compress the canonical RLP-encoded truncated-block payload. The
    blob-proof guest asserts that this output equals the KZG-bound
    `blob_bytes` (§2.2 step 2).
    """
    raise NotImplementedError("compress_lz4")

def parse_as_bls12_381_fr_vec(data: bytes) -> List[Bytes32]:
    """
    parse_as_bls12_381_fr_vec parses the input sequence of bytes into a sequence
    of BLS12-381 field elements. The function must return an error if the input
    cannot be parsed into an array of BLS12-381 field elements.

    The way the function works is that it expects data to be formed as a
    sequence of BLS12-381 scalar field elements all stored over 32 bytes. So
    the function does not really do anything aside slicing the input data and
    checking the field elements are well formed (don't overflow the modulus)
    """
    raise NotImplementedError("parse_as_bls12_381_fr_vec")

def eval_lagrange_bls12_381(poly: List[Bytes32], x: Bytes32) -> Bytes32:
    """
    eval_lagrange_bls12_381 evaluates a polynomial in Lagrange form in the
    BLS12-381 field
    """
    raise NotImplementedError("eval_lagrange_bls12_381")
