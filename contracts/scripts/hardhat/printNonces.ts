import fs from "fs";
import { ethers } from "hardhat";

const monitoredAccount = process.env.MONITORED_ACCOUNT || "/node-data/test/keys/operator_1.acc";

async function main() {
  console.log(getCurrentDate() + "Starting printing nonces script");
  const credentials = JSON.parse(fs.readFileSync(monitoredAccount, "utf8"));
  const privateKey = "0x" + credentials.account_key.priv_key;
  const wallet = new ethers.Wallet(privateKey, ethers.provider);
  console.log(getCurrentDate() + `Going to print nonces for address: ${wallet.address}`);
  const result = await ethers.provider.send("eth_chainId", []);
  console.log(getCurrentDate() + "eth_chainId", { result });

  async function listenLoop() {
    const count = await ethers.provider.getTransactionCount(wallet.address);
    const currentDate = getCurrentDate();
    console.log(`${currentDate} - transaction count for ${wallet.address}: ${count} `);
    setTimeout(listenLoop, 1000);
  }

  await listenLoop();
}

function getCurrentDate() {
  return "[" + new Date().toUTCString() + "] ";
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
