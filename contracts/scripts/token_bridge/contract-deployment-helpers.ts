import { ethers } from "hardhat";
import { JsonRpcProvider } from "@ethersproject/providers";
import { deployFromFactory } from "../hardhat/utils";

async function deployL1Erc20Wrapper(l1TokenAddress: string, bridgeAddress: string, provider?: JsonRpcProvider) {
  const erc20Wrapper = await deployFromFactory("ERC20Wrapper", provider, l1TokenAddress, bridgeAddress);
  console.log(`Contract deployed at ${erc20Wrapper.address}`);
  return erc20Wrapper;
}

async function deployL2WrappedToken(
  bridgeAddress: string,
  tokenName: string,
  tokenSymbol: string,
  provider?: JsonRpcProvider,
) {
  const l2WrappedToken = await deployFromFactory("WrappedERC20", provider, bridgeAddress, tokenName, tokenSymbol);
  console.log(`Contract deployed at ${l2WrappedToken.address}`);
  return l2WrappedToken;
}

async function configureTokenWrapper(l1TokenWrapper: string, l2WrappedToken: string, provider: JsonRpcProvider) {
  const wrapperContract = await ethers.getContractFactory("ERC20Wrapper");
  wrapperContract.connect(provider.getSigner());
  const wrapper = wrapperContract.attach(l1TokenWrapper);
  await wrapper.setL2TokenPair(l2WrappedToken, {
    maxFeePerGas: 3292893616,
    maxPriorityFeePerGas: 2500000000,
  });
  console.log("Updated wrapper with new token pair address.");
}

async function deployL1Token(tokenName: string, tokenSymbol: string, provider = null) {
  const token = await deployFromFactory("Token", provider, tokenName, tokenSymbol);
  console.log(`Contract deployed at ${token.address}`);
  return token;
}

export { deployL1Erc20Wrapper, deployL2WrappedToken, configureTokenWrapper, deployL1Token };
