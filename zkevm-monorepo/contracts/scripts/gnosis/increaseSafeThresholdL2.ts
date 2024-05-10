import Safe, { EthersAdapter } from "@safe-global/protocol-kit";
import { ethers } from "ethers";
import { requireEnv } from "../hardhat/utils";
import { get1559Fees } from "../utils";

const main = async () => {
  const SAFE_ADDRESS = requireEnv("SAFE_ADDRESS");
  const RPC_URL = requireEnv("BLOCKCHAIN_NODE");
  const SAFE_OWNER1_PRIVATE_KEY = requireEnv("SAFE_OWNER1_PRIVATE_KEY");
  const SAFE_NEW_THRESHOLD = requireEnv("SAFE_NEW_THRESHOLD");

  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const signer = new ethers.Wallet(SAFE_OWNER1_PRIVATE_KEY, provider);

  const ethAdapter = new EthersAdapter({ ethers, signerOrProvider: signer });

  const safeSdk = await Safe.create({ ethAdapter, safeAddress: SAFE_ADDRESS });

  const chainId = await ethAdapter.getChainId();
  console.log(`ChainId: ${chainId}`);
  const eip1559Fees = await get1559Fees(provider);

  // Display initial threshold & owners
  const threshold = await safeSdk.getThreshold();
  console.log("Safe's current threshold: ", threshold);

  const currentOwners = await safeSdk.getOwners();
  console.log("Safe's owners :", currentOwners);

  // Threshold change
  const safeTransaction = await safeSdk.createChangeThresholdTx(parseInt(SAFE_NEW_THRESHOLD));
  const txResponse = await safeSdk.executeTransaction(safeTransaction, {
    maxFeePerGas: eip1559Fees.maxFeePerGas?.toString(),
    maxPriorityFeePerGas: eip1559Fees.maxPriorityFeePerGas?.toString(),
  });
  await txResponse.transactionResponse?.wait();

  // Display new threshold
  const newThreshold = await safeSdk.getThreshold();
  console.log("Safe's new threshold: ", newThreshold);
};

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
