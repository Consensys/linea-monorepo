import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployFromFactory, deployUpgradableFromFactoryWithConstructorArgs } from "../scripts/hardhat/utils";
import {
  generateRoleAssignments,
  getEnvVarOrDefault,
  getRequiredEnvVar,
  getDeployedContractAddress,
  LogContractDeployment,
  tryVerifyContractWithConstructorArgs,
} from "../common/helpers";
import {
  DEAD_ADDRESS,
  YIELD_MANAGER_INITIALIZE_SIGNATURE,
  YIELD_MANAGER_OPERATOR_ROLES,
  YIELD_MANAGER_PAUSE_TYPES_ROLES,
  YIELD_MANAGER_SECURITY_COUNCIL_ROLES,
  YIELD_MANAGER_UNPAUSE_TYPES_ROLES,
} from "../common/constants";
import { YieldManagerInitializationData } from "contracts/test/yield/helpers";
import { YieldManager } from "contracts/typechain-types";
import { GI_FIRST_VALIDATOR_CURR, GI_FIRST_VALIDATOR_PREV, PIVOT_SLOT } from "contracts/test/common/constants";

// Deploys YieldManager, ValidatorContainerProofVerifier and LidoStVaultYieldProviderFactory
const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, ethers, getNamedAccounts } = hre;

  const contractName = "YieldManager";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

  // YieldManager DEPLOYED AS UPGRADEABLE PROXY
  const lineaRollupAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");
  const lineaRollupSecurityCouncil = getRequiredEnvVar("LINEA_ROLLUP_SECURITY_COUNCIL");
  const nativeYieldAutomationServiceAddress = getRequiredEnvVar("NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS");
  const vaultHub = getRequiredEnvVar("VAULT_HUB");
  const vaultFactory = getRequiredEnvVar("VAULT_FACTORY");
  const steth = getRequiredEnvVar("STETH");
  const initialMinimumWithdrawalReservePercentageBps = parseInt(
    getRequiredEnvVar("MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS"),
  );
  const initialTargetWithdrawalReservePercentageBps = parseInt(
    getRequiredEnvVar("TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS"),
  );
  const initialMinimumWithdrawalReserveAmount = BigInt(getRequiredEnvVar("MINIMUM_WITHDRAWAL_RESERVE_AMOUNT"));
  // initialTargetWithdrawalReserveAmount must > initialMinimumWithdrawalReserveAmount
  const initialTargetWithdrawalReserveAmount = BigInt(getRequiredEnvVar("TARGET_WITHDRAWAL_RESERVE_AMOUNT"));
  const gIFirstValidatorPrev = getEnvVarOrDefault("GI_FIRST_VALIDATOR_PREV", GI_FIRST_VALIDATOR_PREV);
  const gIFirstValidatorCurr = getEnvVarOrDefault("GI_FIRST_VALIDATOR_CURR", GI_FIRST_VALIDATOR_CURR);
  const pivotSlot = getEnvVarOrDefault("PIVOT_SLOT", PIVOT_SLOT);

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

  if (!existingContractAddress) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  /********************************************************************
   *                          YieldManager                            *
   ********************************************************************/
  const yieldManagerInitData: YieldManagerInitializationData = {
    pauseTypeRoles: pauseTypeRoles,
    unpauseTypeRoles: unpauseTypeRoles,
    roleAddresses: roleAddresses,
    initialL2YieldRecipients: [DEAD_ADDRESS],
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
  await tryVerifyContractWithConstructorArgs(yieldManagerAddress, "contracts/yield/YieldManager.sol:YieldManager", [
    lineaRollupAddress,
  ]);

  /********************************************************************
   *                ValidatorContainerProofVerifier                   *
   ********************************************************************/
  const provider = ethers.provider;
  const verifier = await deployFromFactory(
    "ValidatorContainerProofVerifier",
    provider,
    gIFirstValidatorPrev,
    gIFirstValidatorCurr,
    pivotSlot,
  );
  await LogContractDeployment("ValidatorContainerProofVerifier", verifier);
  const verifierAddress = await verifier.getAddress();
  await tryVerifyContractWithConstructorArgs(
    verifierAddress,
    "contracts/yield/libs/ValidatorContainerProofVerifier.sol:ValidatorContainerProofVerifier",
    [gIFirstValidatorPrev, gIFirstValidatorCurr, pivotSlot],
  );

  /********************************************************************
   *                LidoStVaultYieldProviderFactory                   *
   ********************************************************************/
  const factory = await deployFromFactory(
    "LidoStVaultYieldProviderFactory",
    provider,
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
    "contracts/yield/LidoStVaultYieldProviderFactory.sol:LidoStVaultYieldProviderFactory",
    [lineaRollupAddress, yieldManagerAddress, vaultHub, vaultFactory, steth, verifierAddress],
  );

  /********************************************************************
   *                    LidoStVaultYieldProvider                      *
   ********************************************************************/
  const { deployer } = await getNamedAccounts();
  const signer = await ethers.getSigner(deployer);
  const factoryContract = await ethers.getContractAt("LidoStVaultYieldProviderFactory", factoryAddress, signer);
  const yieldProvider = await factoryContract.createLidoStVaultYieldProvider.staticCall();
  const createYieldProviderTx = await factoryContract.createLidoStVaultYieldProvider();
  await createYieldProviderTx.wait();
  console.log("Created LidoStVaultYieldProvider at ", yieldProvider);
  await tryVerifyContractWithConstructorArgs(
    yieldProvider,
    "contracts/yield/LidoStVaultYieldProvider.sol:LidoStVaultYieldProvider",
    [lineaRollupAddress, yieldManagerAddress, vaultHub, vaultFactory, steth, verifierAddress],
  );
};

export default func;
func.tags = ["YieldManager"];
