/* eslint-disable @typescript-eslint/no-explicit-any */
import { JsonRpcProvider, Signature, Signer, TransactionReceipt, TransactionResponse } from "ethers";
import { ILogger } from "../../core/utils/logging/ILogger";
import {
  TEST_ADDRESS_1,
  TEST_BLOCK_HASH,
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  TEST_MESSAGE_HASH,
  TEST_TRANSACTION_HASH,
} from "./constants";
import {
  DEFAULT_ENFORCE_MAX_GAS_FEE,
  DEFAULT_GAS_ESTIMATION_PERCENTILE,
  DEFAULT_L2_MESSAGE_TREE_DEPTH,
  DEFAULT_MAX_FEE_PER_GAS,
  MESSAGE_SENT_EVENT_SIGNATURE,
} from "../../core/constants";
import { Direction, MessageStatus } from "../../core/enums/MessageEnums";
import { Message, MessageProps } from "../../core/entities/Message";
import {
  LineaRollupClient,
  EthersLineaRollupLogClient,
  EthersL2MessageServiceLogClient,
  L2MessageServiceClient,
  LineaRollupMessageRetriever,
  L2MessageServiceMessageRetriever,
  SDKMode,
  Provider,
  LineaProvider,
  LineaGasProvider,
  DefaultGasProvider,
  MerkleTreeService,
} from "@consensys/linea-sdk";
import { MessageEntity } from "../../application/postman/persistence/entities/Message.entity";

export class TestLogger implements ILogger {
  public readonly name: string;

  constructor(loggerName: string) {
    this.name = loggerName;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public info(error: any): void {}

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public error(error: any): void {}

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public warn(error: any): void {}

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public debug(error: any): void {}

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public warnOrError(error: any): void {}
}

export const getTestProvider = () => {
  return new JsonRpcProvider("http://localhost:8545");
};

export const generateTransactionReceipt = (overrides?: Partial<TransactionReceipt>): TransactionReceipt => {
  return new TransactionReceipt(
    {
      hash: TEST_TRANSACTION_HASH,
      blockHash: TEST_BLOCK_HASH,
      to: TEST_CONTRACT_ADDRESS_1,
      from: TEST_ADDRESS_1,
      contractAddress: TEST_CONTRACT_ADDRESS_1,
      index: 0,
      gasUsed: 70_000n,
      logsBloom: "",
      logs: [
        {
          transactionIndex: 0,
          blockNumber: 100_000,
          removed: false,
          transactionHash: TEST_TRANSACTION_HASH,
          address: TEST_CONTRACT_ADDRESS_1,
          topics: [
            MESSAGE_SENT_EVENT_SIGNATURE,
            `0x000000000000000000000000${TEST_ADDRESS_1.slice(2)}`,
            `0x000000000000000000000000${TEST_ADDRESS_1.slice(2)}`,
            TEST_MESSAGE_HASH,
          ],
          data: "0x00000000000000000000000000000000000000000000000000038d7ea4c68000000000000000000000000000000000000000000000000000015fb7f9b8c3800000000000000000000000000000000000000000000000000000000000000003d700000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000",
          index: 0,
          blockHash: TEST_BLOCK_HASH,
        },
      ],
      blockNumber: 100_000,
      confirmations: 1,
      cumulativeGasUsed: 75_000n,
      effectiveGasPrice: 30_000n,
      root: "",
      type: 2,
      status: 1,
      ...overrides,
    },
    getTestProvider(),
  );
};

export const generateTransactionResponse = (overrides?: Partial<TransactionResponse>): TransactionResponse => {
  return new TransactionResponse(
    {
      hash: TEST_TRANSACTION_HASH,
      type: 2,
      accessList: null,
      index: 0,
      blockHash: TEST_BLOCK_HASH,
      blockNumber: 1077297,
      confirmations: async () => 212238,
      from: TEST_ADDRESS_1,
      gasPrice: BigInt("3492211493612"),
      gasLimit: BigInt("74959"),
      to: TEST_CONTRACT_ADDRESS_2,
      value: BigInt("4313771350571206145"),
      maxPriorityFeePerGas: 50_000_000n,
      maxFeePerGas: 100_000_000n,
      nonce: 40,
      signature: Signature.from(),
      data: `0x9f3ce55a000000000000000000000000${TEST_ADDRESS_1.slice(
        2,
      )}000000000000000000000000000000000000000000000000002386f26fc1000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000`,
      chainId: 59140n,
      ...overrides,
    },
    getTestProvider(),
  );
};

export const generateMessage = (overrides?: Partial<MessageProps>): Message => {
  return new Message({
    id: 1,
    messageSender: TEST_ADDRESS_1,
    destination: TEST_CONTRACT_ADDRESS_1,
    fee: 10n,
    value: 2n,
    messageNonce: 1n,
    calldata: "0x",
    messageHash: TEST_MESSAGE_HASH,
    contractAddress: TEST_CONTRACT_ADDRESS_2,
    sentBlockNumber: 100_000,
    direction: Direction.L1_TO_L2,
    status: MessageStatus.SENT,
    claimNumberOfRetry: 0,
    createdAt: new Date("2023-08-04"),
    updatedAt: new Date("2023-08-04"),
    ...overrides,
  });
};

export const generateMessageEntity = (overrides?: Partial<MessageEntity>): MessageEntity => {
  return {
    id: 1,
    messageSender: TEST_ADDRESS_1,
    destination: TEST_CONTRACT_ADDRESS_1,
    fee: "10",
    value: "2",
    messageNonce: 1,
    calldata: "0x",
    messageHash: TEST_MESSAGE_HASH,
    messageContractAddress: TEST_CONTRACT_ADDRESS_2,
    sentBlockNumber: 100_000,
    direction: Direction.L1_TO_L2,
    status: MessageStatus.SENT,
    claimNumberOfRetry: 0,
    createdAt: new Date("2023-08-04"),
    updatedAt: new Date("2023-08-04"),
    ...overrides,
  };
};

export function generateLineaRollupClient(
  l1Provider: Provider,
  l2Provider: LineaProvider,
  l1ContractAddress: string,
  l2ContractAddres: string,
  mode: SDKMode,
  signer?: Signer,
  gasFeesOptions?: {
    maxFeePerGas?: bigint;
    gasEstimationPercentile?: number;
    enforceMaxGasFee?: boolean;
  },
): {
  lineaRollupClient: LineaRollupClient;
  lineaRollupLogClient: EthersLineaRollupLogClient;
  l2MessageServiceLogClient: EthersL2MessageServiceLogClient;
  gasProvider: DefaultGasProvider;
  messageRetriever: LineaRollupMessageRetriever;
  merkleTreeService: MerkleTreeService;
} {
  const lineaRollupLogClient = new EthersLineaRollupLogClient(l1Provider, l1ContractAddress);
  const l2MessageServiceLogClient = new EthersL2MessageServiceLogClient(l2Provider, l2ContractAddres);
  const gasProvider = new DefaultGasProvider(l1Provider, {
    maxFeePerGas: gasFeesOptions?.maxFeePerGas ?? DEFAULT_MAX_FEE_PER_GAS,
    gasEstimationPercentile: gasFeesOptions?.gasEstimationPercentile ?? DEFAULT_GAS_ESTIMATION_PERCENTILE,
    enforceMaxGasFee: gasFeesOptions?.enforceMaxGasFee ?? DEFAULT_ENFORCE_MAX_GAS_FEE,
  });
  const messageRetriever = new LineaRollupMessageRetriever(l1Provider, lineaRollupLogClient, l1ContractAddress);
  const merkleTreeService = new MerkleTreeService(
    l1Provider,
    l1ContractAddress,
    lineaRollupLogClient,
    l2MessageServiceLogClient,
    DEFAULT_L2_MESSAGE_TREE_DEPTH,
  );
  const lineaRollupClient = new LineaRollupClient(
    l1Provider,
    l1ContractAddress,
    lineaRollupLogClient,
    l2MessageServiceLogClient,
    gasProvider,
    messageRetriever,
    merkleTreeService,
    mode,
    signer,
  );

  return {
    lineaRollupClient,
    lineaRollupLogClient,
    l2MessageServiceLogClient,
    gasProvider,
    messageRetriever,
    merkleTreeService,
  };
}

export function generateL2MessageServiceClient(
  l2Provider: LineaProvider,
  l2ContractAddress: string,
  mode: SDKMode,
  signer?: Signer,
  gasFeesOptions?: {
    maxFeePerGas?: bigint;
    gasEstimationPercentile?: number;
    enforceMaxGasFee?: boolean;
    enableLineaEstimateGas?: boolean;
  },
): {
  l2MessageServiceClient: L2MessageServiceClient;
  l2MessageServiceLogClient: EthersL2MessageServiceLogClient;
  gasProvider: LineaGasProvider | DefaultGasProvider;
  messageRetriever: L2MessageServiceMessageRetriever;
} {
  const l2MessageServiceLogClient = new EthersL2MessageServiceLogClient(l2Provider, l2ContractAddress);

  const messageRetriever = new L2MessageServiceMessageRetriever(
    l2Provider,
    l2MessageServiceLogClient,
    l2ContractAddress,
  );

  const gasProvider = gasFeesOptions?.enableLineaEstimateGas
    ? new LineaGasProvider(l2Provider, {
        maxFeePerGas: gasFeesOptions?.maxFeePerGas ?? DEFAULT_MAX_FEE_PER_GAS,
        enforceMaxGasFee: gasFeesOptions?.enforceMaxGasFee ?? DEFAULT_ENFORCE_MAX_GAS_FEE,
      })
    : new DefaultGasProvider(l2Provider, {
        maxFeePerGas: gasFeesOptions?.maxFeePerGas ?? DEFAULT_MAX_FEE_PER_GAS,
        gasEstimationPercentile: gasFeesOptions?.gasEstimationPercentile ?? DEFAULT_GAS_ESTIMATION_PERCENTILE,
        enforceMaxGasFee: gasFeesOptions?.enforceMaxGasFee ?? DEFAULT_ENFORCE_MAX_GAS_FEE,
      });

  const l2MessageServiceClient = new L2MessageServiceClient(
    l2Provider,
    l2ContractAddress,
    messageRetriever,
    gasProvider,
    mode,
    signer,
  );

  return { l2MessageServiceClient, l2MessageServiceLogClient, gasProvider, messageRetriever };
}
