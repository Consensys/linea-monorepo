import { task } from "hardhat/config";

import { getTaskCliOrEnvValue } from "../../../common/helpers/environmentHelper";
import { getAddressFromRegistry } from "../../../common/helpers/readAddress";
import { getUiSigner, runWithSignerUiSession } from "../../../scripts/hardhat/signer-ui-bridge";

/*
    *******************************************************************************************
    1. Set the MESSAGE_SERVICE_ADDRESS 
    2. Set the MESSAGE_SERVICE_TYPE ( e.g. L2MessageService )
    3. Set the WITHDRAW_LIMIT_IN_WEI value
    *******************************************************************************************

    *******************************************************************************************
    DEPLOYER_PRIVATE_KEY=<key> \
    INFURA_API_KEY=<key> \
    pnpm exec hardhat setRateLimit \
    --message-service-address <address> \
    --message-service-type <string> \
    --withdraw-limit <uint256> \
    --network linea_sepolia
    *******************************************************************************************
*/

export default task("setRateLimit", "Sets the rate limit on a Message Service contract")
  .addOption({ name: "messageServiceAddress", defaultValue: "" })
  .addOption({ name: "messageServiceType", defaultValue: "" })
  .addOption({ name: "withdrawLimit", defaultValue: "" })
  .setInlineAction(async (taskArgs, hre) => {
    const connection = await hre.network.getOrCreate();
    return runWithSignerUiSession(hre, "task:setRateLimit", async () => {
      const { ethers } = connection;
      const networkName = connection.networkName === "default" ? "hardhat" : connection.networkName;

      const messageServiceContractType = getTaskCliOrEnvValue(taskArgs, "messageServiceType", "MESSAGE_SERVICE_TYPE");
      let messageServiceAddress = getTaskCliOrEnvValue(taskArgs, "messageServiceAddress", "MESSAGE_SERVICE_ADDRESS");

      if (messageServiceContractType === undefined) {
        throw "Please specify a Message Service name e.g: --message-service-type LineaRollup or MESSAGE_SERVICE_TYPE=LineaRollup";
      }

      if (messageServiceAddress === undefined) {
        messageServiceAddress = getAddressFromRegistry(networkName, messageServiceContractType);
        if (messageServiceAddress === undefined) {
          throw "messageServiceAddress is undefined";
        }
      }

      const withdrawLimitRaw = getTaskCliOrEnvValue(taskArgs, "withdrawLimit", "WITHDRAW_LIMIT_IN_WEI");
      if (withdrawLimitRaw === undefined || String(withdrawLimitRaw).trim() === "") {
        throw new Error(
          "Missing withdraw limit. Pass --withdraw-limit <wei> or set WITHDRAW_LIMIT_IN_WEI (Hardhat exposes this as withdrawLimit on the task).",
        );
      }
      const withdrawLimitInWei = BigInt(String(withdrawLimitRaw).trim());

      const signer = await getUiSigner(hre);
      const messageService = await ethers.getContractAt(messageServiceContractType, messageServiceAddress, signer);

      // get existing limit
      const limitInWei = await messageService.limitInWei();
      console.log(
        `Starting with rate limit in wei of ${limitInWei} at ${messageServiceAddress} of type ${messageServiceContractType}`,
      );

      // set limit

      const updateTx = await messageService.resetRateLimitAmount(withdrawLimitInWei);
      await updateTx.wait();

      console.log(`Changed rate limit in wei to ${withdrawLimitInWei} at ${messageServiceAddress}`);

      // get new updated limited
      const newLimitInWei = await messageService.limitInWei();
      console.log(`Validated rate limit in wei of ${newLimitInWei} at ${messageServiceAddress}`);
    });
  })
  .build();
