import { HDNodeWallet, Wallet } from "ethers";
import { ethers } from "../../common/connection.js";

export const getWalletForIndex = (index: number) => {
  const accounts = hre.config.networks.hardhat.accounts as { mnemonic: string };
  const signer = HDNodeWallet.fromPhrase(accounts.mnemonic, "", `m/44'/60'/0'/0/${index}`);
  return new Wallet(signer.privateKey, ethers.provider);
};
