import {
  Account,
  Address,
  BaseError,
  Chain,
  Client,
  DeriveChain,
  encodeAbiParameters,
  encodeFunctionData,
  erc20Abi,
  EstimateContractGasParameters,
  FormattedTransactionRequest,
  GetChainParameter,
  Hex,
  keccak256,
  SendTransactionParameters,
  SendTransactionReturnType,
  StateOverride,
  Transport,
  UnionEvaluate,
  UnionOmit,
  zeroAddress,
} from "viem";
import { GetAccountParameter } from "../types/account";
import { parseAccount } from "viem/utils";
import { estimateContractGas, estimateFeesPerGas, multicall, readContract, sendTransaction } from "viem/actions";
import { getContractsAddressesByChainId } from "@consensys/linea-sdk-core";

export type DepositParameters<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  chainL2 extends Chain | undefined = Chain | undefined,
  accountL2 extends Account | undefined = Account | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
> = UnionEvaluate<UnionOmit<FormattedTransactionRequest<derivedChain>, "data" | "to" | "from">> &
  Partial<GetChainParameter<chain, chainOverride>> &
  Partial<GetAccountParameter<account>> & {
    l2Client: Client<Transport, chainL2, accountL2>;
    token: Address;
    to: Address;
    fee?: bigint;
    amount: bigint;
    data?: Hex;
  };

export type DepositReturnType = SendTransactionReturnType;

/**
 * Deposits tokens from L1 to L2 or ETH if `token` is set to `zeroAddress`.
 *
 * @param client - Client to use
 * @param parameters - {@link DepositParameters}
 * @returns hash - The [Transaction](https://viem.sh/docs/glossary/terms#transaction) hash. {@link DepositReturnType}
 *
 * @example
 * import { createPublicClient, http, zeroAddress } from 'viem'
 * import { privateKeyToAccount } from 'viem/accounts'
 * import { mainnet, linea } from 'viem/chains'
 * import { deposit } from '@consensys/linea-sdk-viem'
 *
 * const client = createPublicClient({
 *   chain: mainnet,
 *   transport: http(),
 * });
 *
 * const l2Client = createPublicClient({
 *   chain: linea,
 *   transport: http(),
 * });
 *
 * const hash = await deposit(client, {
 *     l2Client,
 *     account: privateKeyToAccount('0x…'),
 *     amount: 1_000_000_000_000n,
 *     token: zeroAddress, // Use zeroAddress for ETH
 *     to: '0xRecipientAddress',
 *     data: '0x', // Optional data
 *     fee: 100_000_000n, // Optional fee
 * });
 *
 * @example Account Hoisting
 * import { createPublicClient, createWalletClient, http, zeroAddress } from 'viem'
 * import { privateKeyToAccount } from 'viem/accounts'
 * import { mainnet, linea } from 'viem/chains'
 * import { deposit } from '@consensys/linea-sdk-viem'
 *
 * const client = createWalletClient({
 *   account: privateKeyToAccount('0x…'),
 *   chain: mainnet,
 *   transport: http(),
 * });
 *
 * const l2Client = createPublicClient({
 *  chain: linea,
 *  transport: http(),
 * });
 *
 * const hash = await deposit(client, {
 *     l2Client
 *     amount: 1_000_000_000_000n,
 *     token: zeroAddress, // Use zeroAddress for ETH
 *     to: '0xRecipientAddress',
 *     data: '0x', // Optional data
 *     fee: 100_000_000n, // Optional fee
 * });
 */
export async function deposit<
  chain extends Chain | undefined,
  account extends Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  chainL2 extends Chain | undefined = Chain | undefined,
  accountL2 extends Account | undefined = Account | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
>(
  client: Client<Transport, chain, account>,
  parameters: DepositParameters<chain, account, chainOverride, chainL2, accountL2, derivedChain>,
) {
  const { account: account_ = client.account, l2Client, token, amount, data, to, ...tx } = parameters;
  let { fee } = parameters;

  const account = account_ ? parseAccount(account_) : client.account;
  if (!account) {
    throw new BaseError("Account is required to send a transaction");
  }

  const l1ChainId = client.chain?.id;
  const l2ChainId = l2Client.chain?.id;

  if (!l1ChainId || !l2ChainId) {
    throw new BaseError("No chain id found in l1 or l2 client");
  }

  if (!fee || fee === 0n) {
    const { maxFeePerGas } = await estimateFeesPerGas(l2Client, { type: "eip1559", chain: l2Client.chain });

    const nextMessageNumber = await readContract(client, {
      address: getContractsAddressesByChainId(l1ChainId).messageService,
      abi: [
        {
          inputs: [],
          name: "nextMessageNumber",
          outputs: [
            {
              internalType: "uint256",
              name: "",
              type: "uint256",
            },
          ],
          stateMutability: "view",
          type: "function",
        },
      ],
      functionName: "nextMessageNumber",
    });

    if (token === zeroAddress) {
      const l2ClaimingTxGasLimit = await estimateEthBridgingGasUsed(l2Client, {
        chainId: l1ChainId,
        account: account.address,
        recipient: to as Address,
        amount,
        nextMessageNumber,
      });
      fee = maxFeePerGas * (l2ClaimingTxGasLimit + 6_000n);
    } else {
      const l2ClaimingTxGasLimit = await estimateERC20BridgingGasUsed(l2Client, {
        account: account.address,
        token,
        l1ChainId,
        l2ChainId,
        amount,
        recipient: to,
        nextMessageNumber,
      });
      fee = maxFeePerGas * (l2ClaimingTxGasLimit + 6_000n);
    }
  }

  if (token === zeroAddress) {
    const lineaRollupAddress = getContractsAddressesByChainId(l1ChainId).messageService;

    return sendTransaction(client, {
      to: lineaRollupAddress,
      value: amount + fee,
      account: account,
      data: encodeFunctionData({
        abi: [
          {
            inputs: [
              { internalType: "address", name: "_to", type: "address" },
              { internalType: "uint256", name: "_fee", type: "uint256" },
              { internalType: "bytes", name: "_calldata", type: "bytes" },
            ],
            name: "sendMessage",
            outputs: [],
            stateMutability: "payable",
            type: "function",
          },
        ],
        functionName: "sendMessage",
        args: [to, fee, data ?? "0x"],
      }),
      ...tx,
    } as SendTransactionParameters);
  }

  const l1TokenBridgeAddress = getContractsAddressesByChainId(l1ChainId).tokenBridge;

  return sendTransaction(client, {
    to: l1TokenBridgeAddress,
    value: fee,
    account: account,
    data: encodeFunctionData({
      abi: [
        {
          inputs: [
            {
              internalType: "address",
              name: "_token",
              type: "address",
            },
            {
              internalType: "uint256",
              name: "_amount",
              type: "uint256",
            },
            {
              internalType: "address",
              name: "_recipient",
              type: "address",
            },
          ],
          name: "bridgeToken",
          outputs: [],
          stateMutability: "payable",
          type: "function",
        },
      ],
      functionName: "bridgeToken",
      args: [token, amount, to],
    }),
    ...tx,
  } as SendTransactionParameters);
}

async function estimateEthBridgingGasUsed<chain extends Chain | undefined, _account extends Account | undefined>(
  client: Client<Transport, chain, _account>,
  parameters: {
    chainId: number;
    account: Address;
    recipient: Address;
    amount: bigint;
    nextMessageNumber: bigint;
  },
) {
  const { account, recipient, amount, nextMessageNumber, chainId } = parameters;
  const messageHash = computeMessageHash(account, recipient, 0n, amount, nextMessageNumber, "0x");

  const messageServiceAddress = getContractsAddressesByChainId(chainId).messageService;

  const storageSlot = computeMessageStorageSlot(messageHash);
  const stateOverride = createStateOverride(messageServiceAddress, storageSlot);

  return estimateContractGas(client, {
    address: messageServiceAddress as `0x${string}`,
    abi: [
      {
        inputs: [
          { internalType: "address", name: "_from", type: "address" },
          { internalType: "address", name: "_to", type: "address" },
          { internalType: "uint256", name: "_fee", type: "uint256" },
          { internalType: "uint256", name: "_value", type: "uint256" },
          { internalType: "address payable", name: "_feeRecipient", type: "address" },
          { internalType: "bytes", name: "_calldata", type: "bytes" },
          { internalType: "uint256", name: "_nonce", type: "uint256" },
        ],
        name: "claimMessage",
        outputs: [],
        stateMutability: "nonpayable",
        type: "function",
      },
    ] as const,
    functionName: "claimMessage",
    account: account as `0x${string}`,
    args: [account, recipient, 0n, amount, zeroAddress as `0x${string}`, "0x" as `0x${string}`, nextMessageNumber],
    stateOverride,
  } as EstimateContractGasParameters);
}

async function estimateERC20BridgingGasUsed<chain extends Chain | undefined, _account extends Account | undefined>(
  client: Client<Transport, chain, _account>,
  parameters: {
    account: Address;
    token: Address;
    l1ChainId: number;
    l2ChainId: number;
    amount: bigint;
    recipient: Address;
    nextMessageNumber: bigint;
  },
) {
  const { token, l1ChainId, l2ChainId, amount, recipient, nextMessageNumber, account } = parameters;

  const { tokenAddress, chainId, tokenMetadata } = await prepareERC20TokenParams(client, {
    token,
    l1ChainId,
    l2ChainId,
  });

  const encodedData = encodeFunctionData({
    abi: [
      {
        inputs: [
          {
            internalType: "address",
            name: "_nativeToken",
            type: "address",
          },
          {
            internalType: "uint256",
            name: "_amount",
            type: "uint256",
          },
          {
            internalType: "address",
            name: "_recipient",
            type: "address",
          },
          {
            internalType: "uint256",
            name: "_chainId",
            type: "uint256",
          },
          {
            internalType: "bytes",
            name: "_tokenMetadata",
            type: "bytes",
          },
        ],
        name: "completeBridging",
        outputs: [],
        stateMutability: "nonpayable",
        type: "function",
      },
    ],
    functionName: "completeBridging",
    args: [tokenAddress, amount, recipient, BigInt(chainId), tokenMetadata],
  });

  const messageHash = computeMessageHash(
    getContractsAddressesByChainId(l1ChainId).tokenBridge,
    getContractsAddressesByChainId(l2ChainId).tokenBridge,
    0n,
    0n,
    nextMessageNumber,
    encodedData,
  );

  const storageSlot = computeMessageStorageSlot(messageHash);
  const stateOverride = createStateOverride(getContractsAddressesByChainId(l2ChainId).messageService, storageSlot);

  return estimateContractGas(client, {
    address: getContractsAddressesByChainId(l2ChainId).messageService,
    abi: [
      {
        inputs: [
          { internalType: "address", name: "_from", type: "address" },
          { internalType: "address", name: "_to", type: "address" },
          { internalType: "uint256", name: "_fee", type: "uint256" },
          { internalType: "uint256", name: "_value", type: "uint256" },
          { internalType: "address payable", name: "_feeRecipient", type: "address" },
          { internalType: "bytes", name: "_calldata", type: "bytes" },
          { internalType: "uint256", name: "_nonce", type: "uint256" },
        ],
        name: "claimMessage",
        outputs: [],
        stateMutability: "nonpayable",
        type: "function",
      },
    ] as const,
    functionName: "claimMessage",
    account: account,
    args: [
      getContractsAddressesByChainId(l1ChainId).tokenBridge,
      getContractsAddressesByChainId(l2ChainId).tokenBridge,
      0n,
      0n,
      zeroAddress,
      encodedData,
      nextMessageNumber,
    ],
    stateOverride,
  } as EstimateContractGasParameters);
}

async function prepareERC20TokenParams<chain extends Chain | undefined, _account extends Account | undefined>(
  client: Client<Transport, chain, _account>,
  parameters: {
    token: Address;
    l1ChainId: number;
    l2ChainId: number;
  },
): Promise<{ tokenAddress: Address; chainId: number; tokenMetadata: Hex }> {
  const { token, l1ChainId, l2ChainId } = parameters;

  const [tokenName, tokenSymbol, tokenDecimals, nativeToken] = await multicall(client, {
    contracts: [
      {
        address: token,
        abi: erc20Abi,
        functionName: "name",
      },
      {
        address: token,
        abi: erc20Abi,
        functionName: "symbol",
      },
      {
        address: token,
        abi: erc20Abi,
        functionName: "decimals",
      },
      {
        address: getContractsAddressesByChainId(l1ChainId).tokenBridge,
        abi: [
          {
            inputs: [
              {
                internalType: "address",
                name: "bridged",
                type: "address",
              },
            ],
            name: "bridgedToNativeToken",
            outputs: [
              {
                internalType: "address",
                name: "native",
                type: "address",
              },
            ],
            stateMutability: "view",
            type: "function",
          },
        ],
        functionName: "bridgedToNativeToken",
        args: [token],
      },
    ],
    allowFailure: false,
  });

  let tokenAddress = token;
  let chainId = l1ChainId;

  let tokenMetadata = encodeAbiParameters(
    [
      { name: "tokenName", type: "string" },
      { name: "tokenSymbol", type: "string" },
      { name: "tokenDecimals", type: "uint8" },
    ],
    [tokenName, tokenSymbol, tokenDecimals],
  );

  if (nativeToken !== zeroAddress) {
    tokenAddress = nativeToken;
    chainId = l2ChainId;
    tokenMetadata = "0x";
  }

  return { tokenAddress, chainId, tokenMetadata };
}

export function computeMessageHash(
  from: Address,
  to: Address,
  fee: bigint,
  value: bigint,
  nonce: bigint,
  calldata: `0x${string}` = "0x",
) {
  return keccak256(
    encodeAbiParameters(
      [
        { name: "from", type: "address" },
        { name: "to", type: "address" },
        { name: "fee", type: "uint256" },
        { name: "value", type: "uint256" },
        { name: "nonce", type: "uint256" },
        { name: "calldata", type: "bytes" },
      ],
      [from, to, fee, value, nonce, calldata],
    ),
  );
}

export function computeMessageStorageSlot(messageHash: `0x${string}`) {
  return keccak256(
    encodeAbiParameters(
      [
        { name: "messageHash", type: "bytes32" },
        { name: "mappingSlot", type: "uint256" },
      ],
      [messageHash, 176n],
    ),
  );
}

function createStateOverride(messageServiceAddress: Address, storageSlot: Hex): StateOverride {
  return [
    {
      address: messageServiceAddress,
      stateDiff: [
        {
          slot: storageSlot,
          value: "0x0000000000000000000000000000000000000000000000000000000000000001" as Hex,
        },
      ],
    },
  ];
}
