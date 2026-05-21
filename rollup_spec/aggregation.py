from dataclasses import dataclass
from typing import List

from ethereum.crypto.hash import Hash32
from ethereum.state import Address

from .blob import AggregatedPublicInput, BlobProof
from .execution import hash_address_list, hash_hash_list


@dataclass
class AggregationProofPrivateInput:
    """
    Logical aggregation-proof request. The topology is flat across M blob
    proofs, as specified in Readme.md section 2.3.
    """
    blob_proofs: List[BlobProof]


def check_aggregation_proof(aggregation_input: AggregationProofPrivateInput) -> AggregatedPublicInput:
    """
    Aggregation proof: flat recursion over M blob proofs with continuity checks
    and merged L2-to-L1 root/address commitments.
    """
    if len(aggregation_input.blob_proofs) == 0:
        raise Exception("aggregation proof must consume at least one blob proof")

    for proof in aggregation_input.blob_proofs:
        verify_blob_proof(proof)

    for left, right in zip(aggregation_input.blob_proofs, aggregation_input.blob_proofs[1:]):
        assert_blob_continuity(left, right)

    first_proof = aggregation_input.blob_proofs[0]
    last_proof = aggregation_input.blob_proofs[-1]
    merged_l2_l1_roots: List[Hash32] = []
    merged_filtered_addresses: List[Address] = []

    for proof in aggregation_input.blob_proofs:
        merged_l2_l1_roots.extend(proof.l2_l1_roots)
        merged_filtered_addresses.extend(proof.filtered_addresses)

    return AggregatedPublicInput(
        end_block_number=last_proof.public_inputs.end_block_number,
        end_block_timestamp=last_proof.public_inputs.end_block_timestamp,
        l2_l1_bridge_transaction_tree=hash_hash_list(merged_l2_l1_roots),
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
        filtered_addresses_hash=hash_address_list(merged_filtered_addresses),
        parent_shnarf=first_proof.public_inputs.parent_shnarf,
        end_shnarf=last_proof.public_inputs.end_shnarf,
    )


def verify_blob_proof(proof: BlobProof) -> None:
    """
    Placeholder for recursive proof verification plus explicit checks for the
    hash preimages that the aggregation proof consumes.
    """
    if proof.public_inputs.end_block_number != proof.end_block_number:
        raise Exception("blob proof range metadata does not match public inputs")
    if hash_hash_list(proof.l2_l1_roots) != proof.public_inputs.l2_l1_bridge_transaction_tree:
        raise Exception("invalid L2L1BridgeTransactionTree preimage")
    if hash_address_list(proof.filtered_addresses) != proof.public_inputs.filtered_addresses_hash:
        raise Exception("invalid blob filteredAddressesHash preimage")


def assert_blob_continuity(left: BlobProof, right: BlobProof) -> None:
    # Block-number / block-hash continuity is implicit in the shnarf check:
    # `endShnarf = Hash(parentShnarf, lastBlockHash, blobHash)` binds the
    # last block hash. Once the next blob's `parentShnarf` matches, the
    # inner block-hash chain inside that blob anchors block numbers
    # transitively. No separate block-number assertion is needed at this
    # layer (the blob-proof PI does not expose `startBlockNumber` anyway).
    if left.public_inputs.end_shnarf != right.public_inputs.parent_shnarf:
        raise Exception("blob shnarf continuity failed")
    if left.public_inputs.end_l1_l2_bridge_rolling_hash != right.public_inputs.parent_l1_l2_bridge_rolling_hash:
        raise Exception("blob L1-to-L2 rolling-hash continuity failed")
    if (
        left.public_inputs.end_l1_l2_bridge_rolling_hash_message_number !=
        right.public_inputs.parent_l1_l2_bridge_rolling_hash_message_number
    ):
        raise Exception("blob L1-to-L2 rolling-hash-number continuity failed")
    if left.public_inputs.dynamic_chain_config_hash != right.public_inputs.dynamic_chain_config_hash:
        raise Exception("blob dynamic chain configuration continuity failed")
    if left.public_inputs.end_ftx_rolling_hash != right.public_inputs.parent_ftx_rolling_hash:
        raise Exception("blob FTX rolling-hash continuity failed")
