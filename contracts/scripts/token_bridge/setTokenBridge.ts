#!/usr/bin/env node

/*
Usage:
ts-node scripts/token_bridge/setTokenBridge.ts \
    --priv-key SMART_CONTRACTS_PRIVATE_KEY \
    --layer l1 \
    --wrapper-address 0x73feE82ba7f6B98D27BCDc2bEFc1d3f6597fb02D \
    --wrapper-token-pair 0xnewaddress \
    --wrapper-pegged-token 0xnewaddress \
    --wrapped-address 0x964FF70695da981027c81020B1c58d833D49A640 \
    --wrapped-token-pair 0xnewaddress \
    --wrapped-l2-bridge 0xnewaddress \
    --l1-blockchain-url L1_URL \
    --l2-blockchain-url L2_URL

Used to set data in the contracts on the selected layer, each setter is optional.
layer options: l1, l2
*/

import { ethers } from "ethers";
import log from "npmlog";
import yargs from "yargs/yargs";
import { hideBin } from "yargs/helpers";
import { sanitizeAddress, sanitizePrivKey } from "../cli";

const argv = yargs(hideBin(process.argv))
  .option("priv-key", {
    describe: "Smart contract owner private key",
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
  .option("wrapper-token-pair", {
    describe: "L1 token wrapper new L2 token pair address",
    type: "string",
    demandOption: false,
    coerce: sanitizeAddress("wrapper-token-pair"),
  })
  .option("wrapper-pegged-token", {
    describe: "L1 token wrapper new pegged token address",
    type: "string",
    demandOption: false,
    coerce: sanitizeAddress("wrapper-pegged-token"),
  })
  .option("wrapped-address", {
    describe: "L2 wrapped token smart contract address",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("wrapped-address"),
  })
  .option("wrapped-token-pair", {
    describe: "L2 wrapped token new L1 token pair address",
    type: "string",
    demandOption: false,
    coerce: sanitizeAddress("wrapped-token-pair"),
  })
  .option("wrapped-l2-bridge", {
    describe: "L2 bridge address",
    type: "string",
    demandOption: false,
    coerce: sanitizeAddress("wrapped-l2-bridge"),
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
      "function setL2TokenPair(address)",
      "function setPeggedToken(address)",
    ]);
    const wallet = new ethers.Wallet(args.privKey, provider);
    const wrapper = new ethers.Contract(args.wrapperAddress, wrapperInterface, wallet);

    const tokenPair = args.wrapperTokenPair;
    const peggedToken = args.wrapperPeggedToken;

    if (tokenPair) {
      await wrapper.setL2TokenPair(tokenPair, {
        maxFeePerGas: 3292893616,
        maxPriorityFeePerGas: 2500000000,
      });
    }
    if (peggedToken) {
      await wrapper.setPeggedToken(peggedToken, {
        maxFeePerGas: 3292893616,
        maxPriorityFeePerGas: 2500000000,
      });
    }
  } else {
    // ------------------ L2 Wrapped token ------------------
    const provider = new ethers.providers.JsonRpcProvider(args.l2BlockchainUrl);
    const wrappedInterfance = new ethers.utils.Interface([
      "function setERC20Wrapper(address)",
      "function setL2Bridge(address)",
    ]);
    const wallet = new ethers.Wallet(args.privKey, provider);
    const wrapped = new ethers.Contract(args.wrappedAddress, wrappedInterfance, wallet);

    const l1WrapperAddr = args.wrappedTokenPair;
    const l2BridgeAddr = args.wrappedL2Bridge;
    if (l1WrapperAddr) {
      await wrapped.setERC20Wrapper(l1WrapperAddr);
    }
    if (l2BridgeAddr) {
      await wrapped.setERC20Wrapper(l2BridgeAddr);
    }
  }
}

main(argv)
  .then(() => process.exit(0))
  .catch((error) => {
    log.error("", error);
    process.exit(1);
  });
