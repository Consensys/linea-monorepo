import { DeployFunction } from "hardhat-deploy/types";
import { getRequiredEnvVar, deployContractFromArtifacts, getInitializerData } from "../common/helpers";
import {
  PAUSE_STATE_DATA_SUBMISSION_ROLE,
  UNPAUSE_STATE_DATA_SUBMISSION_ROLE,
  STATE_DATA_SUBMISSION_PAUSE_TYPE,
} from "contracts/common/constants";
import {
  contractName as LineaRollupV7ContractName,
  abi as LineaRollupV7Abi,
  bytecode as LineaRollupV7Bytecode,
} from "../deployments/bytecode/2026-01-14/LineaRollup.json";
import { HardhatRuntimeEnvironment } from "hardhat/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { ethers, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const signer = await ethers.getSigner(deployer);

  let upgradePauseTypeRoles = [];
  let upgradeUnpauseTypeRoles = [];
  let upgradeRoleAddresses = [];

  const securityCouncilAddress = getRequiredEnvVar("LINEA_ROLLUP_SECURITY_COUNCIL");

  upgradeRoleAddresses = [
    {
      addressWithRole: securityCouncilAddress,
      role: PAUSE_STATE_DATA_SUBMISSION_ROLE,
    },
    {
      addressWithRole: securityCouncilAddress,
      role: UNPAUSE_STATE_DATA_SUBMISSION_ROLE,
    },
  ];

  upgradePauseTypeRoles = [{ pauseType: STATE_DATA_SUBMISSION_PAUSE_TYPE, role: PAUSE_STATE_DATA_SUBMISSION_ROLE }];
  upgradeUnpauseTypeRoles = [{ pauseType: STATE_DATA_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_STATE_DATA_SUBMISSION_ROLE }];

  const proxyAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");
  const yieldManagerAddress = getRequiredEnvVar("YIELD_MANAGER_ADDRESS");

  const lineaRollupV7Impl = await deployContractFromArtifacts(
    LineaRollupV7ContractName,
    LineaRollupV7Abi,
    LineaRollupV7Bytecode,
    signer,
  );

  const lineaRollupV7Reinitializer = getInitializerData(LineaRollupV7Abi, "reinitializeLineaRollupV7", [
    upgradeRoleAddresses,
    upgradePauseTypeRoles,
    upgradeUnpauseTypeRoles,
    yieldManagerAddress,
  ]);

  // The encoding should be used through the safe.
  // THIS IS JUST A SAMPLE AND WILL BE ADJUSTED WHEN NEEDED FOR GENERATING THE CALLDATA FOR THE UPGRADE CALL
  // https://www.4byte.directory/signatures/?bytes4_signature=0x9623609d
  const upgradeCallWithReinitializationUsingSecurityCouncil = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [proxyAddress, await lineaRollupV7Impl.getAddress(), lineaRollupV7Reinitializer],
    ),
  ]);

  console.log(
    "Encoded Tx Upgrade with Reinitialization from Security Council:",
    "\n",
    upgradeCallWithReinitializationUsingSecurityCouncil,
  );
  console.log("\n");
};

export default func;
func.tags = ["LineaRollupWithReinitialization"];
