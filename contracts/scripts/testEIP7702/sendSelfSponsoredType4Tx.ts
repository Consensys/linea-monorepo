import { ethers } from "ethers";
import { requireEnv, checkDelegation, getAccountInfo, createAuthorization, estimateGasFees } from "../utils";
import * as dotenv from "dotenv";

// Self-sponsored EIP-7702 transaction: the same wallet signs the authorization AND sends the transaction.
// Prerequisite - Deploy a contract with NON-VIEW initialize() function, e.g. TestEIP7702Delegation
// Use this contract for TARGET_ADDRESS env

// RPC_URL=<> DEPLOYER_PRIVATE_KEY=<> TARGET_ADDRESS=<> npx hardhat run scripts/testEIP7702/sendSelfSponsoredType4Tx.ts

dotenv.config();

async function main() {
  const rpcUrl = requireEnv("RPC_URL");
  const privateKey = requireEnv("DEPLOYER_PRIVATE_KEY");
  const targetAddress = requireEnv("TARGET_ADDRESS");

  const provider = new ethers.JsonRpcProvider(rpcUrl);
  const signer = new ethers.Wallet(privateKey, provider);

  const signerInfo = await getAccountInfo(provider, signer.address);
  console.log("Signer info:", signerInfo);

  const delegationStatus = await checkDelegation(provider, signer.address);
  console.log("EOA delegation status:", delegationStatus);

  console.log("\n=== SELF-SPONSORED EIP-7702 TRANSACTION ===");
  console.log("Same wallet signs authorization and sends the transaction.");

  // +1 nonce offset: sender nonce is incremented before authorization processing
  const authorization = await createAuthorization(signer, provider, targetAddress, 1);

  const ABI = ["function initialize() external"];
  const delegatedContract = new ethers.Contract(signer, ABI, signer);

  const fees = await estimateGasFees(provider, rpcUrl, signer.address, signer.address);

  const txParams = {
    type: 4,
    authorizationList: [authorization],
    gasLimit: 500000n,
    value: 0n,
    ...(fees.maxFeePerGas != null && { maxFeePerGas: fees.maxFeePerGas }),
    ...(fees.maxPriorityFeePerGas != null && { maxPriorityFeePerGas: fees.maxPriorityFeePerGas }),
  };

  const tx = await delegatedContract["initialize()"](txParams);
  console.log("Self-sponsored transaction sent:", tx.hash);

  const receipt = await tx.wait();
  console.log("Receipt:", receipt);
}

main().catch((error) => {
  console.error("Error:", error);
  process.exit(1);
});
