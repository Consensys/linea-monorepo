// Test agnostic script for
import fs from "fs";
import path from "path";
import { ethers } from "ethers";
import { getBlockchainNode } from "../common";

async function main() {
  const blockchainNode = getBlockchainNode();
  const provider = new ethers.providers.JsonRpcProvider(blockchainNode);
  const keyFiles = process.argv.slice(2);
  const addresses = keyFiles
    .map((filePath) => JSON.parse(fs.readFileSync(filePath, "utf8")))
    .map((keyPair) => (keyPair.account_key ? keyPair.account_key.addr : keyPair.address));
  const balances = await Promise.all(addresses.map((address) => provider.getBalance(address)));
  for (let i = 0; i < balances.length; i++) {
    const keyFile = path.basename(keyFiles[i]).padStart(20, " ");
    const address = addresses[i];
    const balance = ethers.utils.formatEther(balances[i]).padEnd(25, " ");
    console.log(`Balance of ${keyFile} on ${address} is ${balance} ETH`);
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
