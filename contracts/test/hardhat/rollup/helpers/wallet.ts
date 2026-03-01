import { HDNodeWallet, Wallet } from "ethers";
import hre from "hardhat";
import { ethers } from "../../common/connection.js";

const accounts = hre.config.networks.hardhat.accounts as {
  mnemonic: { get: () => Promise<string> };
};
const mnemonic = await accounts.mnemonic.get();

export const getWalletForIndex = (index: number) => {
  const signer = HDNodeWallet.fromPhrase(mnemonic, "", `m/44'/60'/0'/0/${index}`);
  return new Wallet(signer.privateKey, ethers.provider);
};
