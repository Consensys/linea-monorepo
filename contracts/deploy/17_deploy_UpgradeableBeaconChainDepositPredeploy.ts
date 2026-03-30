import { DeployFunction } from "hardhat-deploy/types";
import { withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployUpgradableFromFactory } from "../scripts/hardhat/utils";
import { tryVerifyContract, LogContractDeployment } from "../common/helpers";
import { EMPTY_INITIALIZE_SIGNATURE } from "../common/constants";

const func: DeployFunction = withSignerUiSession(
  "17_deploy_UpgradeableBeaconChainDepositPredeploy.ts",
  async function () {
    const contractName = "UpgradeableBeaconChainDepositPredeploy";

    const contract = await deployUpgradableFromFactory("UpgradeableBeaconChainDepositPredeploy", [], {
      initializer: EMPTY_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor"],
    });

    await LogContractDeployment(contractName, contract);
    const contractAddress = await contract.getAddress();

    await tryVerifyContract(
      contractAddress,
      "src/predeploy/UpgradeableBeaconChainDepositPredeploy.sol:UpgradeableBeaconChainDepositPredeploy",
    );
  },
);

export default func;
func.tags = ["UpgradeableBeaconChainDepositPredeploy"];
