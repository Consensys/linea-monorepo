/**
 * eth_sendBundle Script
 *
 * A runalone script that uses the eth_sendBundle JSON RPC method to send transaction bundles.
 * Creates a single ETH transfer transaction and submits it to a bundle relay service.
 *
 * Usage:
 * RPC_URL="https://your-eth-rpc-url" \
 * PRIVATE_KEY="0x..." \
 * TO_ADDRESS="0x..." \
 * AMOUNT="0.1" \
 * BUNDLE_URL="https://relay.flashbots.net" \
 * npx hardhat run scripts/testSendBundle/testSendBundle.ts
 *
 * Required Environment Variables:
 * - RPC_URL: Ethereum RPC endpoint for getting chain data
 * - PRIVATE_KEY: Private key for signing transactions (source address derived from this)
 * - TO_ADDRESS: Destination address for ETH transfer
 * - AMOUNT: Amount of ETH to transfer (e.g., "0.1")
 * - BUNDLE_URL: Bundle relay URL (e.g., "https://relay.flashbots.net")
 *
 * Optional Environment Variables:
 * - BLOCK_NUMBER: Target block number (defaults to current + 1)
 *
 * Examples:
 * # Mainnet Flashbots
 * BUNDLE_URL="https://relay.flashbots.net"
 *
 * # Sepolia Flashbots
 * BUNDLE_URL="https://relay-sepolia.flashbots.net"
 */

import { ethers } from "ethers";
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

    const txParams = {
      to: toAddress,
      value: amountWei,
      gasLimit,
      nonce,
      chainId,
      type: 2, // EIP-1559 transaction
      maxFeePerGas,
      maxPriorityFeePerGas,
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

    console.log("Sending bundle:", {
      bundleParams,
      bundleUrl: this.bundleUrl,
      targetBlock: blockNumber,
      currentBlock: currentBlockNumber,
    });

    const requestBody = {
      jsonrpc: "2.0",
      id: 1,
      method: "eth_sendBundle",
      params: [bundleParams],
    };

    const response = await fetch(this.bundleUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result: BundleResponse = await response.json();

    console.log("Bundle response:", result);

    if (result.error) {
      throw new Error(`Bundle error: ${result.error.code} - ${result.error.message}`);
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

    const bundleSender = new BundleSender(rpcUrl, privateKey, bundleUrl);

    const signerInfo = await bundleSender.getSignerInfo();
    const fromAddress = signerInfo.address;

    console.log("Configuration:", {
      rpcUrl: rpcUrl.replace(/\/\/.*@/, "//***:***@"), // Hide credentials in URL
      fromAddress,
      toAddress,
      amount,
      bundleUrl,
      targetBlockNumber: targetBlockNumber || "auto (current + 1)",
    });

    console.log("\n=== Creating Transaction ===");
    const signedTx = await bundleSender.createEthTransferTransaction(fromAddress, toAddress, amount);

    console.log("\n=== Sending Bundle ===");
    const result = await bundleSender.sendBundle([signedTx], targetBlockNumber);

    console.log("\n=== Success ===");
    console.log("Bundle hash:", result.result?.bundleHash);
    if (result.result?.smart) {
      console.log("Smart bundle:", result.result.smart);
    }
    console.log("\nBundle submitted successfully! Monitor the target block to see if it was included.");
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
