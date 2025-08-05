import { ethers, getChainId } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { tryStoreAddress } from "../common/helpers";

// Deploy EIP-2935 Historical Block Hashes system contract - https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2935.md
// Prerequisite - Fund the predetermined sender address with enough ETH to cover the deployment cost
// npx hardhat deploy --tags EIP2935SystemContract

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const contractName = "EIP2935SystemContract";
  const expectedAddress = "0x0000F90827F1C53a10cb7A02335B175320002935";

  // Check if the contract is already deployed at the expected address
  const code = await ethers.provider.getCode(expectedAddress);
  if (code !== "0x") {
    console.log(`EIP-2935 system contract already deployed at ${expectedAddress}`);
    return;
  }

  // EIP-2935 deployment transaction data - https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2935.md#deployment
  const deploymentTx = {
    type: 0x0,
    nonce: 0x0,
    to: null,
    gas: 0x3d090,
    gasPrice: 0xe8d4a51000,
    value: 0x0,
    data: "0x60538060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604657602036036042575f35600143038111604257611fff81430311604257611fff9006545f5260205ff35b5f5ffd5b5f35611fff60014303065500",
    v: 0x1b,
    r: "0x539",
    s: "0xaa12693182426612186309f02cfe8a80a0000",
  };

  // The predetermined sender address from EIP-2935
  const predeterminedSenderAddress = "0x3462413Af4609098e1E27A490f554f260213D685";

  const [deployer] = await ethers.getSigners();

  // Fund the sender address if needed
  const predeterminedSenderBalance = await ethers.provider.getBalance(predeterminedSenderAddress);
  const requiredBalance = BigInt(deploymentTx.gas) * BigInt(deploymentTx.gasPrice);

  if (predeterminedSenderBalance < requiredBalance) {
    const fundingAmount = requiredBalance - predeterminedSenderBalance + 1n; // Add a small buffer
    console.log(
      `Funding predetermined sender address ${predeterminedSenderAddress} with ${ethers.formatEther(fundingAmount)} ETH`,
    );
    const fundingTx = await deployer.sendTransaction({
      to: predeterminedSenderAddress,
      value: fundingAmount,
    });
    await fundingTx.wait();
  }

  // Serialize the raw transaction with signature
  const rawTx = ethers.Transaction.from({
    type: deploymentTx.type,
    nonce: deploymentTx.nonce,
    to: deploymentTx.to,
    gasLimit: deploymentTx.gas,
    gasPrice: deploymentTx.gasPrice,
    value: deploymentTx.value,
    data: deploymentTx.data,
  });

  rawTx.signature = ethers.Signature.from({
    v: deploymentTx.v,
    r: deploymentTx.r,
    s: deploymentTx.s,
  });

  const tx = await ethers.provider.broadcastTransaction(rawTx.serialized);
  const receipt = await tx.wait();
  console.log(
    `contract=${contractName} deployed: address=${receipt?.contractAddress} blockNumber=${receipt?.blockNumber} chainId=${getChainId()}`,
  );
  if (receipt?.contractAddress !== expectedAddress) {
    throw new Error(`Contract deployed to ${receipt?.contractAddress}, expected ${expectedAddress}`);
  }

  await tryStoreAddress(hre.network.name, contractName, expectedAddress, tx.hash);
};
export default func;
func.tags = ["EIP2935SystemContract"];
