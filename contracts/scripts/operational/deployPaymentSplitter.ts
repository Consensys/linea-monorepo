import { task } from "hardhat/config";
import { getTaskCliOrEnvValue } from "../../common/helpers/environmentHelper";
import { isAddress } from "ethers";

/*
  *******************************************************************************************
  Deploys PaymentSplitter.sol from OpenZeppelin

  Constructor requires:
  - address[] memory payees - array of payee addresses
  - uint256[] memory shares_ - array of shares for each payee

  Payees and shares should be provided as comma-separated values.
  Both arrays must have the same length and must not be empty.

  -------------------------------------------------------------------------------------------
  Example (Hoodi):
  -------------------------------------------------------------------------------------------
  CUSTOM_PRIVATE_KEY=0000000000000000000000000000000000000000000000000000000000000002 \
  CUSTOM_BLOCKCHAIN_URL=https://0xrpc.io/hoodi \
  npx hardhat deployPaymentSplitter \
    --payees "0x123...,0x456..." \
    --shares "100,200" \
    --network custom
  -------------------------------------------------------------------------------------------

  Env var alternatives (used if CLI params omitted):
    PAYEES (comma-separated addresses)
    SHARES (comma-separated numbers)
  *******************************************************************************************
*/

task("deployPaymentSplitter", "Deploys PaymentSplitter contract from OpenZeppelin")
  .addOptionalParam("payees")
  .addOptionalParam("shares")
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;

    // --- Resolve inputs from CLI or ENV ---
    const payeesRaw = getTaskCliOrEnvValue(taskArgs, "payees", "PAYEES");
    const sharesRaw = getTaskCliOrEnvValue(taskArgs, "shares", "SHARES");

    // --- Basic required fields check ---
    const missing: string[] = [];
    if (!payeesRaw) missing.push("payees / PAYEES");
    if (!sharesRaw) missing.push("shares / SHARES");
    if (missing.length) {
      throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
    }

    // --- Parse comma-separated strings into arrays ---
    const payeesArray = payeesRaw!
      .split(",")
      .map((p) => p.trim())
      .filter((p) => p.length > 0);
    const sharesArray = sharesRaw!
      .split(",")
      .map((s) => s.trim())
      .filter((s) => s.length > 0);

    // --- Validate arrays ---
    if (payeesArray.length === 0) {
      throw new Error("Payees array cannot be empty");
    }
    if (sharesArray.length === 0) {
      throw new Error("Shares array cannot be empty");
    }
    if (payeesArray.length !== sharesArray.length) {
      throw new Error(
        `Payees and shares arrays must have the same length. Got ${payeesArray.length} payees and ${sharesArray.length} shares`,
      );
    }

    // Validate addresses
    for (let i = 0; i < payeesArray.length; i++) {
      if (!isAddress(payeesArray[i]!)) {
        throw new Error(`Invalid address at index ${i}: ${payeesArray[i]}`);
      }
    }

    // Validate and parse shares
    const sharesBigInt: bigint[] = [];
    for (let i = 0; i < sharesArray.length; i++) {
      const shareStr = sharesArray[i]!;
      const shareNum = BigInt(shareStr);
      if (shareNum < 0n) {
        throw new Error(`Share at index ${i} must be non-negative: ${shareStr}`);
      }
      sharesBigInt.push(shareNum);
    }

    // --- Log params ---
    console.log("Params:");
    console.log("  payees:", payeesArray);
    console.log(
      "  shares:",
      sharesBigInt.map((s) => s.toString()),
    );

    // --- Deploy contract ---
    console.log("Deploying PaymentSplitter...");
    const PaymentSplitterFactory = await ethers.getContractFactory("PaymentSplitter");
    const paymentSplitter = await PaymentSplitterFactory.deploy(payeesArray, sharesBigInt);

    console.log("Transaction sent, hash:", paymentSplitter.deploymentTransaction()?.hash);
    await paymentSplitter.waitForDeployment();

    const address = await paymentSplitter.getAddress();
    console.log("PaymentSplitter deployed at:", address);
    console.log("Deployment successful!");
  });
