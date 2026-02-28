import hre from "hardhat";
import { MockERC20MintBurn } from "../../../typechain-types";

const connection = await hre.network.connect();
const { ethers } = connection;

const tokenNames = ["L1USDT", "L1DAI", "L1WETH", "L2UNI", "L2SHIBA"];

export async function deployTokens(verbose = false) {
  const ERC20 = await ethers.getContractFactory("MockERC20MintBurn");

  const tokens: { [name: string]: MockERC20MintBurn } = {};

  for (const name of tokenNames) {
    tokens[name] = await ERC20.deploy(name, name);
    await tokens[name].waitForDeployment();
    if (verbose) {
      console.log(name, "deployed");
    }
  }

  return tokens;
}
