import { HDNodeWallet, Wallet } from "ethers";
import { config, network as hardhatNetwork } from "hardhat";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;

type HardhatNetworkHDAccountsConfig = {
  mnemonic: string | { get(): Promise<string> };
  passphrase?: string | { get(): Promise<string> };
  path?: string;
  initialIndex?: number;
};

async function resolveSensitiveString(value: string | { get(): Promise<string> } | undefined, fallback = "") {
  if (value === undefined) {
    return fallback;
  }

  if (typeof value === "string") {
    return value;
  }

  return value.get();
}

export const getWalletForIndex = async (index: number) => {
  const accounts = config.networks.hardhat.accounts as HardhatNetworkHDAccountsConfig;
  const path = `${accounts.path ?? "m/44'/60'/0'/0"}/${(accounts.initialIndex ?? 0) + index}`;
  const signer = HDNodeWallet.fromPhrase(
    await resolveSensitiveString(accounts.mnemonic),
    await resolveSensitiveString(accounts.passphrase),
    path,
  );
  return new Wallet(signer.privateKey, ethers.provider);
};
