import { HDNodeWallet, Wallet } from "ethers";
import hre from "hardhat";
import { ethers } from "../../common/hardhat-ethers.js";
import { HardhatNetworkHDAccountsConfig } from "hardhat/types";

export const getWalletForIndex = (index: number) => {
  const accounts = hre.config.networks.hardhat.accounts as HardhatNetworkHDAccountsConfig;
  const signer = HDNodeWallet.fromPhrase(accounts.mnemonic, "", `m/44'/60'/0'/0/${index}`);
  return new Wallet(signer.privateKey, ethers.provider);
};
