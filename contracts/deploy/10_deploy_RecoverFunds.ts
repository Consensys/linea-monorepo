import { network as hardhatNetwork } from "hardhat";

import { tryVerifyContract, requireAddressFromRegistryOrEnv, LogContractDeployment } from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployUpgradableFromFactory } from "../scripts/hardhat/utils";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

const func = withSignerUiSession("10_deploy_RecoverFunds.ts", async function () {
  const contractName = "RecoverFunds";

  // RecoverFunds DEPLOYED AS UPGRADEABLE PROXY
  const RecoverFunds_securityCouncil = requireAddressFromRegistryOrEnv(
    networkName,
    "L1_SECURITY_COUNCIL",
    "L1_SECURITY_COUNCIL",
  );
  const RecoverFunds_executorAddress = requireAddressFromRegistryOrEnv(
    networkName,
    "RECOVERFUNDS_EXECUTOR_ADDRESS",
    "RECOVERFUNDS_EXECUTOR_ADDRESS",
  );

  const contract = await deployUpgradableFromFactory(
    "RecoverFunds",
    [RecoverFunds_securityCouncil, RecoverFunds_executorAddress],
    {
      initializer: "initialize(address, address)",
      unsafeAllow: ["constructor"],
    },
  );

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  await tryVerifyContract(contractAddress);
});

export default deployScript(func, { tags: ["RecoverFunds"] });
