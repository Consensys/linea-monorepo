import { localL1Network, localL2Network } from "@/constants";
import { test as setup } from "@playwright/test";
import {
  createPublicClient,
  createWalletClient,
  decodeEventLog,
  encodeFunctionData,
  erc20Abi,
  formatEther,
  http,
  parseEther,
} from "viem";
import { mnemonicToAccount, privateKeyToAccount } from "viem/accounts";
import { waitForTransactionReceipt } from "viem/actions";
import { sendTransactionsToGenerateTrafficWithInterval } from "./utils/utils";
import { LOCAL_L1_NETWORK, LOCAL_L2_NETWORK } from "./constants";
import { estimateGas } from "viem/linea";

setup.setTimeout(200_000);
setup("Global setup", async () => {
  console.log("Generating L2 traffic...");
  await generateL2Traffic();

  console.log("Funding L2 account...");
  await fundL2Account();

  console.log("Sending ERC20 token to account...");
  await sendERC20TokenToAccount();
});

async function generateL2Traffic() {
  // FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  const pollingAccount = privateKeyToAccount("0xb17202c37cce9498e6f7dcdc1abd207802d09b5eee96677ea219ac867a198b91");

  const walletClient = createWalletClient({
    chain: localL2Network,
    transport: http(LOCAL_L2_NETWORK.rpcUrl),
    account: pollingAccount,
  });
  await sendTransactionsToGenerateTrafficWithInterval(walletClient, 2_000);
}

async function fundL2Account() {
  const account = mnemonicToAccount(process.env.E2E_TEST_SEED_PHRASE, { accountIndex: 0, addressIndex: 6 });

  const walletClient = createWalletClient({
    chain: localL1Network,
    transport: http(LOCAL_L1_NETWORK.rpcUrl),
    account,
  });

  const publicClient = createPublicClient({
    chain: localL1Network,
    transport: http(LOCAL_L1_NETWORK.rpcUrl),
  });

  const { maxPriorityFeePerGas, maxFeePerGas } = await publicClient.estimateFeesPerGas({ type: "eip1559" });

  const transactionHash = await walletClient.sendTransaction({
    to: "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9",
    value: parseEther("10000"),
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
      ] as const,
      functionName: "sendMessage",
      args: [account.address, 0n, "0x"],
    }),
    maxPriorityFeePerGas,
    maxFeePerGas,
  });

  console.log(`Transaction sent. Hash: ${transactionHash}`);

  const receipt = await waitForTransactionReceipt(walletClient, { hash: transactionHash, confirmations: 1 });

  const messageSentEventLog = receipt.logs.find(
    (log) => log.topics[0] === "0xe856c2b8bd4eb0027ce32eeaf595c21b0b6b4644b326e5b7bd80a1cf8db72e6c",
  );

  if (!messageSentEventLog) {
    throw new Error("MessageSent event not found in transaction receipt logs");
  }

  const messageSentEvent = decodeEventLog({
    abi: [
      {
        anonymous: false,
        inputs: [
          { indexed: true, internalType: "address", name: "_from", type: "address" },
          { indexed: true, internalType: "address", name: "_to", type: "address" },
          { indexed: false, internalType: "uint256", name: "_fee", type: "uint256" },
          { indexed: false, internalType: "uint256", name: "_value", type: "uint256" },
          { indexed: false, internalType: "uint256", name: "_nonce", type: "uint256" },
          { indexed: false, internalType: "bytes", name: "_calldata", type: "bytes" },
          { indexed: true, internalType: "bytes32", name: "_messageHash", type: "bytes32" },
        ],
        name: "MessageSent",
        type: "event",
      },
    ],
    data: messageSentEventLog.data,
    topics: messageSentEventLog.topics,
  });

  const messageHash = messageSentEvent.args._messageHash;

  console.log(`Message hash: ${messageHash}`);

  const l2Client = createPublicClient({
    chain: localL2Network,
    transport: http(LOCAL_L2_NETWORK.rpcUrl),
  });

  let events = [];

  while (events.length === 0) {
    events = await l2Client.getContractEvents({
      address: "0xe537D669CA013d86EBeF1D64e40fC74CADC91987",
      abi: [
        {
          anonymous: false,
          inputs: [{ indexed: true, internalType: "bytes32", name: "_messageHash", type: "bytes32" }],
          name: "MessageClaimed",
          type: "event",
        },
      ] as const,
      eventName: "MessageClaimed",
      args: {
        _messageHash: messageHash,
      },
      fromBlock: 0n,
      toBlock: "latest",
    });

    console.log(`Waiting for MessageClaimed event on L2... Found ${events.length} events so far.`);

    await new Promise((resolve) => setTimeout(resolve, 2000));
  }

  const l2AccountBalance = await l2Client.getBalance({
    address: account.address,
    blockTag: "latest",
  });

  console.log(`L2 account balance after sending ETH: ${formatEther(l2AccountBalance)} ETH`);

  if (l2AccountBalance === 0n) {
    throw new Error("L2 account balance is zero after sending ETH");
  }
}

async function sendERC20TokenToAccount() {
  const account = mnemonicToAccount(process.env.E2E_TEST_SEED_PHRASE, { accountIndex: 0, addressIndex: 6 });

  const walletClient = createWalletClient({
    chain: localL1Network,
    transport: http(LOCAL_L1_NETWORK.rpcUrl),
    account,
  });

  const publicClient = createPublicClient({
    chain: localL1Network,
    transport: http(LOCAL_L1_NETWORK.rpcUrl),
  });

  const { maxPriorityFeePerGas, maxFeePerGas } = await publicClient.estimateFeesPerGas({ type: "eip1559" });

  const transactionHash = await walletClient.sendTransaction({
    to: "0x8A791620dd6260079BF849Dc5567aDC3F2FdC318",
    value: 0n,
    data: encodeFunctionData({
      abi: [
        {
          inputs: [
            {
              internalType: "address",
              name: "_to",
              type: "address",
            },
            {
              internalType: "uint256",
              name: "_amount",
              type: "uint256",
            },
          ],
          name: "mint",
          outputs: [],
          stateMutability: "nonpayable",
          type: "function",
        },
      ],
      functionName: "mint",
      args: [account.address, parseEther("1000")],
    }),
    maxPriorityFeePerGas: maxPriorityFeePerGas,
    maxFeePerGas: maxFeePerGas,
  });

  await waitForTransactionReceipt(walletClient, { hash: transactionHash, confirmations: 1 });

  const l2WalletClient = createWalletClient({
    chain: localL2Network,
    transport: http(LOCAL_L2_NETWORK.rpcUrl),
    account,
  });

  const l2PublicClient = createPublicClient({
    chain: localL2Network,
    transport: http(LOCAL_L2_NETWORK.rpcUrl),
  });

  const { priorityFeePerGas, baseFeePerGas, gasLimit } = await estimateGas(l2WalletClient, {
    account: l2WalletClient.account,
    to: "0xCC1B08B17301e090cbb4c1F5598Cbaa096d591FB",
    value: 0n,
    data: encodeFunctionData({
      abi: [
        {
          inputs: [
            {
              internalType: "address",
              name: "_to",
              type: "address",
            },
            {
              internalType: "uint256",
              name: "_amount",
              type: "uint256",
            },
          ],
          name: "mint",
          outputs: [],
          stateMutability: "nonpayable",
          type: "function",
        },
      ],
      functionName: "mint",
      args: [account.address, parseEther("1000")],
    }),
  });

  const transactionHashL2 = await l2WalletClient.sendTransaction({
    to: "0xCC1B08B17301e090cbb4c1F5598Cbaa096d591FB",
    value: 0n,
    data: encodeFunctionData({
      abi: [
        {
          inputs: [
            {
              internalType: "address",
              name: "_to",
              type: "address",
            },
            {
              internalType: "uint256",
              name: "_amount",
              type: "uint256",
            },
          ],
          name: "mint",
          outputs: [],
          stateMutability: "nonpayable",
          type: "function",
        },
      ],
      functionName: "mint",
      args: [account.address, parseEther("1000")],
    }),
    maxPriorityFeePerGas: priorityFeePerGas,
    maxFeePerGas: priorityFeePerGas + baseFeePerGas,
    gas: gasLimit,
  });

  await waitForTransactionReceipt(l2WalletClient, { hash: transactionHashL2, confirmations: 1 });

  const l1Balance = await publicClient.readContract({
    address: "0x8A791620dd6260079BF849Dc5567aDC3F2FdC318",
    abi: erc20Abi,
    functionName: "balanceOf",
    args: [account.address],
  });

  const l2Balance = await l2PublicClient.readContract({
    address: "0xCC1B08B17301e090cbb4c1F5598Cbaa096d591FB",
    abi: erc20Abi,
    functionName: "balanceOf",
    args: [account.address],
  });

  console.log(`L1 Token balance after minting: ${formatEther(l1Balance)} tokens`);
  console.log(`L2 Token balance after minting: ${formatEther(l2Balance)} tokens`);
}
