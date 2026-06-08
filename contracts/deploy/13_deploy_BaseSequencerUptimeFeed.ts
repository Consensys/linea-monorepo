import { tryVerifyContractWithConstructorArgs, getRequiredEnvVar, LogContractDeployment } from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployFromFactory } from "../scripts/hardhat/utils";

const func = withSignerUiSession("13_deploy_BaseSequencerUptimeFeed.ts", async function () {
  const contractName = "LineaSequencerUptimeFeed";
  const signer = await getUiSigner();

  const initialStatus = getRequiredEnvVar("LINEA_SEQUENCER_UPTIME_FEED_INITIAL_STATUS");
  const adminAddress = getRequiredEnvVar("LINEA_SEQUENCER_UPTIME_FEED_ADMIN");
  const feedUpdaterAddress = getRequiredEnvVar("LINEA_SEQUENCER_UPTIME_FEED_UPDATER");

  const args = [initialStatus, adminAddress, feedUpdaterAddress];

  const contract = await deployFromFactory(contractName, signer, ...args);

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  await tryVerifyContractWithConstructorArgs(
    contractAddress,
    "src/operational/LineaSequencerUptimeFeed.sol:LineaSequencerUptimeFeed",
    args,
  );
});
export default deployScript(func, { tags: ["LineaSequencerUptimeFeed"] });
