import { network as hardhatNetwork } from "hardhat";

import { getRequiredEnvVar, LogContractDeployment, tryVerifyContractWithConstructorArgs } from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;

const func = withSignerUiSession("20_deploy_V3DexSwapAdapter.ts", async function () {
  const contractName = "V3DexSwapAdapter";
  const signer = await getUiSigner();

  const router = getRequiredEnvVar("V3_DEX_SWAP_ADAPTER_ROUTER");
  const wethToken = getRequiredEnvVar("V3_DEX_SWAP_ADAPTER_WETH_TOKEN");
  const lineaToken = getRequiredEnvVar("V3_DEX_SWAP_ADAPTER_LINEA_TOKEN");
  const poolTickSpacing = getRequiredEnvVar("V3_DEX_SWAP_ADAPTER_POOL_TICK_SPACING");

  const factory = await ethers.getContractFactory(contractName, signer);
  const contract = await factory.deploy(router, wethToken, lineaToken, poolTickSpacing);

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  const args = [router, wethToken, lineaToken, poolTickSpacing];
  await tryVerifyContractWithConstructorArgs(
    contractAddress,
    "src/operational/V3DexSwapAdapter.sol:V3DexSwapAdapter",
    args,
  );
});

export default deployScript(func, { tags: ["V3DexSwapAdapter"] });
