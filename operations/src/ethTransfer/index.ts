import { ethers, TransactionLike } from "ethers";
import yargs from "yargs";
import { hideBin } from "yargs/helpers";
import { sanitizeAddress, sanitizeHexString, sanitizeUrl, sanitizeETHThreshold } from "./cli";
import { estimateTransactionGas, executeTransaction, get1559Fees, getWeb3SignerSignature } from "../common";
import { Config } from "./types";
import { calculateRewards } from "./utils";

const WEB3_SIGNER_PUBLIC_KEY_LENGTH = 64;

const DEFAULT_MAX_FEE_PER_GAS = "100000000000";
const DEFAULT_GAS_ESTIMATION_PERCENTILE = "10";

const argv = yargs(hideBin(process.argv))
  .option("sender-address", {
    describe: "Sender address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("sender-address"),
  })
  .option("destination-address", {
    describe: "Destination address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("destination-address"),
  })
  .option("threshold", {
    describe: "Balance threshold of Validator address",
    type: "string",
    demandOption: true,
    coerce: sanitizeETHThreshold(),
  })
  .option("blockchain-rpc-url", {
    describe: "Blockchain rpc url",
    type: "string",
    demandOption: true,
    coerce: sanitizeUrl("blockchain-rpc-url", ["http:", "https:"]),
  })
  .option("web3-signer-url", {
    describe: "Web3 Signer URL",
    type: "string",
    demandOption: true,
    coerce: sanitizeUrl("web3-signer-url", ["http:", "https:"]),
  })
  .option("web3-signer-public-key", {
    describe: "Web3 Signer Public Key",
    type: "string",
    demandOption: true,
    coerce: sanitizeHexString("web3-signer-public-key", WEB3_SIGNER_PUBLIC_KEY_LENGTH),
  })
  .option("dry-run", {
    describe: "Dry run flag",
    type: "string",
    demandOption: false,
  })
  .option("max-fee-per-gas", {
    describe: "MaxFeePerGas in wei",
    type: "string",
    default: DEFAULT_MAX_FEE_PER_GAS,
    demandOption: false,
  })
  .option("gas-estimation-percentile", {
    describe: "Gas estimation percentile",
    type: "string",
    default: DEFAULT_GAS_ESTIMATION_PERCENTILE,
    demandOption: false,
  })
  .parseSync();

function getConfig(args: typeof argv): Config {
  const destinationAddress: string = args.destinationAddress;
  const senderAddress = args.senderAddress;
  const threshold = args.threshold;
  const blockchainRpcUrl = args.blockchainRpcUrl;
  const web3SignerUrl = args.web3SignerUrl;
  const web3SignerPublicKey = args.web3SignerPublicKey;
  const dryRunCheck = args.dryRun !== "false";
  const maxFeePerGasArg = args.maxFeePerGas;
  const gasEstimationPercentileArg = args.gasEstimationPercentile;

  return {
    destinationAddress,
    senderAddress,
    threshold,
    blockchainRpcUrl,
    web3SignerUrl,
    web3SignerPublicKey,
    maxFeePerGas: BigInt(maxFeePerGasArg),
    gasEstimationPercentile: parseInt(gasEstimationPercentileArg),
    dryRun: dryRunCheck,
  };
}

const main = async (args: typeof argv) => {
  const {
    destinationAddress,
    senderAddress,
    threshold,
    blockchainRpcUrl,
    web3SignerUrl,
    web3SignerPublicKey,
    maxFeePerGas,
    gasEstimationPercentile,
    dryRun,
  } = getConfig(args);

  const provider = new ethers.JsonRpcProvider(blockchainRpcUrl);

  const [{ chainId }, senderBalance, fees, nonce] = await Promise.all([
    provider.getNetwork(),
    provider.getBalance(senderAddress),
    get1559Fees(provider, maxFeePerGas, gasEstimationPercentile),
    provider.getTransactionCount(senderAddress),
  ]);

  if (senderBalance <= ethers.parseEther(threshold)) {
    console.log(`Sender balance (${ethers.formatEther(senderBalance)} ETH) is less than threshold. No action needed.`);
    return;
  }

  const rewards = calculateRewards(senderBalance);

  if (rewards == 0n) {
    console.log(`No rewards to send.`);
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
    console.log("Dryrun enabled: Skipping transaction submission to blockchain.");
    console.log(`Here is the expected rewards: ${ethers.formatEther(rewards)} ETH`);
    return;
  }

  const receipt = await executeTransaction(provider, { ...transaction, signature });

  if (!receipt) {
    throw new Error(`Transaction receipt not found for this transaction ${JSON.stringify(transaction)}`);
  }

  if (receipt.status == 0) {
    throw new Error(`Transaction reverted. Receipt: ${JSON.stringify(receipt)}`);
  }

  console.log(
    `Transaction succeed. Rewards sent: ${ethers.formatEther(rewards)} ETH. Receipt: ${JSON.stringify(receipt)}`,
  );
  console.log(`Rewards sent: ${ethers.formatEther(rewards)} ETH`);
};

main(argv)
  .then(() => process.exit(0))
  .catch((error) => {
    console.error("The ETH transfer script failed with the following error:", error);
    process.exit(1);
  });
