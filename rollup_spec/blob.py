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
class RollupDataWitness:
    """
    RollupSubmittedData represents the witness of a data-submission; e.g. the
    inputs of the prover relative to a particular blob submission.
    """
    blob_bytes: bytes
    block_number_range: tuple[int, int]
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
        is_authenticated_blob_bytes checks the KZG proofs and returns true if
        it successfully validates that the blob_hash relates to the blob_data.

        The function also returns the blob_hash.
        """
        blob_hash = self.blob_hash()
        if self.expected_blob_hash is not None and self.expected_blob_hash != blob_hash:
            return False, blob_hash
        blob_kzg_x = Bytes32(keccak256(self.blob_bytes + blob_hash))
        blob_poly = parse_as_bls12_381_fr_vec(self.blob_bytes)
        blob_kzg_y = eval_lagrange_bls12_381(blob_poly, blob_kzg_x)
        return verify_kzg_proof(self.blob_kzg_commitment, blob_kzg_x, blob_kzg_y, self.blob_kzg_proof), blob_hash
    
    def parse_block_data(self) -> List[TruncatedEthereumBlock]:
        """
        This parts parses the blob-data and checks the relevant execution
        proof for each transaction.
        """
        blob_data_uncompressed = decompress_lz4(self.blob_bytes)
        return parse_public_da_block_data(blob_data_uncompressed)


@dataclass
class AggregatedPublicInput:
    """
    The 13-field blob-proof / aggregation-proof public input tuple from
    Readme.md section 2.4.
    """
    end_block_number: U64
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
    blobs: List[RollupDataWitness]
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
        blob_auth, blob_hash = blob.is_authenticated_blob_bytes()
        if not blob_auth:
            raise Exception("invalid KZG proof")

        blob_blocks = blob.parse_block_data()
        start_block_number, end_block_number = blob.block_number_range
        expected_block_count = end_block_number + 1 - start_block_number
        if len(blob_blocks) != expected_block_count:
            raise Exception("blob block range is inconsistent with decompressed block data")
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


def decompress_lz4(data: bytes) -> bytes:
    """
    decompress_lz4 decompressed with LZ4 decompresses and returns the bytes
    """
    raise NotImplementedError("decompress_lz4")

def parse_public_da_block_data(data: bytes) -> List[TruncatedEthereumBlock]:
    """
    parse_public_da_block_data parses the blockdata coming from a DA blob. The
    encoding of the block data relies on the RLP encoding of the block.
    """
    raise NotImplementedError("parse_public_da_block_data")

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
