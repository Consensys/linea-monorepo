import { task } from "hardhat/config";
import { getTaskCliOrEnvValue } from "../../common/helpers/environmentHelper";
import { generateRoleAssignments } from "../../common/helpers/roles";
import { buildVendorInitializationData } from "../../common/helpers/buildVendorInitializationData";
import { LIDO_DASHBOARD_OPERATIONAL_ROLES } from "../../common/constants";
import { ONE_ETHER } from "../../common/constants/general";

/*
  *******************************************************************************************
  Generates calldata for adding and configuring a new LidoStVaultYieldProvider.

  This script generates calldata for two transactions:
  1) transferFundsForNativeYield - Transfer funds from LineaRollup to YieldManager
  2) addYieldProvider - Add the yield provider to YieldManager with initialization data

  The calldata can be used to create Safe multisig transactions or execute via other methods.

  1) Prerequisite - 14_deploy_YieldManager.ts script run
  2) Run this task with the right params or env vars.
  3) No transactions are executed - only calldata is generated

  -------------------------------------------------------------------------------------------
  Example (Hoodi):
  -------------------------------------------------------------------------------------------
  CUSTOM_PRIVATE_KEY=0000000000000000000000000000000000000000000000000000000000000002 \
  CUSTOM_BLOCKCHAIN_URL=https://0xrpc.io/hoodi \
  npx hardhat addLidoStVaultYieldProvider \
    --yield-manager <address> \
    --yield-provider <address> \
    --linea-rollup <address> \
    --node-operator <address> \
    --security-council <address> \
    --automation-service-address <address> \
    --node-operator-fee <uint256> \
    --confirm-expiry <uint256> \
    --network custom
  -------------------------------------------------------------------------------------------

  Env var alternatives (used if CLI params omitted):
    YIELD_MANAGER
    YIELD_PROVIDER
    NODE_OPERATOR
    SECURITY_COUNCIL
    AUTOMATION_SERVICE_ADDRESS
    NODE_OPERATOR_FEE
    CONFIRM_EXPIRY
    LINEA_ROLLUP_ADDRESS
  *******************************************************************************************
*/
task("addLidoStVaultYieldProvider", "Generates calldata for adding and configuring a new LidoStVaultYieldProvider")
  .addOptionalParam("yieldManager")
  .addOptionalParam("yieldProvider")
  .addOptionalParam("nodeOperator")
  .addOptionalParam("securityCouncil")
  .addOptionalParam("automationServiceAddress")
  .addOptionalParam("nodeOperatorFee")
  .addOptionalParam("confirmExpiry")
  .addOptionalParam("lineaRollup")
  .setAction(async (taskArgs, hre) => {
    const { ethers, deployments } = hre;
    const { get } = deployments;

    // --- Resolve inputs from CLI or ENV (with sensible fallbacks to deployments) ---
    let yieldManager = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER");
    const yieldProvider = getTaskCliOrEnvValue(taskArgs, "yieldProvider", "YIELD_PROVIDER");
    const nodeOperator = getTaskCliOrEnvValue(taskArgs, "nodeOperator", "NODE_OPERATOR");
    const securityCouncil = getTaskCliOrEnvValue(taskArgs, "securityCouncil", "SECURITY_COUNCIL");
    const nodeOperatorFeeRaw = getTaskCliOrEnvValue(taskArgs, "nodeOperatorFee", "NODE_OPERATOR_FEE");
    const confirmExpiryRaw = getTaskCliOrEnvValue(taskArgs, "confirmExpiry", "CONFIRM_EXPIRY");
    const lineaRollup = getTaskCliOrEnvValue(taskArgs, "lineaRollup", "LINEA_ROLLUP_ADDRESS");

    // --- Use address from artifacts ---
    if (yieldManager === undefined) {
      yieldManager = (await get("YieldManager")).address;
    }

    // --- Basic required fields check (adjust as needed) ---
    const missing: string[] = [];
    if (!yieldProvider) missing.push("yieldProvider / YIELD_PROVIDER");
    if (!nodeOperator) missing.push("nodeOperator / NODE_OPERATOR");
    if (!securityCouncil) missing.push("securityCouncil / SECURITY_COUNCIL");
    if (!lineaRollup) missing.push("lineaRollup / LINEA_ROLLUP_ADDRESS");
    if (missing.length) {
      throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
    }

    // --- Parse numeric params ---
    const nodeOperatorFee = nodeOperatorFeeRaw ? BigInt(nodeOperatorFeeRaw) : 0n;
    const confirmExpiry = confirmExpiryRaw ? BigInt(confirmExpiryRaw) : 0n;

    // --- Log params ---
    console.log("Params:");
    console.log("  yieldManager:", yieldManager);
    console.log("  yieldProvider:", yieldProvider);
    console.log("  nodeOperator:", nodeOperator);
    console.log("  securityCouncil:", securityCouncil);
    console.log("  nodeOperatorFee:", nodeOperatorFee.toString());
    console.log("  confirmExpiry:", confirmExpiry.toString());
    console.log("  lineaRollup:", lineaRollup);

    /********************************************************************
     *                Below here requires Security Council              *
     ********************************************************************/
    // Generate calldata for Safe tx

    const lineaRollupContract = await ethers.getContractAt("LineaRollup", lineaRollup!);
    const transferFundsCalldata = lineaRollupContract.interface.encodeFunctionData("transferFundsForNativeYield", [
      ONE_ETHER,
    ]);

    // --- Add YieldProvider ---
    const yieldManagerContract = await ethers.getContractAt("YieldManager", yieldManager);
    const yieldProviderInitData = buildVendorInitializationData({
      defaultAdmin: securityCouncil,
      nodeOperator,
      nodeOperatorManager: securityCouncil,
      nodeOperatorFeeBP: nodeOperatorFee,
      confirmExpiry,
      roleAssignments: generateRoleAssignments(LIDO_DASHBOARD_OPERATIONAL_ROLES, yieldManager, []),
    });
    const addYieldProviderCalldata = yieldManagerContract.interface.encodeFunctionData("addYieldProvider", [
      yieldProvider!,
      yieldProviderInitData,
    ]);

    // --- Output calldata ---
    console.log("\n" + "=".repeat(80));
    console.log("Transaction Calldata for Safe Multisig:");
    console.log("=".repeat(80));
    console.log("\n1. transferFundsForNativeYield");
    console.log("   Target Contract:", lineaRollup);
    console.log("   Function: transferFundsForNativeYield(uint256)");
    console.log("   Calldata:", transferFundsCalldata);
    console.log("\n2. addYieldProvider");
    console.log("   Target Contract:", yieldManager);
    console.log("   Function: addYieldProvider(address,bytes)");
    console.log("   Calldata:", addYieldProviderCalldata);
    console.log("\n" + "=".repeat(80));
    console.log("\n⚠️  IMPORTANT: After executing the transactions, verify the LidoStVaultYieldProvider contract:");
    console.log("\n   npx hardhat verify --network <NETWORK> <YIELD_PROVIDER_ADDRESS> \\");
    console.log("     <L1_MESSAGE_SERVICE_ADDRESS> \\");
    console.log("     <YIELD_MANAGER_ADDRESS> \\");
    console.log("     <VAULT_HUB_ADDRESS> \\");
    console.log("     <VAULT_FACTORY_ADDRESS> \\");
    console.log("     <STETH_ADDRESS> \\");
    console.log("     <VALIDATOR_CONTAINER_PROOF_VERIFIER_ADDRESS>");
    console.log("\n   Example:");
    console.log(`   npx hardhat verify --network sepolia ${yieldProvider} \\`);
    console.log("     <L1_MESSAGE_SERVICE_ADDRESS> \\");
    console.log(`     ${yieldManager} \\`);
    console.log("     <VAULT_HUB_ADDRESS> \\");
    console.log("     <VAULT_FACTORY_ADDRESS> \\");
    console.log("     <STETH_ADDRESS> \\");
    console.log("     <VALIDATOR_CONTAINER_PROOF_VERIFIER_ADDRESS>");
    console.log("\n" + "=".repeat(80));
  });
