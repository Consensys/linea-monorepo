import { network as hardhatNetwork } from "hardhat";

import {
  LogContractDeployment,
  requireAddressFromRegistryOrEnv,
  tryVerifyContractWithConstructorArgs,
} from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployFromFactory } from "../scripts/hardhat/utils";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

const func = withSignerUiSession("12_deploy_CallForwardingProxy.ts", async function () {
  const contractName = "CallForwardingProxy";
  const signer = await getUiSigner();

  // This should be the LineaRollup
  const targetAddress = requireAddressFromRegistryOrEnv(networkName, "LineaRollup", "LINEA_ROLLUP_ADDRESS");

  const contract = await deployFromFactory(contractName, signer, targetAddress);

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  const args = [targetAddress];

  await tryVerifyContractWithConstructorArgs(
    contractAddress,
    "contracts/lib/CallForwardingProxy.sol:CallForwardingProxy",
    args,
  );
});
export default deployScript(func, { tags: ["CallForwardingProxy"] });
