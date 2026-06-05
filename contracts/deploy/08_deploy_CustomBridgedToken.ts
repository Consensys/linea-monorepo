import { network as hardhatNetwork } from "hardhat";

import { tryVerifyContract, getRequiredEnvVar, requireAddressFromRegistryOrEnv } from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployUpgradableFromFactory } from "../scripts/hardhat/utils";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

const func = withSignerUiSession("08_deploy_CustomBridgedToken.ts", async function () {
  const contractName = "CustomBridgedToken";

  const CustomTokenBridge_name = getRequiredEnvVar("CUSTOMTOKENBRIDGE_NAME");
  const CustomTokenBridge_symbol = getRequiredEnvVar("CUSTOMTOKENBRIDGE_SYMBOL");
  const CustomTokenBridge_decimals = getRequiredEnvVar("CUSTOMTOKENBRIDGE_DECIMALS");
  const CustomTokenBridge_bridge_address = requireAddressFromRegistryOrEnv(
    networkName,
    "CUSTOMTOKENBRIDGE_BRIDGE_ADDRESS",
    "CUSTOMTOKENBRIDGE_BRIDGE_ADDRESS",
  );

  const chainId = (await ethers.provider.getNetwork()).chainId;
  console.log(`Current network's chainId is ${chainId}`);

  // Deploy proxy for custom bridged token
  const customBridgedToken = await deployUpgradableFromFactory(
    contractName,
    [CustomTokenBridge_name, CustomTokenBridge_symbol, CustomTokenBridge_decimals, CustomTokenBridge_bridge_address],
    {
      initializer: "initializeV2(string,string,uint8,address)",
      unsafeAllow: ["constructor"],
    },
  );

  const txReceipt = await customBridgedToken.deploymentTransaction()?.wait();
  if (!txReceipt) {
    throw "Deployment transaction not found.";
  }

  const contractAddress = await customBridgedToken.getAddress();

  console.log(
    `contract=${contractName} deployed: address=${contractAddress} blockNumber=${txReceipt.blockNumber} chainId=${chainId}`,
  );

  await tryVerifyContract(contractAddress);
});

export default deployScript(func, { tags: ["CustomBridgedToken"] });
