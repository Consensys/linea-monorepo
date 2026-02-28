import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";
import { generateRoleAssignments, buildVendorInitializationData } from "../../../../common/helpers/index.js";
import { LIDO_DASHBOARD_OPERATIONAL_ROLES } from "../../../../common/constants/index.js";

interface TaskArgs {
  yieldManager?: string;
  yieldProvider?: string;
  nodeOperator?: string;
  securityCouncil?: string;
  nodeOperatorFee?: string;
  confirmExpiry?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs) => {
  const yieldManager = getTaskCliOrEnvValue(taskArgs, "yieldManager", "YIELD_MANAGER_ADDRESS");
  const yieldProvider = getTaskCliOrEnvValue(taskArgs, "yieldProvider", "YIELD_PROVIDER_ADDRESS");
  const nodeOperator = getTaskCliOrEnvValue(taskArgs, "nodeOperator", "NODE_OPERATOR");
  const securityCouncil = getTaskCliOrEnvValue(taskArgs, "securityCouncil", "L1_SECURITY_COUNCIL");
  const nodeOperatorFeeRaw = getTaskCliOrEnvValue(taskArgs, "nodeOperatorFee", "NODE_OPERATOR_FEE");
  const confirmExpiryRaw = getTaskCliOrEnvValue(taskArgs, "confirmExpiry", "CONFIRM_EXPIRY");

  const missing: string[] = [];
  if (!yieldManager) missing.push("yieldManager / YIELD_MANAGER_ADDRESS");
  if (!yieldProvider) missing.push("yieldProvider / YIELD_PROVIDER_ADDRESS");
  if (!nodeOperator) missing.push("nodeOperator / NODE_OPERATOR");
  if (!securityCouncil) missing.push("securityCouncil / L1_SECURITY_COUNCIL");
  if (missing.length) {
    throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
  }

  const nodeOperatorFee = nodeOperatorFeeRaw ? BigInt(nodeOperatorFeeRaw) : 0n;
  const confirmExpiry = confirmExpiryRaw ? BigInt(confirmExpiryRaw) : 0n;

  console.log("Params:");
  console.log("  yieldManager:", yieldManager);
  console.log("  yieldProvider:", yieldProvider);
  console.log("  nodeOperator:", nodeOperator);
  console.log("  securityCouncil:", securityCouncil);
  console.log("  nodeOperatorFee:", nodeOperatorFee.toString());
  console.log("  confirmExpiry:", confirmExpiry.toString());

  const yieldProviderInitData = buildVendorInitializationData({
    defaultAdmin: securityCouncil!,
    nodeOperator: nodeOperator!,
    nodeOperatorManager: securityCouncil!,
    nodeOperatorFeeBP: nodeOperatorFee,
    confirmExpiry,
    roleAssignments: generateRoleAssignments(LIDO_DASHBOARD_OPERATIONAL_ROLES, yieldManager!, []),
  });

  console.log("\n" + "=".repeat(80));
  console.log("Parameters for YieldManager.addYieldProvider:");
  console.log("=".repeat(80));
  console.log("\nFunction: addYieldProvider(address _yieldProvider, bytes _initializationData)");
  console.log("\nParameters:");
  console.log("  _yieldProvider:", yieldProvider);
  console.log("  _initializationData:", yieldProviderInitData);
  console.log("\n" + "=".repeat(80));
};

export default action;
