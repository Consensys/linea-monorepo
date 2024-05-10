import {
  Overrides,
  ContractTransactionResponse,
  JsonRpcProvider,
  TransactionReceipt,
  TransactionRequest,
  TransactionResponse,
  Signer,
} from "ethers";
import { LineaRollup, LineaRollup__factory } from "../typechain";
import { EIP1559GasProvider } from "../EIP1559GasProvider";
import { GasEstimationError } from "../../../core/errors/GasFeeErrors";
import { MessageProps, MessageWithProofProps } from "../../../core/entities/Message";
import { OnChainMessageStatus } from "../../../core/enums/MessageEnums";
import {
  MESSAGE_SENT_EVENT_SIGNATURE,
  L2_MERKLE_TREE_ADDED_EVENT_SIGNATURE,
  L2_MESSAGING_BLOCK_ANCHORED_EVENT_SIGNATURE,
  MESSAGE_UNKNOWN_STATUS,
  MESSAGE_CLAIMED_STATUS,
  DEFAULT_L2_MESSAGE_TREE_DEPTH,
  ZERO_ADDRESS,
  ZERO_HASH,
} from "../../../core/constants";
import { SparseMerkleTreeFactory } from "../../../services/merkleTree/MerkleTreeFactory";
import { Proof, FinalizationMessagingInfo } from "../../../services/merkleTree/types";
import { ILineaRollupClient } from "../../../core/clients/blockchain/ethereum/ILineaRollupClient";
import { ILineaRollupLogClient } from "../../../core/clients/blockchain/ethereum/ILineaRollupLogClient";
import { MessageSent } from "../../../core/types/Events";
import { IL2MessageServiceLogClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";
import { BaseError } from "../../../core/errors/Base";
import { isNull } from "../../../core/utils/shared";
import { SDKMode } from "../../../sdk/config";
import { formatMessageStatus } from "../../../core/utils/message";

export class LineaRollupClient
  extends EIP1559GasProvider
  implements ILineaRollupClient<Overrides, ContractTransactionResponse>
{
  private readonly contract: LineaRollup;
  private readonly l2MessageTreeDepth: number;

  /**
   * Initializes a new instance of the `LineaRollupClient`.
   *
   * @param {JsonRpcProvider} provider - The JSON RPC provider for interacting with the Ethereum network.
   * @param {string} contractAddress - The address of the Linea Rollup contract.
   * @param {ILineaRollupLogClient} lineaRollupLogClient - An instance of a class implementing the `ILineaRollupLogClient` interface for fetching events from the blockchain.
   * @param {IL2MessageServiceLogClient} l2MessageServiceLogClient - An instance of a class implementing the `IL2MessageServiceLogClient` interface for fetching events from the blockchain.
   * @param {SDKMode} mode - The mode in which the SDK is operating, e.g., `read-only` or `read-write`.
   * @param {Signer} [signer] - An optional Ethers.js signer object for signing transactions.
   * @param {bigint} [maxFeePerGas] - The maximum fee per gas that the user is willing to pay.
   * @param {number} [gasEstimationPercentile] - The percentile to sample from each block's effective priority fees.
   * @param {boolean} [isMaxGasFeeEnforced] - Whether to enforce the maximum gas fee.
   * @param {number} [l2MessageTreeDepth] - The depth of the L2 message tree, used for merkle proof calculations.
   */
  constructor(
    provider: JsonRpcProvider,
    private readonly contractAddress: string,
    private readonly lineaRollupLogClient: ILineaRollupLogClient,
    private readonly l2MessageServiceLogClient: IL2MessageServiceLogClient,
    private readonly mode: SDKMode,
    private readonly signer?: Signer,
    maxFeePerGas?: bigint,
    gasEstimationPercentile?: number,
    isMaxGasFeeEnforced?: boolean,
    l2MessageTreeDepth?: number,
  ) {
    super(provider, maxFeePerGas, gasEstimationPercentile, isMaxGasFeeEnforced);
    this.contract = this.getContract(this.contractAddress, this.signer);
    this.l2MessageTreeDepth = l2MessageTreeDepth ?? DEFAULT_L2_MESSAGE_TREE_DEPTH;
  }

  /**
   * Retrieves message information by message hash.
   * @param {string} messageHash - The hash of the message sent on L1.
   * @returns {Promise<MessageSent | null>} The message information or null if not found.
   */
  public async getMessageByMessageHash(messageHash: string): Promise<MessageSent | null> {
    const [event] = await this.lineaRollupLogClient.getMessageSentEvents({
      filters: { messageHash },
      fromBlock: 0,
      toBlock: "latest",
    });
    return event ?? null;
  }

  /**
   * Retrieves messages information by the transaction hash.
   * @param {string} transactionHash - The hash of the `sendMessage` transaction on L1.
   * @returns {Promise<MessageSent[] | null>} An array of message information or null if not found.
   */
  public async getMessagesByTransactionHash(transactionHash: string): Promise<MessageSent[] | null> {
    const receipt = await this.provider.getTransactionReceipt(transactionHash);
    if (!receipt) {
      return null;
    }

    const messageSentEvents = await Promise.all(
      receipt.logs
        .filter((log) => log.address === this.contractAddress && log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE)
        .map((log) => this.contract.interface.parseLog(log))
        .filter((log) => !isNull(log))
        .map((log) => this.getMessageByMessageHash(log!.args._messageHash)),
    );
    return messageSentEvents.filter((log) => !isNull(log)) as MessageSent[];
  }

  /**
   * Retrieves the transaction receipt by message hash.
   * @param {string} messageHash - The hash of the message sent on L1.
   * @returns {Promise<TransactionReceipt | null>} The transaction receipt or null if not found.
   */
  public async getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null> {
    const [event] = await this.lineaRollupLogClient.getMessageSentEvents({
      filters: { messageHash },
      fromBlock: 0,
      toBlock: "latest",
    });

    if (!event) {
      return null;
    }

    const receipt = await this.provider.getTransactionReceipt(event.transactionHash);
    if (!receipt) {
      return null;
    }
    return receipt;
  }

  /**
   * Retrieves the LineaRollup contract instance.
   * @param {string} contractAddress - Address of the L1 contract.
   * @param {Signer} [signer] - The signer instance.
   * @returns {LineaRollup} The LineaRollup contract instance.
   */
  private getContract(contractAddress: string, signer?: Signer): LineaRollup {
    if (this.mode === "read-only") {
      return LineaRollup__factory.connect(contractAddress, this.provider);
    }

    if (!signer) {
      throw new BaseError("Please provide a signer.");
    }

    return LineaRollup__factory.connect(contractAddress, signer);
  }

  /**
   * Retrieves the L2 message status on L1 using the message hash (for messages sent before migration).
   * @param {string} messageHash - The hash of the message sent on L2.
   * @param {Overrides} [overrides={}] - Ethers call overrides. Defaults to `{}` if not specified.
   * @returns {Promise<OnChainMessageStatus>} The message status (CLAIMED, CLAIMABLE, UNKNOWN).
   */
  public async getMessageStatusUsingMessageHash(
    messageHash: string,
    overrides: Overrides = {},
  ): Promise<OnChainMessageStatus> {
    let status = await this.contract.inboxL2L1MessageStatus(messageHash, overrides);
    if (status === BigInt(MESSAGE_UNKNOWN_STATUS)) {
      const events = await this.lineaRollupLogClient.getMessageClaimedEvents({
        filters: { messageHash },
        fromBlock: 0,
        toBlock: "latest",
      });
      if (events.length > 0) {
        status = BigInt(MESSAGE_CLAIMED_STATUS);
      }
    }
    return formatMessageStatus(status);
  }

  /**
   * Retrieves the L2 message status on L1.
   * @param {string} messageHash - The hash of the message sent on L2.
   * @param {Overrides} [overrides={}] - Ethers call overrides. Defaults to `{}` if not specified.
   * @returns {Promise<OnChainMessageStatus>} The message status (CLAIMED, CLAIMABLE, UNKNOWN).
   */
  public async getMessageStatus(messageHash: string, overrides: Overrides = {}): Promise<OnChainMessageStatus> {
    return this.getMessageStatusUsingMerkleTree(messageHash, overrides);
  }

  /**
   * Retrieves the L2 message status on L1 using merkle tree (for messages sent after migration).
   * @param {string} messageHash - The hash of the message sent on L2.
   * @param {Overrides} [overrides={}] - Ethers call overrides. Defaults to `{}` if not specified.
   * @returns {Promise<OnChainMessageStatus>} The message status (CLAIMED, CLAIMABLE, UNKNOWN).
   */
  public async getMessageStatusUsingMerkleTree(
    messageHash: string,
    overrides: Overrides = {},
  ): Promise<OnChainMessageStatus> {
    const [messageEvent] = await this.l2MessageServiceLogClient.getMessageSentEventsByMessageHash({ messageHash });

    if (!messageEvent) {
      throw new BaseError(`Message hash does not exist on L2. Message hash: ${messageHash}`);
    }

    const [[l2MessagingBlockAnchoredEvent], isMessageClaimed] = await Promise.all([
      this.lineaRollupLogClient.getL2MessagingBlockAnchoredEvents({
        filters: { l2Block: BigInt(messageEvent.blockNumber) },
      }),
      this.contract.isMessageClaimed(messageEvent.messageNonce, overrides),
    ]);

    if (isMessageClaimed) {
      return OnChainMessageStatus.CLAIMED;
    }

    if (l2MessagingBlockAnchoredEvent) {
      return OnChainMessageStatus.CLAIMABLE;
    }

    return OnChainMessageStatus.UNKNOWN;
  }

  /**
   * Estimates the gas required for the claimMessage transaction.
   * @param {MessageProps & { feeRecipient?: string }} message - The message information.
   * @param {Overrides} [overrides={}] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<bigint>} The estimated transaction gas.
   */
  public async estimateClaimWithoutProofGas(
    message: MessageProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<bigint> {
    if (this.mode === "read-only") {
      throw new BaseError("'EstimateClaimGas' function not callable using readOnly mode.");
    }

    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l1FeeRecipient = feeRecipient ?? ZERO_ADDRESS;
    try {
      return await this.contract.claimMessage.estimateGas(
        messageSender,
        destination,
        fee,
        value,
        l1FeeRecipient,
        calldata,
        messageNonce,
        {
          ...(await this.get1559Fees()),
          ...overrides,
        },
      );
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      throw new GasEstimationError(e, message);
    }
  }

  /**
   * Estimates the gas required for the claimMessageWithProof transaction.
   * @param {MessageProps & { feeRecipient?: string }} message - The message information.
   * @param {Overrides} [overrides={}] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<bigint>} The estimated transaction gas.
   */
  public async estimateClaimGas(
    message: MessageProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<bigint> {
    if (this.mode === "read-only") {
      throw new BaseError("'EstimateClaimGas' function not callable using readOnly mode.");
    }

    try {
      const merkleTreeInfo = await this.getMessageProof(message.messageHash);

      return await this.estimateClaimWithProofGas(
        {
          ...message,
          merkleRoot: merkleTreeInfo.root,
          proof: merkleTreeInfo.proof,
          leafIndex: merkleTreeInfo.leafIndex,
        },
        overrides,
      );
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      throw new GasEstimationError(e, message);
    }
  }

  /**
   * Claims the message on L1 without merkle tree (for message sent before the migration).
   * @param {MessageProps & { feeRecipient?: string }} message - The message information.
   * @param {Overrides} [overrides={}] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<ContractTransactionResponse>} The transaction response.
   */
  public async claimWithoutProof(
    message: MessageProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<ContractTransactionResponse> {
    if (this.mode === "read-only") {
      throw new BaseError("'claim' function not callable using readOnly mode.");
    }

    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l1FeeRecipient = feeRecipient ?? ZERO_ADDRESS;

    return await this.contract.claimMessage(
      messageSender,
      destination,
      fee,
      value,
      l1FeeRecipient,
      calldata,
      messageNonce,
      {
        ...(await this.get1559Fees()),
        ...overrides,
      },
    );
  }

  /**
   * Claims the message with proof on L1.
   * @param {MessageProps & { feeRecipient?: string }} message - The message information.
   * @param {Overrides} [overrides={}] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<ContractTransactionResponse>} The transaction response.
   */
  public async claim(
    message: MessageProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<ContractTransactionResponse> {
    if (this.mode === "read-only") {
      throw new BaseError("'claim' function not callable using readOnly mode.");
    }

    const merkleTreeInfo = await this.getMessageProof(message.messageHash);

    return this.claimWithProof(
      {
        ...message,
        merkleRoot: merkleTreeInfo.root,
        proof: merkleTreeInfo.proof,
        leafIndex: merkleTreeInfo.leafIndex,
      },
      overrides,
    );
  }

  /**
   * Retrieves the finalization messaging info. This function is used in the L1 claiming flow.
   * @param {string} transactionHash - The finalization transaction hash.
   * @returns {Promise<FinalizationMessagingInfo>} The finalization messaging info: l2MessagingBlocksRange, l2MerkleRoots, treeDepth.
   */
  public async getFinalizationMessagingInfo(transactionHash: string): Promise<FinalizationMessagingInfo> {
    const receipt = await this.provider.getTransactionReceipt(transactionHash);

    if (!receipt || receipt.logs.length === 0) {
      throw new BaseError(`Transaction does not exist or no logs found in this transaction: ${transactionHash}.`);
    }

    let treeDepth = 0;
    const l2MerkleRoots: string[] = [];
    const blocksNumber: number[] = [];

    const filteredLogs = receipt.logs.filter((log) => log.address === this.contractAddress);

    for (const log of filteredLogs) {
      const parsedLog = this.contract.interface.parseLog(log);

      if (log.topics[0] === L2_MERKLE_TREE_ADDED_EVENT_SIGNATURE) {
        treeDepth = parseInt(parsedLog?.args.treeDepth);
        l2MerkleRoots.push(parsedLog?.args.l2MerkleRoot);
      } else if (log.topics[0] === L2_MESSAGING_BLOCK_ANCHORED_EVENT_SIGNATURE) {
        blocksNumber.push(Number(parsedLog?.args.l2Block));
      }
    }

    if (l2MerkleRoots.length === 0) {
      throw new BaseError(`No L2MerkleRootAdded events found in this transaction.`);
    }

    if (blocksNumber.length === 0) {
      throw new BaseError(`No L2MessagingBlocksAnchored events found in this transaction.`);
    }

    return {
      l2MessagingBlocksRange: {
        startingBlock: Math.min(...blocksNumber),
        endBlock: Math.max(...blocksNumber),
      },
      l2MerkleRoots,
      treeDepth,
    };
  }

  /**
   * Retrieves L2 message hashes in a specific L2 block range.
   * @param {number} fromBlock - The starting block number.
   * @param {number} toBlock - The ending block number.
   * @returns {Promise<string[]>} A list of all L2 message hashes in the specified block range.
   */
  public async getL2MessageHashesInBlockRange(fromBlock: number, toBlock: number): Promise<string[]> {
    const events = await this.l2MessageServiceLogClient.getMessageSentEventsByBlockRange(fromBlock, toBlock);

    if (events.length === 0) {
      throw new BaseError(`No MessageSent events found in this block range on L2.`);
    }

    return events.map((event) => event.messageHash);
  }

  /**
   * Retrieves message siblings for building the merkle tree. This merkle tree will be used to generate the proof to be able to claim the message on L1.
   * @param {string} messageHash - The message hash.
   * @param {string[]} messageHashes - The list of all L2 message hashes finalized in the same finalization transaction on L1.
   * @param {number} treeDepth - The merkle tree depth.
   * @returns {string[]} The message siblings.
   */
  public getMessageSiblings(messageHash: string, messageHashes: string[], treeDepth: number): string[] {
    const numberOfMessagesInTrees = 2 ** treeDepth;
    const messageHashesLength = messageHashes.length;

    const messageHashIndex = messageHashes.indexOf(messageHash);

    if (messageHashIndex === -1) {
      throw new BaseError("Message hash not found in messages.");
    }

    const start = Math.floor(messageHashIndex / numberOfMessagesInTrees) * numberOfMessagesInTrees;
    const end = Math.min(messageHashesLength, start + numberOfMessagesInTrees);

    const siblings = messageHashes.slice(start, end);

    const remainder = siblings.length % numberOfMessagesInTrees;
    if (remainder !== 0) {
      siblings.push(...Array(numberOfMessagesInTrees - remainder).fill(ZERO_HASH));
    }

    return siblings;
  }

  /**
   * Retrieves the message proof for claiming the message on L1.
   * @param {string} messageHash - The message hash.
   * @returns {Promise<Proof>} The merkle root, the merkle proof and the message leaf index.
   */
  public async getMessageProof(messageHash: string): Promise<Proof> {
    const [messageEvent] = await this.l2MessageServiceLogClient.getMessageSentEventsByMessageHash({ messageHash });

    if (!messageEvent) {
      throw new BaseError(`Message hash does not exist on L2. Message hash: ${messageHash}`);
    }

    const [l2MessagingBlockAnchoredEvent] = await this.lineaRollupLogClient.getL2MessagingBlockAnchoredEvents({
      filters: { l2Block: BigInt(messageEvent.blockNumber) },
    });

    if (!l2MessagingBlockAnchoredEvent) {
      throw new BaseError(`L2 block number ${messageEvent.blockNumber} has not been finalized on L1.`);
    }

    const finalizationInfo = await this.getFinalizationMessagingInfo(l2MessagingBlockAnchoredEvent.transactionHash);

    const l2MessageHashesInBlockRange = await this.getL2MessageHashesInBlockRange(
      finalizationInfo.l2MessagingBlocksRange.startingBlock,
      finalizationInfo.l2MessagingBlocksRange.endBlock,
    );

    const l2messages = this.getMessageSiblings(messageHash, l2MessageHashesInBlockRange, finalizationInfo.treeDepth);

    const merkleTreeFactory = new SparseMerkleTreeFactory(this.l2MessageTreeDepth);
    const tree = merkleTreeFactory.createAndAddLeaves(l2messages);

    if (!finalizationInfo.l2MerkleRoots.includes(tree.getRoot())) {
      throw new BaseError("Merkle tree build failed.");
    }

    return tree.getProof(l2messages.indexOf(messageHash));
  }

  /**
   * Retries a specific transaction with a higher fee.
   * @param {string} transactionHash - The transaction hash.
   * @param {number} [priceBumpPercent=10] - The percentage of price increase. Defaults to `10` if not specified.
   * @returns {Promise<TransactionResponse>} The transaction response.
   */
  public async retryTransactionWithHigherFee(
    transactionHash: string,
    priceBumpPercent: number = 10,
  ): Promise<TransactionResponse> {
    if (!Number.isInteger(priceBumpPercent)) {
      throw new Error("'priceBumpPercent' must be an integer");
    }

    if (this.mode === "read-only") {
      throw new BaseError("'retryTransactionWithHigherFee' function not callable using readOnly mode.");
    }

    const transaction = await this.provider.getTransaction(transactionHash);

    if (!transaction) {
      throw new BaseError(`Transaction with hash ${transactionHash} not found.`);
    }

    let maxPriorityFeePerGas;
    let maxFeePerGas;

    if (!transaction.maxPriorityFeePerGas || !transaction.maxFeePerGas) {
      const txFees = await this.get1559Fees();
      maxPriorityFeePerGas = txFees.maxPriorityFeePerGas;
      maxFeePerGas = txFees.maxFeePerGas;
    } else {
      maxPriorityFeePerGas = (transaction.maxPriorityFeePerGas * (BigInt(priceBumpPercent) + 100n)) / 100n;
      maxFeePerGas = (transaction.maxFeePerGas * (BigInt(priceBumpPercent) + 100n)) / 100n;
      if (maxPriorityFeePerGas > this.maxFeePerGas) {
        maxPriorityFeePerGas = this.maxFeePerGas;
      }
      if (maxFeePerGas > this.maxFeePerGas) {
        maxFeePerGas = this.maxFeePerGas;
      }
    }

    const updatedTransaction: TransactionRequest = {
      to: transaction.to,
      value: transaction.value,
      data: transaction.data,
      nonce: transaction.nonce,
      gasLimit: transaction.gasLimit,
      chainId: transaction.chainId,
      type: 2,
      maxPriorityFeePerGas,
      maxFeePerGas,
    };
    const signedTransaction = await this.signer!.signTransaction(updatedTransaction);
    return await this.provider.broadcastTransaction(signedTransaction);
  }

  /**
   * Checks if the withdrawal rate limit has been exceeded.
   * @param {bigint} _messageFee - The message fee.
   * @param {bigint} _messageValue - The message value.
   * @returns {Promise<boolean>} True if the rate limit has been exceeded, false otherwise.
   */
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public async isRateLimitExceeded(_messageFee: bigint, _messageValue: bigint): Promise<boolean> {
    return false;
  }

  /**
   * Checks if an error is of type 'RateLimitExceeded'.
   * @param {string} transactionHash - The transaction hash.
   * @returns {Promise<boolean>} True if the error type is 'RateLimitExceeded', false otherwise.
   */
  public async isRateLimitExceededError(transactionHash: string): Promise<boolean> {
    try {
      const tx = await this.provider.getTransaction(transactionHash);
      const errorEncodedData = await this.provider.call({
        to: tx?.to,
        from: tx?.from,
        nonce: tx?.nonce,
        gasLimit: tx?.gasLimit,
        data: tx?.data,
        value: tx?.value,
        chainId: tx?.chainId,
        accessList: tx?.accessList,
        maxPriorityFeePerGas: tx?.maxPriorityFeePerGas,
        maxFeePerGas: tx?.maxFeePerGas,
      });
      const error = this.contract.interface.parseError(errorEncodedData);

      return error?.name === "RateLimitExceeded";
    } catch (e) {
      return false;
    }
  }

  /**
   * Estimates the gas required for the claimMessageWithProof transaction.
   * @param {MessageWithProofProps & { feeRecipient?: string }} messageWithProof - The message information with proof.
   * @param {Overrides} [overrides={}] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<bigint>} The estimated gas.
   */
  public async estimateClaimWithProofGas(
    messageWithProof: MessageWithProofProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<bigint> {
    if (this.mode === "read-only") {
      throw new BaseError("'EstimateClaimWithProofGas' function not callable using readOnly mode.");
    }

    const {
      messageSender,
      destination,
      fee,
      value,
      calldata,
      messageNonce,
      feeRecipient,
      proof,
      leafIndex,
      merkleRoot,
    } = messageWithProof;

    const l1FeeRecipient = feeRecipient ?? ZERO_ADDRESS;
    try {
      return await this.contract.claimMessageWithProof.estimateGas(
        {
          from: messageSender,
          to: destination,
          fee,
          value,
          data: calldata,
          messageNumber: messageNonce,
          proof,
          leafIndex,
          merkleRoot,
          feeRecipient: l1FeeRecipient,
        },

        {
          ...(await this.get1559Fees()),
          ...overrides,
        },
      );
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      throw new GasEstimationError(e, messageWithProof);
    }
  }

  /**
   * Claims the message using merkle proof on L1.
   * @param {MessageWithProofProps & { feeRecipient?: string }} message - The message information with proof.
   * @param {Overrides} [overrides={}] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<ContractTransactionResponse>} The transaction response.
   */
  public async claimWithProof(
    message: MessageWithProofProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<ContractTransactionResponse> {
    if (this.mode === "read-only") {
      throw new BaseError("'claimWithProof' function not callable using readOnly mode.");
    }

    const {
      messageSender,
      destination,
      fee,
      value,
      calldata,
      messageNonce,
      feeRecipient,
      proof,
      leafIndex,
      merkleRoot,
    } = message;

    const l1FeeRecipient = feeRecipient ?? ZERO_ADDRESS;

    return await this.contract.claimMessageWithProof(
      {
        from: messageSender,
        to: destination,
        fee,
        value,
        data: calldata,
        messageNumber: messageNonce,
        proof,
        leafIndex,
        merkleRoot,
        feeRecipient: l1FeeRecipient,
      },
      {
        ...(await this.get1559Fees()),
        ...overrides,
      },
    );
  }
}
