import { network as hardhatNetwork } from "hardhat";

import {
  requireAddressFromRegistryOrEnv,
  LogContractDeployment,
  tryVerifyContractWithConstructorArgs,
} from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployFromFactory } from "../scripts/hardhat/utils";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

const func = withSignerUiSession("22_deploy_YieldManagerUpgrade.ts", async function () {
  const lineaRollupAddress = requireAddressFromRegistryOrEnv(networkName, "LineaRollup", "LINEA_ROLLUP_ADDRESS");
  const yieldManagerProxyAddress = requireAddressFromRegistryOrEnv(
    networkName,
    "YieldManager",
    "YIELD_MANAGER_ADDRESS",
  );

  console.log("Deploying Contract...");
  const signer = await getUiSigner();
  const contractName = "YieldManager";
  const contract = await deployFromFactory(contractName, signer, lineaRollupAddress);
  const newYieldManagerImplementationAddress = await contract.getAddress();
  await LogContractDeployment(contractName, contract);
  await tryVerifyContractWithConstructorArgs(
    newYieldManagerImplementationAddress,
    "src/yield/YieldManager.sol:YieldManager",
    [lineaRollupAddress],
  );

  // Encodes the upgrade calldata to be executed through the Security Council Safe.
  // upgrade(address proxy, address implementation) - selector 0x99a88ec4
  // https://www.4byte.directory/signatures/?bytes4_signature=0x99a88ec4
  const upgradeCallUsingSecurityCouncil = ethers.concat([
    "0x99a88ec4",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address"],
      [yieldManagerProxyAddress, newYieldManagerImplementationAddress],
    ),
  ]);

  console.log("Encoded Tx Upgrade from Security Council:", "\n", upgradeCallUsingSecurityCouncil);
  console.log("\n");
});

export default deployScript(func, { tags: ["YieldManagerImplementation"] });
