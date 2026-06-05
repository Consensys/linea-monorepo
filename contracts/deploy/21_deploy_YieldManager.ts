import { YieldManager } from "contracts/typechain-types";
import { network as hardhatNetwork } from "hardhat";

import {
  YIELD_MANAGER_INITIALIZE_SIGNATURE,
  YIELD_MANAGER_OPERATOR_ROLES,
  YIELD_MANAGER_PAUSE_TYPES_ROLES,
  YIELD_MANAGER_SECURITY_COUNCIL_ROLES,
  YIELD_MANAGER_UNPAUSE_TYPES_ROLES,
} from "../common/constants";
import {
  generateRoleAssignments,
  getEnvVarOrDefault,
  getRequiredEnvVar,
  requireAddressFromRegistryOrEnv,
  LogContractDeployment,
  tryVerifyContractWithConstructorArgs,
} from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployFromFactory, deployUpgradableFromFactoryWithConstructorArgs } from "../scripts/hardhat/utils";
import { GI_FIRST_VALIDATOR, GI_PENDING_PARTIAL_WITHDRAWALS_ROOT } from "../test/hardhat/common/constants";
import { YieldManagerInitializationData } from "../test/hardhat/yield/helpers/types";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

// Deploys YieldManager, ValidatorContainerProofVerifier and LidoStVaultYieldProviderFactory
const func = withSignerUiSession("21_deploy_YieldManager.ts", async function () {
  const signer = await getUiSigner();

  const contractName = "YieldManager";

  // YieldManager DEPLOYED AS UPGRADEABLE PROXY
  const lineaRollupAddress = requireAddressFromRegistryOrEnv(networkName, "LineaRollup", "LINEA_ROLLUP_ADDRESS");
  const lineaRollupSecurityCouncil = requireAddressFromRegistryOrEnv(
    networkName,
    "L1_SECURITY_COUNCIL",
    "L1_SECURITY_COUNCIL",
  );
  const nativeYieldAutomationServiceAddress = requireAddressFromRegistryOrEnv(
    networkName,
    "NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS",
    "NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS",
  );
  const vaultHub = requireAddressFromRegistryOrEnv(networkName, "VAULT_HUB", "VAULT_HUB");
  const vaultFactory = requireAddressFromRegistryOrEnv(networkName, "VAULT_FACTORY", "VAULT_FACTORY");
  const steth = requireAddressFromRegistryOrEnv(networkName, "STETH", "STETH");
  const initialMinimumWithdrawalReservePercentageBps = parseInt(
    getRequiredEnvVar("MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS"),
  );
  const initialTargetWithdrawalReservePercentageBps = parseInt(
    getRequiredEnvVar("TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS"),
  );
  const initialMinimumWithdrawalReserveAmount = BigInt(getRequiredEnvVar("MINIMUM_WITHDRAWAL_RESERVE_AMOUNT"));
  const initialTargetWithdrawalReserveAmount = BigInt(getRequiredEnvVar("TARGET_WITHDRAWAL_RESERVE_AMOUNT"));
  const gIFirstValidator = getEnvVarOrDefault("GI_FIRST_VALIDATOR", GI_FIRST_VALIDATOR);
  const gIPendingPartialWithdrawalsRoot = getEnvVarOrDefault(
    "GI_PENDING_PARTIAL_WITHDRAWALS_ROOT",
    GI_PENDING_PARTIAL_WITHDRAWALS_ROOT,
  );
  const verifierAdmin = getEnvVarOrDefault("VALIDATOR_CONTAINER_PROOF_VERIFIER_ADMIN", lineaRollupSecurityCouncil);

  const securityCouncilRoles = generateRoleAssignments(
    YIELD_MANAGER_SECURITY_COUNCIL_ROLES,
    lineaRollupSecurityCouncil,
    [],
  );
  const automationServiceRoles = generateRoleAssignments(
    YIELD_MANAGER_OPERATOR_ROLES,
    nativeYieldAutomationServiceAddress,
    [],
  );
  const roleAddresses = [...securityCouncilRoles, ...automationServiceRoles];

  const pauseTypeRoles = getEnvVarOrDefault("YIELD_MANAGER_PAUSE_TYPES_ROLES", YIELD_MANAGER_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault("YIELD_MANAGER_UNPAUSE_TYPES_ROLES", YIELD_MANAGER_UNPAUSE_TYPES_ROLES);

  /********************************************************************
   *                          YieldManager                            *
   ********************************************************************/
  const yieldManagerInitData: YieldManagerInitializationData = {
    pauseTypeRoles: pauseTypeRoles,
    unpauseTypeRoles: unpauseTypeRoles,
    roleAddresses: roleAddresses,
    initialL2YieldRecipients: [],
    defaultAdmin: lineaRollupSecurityCouncil,
    initialMinimumWithdrawalReservePercentageBps: initialMinimumWithdrawalReservePercentageBps,
    initialTargetWithdrawalReservePercentageBps: initialTargetWithdrawalReservePercentageBps,
    initialMinimumWithdrawalReserveAmount: initialMinimumWithdrawalReserveAmount,
    initialTargetWithdrawalReserveAmount: initialTargetWithdrawalReserveAmount,
  };

  const yieldManager = (await deployUpgradableFromFactoryWithConstructorArgs(
    "YieldManager",
    [lineaRollupAddress],
    [yieldManagerInitData],
    {
      initializer: YIELD_MANAGER_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor", "incorrect-initializer-order", "state-variable-immutable", "delegatecall"],
    },
  )) as unknown as YieldManager;

  await LogContractDeployment(contractName, yieldManager);
  const yieldManagerAddress = await yieldManager.getAddress();
  await tryVerifyContractWithConstructorArgs(yieldManagerAddress, "src/yield/YieldManager.sol:YieldManager", [
    lineaRollupAddress,
  ]);

  /********************************************************************
   *                ValidatorContainerProofVerifier                   *
   ********************************************************************/
  const verifier = await deployFromFactory(
    "ValidatorContainerProofVerifier",
    signer,
    verifierAdmin,
    gIFirstValidator,
    gIPendingPartialWithdrawalsRoot,
  );
  await LogContractDeployment("ValidatorContainerProofVerifier", verifier);
  const verifierAddress = await verifier.getAddress();
  await tryVerifyContractWithConstructorArgs(
    verifierAddress,
    "src/yield/libs/ValidatorContainerProofVerifier.sol:ValidatorContainerProofVerifier",
    [verifierAdmin, gIFirstValidator, gIPendingPartialWithdrawalsRoot],
  );

  /********************************************************************
   *                LidoStVaultYieldProviderFactory                   *
   ********************************************************************/
  const factory = await deployFromFactory(
    "LidoStVaultYieldProviderFactory",
    signer,
    lineaRollupAddress,
    yieldManagerAddress,
    vaultHub,
    vaultFactory,
    steth,
    verifierAddress,
  );
  await LogContractDeployment("LidoStVaultYieldProviderFactory", factory);
  const factoryAddress = await factory.getAddress();
  await tryVerifyContractWithConstructorArgs(
    factoryAddress,
    "src/yield/LidoStVaultYieldProviderFactory.sol:LidoStVaultYieldProviderFactory",
    [lineaRollupAddress, yieldManagerAddress, vaultHub, vaultFactory, steth, verifierAddress],
  );

  /********************************************************************
   *                    LidoStVaultYieldProvider                      *
   ********************************************************************/
  const factoryContract = await ethers.getContractAt("LidoStVaultYieldProviderFactory", factoryAddress, signer);
  const yieldProvider = await factoryContract.createLidoStVaultYieldProvider.staticCall();
  const createYieldProviderTx = await factoryContract.createLidoStVaultYieldProvider();
  await createYieldProviderTx.wait(5);
  console.log("Created LidoStVaultYieldProvider at ", yieldProvider);
  await tryVerifyContractWithConstructorArgs(
    yieldProvider,
    "src/yield/LidoStVaultYieldProvider.sol:LidoStVaultYieldProvider",
    [lineaRollupAddress, yieldManagerAddress, vaultHub, vaultFactory, steth, verifierAddress],
  );
});

export default deployScript(func, { tags: ["YieldManager"] });
