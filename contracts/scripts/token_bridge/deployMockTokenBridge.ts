import { deployFromFactory } from "../hardhat/utils";
import { deployL1Erc20Wrapper, deployL2WrappedToken, deployL1Token } from "./contract-deployment-helpers";

async function main() {
  const tokenName = "USD Coin";
  const tokenSymbol = "USDC";

  console.log("Deploying L1 Token...");
  const l1Token = await deployL1Token(tokenName, tokenSymbol);

  console.log("Deploying bridge...");
  const bridge = await deployFromFactory("MockBridge");
  console.log(`Contract deployed at ${bridge.address}`);

  console.log("Deploying L1 ERC20 wrapper...");
  const erc20Wrapper = await deployL1Erc20Wrapper(l1Token.address, bridge.address);

  console.log("Deploying L2 wrapped ERC20 token...");
  const l2Token = await deployL2WrappedToken(bridge.address, tokenName, tokenSymbol);

  console.log("Setting token pairs...");
  await erc20Wrapper.setL2TokenPair(l2Token.address);
  await l2Token.setERC20Wrapper(erc20Wrapper.address);
}

main().catch((error) => console.error(error));
