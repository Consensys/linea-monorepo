import { ethers, network, upgrades } from "hardhat";
import { default as deployments } from "../../deployments.json";
import { SupportedChainIds } from "./supportedNetworks";

export async function main() {
  const [owner] = await ethers.getSigners();
  const chainId = await owner.getChainId();

  if (!(chainId in SupportedChainIds)) {
    throw `Chaind Id ${chainId} not supported`;
  }

  if (!deployments.zkevm_dev.TokenBridge || !deployments.l2.TokenBridge) {
    throw "The TokenBridge needs to be deployed on both layers first";
  }

  let tokenBridgeAddress;
  let remoteTokenBridgeAddress;
  let lineaSafeAddr;
  switch (chainId) {
    case SupportedChainIds.MAINNET:
    case SupportedChainIds.GOERLI:
      tokenBridgeAddress = deployments.zkevm_dev.TokenBridge;
      remoteTokenBridgeAddress = deployments.l2.TokenBridge;
      lineaSafeAddr = process.env.LINEA_SAFE_L1 ? process.env.LINEA_SAFE_L1 : "";
      break;
    case SupportedChainIds.LINEA:
    case SupportedChainIds.LINEA_TESTNET:
      tokenBridgeAddress = deployments.l2.TokenBridge;
      remoteTokenBridgeAddress = deployments.zkevm_dev.TokenBridge;
      lineaSafeAddr = process.env.LINEA_SAFE_L2 ? process.env.LINEA_SAFE_L2 : "";
      break;
  }

  if (!lineaSafeAddr) {
    throw `Linea Safe address is not initialized`;
  }

  const TokenBridge = await ethers.getContractFactory("TokenBridge");
  const tokenBridge = await TokenBridge.attach(tokenBridgeAddress);
  let tx = await tokenBridge.setRemoteTokenBridge(remoteTokenBridgeAddress);

  await tx.wait();

  console.log(`RemoteTokenBridge set for the TokenBridge on: ${network.name}`);

  tx = await tokenBridge.transferOwnership(lineaSafeAddr);
  await tx.wait();
  await upgrades.admin.transferProxyAdminOwnership(lineaSafeAddr);

  console.log(`TokenBridge ownership and proxy admin set to: ${lineaSafeAddr}`);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main()
  .then(() => {
    process.exitCode = 0;
    process.exit();
  })
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
    process.exit();
  });
