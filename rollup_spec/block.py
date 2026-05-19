from ethereum.forks.osaka.blocks import Block as EthereumBlock, Header
from ethereum.forks.osaka.transactions import (
    Transaction,
    recover_sender,
    validate_transaction,
    FeeMarketTransaction,
    BlobTransaction,
    SetCodeTransaction,
    LegacyTransaction,
    decode_transaction,
)
from ethereum.forks.osaka.fork import (
    BLOB_COUNT_LIMIT,
    VERSIONED_HASH_VERSION_KZG,
)
from ethereum.forks.osaka.vm.eoa_delegation import is_valid_delegation
from ethereum.forks.osaka.vm.gas import calculate_total_blob_gas
from ethereum.state import Address, Account, EMPTY_CODE_HASH
from dataclasses import dataclass
from ethereum_types.numeric import U64, Uint
from ethereum.crypto.hash import Hash32, keccak256
from ethereum_types.bytes import Bytes
from ethereum_rlp import rlp
from enum import Enum
from typing import List, Optional, TypeAlias, Tuple


BlockTransaction: TypeAlias = bytes | LegacyTransaction


@dataclass
class ChainConfig:
    l2_message_service_address: Address
    base_fee: Uint
    coinbase: Address
    chain_id: U64

    def hash(self) -> Hash32:
        """
        Hash the dynamic chain configuration exposed by the execution proof.

        This is plain concatenation of:
            uint256_be(chainID) ||
            coinbase ||
            L2MessageServiceContract ||
            uint256_be(baseFee)
        """
        return keccak256(
            int(self.chain_id).to_bytes(32, "big") +
            bytes(self.coinbase) +
            bytes(self.l2_message_service_address) +
            int(self.base_fee).to_bytes(32, "big")
        )

class ForcedTransactionAcceptance(Enum):
    """
    ForcedTransactionAcceptance is an enum indicating if a forced transaction has
    been rejected. (rejection = arbitrary decision of the prover, but has to
    report. The L1 contract will check if this is a reasonable decision). The
    enum can be either

    - ACCEPTED
    - REJECTED_BECAUSE_FROM
    - REJECTED_BECAUSE_TO

    Note: if a transaction is NOT valid because of nonce/balance, it will be
    tagged as accepted. The rejection is only for sanctioned addresses.
    """
    ACCEPTED=0
    REJECTED_BECAUSE_FROM=1
    REJECTED_BECAUSE_TO=2

@dataclass
class AccountProof:
    """
    Ethereum MPT account proof, shaped like `eth_getProof`'s account output
    (Type-1 state model). Carries the account address, the RLP-encoded
    account value (`[nonce, balance, storageRoot, codeHash]`), and the
    ordered list of RLP-encoded trie nodes from the state root down to the
    account leaf.
    """
    address: Address
    value: Bytes
    proof: List[Bytes]

    def check_shape(self) -> bool:
        """
        Schema-level shape checks. Full verification walks `proof` against
        `state_root` with the standard Ethereum MPT verifier; that verifier is
        part of the production guest and is not duplicated here.
        """
        if len(self.proof) == 0:
            return False
        return all(len(node) > 0 for node in self.proof)


@dataclass
class StorageProof:
    """
    Ethereum MPT storage proof for one slot, shaped like a single entry of
    `eth_getProof`'s `storageProof` array. The guest verifies the ordered
    RLP-encoded trie nodes against the contract's `storageRoot`.
    """
    key: Bytes
    value: Bytes
    proof: List[Bytes]

    def check_shape(self) -> bool:
        if len(self.key) != 32:
            return False
        return all(len(node) > 0 for node in self.proof)


@dataclass
class AccountWitness:
    """
    AccountWitness represents an account and its Ethereum MPT inclusion proof.
    """
    account: Account
    address: Address
    proof: AccountProof
    code: Bytes
    """
    code is the code of the contract in case the sender account of the forced
    transaction is a contract.
    """

    def check_inclusion(self, state_root: Hash32) -> bool:
        """
        check_inclusion verifies that the account witness is tied to the
        requested address and is shaped like `eth_getProof`'s output.

        If the code is provided, it will also check that its hash matches the
        codehash in the account.
        """
        if not self.proof.check_shape():
            return False
        if len(state_root) != 32:
            return False
        if self.proof.address != self.address:
            return False
        if len(self.code) > 0 and self.account.code_hash != keccak256(self.code):
            return False

        # The production guest verifies `proof` against `state_root` with the
        # standard Ethereum MPT verifier. The Python reference models the
        # witness shape and cheap consistency checks, but does not duplicate
        # that verifier.
        return True


@dataclass
class ForcedTransactionWitness:
    """
    ForcedTransactionWitness is the collection of canonical input needed to
    validate the processing of a forced transaction. It consists of the signed
    transaction bytes, an indication on whether it is a sanctioned address, and
    an optional state witness from the sender account to justify invalidity.

    We expect the following forms:

    - acceptance == ACCEPTED && state_witness == NONE -> accepted and valid
    - acceptance != ACCEPTED && state_witness == NONE -> rejected
    - acceptance == ACCEPTED && state_witness != NONE -> invalid
    """
    number: U64
    signed_tx_rlp: bytes
    acceptance: ForcedTransactionAcceptance
    state_witness: Optional[AccountWitness]
    deadline: U64
    """
    deadline is an L2 block number. The forced transaction must be included in
    a block whose number is <= deadline; if not, it may be rejected and the
    sequencer is expected to report it.

    state_witness is provided to indicate and justify that the forced
    transaction was invalid. If the transaction is valid, then the value should
    be None.
    """


@dataclass(frozen=True)
class ResolvedForcedTransaction:
    """
    Forced transaction after decoding its signed bytes and recovering sender.

    `ForcedTransactionWitness` carries only canonical inputs. This derived view
    is created once per witness and passed to downstream checks so sender
    recovery has a single owner.
    """
    number: U64
    signed_tx_rlp: bytes
    transaction: Transaction
    from_address: Address
    acceptance: ForcedTransactionAcceptance
    state_witness: Optional[AccountWitness]
    deadline: U64

    @property
    def tx_hash(self) -> Hash32:
        return keccak256(self.signed_tx_rlp)

@dataclass
class RollupBlock:
    """
    Logical execution-proof block witness.

    `block_rlp` is the canonical private input supplied by the coordinator. The
    Python reference decodes an `EthereumBlock` from these bytes internally when
    it needs an execution view.
    """
    block_rlp: bytes
    forced_transactions: List[ForcedTransactionWitness]

def block_hash(header: Header) -> Hash32:
    """
    block_hash computes the hash of a block header
    """
    return keccak256(rlp.encode(header))

def decode_block_rlp(block_rlp: bytes) -> EthereumBlock:
    """
    Decode the canonical block RLP carried by the execution-proof private input
    into the Ethereum execution-specs block view.
    """
    return rlp.decode_to(EthereumBlock, block_rlp)

def decode_signed_transaction_rlp(signed_tx_rlp: bytes) -> Transaction:
    """
    Decode the canonical signed transaction bytes used for forced transactions.

    Typed EIP-2718 transactions are encoded as type byte || RLP(payload). Legacy
    transactions are plain RLP.
    """
    if len(signed_tx_rlp) == 0:
        raise Exception("empty signed transaction RLP")
    if signed_tx_rlp[0] in (1, 2, 3, 4):
        return decode_transaction(Bytes(signed_tx_rlp))
    return rlp.decode_to(LegacyTransaction, signed_tx_rlp)


def resolve_forced_transaction(ftx: ForcedTransactionWitness, chain_id: U64) -> ResolvedForcedTransaction:
    """
    Decode a forced transaction witness and recover its sender exactly once.
    """
    transaction = decode_signed_transaction_rlp(ftx.signed_tx_rlp)
    from_address = recover_sender(chain_id, transaction)
    return ResolvedForcedTransaction(
        number=ftx.number,
        signed_tx_rlp=ftx.signed_tx_rlp,
        transaction=transaction,
        from_address=from_address,
        acceptance=ftx.acceptance,
        state_witness=ftx.state_witness,
        deadline=ftx.deadline,
    )


def _canonical_transaction_bytes(transaction: BlockTransaction) -> bytes:
    """
    Return the canonical signed bytes for a transaction from an execution-spec
    block view.

    `Block.transactions` is `Tuple[bytes | LegacyTransaction, ...]`: typed
    transactions are already EIP-2718 bytes, while legacy transactions are
    decoded objects and need regular RLP encoding.
    """
    match transaction:
        case bytes():
            return transaction
        case LegacyTransaction():
            return rlp.encode(transaction)


def parse_block_transaction_rlps(block_rlp: bytes) -> List[bytes]:
    """
    Return the signed transaction RLP byte strings decoded from `block_rlp`.

    `blockRlp` is the canonical witness bytes in the logical schema. Decoded
    `EthereumBlock` objects are only execution views, so hash checks over block
    transactions must use bytes extracted from `block_rlp` rather than
    re-encoding decoded transaction objects.
    """
    block = decode_block_rlp(block_rlp)
    return [
        _canonical_transaction_bytes(transaction)
        for transaction in block.transactions
    ]

def validate_forced_transactions(curr_rolling_hash: Hash32,
                       last_processed_ftx_number: U64,
                       chain_config: ChainConfig, base_fee: Uint,
                       parent_block_number: U64, state_root: Hash32, block: RollupBlock,
                       ) -> Tuple[List[Address], Hash32, U64]:
    """
    validate_forced_tx scans the forced transactions and checks they have been
    correctly executed/rejected.

    3 things can happen to a forced transaction:
        - It is REJECTED -> then we have to bubble-up the problematic address
        - It is included -> then we just check for inclusion in the block-list
        - Its Nonce/Balance make it an invalid transaction:
    """

    rejected_addresses = []

    for ftx in block.forced_transactions:

        # The execution guest processes FTXs in ascending L1-assigned number.
        if ftx.number != last_processed_ftx_number + 1:
            raise Exception("forced transactions must be processed in ascending sequence")

        # A FTX whose deadline is before the parent block of this range should
        # already have been handled by an earlier proof range.
        if ftx.deadline < parent_block_number:
            raise Exception("deadline exceeded")

        resolved_ftx = resolve_forced_transaction(ftx, chain_config.chain_id)
        transaction = resolved_ftx.transaction
        from_address = resolved_ftx.from_address

        #
        # The rolling hash is updated with the current forced transaction
        # regardless of whether it was actually included or not.
        curr_rolling_hash, last_processed_ftx_number = add_to_forced_tx_rolling_hash(
            curr_rolling_hash, resolved_ftx)

        if resolved_ftx.acceptance == ForcedTransactionAcceptance.REJECTED_BECAUSE_FROM:
            # The sequencer refuses the transaction on compliance grounds (sanctioned
            # sender). The from address is bubbled up; the L1 contract verifies a
            # posteriori that it appears in its reference sanction list. If it does
            # not, the finalization call is aborted — no explicit governance witness
            # is needed inside the proof.
            rejected_addresses.append(from_address)
            continue

        if resolved_ftx.acceptance == ForcedTransactionAcceptance.REJECTED_BECAUSE_TO:
            # Same as above, but the sanctioned party is the recipient. The to
            # address is bubbled up instead of the sender.
            #
            # Contract-creation transactions have no recipient (to == None).
            # REJECTED_BECAUSE_TO is meaningless for them and indicates a
            # malformed witness.
            if not isinstance(transaction.to, Address):
                raise Exception("REJECTED_BECAUSE_TO on a contract-creation transaction")
            to_address: Address = transaction.to
            rejected_addresses.append(to_address)
            continue

        if resolved_ftx.state_witness is not None:

            if resolved_ftx.state_witness.address != from_address:
                raise Exception("state_witness provided for the wrong address")

            if not resolved_ftx.state_witness.check_inclusion(state_root):
                raise Exception("provided an invalid state-inclusion witness for forced transactions")

            # This call will raise an exception if the transaction is
            # intrinsically invalid. But this does not cover situations where the
            # the transaction was invalid due to
            #
            # @alex: normally these checks could be carried by the L1 contract
            # as they don't require state access. So we could change the design.
            # I still leave it here to make it explicit that we want to check
            # that one way or another.
            try:
                validate_transaction(transaction)
            except Exception:
                continue

            if is_valid_forced_transaction(transaction, resolved_ftx.state_witness, base_fee):
                raise Exception("the forced transaction was indicated to be invalid, but was found to be valid")
            continue

        # Otherwise, the transaction should be included somewhere in the block.
        # Position is intentionally not part of the protocol rule: a sequencer
        # implementation may choose to place FTXs at the beginning of a block,
        # but future implementations can support other positions without
        # changing this proof statement.
        if resolved_ftx.acceptance != ForcedTransactionAcceptance.ACCEPTED:
            raise Exception("forced transaction has an unknown acceptance value")
        block_tx_hashes = [
            keccak256(tx_rlp)
            for tx_rlp in parse_block_transaction_rlps(block.block_rlp)
        ]
        if resolved_ftx.tx_hash not in block_tx_hashes:
            raise Exception("forced transaction was allegedly valid but not included")

    return rejected_addresses, curr_rolling_hash, last_processed_ftx_number

def is_valid_forced_transaction(
        tx: Transaction,
        sender_account_witness: AccountWitness,
        base_fee: Uint,
    ) -> bool:
    """
    This function computes the cost of a forced transaction. It is essentially
    a copy-paste from [fork.check_transaction].

    It does essentially the same checks as [fork.check_transaction] except, all
    the checks relying on the remain gas/blob-gas in the block. Also, it returns
    a boolean indicating whether or not the transaction was valid. It does not
    raise errors.
    """
    sender_account = sender_account_witness.account
    if isinstance(
        tx, (FeeMarketTransaction, BlobTransaction, SetCodeTransaction)
    ):
        # @alex:
        # What if the forced transaction does not respect that? Shouldn't that
        # be part of an ad-hoc check? It could as well be done at L1 as we have
        # static baseFee.
        #
        if tx.max_fee_per_gas < tx.max_priority_fee_per_gas:
            return False

        if tx.max_fee_per_gas < base_fee:
            return False
        max_gas_fee = tx.gas * tx.max_fee_per_gas
    else:
        if tx.gas_price < base_fee:
            return False
        max_gas_fee = tx.gas * tx.gas_price

    if isinstance(tx, BlobTransaction):
        blob_count = len(tx.blob_versioned_hashes)
        if blob_count == 0:
            return False
        if blob_count > BLOB_COUNT_LIMIT:
            return False
        for blob_versioned_hash in tx.blob_versioned_hashes:
            if blob_versioned_hash[0:1] != VERSIONED_HASH_VERSION_KZG:
                return False

        max_gas_fee += Uint(calculate_total_blob_gas(tx)) * Uint(tx.max_fee_per_blob_gas)

    if isinstance(tx, (BlobTransaction, SetCodeTransaction)):
        if not isinstance(tx.to, Address):
            return False

    if isinstance(tx, SetCodeTransaction):
        if not any(tx.authorizations):
            return False

    if sender_account.nonce != Uint(tx.nonce):
        return False

    if Uint(sender_account.balance) < max_gas_fee + Uint(tx.value):
        return False
    sender_code = sender_account_witness.code
    if sender_account.code_hash != EMPTY_CODE_HASH and not is_valid_delegation(
        sender_code
    ):
        return False

    return True

def add_to_forced_tx_rolling_hash(forced_tx_rolling_hash: Hash32,
                                  ftx: ResolvedForcedTransaction) -> Tuple[Hash32, U64]:
    """
    add_to_forced_tx_rolling_hash updates the forced transaction rolling hash
    with an already-resolved FTX. Sender recovery is intentionally outside this
    helper so the dependency on signed transaction bytes is explicit and only
    performed once.
    """
    return keccak256(
        forced_tx_rolling_hash +
        ftx.tx_hash +
        int(ftx.deadline).to_bytes(32, "big") +
        bytes(ftx.from_address)
    ), ftx.number
