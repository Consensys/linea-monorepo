import { task } from "hardhat/config";

import { getTaskCliOrEnvValue } from "../../../common/helpers/environmentHelper";
import { getAddressFromRegistry } from "../../../common/helpers/readAddress";
import { getUiSigner, runWithSignerUiSession } from "../../../scripts/hardhat/signer-ui-bridge";

import type { TokenBridge } from "../../../typechain-types";

/*
    *******************************************************************************************
    1. Deploy the TokenBridge + MessageService contracts on both networks and get the addresses
    2. Run this script on both addresses with the correct variables set.
    *******************************************************************************************
    DEPLOYER_PRIVATE_KEY=<key> \
    INFURA_API_KEY=<key> \
    pnpm exec hardhat setMessageServiceOnTokenBridge \
    --message-service-address <address> \
    --token-bridge-address <address> \
    --network sepolia
    *******************************************************************************************
*/

export default task("setMessageServiceOnTokenBridge", "Sets The Message Service On A TokenBridge")
  .addOption({ name: "messageServiceAddress", defaultValue: "" })
  .addOption({ name: "tokenBridgeAddress", defaultValue: "" })
  .setInlineAction(async (taskArgs, hre) => {
    const connection = await hre.network.getOrCreate();
    return runWithSignerUiSession(hre, "task:setMessageServiceOnTokenBridge", async () => {
      const { ethers } = connection;
      const networkName = connection.networkName === "default" ? "hardhat" : connection.networkName;

      const messageServiceAddress = getTaskCliOrEnvValue(taskArgs, "messageServiceAddress", "MESSAGE_SERVICE_ADDRESS");
      if (!messageServiceAddress) {
        throw "messageServiceAddress is undefined";
      }

      let tokenBridgeAddress = getTaskCliOrEnvValue(taskArgs, "tokenBridgeAddress", "TOKEN_BRIDGE_ADDRESS");
      if (!tokenBridgeAddress) {
        tokenBridgeAddress = await getAddressFromRegistry(networkName, "TokenBridge");
        if (!tokenBridgeAddress) {
          throw "tokenBridgeAddress is undefined";
        }
      }

      const chainId = (await ethers.provider.getNetwork()).chainId;
      console.log(`Current network's chainId is ${chainId}`);

      const signer = await getUiSigner(hre);
      const tokenBridge = (await ethers.getContractAt(
        "TokenBridge",
        tokenBridgeAddress,
        signer,
      )) as unknown as TokenBridge;
      const tx = await tokenBridge.setMessageService(messageServiceAddress!);

      await tx.wait();

      console.log(`MessageService set for the TokenBridge on: ${networkName}`);
    });
  })
  .build();
