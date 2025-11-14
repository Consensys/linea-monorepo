import { task } from "hardhat/config";
import { getTaskCliOrEnvValue } from "../../common/helpers/environmentHelper";
import { generateRoleAssignments } from "contracts/common/helpers";
import { LIDO_DASHBOARD_OPERATIONAL_ROLES } from "contracts/common/constants";
import { buildVendorInitializationData } from "contracts/test/yield/helpers";

/*
  *******************************************************************************************
  Creates and configures a new LidoStVaultYieldProvider.

  1) Prerequisite - 14_deploy_YieldManager.ts script run
  2) Run this task with the right params or env vars.
  3) Signer must have SET_YIELD_PROVIDER_ROLE role for YieldManager

  -------------------------------------------------------------------------------------------
  Example (Hoodi):
  -------------------------------------------------------------------------------------------
  CUSTOM_PRIVATE_KEY=<key> \
  CUSTOM_BLOCKCHAIN_URL=https://0xrpc.io/hoodi \
  npx hardhat addLidoStVaultYieldProvider \
    --yield-provider-factory <address> \
    --yield-manager <address> \
    --node-operator <address> \
    --security-council <address> \
    --automation-service-address <address> \
    --node-operator-fee <uint256> \
    --confirm-expiry <uint256> \
    --network custom
  -------------------------------------------------------------------------------------------

  Env var alternatives (used if CLI params omitted):
    YIELD_PROVIDER_FACTORY
    YIELD_MANAGER
    NODE_OPERATOR
    SECURITY_COUNCIL
    AUTOMATION_SERVICE_ADDRESS
    NODE_OPERATOR_FEE
    CONFIRM_EXPIRY
  *******************************************************************************************
*/
task("addLidoStVaultYieldProvider", "Creates and configures a new LidoStVaultYieldProvider")
  .addOptionalParam("yieldProviderFactory")
  .addOptionalParam("yieldManager")
  .addOptionalParam("nodeOperator")
  .addOptionalParam("securityCouncil")
  .addOptionalParam("automationServiceAddress")
  .addOptionalParam("nodeOperatorFee")
  .addOptionalParam("confirmExpiry")
  .setAction(async (taskArgs, hre) => {
    const { ethers, deployments, getNamedAccounts } = hre;
    const { get } = deployments;
    const { deployer } = await getNamedAccounts();
    const signer = await ethers.getSigner(deployer);

    // --- Resolve inputs from CLI or ENV (with sensible fallbacks to deployments) ---
    let yieldProviderFactory = getTaskCliOrEnvValue(taskArgs, "yieldProviderFactory", "YIELD_PROVIDER_FACTORY");
    let yieldManager = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER");
    const nodeOperator = getTaskCliOrEnvValue(taskArgs, "nodeOperator", "NODE_OPERATOR");
    const securityCouncil = getTaskCliOrEnvValue(taskArgs, "securityCouncil", "SECURITY_COUNCIL");
    const automationServiceAddress = getTaskCliOrEnvValue(
      taskArgs,
      "automationServiceAddress",
      "AUTOMATION_SERVICE_ADDRESS",
    );
    const nodeOperatorFeeRaw = getTaskCliOrEnvValue(taskArgs, "nodeOperatorFee", "NODE_OPERATOR_FEE");
    const confirmExpiryRaw = getTaskCliOrEnvValue(taskArgs, "confirmExpiry", "CONFIRM_EXPIRY");

    // --- Use address from artifacts ---
    if (yieldProviderFactory === undefined) {
      yieldProviderFactory = (await get("LidoStVaultYieldProviderFactory")).address;
    }
    if (yieldManager === undefined) {
      yieldManager = (await get("YieldManager")).address;
    }

    // --- Basic required fields check (adjust as needed) ---
    const missing: string[] = [];
    if (!nodeOperator) missing.push("nodeOperator / NODE_OPERATOR");
    if (!securityCouncil) missing.push("securityCouncil / SECURITY_COUNCIL");
    if (!automationServiceAddress) missing.push("automationServiceAddress / AUTOMATION_SERVICE_ADDRESS");
    if (missing.length) {
      throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
    }

    // --- Parse numeric params ---
    const nodeOperatorFee = nodeOperatorFeeRaw ? BigInt(nodeOperatorFeeRaw) : 0n;
    const confirmExpiry = confirmExpiryRaw ? BigInt(confirmExpiryRaw) : 0n;

    // --- Log params ---
    console.log("Using factory:", yieldProviderFactory);
    console.log("Params:");
    console.log("  lidoStVaultYieldProviderFactory:", yieldProviderFactory);
    console.log("  yieldManager:", yieldManager);
    console.log("  nodeOperator:", nodeOperator);
    console.log("  securityCouncil:", securityCouncil);
    console.log("  automationServiceAddress:", automationServiceAddress);
    console.log("  nodeOperatorFee:", nodeOperatorFee.toString());
    console.log("  confirmExpiry:", confirmExpiry.toString());

    // --- Create LidoStVaultYieldProvider factory ---
    const factory = await ethers.getContractAt("LidoStVaultYieldProviderFactory", yieldProviderFactory, signer);
    const yieldProvider = await factory.createLidoStVaultYieldProvider.staticCall();
    const createYieldProviderTx = await factory.createLidoStVaultYieldProvider();
    await createYieldProviderTx.wait();
    console.log("Created LidoStVaultYieldProvider at ", yieldProvider);

    // --- Add YieldProvider ---
    const yieldManagerContract = await ethers.getContractAt("YieldManager", yieldManager, signer);
    const yieldProviderInitData = buildVendorInitializationData({
      defaultAdmin: securityCouncil,
      nodeOperator,
      nodeOperatorManager: securityCouncil,
      nodeOperatorFeeBP: nodeOperatorFee,
      confirmExpiry,
      roleAssignments: generateRoleAssignments(LIDO_DASHBOARD_OPERATIONAL_ROLES, automationServiceAddress!, []),
    });
    const addYieldProviderTx = await yieldManagerContract.addYieldProvider(yieldProvider, yieldProviderInitData);
    await addYieldProviderTx.wait();
    console.log(`LidoStVaultYieldProvider address=${yieldProvider} added to YieldManager address=${yieldManager}`);
  });
