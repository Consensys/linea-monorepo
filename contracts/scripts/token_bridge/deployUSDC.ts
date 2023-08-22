#!/usr/bin/env node

/*
Usage:
ts-node scripts/token_bridge/deployUSDC.ts \
--priv-key YOUR_PRIVATE_KEY \
--l1-bridge-address L1 BRIDGE ADDRESS \
--l2-bridge-address L2 BRIDGE ADDRESS \
--l1-token-address L1 TOKEN ADDRESS TO WRAP \
--l1-blockchain-url L1_URL \
--l2-blockchain-url L2_URL

Will deploy the token bridge contracts for a given L1 token across both layers.
layer options: l1, l2
*/

import { ethers } from "ethers";
import log from "npmlog";
import yargs from "yargs/yargs";
import { hideBin } from "yargs/helpers";
import { deployL1Erc20Wrapper, deployL2WrappedToken, configureTokenWrapper } from "./contract-deployment-helpers";
import { sanitizeAddress, sanitizePrivKey } from "../cli";

const argv = yargs(hideBin(process.argv))
  .option("priv-key", {
    describe: "Bridge operator private key",
    type: "string",
    demandOption: true,
    coerce: sanitizePrivKey("priv-key"),
  })
  .option("l1-bridge-address", {
    describe: "Address of the Bridge smart contract on L2",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("l1-bridge-address"),
  })
  .option("l2-bridge-address", {
    describe: "Address of the Bridge smart contract on L2",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("l2-bridge-address"),
  })
  .option("l1-token-address", {
    describe: "Address of the L1 token to wrap",
    type: "string",
    demandOption: true,
    coerce: sanitizeAddress("l1-token-address"),
  })
  .option("l1-blockchain-url", {
    describe: "RPC url for the l1 blockchain",
    type: "string",
    demandOption: true,
  })
  .option("l2-blockchain-url", {
    describe: "RPC url for the l2 blockchain",
    type: "string",
    demandOption: true,
  })
  .parseSync();

async function main(args: typeof argv) {
  const l1TokenAddress = args.l1TokenAddress;
  const l1BridgeAddress = args.l1BridgeAddress;
  const l2BridgeAddress = args.l2BridgeAddress;

  const l1Provider = new ethers.providers.JsonRpcProvider(args.l1BlockchainUrl);
  const l2Provider = new ethers.providers.JsonRpcProvider(args.l2BlockchainUrl);

  const l1UsdcWrapper = await deployL1Erc20Wrapper(l1TokenAddress, l1BridgeAddress, l1Provider);
  const l2WrappedUsdc = await deployL2WrappedToken(l2BridgeAddress, "Token Name", "Token Symbol", l2Provider);
  await configureTokenWrapper(l1UsdcWrapper.address, l2WrappedUsdc.address, l1Provider);
}

main(argv)
  .then(() => process.exit(0))
  .catch((error) => {
    log.error("", error);
    process.exit(1);
  });
