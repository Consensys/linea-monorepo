import { ethers } from "hardhat";
import { Contract } from "ethers";

const tokenNames = ["L1USDT", "L1DAI", "L1WETH", "L2UNI", "L2SHIBA"];

export async function deployTokens(verbose = false) {
  const ERC20 = await ethers.getContractFactory("MockERC20MintBurn");

  const tokens: { [name: string]: Contract } = {};

  for (const name of tokenNames) {
    tokens[name] = await ERC20.deploy(name, name);
    await tokens[name].deployed();
    if (verbose) {
      console.log(name, "deployed");
    }
  }

  return tokens;
}
