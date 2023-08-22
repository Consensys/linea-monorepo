#!/usr/bin/env node

/*
Usage:
ts-node scripts/token_bridge/checkTokenBridge.ts \
--priv-key YOUR_PRIVATE_KEY \
--layer l1 \
--wrapper-address 0x73feE82ba7f6B98D27BCDc2bEFc1d3f6597fb02D \
--wrapped-address 0x964FF70695da981027c81020B1c58d833D49A640 \
--l1-blockchain-url L1_URL \
--l2-blockchain-url L2_URL

Will print out data about the contracts on the selected layer.
layer options: l1, l2
*/

import { ethers } from "ethers";
import log from "npmlog";
import yargs from "yargs/yargs";
import { hideBin } from "yargs/helpers";
import { sanitizeAddress, sanitizePrivKey } from "../cli";

const argv = yargs(hideBin(process.argv))
  .option("priv-key", {
    describe: "Your private key",
    type: "string",
    demandOption: true,
    coerce: sanitizePrivKey("priv-key"),
  })
  .option("layer", {
    describe: "Option to check either L1 or L2 token bridge contracts",
    choices: ["l1", "l2"],
    type: "string",
    demandOption: true,
  })
  .option("wrapper-address", {
    describe: "L1 token wrapper smart contract address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("wrapper-address"),
  })
  .option("wrapped-address", {
    describe: "L2 wrapped token smart contract address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("wrapped-address"),
  })
  .option("l1-blockchain-url", {
    describe: "RPC url for the l1 blockchain",
    type: "string",
    demandOption: false,
  })
  .option("l2-blockchain-url", {
    describe: "RPC url for the l2 blockchain",
    type: "string",
    demandOption: false,
  })
  .parseSync();

async function main(args: typeof argv) {
  const layer = args.layer;

  if (layer === "l1") {
    // ------------------ L1 token wrapper ------------------
    const provider = new ethers.providers.JsonRpcProvider(args.l1BlockchainUrl);
    const wrapperInterface = new ethers.utils.Interface([
      "function l1Token() public view returns (address)",
      "function l1Bridge() public view returns (address)",
      "function l2TokenPair() public view returns (address)",
    ]);
    const wallet = new ethers.Wallet(args.privKey, provider);
    const wrapper = new ethers.Contract(args.wrapperAddress, wrapperInterface, wallet);

    const wrapperL1Token = await wrapper.l1Token();
    const wrapperl1Bridge = await wrapper.l1Bridge();
    const wrapperL2TokenPair = await wrapper.l2TokenPair();

    console.log("Wrapper L1 token: " + wrapperL1Token);
    console.log("Wrapper L1 Bridge: " + wrapperl1Bridge);
    console.log("Wrapper L2 Token Pair: " + wrapperL2TokenPair);
  } else {
    // ------------------ L2 Wrapped token ------------------
    const provider = new ethers.providers.JsonRpcProvider(argv["l2-blockchain-url"]);
    const wrappedInterfance = new ethers.utils.Interface([
      "function erc20Wrapper() public view returns (address)",
      "function l2Bridge() public view returns (address)",
    ]);
    const wallet = new ethers.Wallet(argv.privKey, provider);
    const wrapped = new ethers.Contract(args.wrappedAddress, wrappedInterfance, wallet);

    const wrappedErc20Wrapper = await wrapped.erc20Wrapper();
    const wrappedL2Bridge = await wrapped.l2Bridge();

    console.log("wrapped_erc20Wrapper: " + wrappedErc20Wrapper);
    console.log("wrappedL2Bridge: " + wrappedL2Bridge);
  }
}

main(argv)
  .then(() => process.exit(0))
  .catch((error) => {
    log.error("", error);
    process.exit(1);
  });
