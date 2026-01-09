import { task } from "hardhat/config";
import { getTaskCliOrEnvValue } from "../../common/helpers/environmentHelper";

/*
  *******************************************************************************************
  Generates calldata for YieldManager::addL2YieldRecipient

  Must be called by SET_L2_YIELD_RECIPIENT_ROLE

  The calldata can be used to create Safe multisig transactions or execute via other methods.

  1) Run this task with the right params or env vars.
  2) No transactions are executed - only calldata is generated

  -------------------------------------------------------------------------------------------
  Example (Hoodi):
  -------------------------------------------------------------------------------------------
  CUSTOM_PRIVATE_KEY=0000000000000000000000000000000000000000000000000000000000000002 \
  CUSTOM_BLOCKCHAIN_URL=https://0xrpc.io/hoodi \
  npx hardhat addL2YieldRecipient \
    --yield-manager <address> \
    --yield-recipient <address> \
    --network custom
  -------------------------------------------------------------------------------------------

  Env var alternatives (used if CLI params omitted):
    YIELD_MANAGER
    YIELD_RECIPIENT
  *******************************************************************************************
*/

task("addL2YieldRecipient", "Generates calldata for YieldManager::addL2YieldRecipient")
  .addOptionalParam("yieldManager")
  .addOptionalParam("yieldRecipient")
  .setAction(async (taskArgs, hre) => {
    const { ethers } = hre;

    // --- Resolve inputs from CLI or ENV ---
    const yieldManagerAddress = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER");
    const yieldRecipientAddress = getTaskCliOrEnvValue(taskArgs, "yieldRecipient", "YIELD_RECIPIENT");

    // --- Basic required fields check ---
    const missing: string[] = [];
    if (!yieldManagerAddress) missing.push("yieldManager / YIELD_MANAGER");
    if (!yieldRecipientAddress) missing.push("yieldRecipient / YIELD_RECIPIENT");
    if (missing.length) {
      throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
    }

    // --- Log params ---
    console.log("Params:");
    console.log("  yieldManager:", yieldManagerAddress);
    console.log("  yieldRecipient:", yieldRecipientAddress);

    // Define minimal ABI for YieldManager
    const YIELD_MANAGER_ABI = [
      {
        inputs: [{ internalType: "address", name: "_l2YieldRecipient", type: "address" }],
        name: "addL2YieldRecipient",
        outputs: [],
        stateMutability: "nonpayable",
        type: "function",
      },
    ];

    // Generate calldata for Safe tx
    const yieldManagerContract = await ethers.getContractAt(YIELD_MANAGER_ABI, yieldManagerAddress!);
    const addL2YieldRecipientCalldata = yieldManagerContract.interface.encodeFunctionData("addL2YieldRecipient", [
      yieldRecipientAddress!,
    ]);

    // --- Output calldata ---
    console.log("\n" + "=".repeat(80));
    console.log("Transaction Calldata for Safe Multisig:");
    console.log("=".repeat(80));
    console.log("\n1. addL2YieldRecipient");
    console.log("   Target Contract:", yieldManagerAddress);
    console.log("   Function: addL2YieldRecipient(address)");
    console.log("   Calldata:", addL2YieldRecipientCalldata);
    console.log("\n" + "=".repeat(80));
  });
