import { Command, Flags } from "@oclif/core";
import { ethers, TransactionLike } from "ethers";
import { sanitizeAddress, sanitizeHexString, sanitizeUrl, sanitizeETHThreshold } from "../../common/cli.js";
import { estimateTransactionGas, executeTransaction, get1559Fees, getWeb3SignerSignature } from "../../common/index.js";

const WEB3_SIGNER_PUBLIC_KEY_LENGTH = 64;

const DEFAULT_MAX_FEE_PER_GAS = "100000000000";
const DEFAULT_GAS_ESTIMATION_PERCENTILE = 10;

export type Config = {
  senderAddress: string;
  destinationAddress: string;
  threshold: string;
  blockchainRpcUrl: string;
  web3SignerUrl: string;
  web3SignerPublicKey: string;
  maxFeePerGas: bigint;
  gasEstimationPercentile: number;
  dryRun: boolean;
};

type FlagOutput = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [name: string]: any;
};

export function calculateRewards(balance: bigint): bigint {
  const oneEth = ethers.parseEther("1");

  if (balance < oneEth) {
    return 0n;
  }

  const quotient = (balance - oneEth) / oneEth;
  const flooredBalance = quotient * oneEth;
  return flooredBalance;
}

function validateFlags(flags: FlagOutput): Config {
  const {
    senderAddress,
    destinationAddress,
    threshold,
    blockchainRpcUrl,
    web3SignerUrl,
    web3SignerPublicKey,
    dryRun,
    maxFeePerGas: maxFeePerGasArg,
    gasEstimationPercentile,
  } = flags;

  const requiredFlags = [
    "senderAddress",
    "destinationAddress",
    "threshold",
    "blockchainRpcUrl",
    "web3SignerUrl",
    "web3SignerPublicKey",
  ];

  for (const flagName of requiredFlags) {
    if (!flags[flagName]) {
      throw new Error(`Missing required flag: ${flagName}`);
    }
  }

  let maxFeePerGas: bigint;
  try {
    maxFeePerGas = BigInt(maxFeePerGasArg);
    if (maxFeePerGas <= 0n) {
      throw new Error();
    }
  } catch {
    throw new Error(`Invalid value for --max-fee-per-gas: ${maxFeePerGasArg}. Must be a positive integer in wei.`);
  }

  if (gasEstimationPercentile < 0 || gasEstimationPercentile > 100) {
    throw new Error(
      `Invalid value for --gas-estimation-percentile: ${gasEstimationPercentile}. Must be an integer between 0 and 100.`,
    );
  }

  return {
    senderAddress,
    destinationAddress,
    threshold,
    blockchainRpcUrl,
    web3SignerUrl,
    web3SignerPublicKey,
    maxFeePerGas,
    gasEstimationPercentile,
    dryRun,
  };
}

export default class EthTransfer extends Command {
  static examples = [
    // Example 1: Basic usage with required flags
    `<%= config.bin %> <%= command.id %> \\
      --sender-address=0xYourSenderAddress \\
      --destination-address=0xYourDestinationAddress \\
      --threshold=10 \\
      --blockchain-rpc-url=https://mainnet.infura.io/v3/YOUR-PROJECT-ID \\
      --web3-signer-url=http://localhost:8545 \\
      --web3-signer-public-key=0xYourWeb3SignerPublicKey`,

    // Example 2: Including optional flags with custom values
    `<%= config.bin %> <%= command.id %> \\
      --sender-address=0xYourSenderAddress \\
      --destination-address=0xYourDestinationAddress \\
      --threshold=5 \\
      --blockchain-rpc-url=https://mainnet.infura.io/v3/YOUR-PROJECT-ID \\
      --web3-signer-url=http://localhost:8545 \\
      --web3-signer-public-key=0xYourWeb3SignerPublicKey \\
      --max-fee-per-gas=150000000000 \\
      --gas-estimation-percentile=20`,

    // Example 3: Using the dry-run flag to simulate the transaction
    `<%= config.bin %> <%= command.id %> \\
      --sender-address=0xYourSenderAddress \\
      --destination-address=0xYourDestinationAddress \\
      --threshold=1.5 \\
      --blockchain-rpc-url=http://127.0.0.1:8545 \\
      --web3-signer-url=http://127.0.0.1:8546 \\
      --web3-signer-public-key=0xYourWeb3SignerPublicKey \\
      --dry-run`,
  ];

  static flags = {
    "sender-address": Flags.string({
      char: "s",
      description: "Sender address",
      required: true,
      parse: async (input) => sanitizeAddress("sender-address")(input),
      env: "SENDER_ADDRESS",
    }),
    "destination-address": Flags.string({
      char: "d",
      description: "Destination address",
      required: true,
      parse: async (input) => sanitizeAddress("destination-address")(input),
      env: "DESTINATION_ADDRESS",
    }),
    threshold: Flags.string({
      char: "t",
      description: "Balance threshold of Validator address",
      required: true,
      parse: async (input) => sanitizeETHThreshold()(input),
      env: "THRESHOLD",
    }),
    "blockchain-rpc-url": Flags.string({
      description: "Blockchain RPC URL",
      required: true,
      parse: async (input) => sanitizeUrl("blockchain-rpc-url", ["http:", "https:"])(input),
      env: "BLOCKCHAIN_RPC_URL",
    }),
    "web3-signer-url": Flags.string({
      description: "Web3 Signer URL",
      required: true,
      parse: async (input) => sanitizeUrl("web3-signer-url", ["http:", "https:"])(input),
      env: "WEB3_SIGNER_URL",
    }),
    "web3-signer-public-key": Flags.string({
      description: "Web3 Signer Public Key",
      required: true,
      parse: async (input) => sanitizeHexString("web3-signer-public-key", WEB3_SIGNER_PUBLIC_KEY_LENGTH)(input),
      env: "WEB3_SIGNER_PUBLIC_KEY",
    }),
    "dry-run": Flags.boolean({
      description: "Dry run flag",
      required: false,
      default: false,
      env: "DRY_RUN",
    }),
    "max-fee-per-gas": Flags.string({
      description: "MaxFeePerGas in wei",
      required: false,
      default: DEFAULT_MAX_FEE_PER_GAS,
      env: "MAX_FEE_PER_GAS",
    }),
    "gas-estimation-percentile": Flags.integer({
      description: "Gas estimation percentile (0-100)",
      required: false,
      default: DEFAULT_GAS_ESTIMATION_PERCENTILE,
      env: "GAS_ESTIMATION_PERCENTILE",
    }),
  };

  public async run(): Promise<void> {
    const { flags } = await this.parse(EthTransfer);

    const {
      senderAddress,
      destinationAddress,
      threshold,
      blockchainRpcUrl,
      web3SignerUrl,
      web3SignerPublicKey,
      maxFeePerGas,
      gasEstimationPercentile,
      dryRun,
    } = validateFlags(flags);

    const provider = new ethers.JsonRpcProvider(blockchainRpcUrl);

    const [{ chainId }, senderBalance, fees, nonce] = await Promise.all([
      provider.getNetwork(),
      provider.getBalance(senderAddress),
      get1559Fees(provider, maxFeePerGas, gasEstimationPercentile),
      provider.getTransactionCount(senderAddress),
    ]);

    if (senderBalance <= ethers.parseEther(threshold)) {
      this.log(`Sender balance (${ethers.formatEther(senderBalance)} ETH) is less than threshold. No action needed.`);
      return;
    }

    const rewards = calculateRewards(senderBalance);

    if (rewards === 0n) {
      this.log(`No rewards to send.`);
      return;
    }

    const transactionRequest: TransactionLike = {
      to: destinationAddress,
      value: rewards,
      type: 2,
      chainId,
      maxFeePerGas: fees.maxFeePerGas,
      maxPriorityFeePerGas: fees.maxPriorityFeePerGas,
      nonce: nonce,
    };

    const transactionGasLimit = await estimateTransactionGas(provider, {
      ...transactionRequest,
      from: senderAddress,
    } as ethers.TransactionRequest);

    const transaction: TransactionLike = {
      ...transactionRequest,
      gasLimit: transactionGasLimit,
    };

    const signature = await getWeb3SignerSignature(web3SignerUrl, web3SignerPublicKey, transaction);

    if (dryRun) {
      this.log("Dry run enabled: Skipping transaction submission to blockchain.");
      this.log(`Here is the expected rewards: ${ethers.formatEther(rewards)} ETH`);
      return;
    }

    const receipt = await executeTransaction(provider, {
      ...transaction,
      signature,
    });

    if (!receipt) {
      throw new Error(`Transaction receipt not found for this transaction ${JSON.stringify(transaction)}`);
    }

    if (receipt.status === 0) {
      throw new Error(`Transaction reverted. Receipt: ${JSON.stringify(receipt)}`);
    }

    this.log(
      `Transaction succeeded. Rewards sent: ${ethers.formatEther(rewards)} ETH. Receipt: ${JSON.stringify(receipt)}`,
    );
    this.log(`Rewards sent: ${ethers.formatEther(rewards)} ETH`);
  }
}
