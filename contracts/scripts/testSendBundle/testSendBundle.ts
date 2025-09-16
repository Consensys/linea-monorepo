/**
 * eth_sendBundle Script
 *
 * A runalone script that uses the eth_sendBundle JSON RPC method to send transaction bundles.
 * Creates a single ETH transfer transaction and submits it to a bundle relay service.
 *
 * Usage:

RPC_URL=https://your-eth-rpc-url \
PRIVATE_KEY=0x... \
TO_ADDRESS="0x... \
AMOUNT=0.001 \
BUNDLE_URL=https://relay-sepolia.flashbots.net \
npx hardhat run scripts/testSendBundle/testSendBundle.ts

 * Required Environment Variables:
 * - RPC_URL: Ethereum RPC endpoint for getting chain data
 * - PRIVATE_KEY: Private key for signing transactions (source address derived from this)
 * - TO_ADDRESS: Destination address for ETH transfer
 * - AMOUNT: Amount of ETH to transfer (e.g., "0.1")
 * - BUNDLE_URL: Bundle relay URL (e.g., "https://relay-sepolia.flashbots.net")
 *
 * Optional Environment Variables:
 * - BLOCK_NUMBER: Target block number (defaults to current + 1, only used for single bundle)
 * - BUNDLE_COUNT: Number of bundles to send to consecutive blocks (defaults to 10)
 */

import { ethers, id } from "ethers";
import { get1559Fees, isLineaChainId, LineaEstimateGasClient } from "../utils";
import * as dotenv from "dotenv";

dotenv.config();

interface BundleParams {
  txs: string[];
  blockNumber: string;
  minTimestamp?: number;
  maxTimestamp?: number;
  revertingTxHashes?: string[];
  replacementUuid?: string;
  builders?: string[];
}

interface BundleResponse {
  jsonrpc: string;
  id: number;
  result?: {
    bundleHash: string;
    smart?: string;
  };
  error?: {
    code: number;
    message: string;
  };
}

interface CallBundleParams {
  txs: string[];
  blockNumber: string;
  stateBlockNumber: string;
  timestamp?: number;
}

interface CallBundleResult {
  bundleGasPrice: string;
  bundleHash: string;
  coinbaseDiff: string;
  ethSentToCoinbase: string;
  gasFees: string;
  results: Array<{
    coinbaseDiff: string;
    ethSentToCoinbase: string;
    fromAddress: string;
    gasFees: string;
    gasPrice: string;
    gasUsed: number;
    toAddress: string;
    txHash: string;
    value: string;
  }>;
  stateBlockNumber: number;
  totalGasUsed: number;
}

interface CallBundleResponse {
  jsonrpc: string;
  id: number;
  result?: CallBundleResult;
  error?: {
    code: number;
    message: string;
  };
}

class BundleSender {
  private provider: ethers.Provider;
  private signer: ethers.Wallet;
  private bundleUrl: string;
  private lineaEstimateGasClient: LineaEstimateGasClient;

  constructor(rpcUrl: string, privateKey: string, bundleUrl: string) {
    this.provider = new ethers.JsonRpcProvider(rpcUrl);
    this.signer = new ethers.Wallet(privateKey, this.provider);
    this.bundleUrl = bundleUrl;
    this.lineaEstimateGasClient = new LineaEstimateGasClient(new URL(rpcUrl), this.signer.address);
  }

  async createEthTransferTransaction(fromAddress: string, toAddress: string, amount: string): Promise<string> {
    const network = await this.provider.getNetwork();
    const chainId = Number(network.chainId);
    const nonce = await this.provider.getTransactionCount(fromAddress);
    const amountWei = ethers.parseEther(amount);

    const gasLimit = 21000n; // Standard ETH transfer gas limit

    let maxFeePerGas: bigint;
    let maxPriorityFeePerGas: bigint;

    if (isLineaChainId(chainId)) {
      const fees = await this.lineaEstimateGasClient.lineaEstimateGas(toAddress);
      maxFeePerGas = fees.maxFeePerGas;
      maxPriorityFeePerGas = fees.maxPriorityFeePerGas;
    } else {
      const fees = await get1559Fees(this.provider);
      maxFeePerGas = fees.maxFeePerGas || fees.gasPrice || 0n;
      maxPriorityFeePerGas = fees.maxPriorityFeePerGas || 0n;
    }

    const gasMultiplierFactor = 10000000n;

    const txParams = {
      to: toAddress,
      value: amountWei,
      gasLimit,
      nonce,
      chainId,
      type: 2, // EIP-1559 transaction
      maxFeePerGas: maxFeePerGas * gasMultiplierFactor,
      maxPriorityFeePerGas: maxPriorityFeePerGas * gasMultiplierFactor,
    };

    console.log("Transaction parameters:", {
      from: fromAddress,
      to: toAddress,
      value: ethers.formatEther(amountWei),
      gasLimit: gasLimit.toString(),
      maxFeePerGas: ethers.formatUnits(maxFeePerGas, "gwei"),
      maxPriorityFeePerGas: ethers.formatUnits(maxPriorityFeePerGas, "gwei"),
      nonce,
      chainId,
    });

    const transaction = await this.signer.signTransaction(txParams);
    return transaction;
  }

  async sendBundle(signedTransactions: string[], targetBlockNumber?: number): Promise<BundleResponse> {
    const currentBlockNumber = await this.provider.getBlockNumber();
    const blockNumber = targetBlockNumber || currentBlockNumber + 1;

    const bundleParams: BundleParams = {
      txs: signedTransactions,
      blockNumber: `0x${blockNumber.toString(16)}`,
    };

    const requestBody = {
      jsonrpc: "2.0",
      id: 1,
      method: "eth_sendBundle",
      params: [bundleParams],
    };

    const requestBodyString = JSON.stringify(requestBody);
    const bundleSignature = await this.generateBundleSignature(requestBodyString);

    const response = await fetch(this.bundleUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Flashbots-Signature": bundleSignature,
      },
      body: requestBodyString,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result: BundleResponse = await response.json();

    if (result.error) {
      throw new Error(`Bundle error: ${result.error.code} - ${result.error.message}`);
    }

    return result;
  }

  async callBundle(signedTransactions: string[], targetBlockNumber?: number): Promise<CallBundleResponse> {
    const currentBlockNumber = await this.provider.getBlockNumber();
    const blockNumber = targetBlockNumber || currentBlockNumber + 1;

    const callBundleParams: CallBundleParams = {
      txs: signedTransactions,
      blockNumber: `0x${blockNumber.toString(16)}`,
      stateBlockNumber: "latest",
    };

    const requestBody = {
      jsonrpc: "2.0",
      id: 1,
      method: "eth_callBundle",
      params: [callBundleParams],
    };

    const requestBodyString = JSON.stringify(requestBody);
    const bundleSignature = await this.generateBundleSignature(requestBodyString);

    const response = await fetch(this.bundleUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Flashbots-Signature": bundleSignature,
      },
      body: requestBodyString,
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result: CallBundleResponse = await response.json();

    if (result.error) {
      throw new Error(`Call bundle error: ${result.error.code} - ${result.error.message}`);
    }

    return result;
  }

  async getSignerInfo(): Promise<{ address: string; balance: string; nonce: number }> {
    const address = this.signer.address;
    const balance = await this.provider.getBalance(address);
    const nonce = await this.provider.getTransactionCount(address);

    return {
      address,
      balance: ethers.formatEther(balance),
      nonce,
    };
  }

  async getCurrentBlockNumber(): Promise<number> {
    return await this.provider.getBlockNumber();
  }

  private async generateBundleSignature(requestBody: string): Promise<string> {
    const signature = await this.signer.signMessage(id(requestBody));
    return `${this.signer.address}:${signature}`;
  }
}

function requireEnv(name: string): string {
  const envVariable = process.env[name];
  if (!envVariable) {
    throw new Error(`Missing ${name} environment variable`);
  }
  return envVariable;
}

async function main() {
  try {
    console.log("=== ETH Bundle Sender ===\n");

    const rpcUrl = requireEnv("RPC_URL");
    const privateKey = requireEnv("PRIVATE_KEY");
    const toAddress = requireEnv("TO_ADDRESS");
    const amount = requireEnv("AMOUNT");
    const bundleUrl = requireEnv("BUNDLE_URL");

    // Optional environment variables
    const targetBlockNumber = process.env.BLOCK_NUMBER ? parseInt(process.env.BLOCK_NUMBER) : undefined;
    const bundleCount = process.env.BUNDLE_COUNT ? parseInt(process.env.BUNDLE_COUNT) : 10;

    const bundleSender = new BundleSender(rpcUrl, privateKey, bundleUrl);

    const signerInfo = await bundleSender.getSignerInfo();
    const fromAddress = signerInfo.address;

    console.log("Configuration:", {
      fromAddress,
      toAddress,
      amount,
      bundleUrl,
      bundleCount,
      singleTarget: targetBlockNumber ? `block ${targetBlockNumber}` : `next ${bundleCount} blocks`,
    });

    console.log("\n=== Creating Transaction ===");
    const signedTx = await bundleSender.createEthTransferTransaction(fromAddress, toAddress, amount);

    if (targetBlockNumber) {
      // Single bundle mode
      console.log("\n=== Simulating Bundle ===");
      const callResult = await bundleSender.callBundle([signedTx], targetBlockNumber);

      if (callResult.result) {
        console.log("Bundle simulation successful:");
        console.log("  Total gas used:", callResult.result.totalGasUsed);
        console.log("  Bundle gas price:", ethers.formatUnits(callResult.result.bundleGasPrice, "gwei"), "gwei");
        console.log("  Gas fees:", ethers.formatEther(callResult.result.gasFees), "ETH");
        console.log("  Coinbase diff:", ethers.formatEther(callResult.result.coinbaseDiff), "ETH");
        console.log("  State block number:", callResult.result.stateBlockNumber);

        console.log("\n=== Sending Single Bundle ===");
        const result = await bundleSender.sendBundle([signedTx], targetBlockNumber);
        console.log("\n=== Success ===");
        console.log("Bundle hash:", result.result?.bundleHash);
        if (result.result?.smart) {
          console.log("Smart bundle:", result.result.smart);
        }
        console.log("\nBundle submitted successfully! Monitor the target block to see if it was included.");
      } else {
        console.log("Bundle simulation failed - skipping submission");
      }
    } else {
      // Multiple bundles mode
      console.log(`\n=== Sending ${bundleCount} Bundles ===`);

      const currentBlock = await bundleSender.getCurrentBlockNumber();
      const targetBlocks = Array.from({ length: bundleCount }, (_, i) => currentBlock + i + 1);

      console.log(`Target blocks: ${targetBlocks.join(", ")}`);
      console.log(`\nSending all bundles concurrently...`);

      // Create promises for all bundle submissions
      const bundlePromises = targetBlocks.map(async (targetBlock) => {
        try {
          // Simulate bundle first
          const callResult = await bundleSender.callBundle([signedTx], targetBlock);

          if (!callResult.result) {
            console.log(`❌ Bundle simulation failed for block ${targetBlock} - skipping submission`);
            return {
              block: targetBlock,
              success: false,
              error: "Bundle simulation failed",
            };
          }

          console.log(
            `📊 Bundle simulation for block ${targetBlock}: ${callResult.result.totalGasUsed} gas, ${ethers.formatEther(callResult.result.gasFees)} ETH fees`,
          );

          const result = await bundleSender.sendBundle([signedTx], targetBlock);
          console.log(`✅ Bundle submitted - Hash: ${result.result?.bundleHash || "unknown"}`);
          return {
            block: targetBlock,
            success: true,
            bundleHash: result.result?.bundleHash || "",
          };
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : String(error);
          console.log(`❌ Bundle failed: ${errorMessage}`);
          return {
            block: targetBlock,
            success: false,
            error: errorMessage,
          };
        }
      });

      // Wait for all bundles to complete
      const results = await Promise.all(bundlePromises);

      // Results summary
      console.log("\n=== Results Summary ===");
      const successful = results.filter((r) => r.success);
      const failed = results.filter((r) => !r.success);

      console.log(`Successful bundles: ${successful.length}/${bundleCount}`);
      console.log(`Failed bundles: ${failed.length}/${bundleCount}`);

      console.log(
        `\nAll bundles submitted! Monitor blocks ${currentBlock + 1} to ${currentBlock + bundleCount} for inclusion.`,
      );
    }
  } catch (error) {
    console.error("\n=== Error ===");
    console.error("Error:", error);
    process.exit(1);
  }
}

if (require.main === module) {
  main()
    .then(() => process.exit(0))
    .catch((error) => {
      console.error(error);
      process.exit(1);
    });
}
