import { Wallet } from "ethers";
import { JsonRpcProvider } from "@ethersproject/providers";
import { TEST_L1_SIGNER_PRIVATE_KEY, TEST_L2_SIGNER_PRIVATE_KEY } from "./constants";

export const getTestProvider = () => {
  return new JsonRpcProvider("http://localhost:8545");
};

export const getTestL1Signer = () => {
  return new Wallet(TEST_L1_SIGNER_PRIVATE_KEY, getTestProvider());
};

export const getTestL2Signer = () => {
  return new Wallet(TEST_L2_SIGNER_PRIVATE_KEY, getTestProvider());
};
