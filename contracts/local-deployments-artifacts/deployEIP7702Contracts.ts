import { ethers } from "ethers";
import {
  contractName as eip7702TestNestedName,
  abi as eip7702TestNestedAbi,
  bytecode as eip7702TestNestedBytecode,
} from "./static-artifacts/eip7702/Eip7702TestNested.json";
import {
  contractName as eip77022DeletegatedName,
  abi as eip77022DeletegatedAbi,
  bytecode as eip77022DeletegatedBytecode,
} from "./static-artifacts/eip7702/Eip77022Deletegated.json";
import {
  contractName as eip7702TestEntrypointName,
  abi as eip7702TestEntrypointAbi,
  bytecode as eip7702TestEntrypointBytecode,
} from "./static-artifacts/eip7702/Eip7702TestEntrypoint.json";
import { deployContractFromArtifacts } from "../common/helpers/deployments";

async function main() {
  const ORDERED_NONCE_POST_L2MESSAGESERVICE = 3;
  const ORDERED_NONCE_POST_TOKENBRIDGE = 5;
  const ORDERED_NONCE_POST_L2_TEST_ERC20 = 1;

  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const wallet = new ethers.Wallet(process.env.DEPLOYER_PRIVATE_KEY!, provider);

  let baseNonce: number;
  if (!process.env.L2_NONCE) {
    baseNonce = await wallet.getNonce();
  } else {
    baseNonce =
      parseInt(process.env.L2_NONCE) +
      ORDERED_NONCE_POST_L2MESSAGESERVICE +
      ORDERED_NONCE_POST_TOKENBRIDGE +
      ORDERED_NONCE_POST_L2_TEST_ERC20;
  }

  const fees = {
    maxFeePerGas: 7_200_000_000_000n,
    maxPriorityFeePerGas: 7_000_000_000_000n,
  };

  console.log(`Deploying EIP-7702 test contracts with deployer ${wallet.address} (baseNonce: ${baseNonce})`);

  // Deploy in dependency order: Nested first, then Delegated, then Entrypoint
  const nestedAddress = await deployEIP7702TestNested(wallet, baseNonce, fees);
  const delegatedAddress = await deployEIP77022Deletegated(wallet, baseNonce + 1, fees);
  const entrypointAddress = await deployEIP7702TestEntrypoint(wallet, baseNonce + 2, fees);

  console.log(`\n=== EIP-7702 Contracts Deployed Successfully ===`);
  console.log(`Nested:     ${nestedAddress}`);
  console.log(`Delegated:  ${delegatedAddress}`);
  console.log(`Entrypoint: ${entrypointAddress}`);
}

async function deployEIP7702TestNested(
  wallet: ethers.Wallet,
  nonce: number,
  fees: { maxFeePerGas: bigint; maxPriorityFeePerGas: bigint },
): Promise<string> {
  console.log(`\nDeploying Eip7702TestNested (nonce: ${nonce})`);
  const nestedContract = await deployContractFromArtifacts(
    eip7702TestNestedName,
    eip7702TestNestedAbi,
    eip7702TestNestedBytecode,
    wallet,
    {
      nonce,
      ...fees,
    },
  );

  const nestedAddress = await nestedContract.getAddress();
  return nestedAddress;
}

async function deployEIP77022Deletegated(
  wallet: ethers.Wallet,
  nonce: number,
  fees: { maxFeePerGas: bigint; maxPriorityFeePerGas: bigint },
): Promise<string> {
  console.log(`\nDeploying Eip77022Deletegated (nonce: ${nonce})`);
  const delegatedContract = await deployContractFromArtifacts(
    eip77022DeletegatedName,
    eip77022DeletegatedAbi,
    eip77022DeletegatedBytecode,
    wallet,
    {
      nonce,
      ...fees,
    },
  );

  const delegatedAddress = await delegatedContract.getAddress();
  return delegatedAddress;
}

async function deployEIP7702TestEntrypoint(
  wallet: ethers.Wallet,
  nonce: number,
  fees: { maxFeePerGas: bigint; maxPriorityFeePerGas: bigint },
): Promise<string> {
  console.log(`\nDeploying Eip7702TestEntrypoint (nonce: ${nonce})`);
  const entrypointContract = await deployContractFromArtifacts(
    eip7702TestEntrypointName,
    eip7702TestEntrypointAbi,
    eip7702TestEntrypointBytecode,
    wallet,
    {
      nonce,
      ...fees,
    },
  );

  const entrypointAddress = await entrypointContract.getAddress();
  return entrypointAddress;
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
