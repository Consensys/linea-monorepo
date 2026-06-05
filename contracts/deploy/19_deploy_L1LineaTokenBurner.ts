import { network as hardhatNetwork } from "hardhat";

import {
  requireAddressFromRegistryOrEnv,
  LogContractDeployment,
  tryVerifyContractWithConstructorArgs,
} from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

const func = withSignerUiSession("19_deploy_L1LineaTokenBurner.ts", async function () {
  const contractName = "L1LineaTokenBurner";
  const signer = await getUiSigner();

  const messageService = requireAddressFromRegistryOrEnv(networkName, "LineaRollup", "LINEA_ROLLUP_ADDRESS");
  const lineaToken = requireAddressFromRegistryOrEnv(networkName, "LINEA_TOKEN", "LINEA_TOKEN");

  const factory = await ethers.getContractFactory(contractName, signer);
  const contract = await factory.deploy(messageService, lineaToken);

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  const args = [messageService, lineaToken];
  await tryVerifyContractWithConstructorArgs(
    contractAddress,
    "src/operational/L1LineaTokenBurner.sol:L1LineaTokenBurner",
    args,
  );
});

export default deployScript(func, { tags: ["L1LineaTokenBurner"] });
