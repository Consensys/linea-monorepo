import { task } from "hardhat/config";
import { getTaskCliOrEnvValue } from "../../common/helpers/environmentHelper";

/*
  *******************************************************************************************
  Generates calldata for setting node operator fee recipient for Native Yield system by calling Dashboard::setFeeRecipient

  Must be called by NODE_OPERATOR_MANAGER_ROLE or its role admin for given Dashboard.sol contract

  The calldata can be used to create Safe multisig transactions or execute via other methods.

  1) Run this task with the right params or env vars.
  2) No transactions are executed - only calldata is generated

  -------------------------------------------------------------------------------------------
  Example (Hoodi):
  -------------------------------------------------------------------------------------------
  CUSTOM_PRIVATE_KEY=0000000000000000000000000000000000000000000000000000000000000002 \
  CUSTOM_BLOCKCHAIN_URL=https://0xrpc.io/hoodi \
  npx hardhat setNodeOperatorFeeRecipient \
    --dashboard <address> \
    --fee-recipient <address> \
    --network custom
  -------------------------------------------------------------------------------------------

  Env var alternatives (used if CLI params omitted):
    DASHBOARD
    FEE_RECIPIENT
  *******************************************************************************************
*/

task(
  "setNodeOperatorFeeRecipient",
  "Generates calldata for setting node operator fee recipient for Native Yield system",
)
  .addOptionalParam("dashboard")
  .addOptionalParam("feeRecipient")
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;

    // --- Resolve inputs from CLI or ENV ---
    const dashboardAddress = getTaskCliOrEnvValue(taskArgs, "dashboard", "DASHBOARD");
    const feeRecipientAddress = getTaskCliOrEnvValue(taskArgs, "feeRecipient", "FEE_RECIPIENT");

    // --- Basic required fields check ---
    const missing: string[] = [];
    if (!dashboardAddress) missing.push("dashboard / DASHBOARD");
    if (!feeRecipientAddress) missing.push("feeRecipient / FEE_RECIPIENT");
    if (missing.length) {
      throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
    }

    // --- Log params ---
    console.log("Params:");
    console.log("  dashboard:", dashboardAddress);
    console.log("  feeRecipient:", feeRecipientAddress);

    // Define minimal ABI for Dashboard
    const DASHBOARD_ABI = [
      {
        inputs: [{ internalType: "address", name: "_newFeeRecipient", type: "address" }],
        name: "setFeeRecipient",
        outputs: [],
        stateMutability: "nonpayable",
        type: "function",
      },
    ];

    // Generate calldata for Safe tx
    const dashboardContract = await ethers.getContractAt(DASHBOARD_ABI, dashboardAddress!);
    const setFeeRecipientCalldata = dashboardContract.interface.encodeFunctionData("setFeeRecipient", [
      feeRecipientAddress!,
    ]);

    // --- Output calldata ---
    console.log("\n" + "=".repeat(80));
    console.log("Transaction Calldata for Safe Multisig:");
    console.log("=".repeat(80));
    console.log("\n1. setFeeRecipient");
    console.log("   Target Contract:", dashboardAddress);
    console.log("   Function: setFeeRecipient(address)");
    console.log("   Calldata:", setFeeRecipientCalldata);
    console.log("\n" + "=".repeat(80));
  });
