import { ethers } from "ethers";
import {
  contractName as eip7702TestNestedName,
  abi as eip7702TestNestedAbi,
  bytecode as eip7702TestNestedBytecode,
} from "./static-artifacts/eip7702/Eip7702TestNested.json";
import {
  contractName as eip77022DelegatedName,
  abi as eip77022DelegatedAbi,
  bytecode as eip77022DelegatedBytecode,
} from "./static-artifacts/eip7702/Eip77022Delegated.json";
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

  await deployEIP7702TestNested(wallet, baseNonce, fees);
  await deployEIP77022Delegated(wallet, baseNonce + 1, fees);
  await deployEIP7702TestEntrypoint(wallet, baseNonce + 2, fees);
}

async function deployEIP7702TestNested(
  wallet: ethers.Wallet,
  nonce: number,
  fees: { maxFeePerGas: bigint; maxPriorityFeePerGas: bigint },
): Promise<string> {
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

  return nestedContract.getAddress();
}

async function deployEIP77022Delegated(
  wallet: ethers.Wallet,
  nonce: number,
  fees: { maxFeePerGas: bigint; maxPriorityFeePerGas: bigint },
): Promise<string> {
  const delegatedContract = await deployContractFromArtifacts(
    eip77022DelegatedName,
    eip77022DelegatedAbi,
    eip77022DelegatedBytecode,
    wallet,
    {
      nonce,
      ...fees,
    },
  );

  return delegatedContract.getAddress();
}

async function deployEIP7702TestEntrypoint(
  wallet: ethers.Wallet,
  nonce: number,
  fees: { maxFeePerGas: bigint; maxPriorityFeePerGas: bigint },
): Promise<string> {
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

  return entrypointContract.getAddress();
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
