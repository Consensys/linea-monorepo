import { ethers, upgrades, network, run } from "hardhat";
import { delay, storeAddress } from "../../utils/storeAddress";
import { requireEnv } from "../hardhat/utils";
import { SupportedChainIds } from "./supportedNetworks";

export async function main() {
  const messageServiceAddress = requireEnv("MESSAGE_SERVICE_ADDRESS");
  const zkEvmV2Address = requireEnv("ZKEVMV2_ADDRESS");
  const [owner] = await ethers.getSigners();
  const chainId = await owner.getChainId();

  if (!(chainId in SupportedChainIds)) {
    throw `Chaind Id ${chainId} not supported`;
  }

  let messageServiceAddr;
  let reservedAddresses;
  let chainIds = [SupportedChainIds.MAINNET, SupportedChainIds.LINEA];
  switch (chainId) {
    case SupportedChainIds.LINEA:
    case SupportedChainIds.LINEA_TESTNET:
      messageServiceAddr = messageServiceAddress;
      reservedAddresses = process.env.L2_RESERVED_TOKEN_ADDRESSES
        ? process.env.L2_RESERVED_TOKEN_ADDRESSES.split(" ")
        : [];
      if (chainId === SupportedChainIds.LINEA) {
        chainIds = [SupportedChainIds.LINEA, SupportedChainIds.MAINNET];
      } else {
        chainIds = [SupportedChainIds.LINEA_TESTNET, SupportedChainIds.GOERLI];
      }
      break;
    case SupportedChainIds.GOERLI:
    case SupportedChainIds.MAINNET:
      messageServiceAddr = zkEvmV2Address;
      reservedAddresses = process.env.L1_RESERVED_TOKEN_ADDRESSES
        ? process.env.L1_RESERVED_TOKEN_ADDRESSES.split(" ")
        : [];
      if (chainId === SupportedChainIds.GOERLI) {
        chainIds = [SupportedChainIds.GOERLI, SupportedChainIds.LINEA_TESTNET];
      }
      break;
  }

  // Deploy beacon for bridged token
  const BridgedToken = await ethers.getContractFactory("BridgedToken");

  const bridgedToken = await upgrades.deployBeacon(BridgedToken);
  await bridgedToken.deployed();
  storeAddress("BridgedToken", bridgedToken.address, network.name);

  console.log(`BridgedToken beacon deployed on ${network.name}, at address:`, bridgedToken.address);

  // Deploying TokenBridge
  const TokenBridgeFactory = await ethers.getContractFactory("TokenBridge");

  const tokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
    owner.address,
    messageServiceAddr,
    bridgedToken.address,
    chainIds[0],
    chainIds[1],
    reservedAddresses,
  ]);
  await tokenBridge.deployed();
  storeAddress("TokenBridge", tokenBridge.address, network.name);

  console.log(`TokenBridge deployed on ${network.name}, at address: ${tokenBridge.address}`);

  // Verify contracts on etherscan, we wait some time so that the contracts can be
  // propagated to the etherscan backend
  // We have to run the verify on the implementation contracts
  await delay(30000);
  console.log("Etherscan verification ongoing...");
  // Verify TokenBridge
  try {
    const bridgedTokenImplAddr = await upgrades.beacon.getImplementationAddress(bridgedToken.address);
    await run("verify", {
      address: bridgedTokenImplAddr,
    });
    const tokenBridgeImplAddr = await upgrades.erc1967.getImplementationAddress(tokenBridge.address);
    await run("verify", {
      address: tokenBridgeImplAddr,
    });
  } catch (err) {
    console.log(`Error happened during verification: ${err}`);
  }

  console.log("Etherscan verification done.");
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
