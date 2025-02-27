import { ethers } from "hardhat";
import { CallForwardingProxy } from "../../../../typechain-types";

export async function deployRevertingVerifier(scenario: bigint): Promise<string> {
  const revertingVerifierFactory = await ethers.getContractFactory("RevertingVerifier");
  const verifier = await revertingVerifierFactory.deploy(scenario);
  await verifier.waitForDeployment();
  return await verifier.getAddress();
}

export async function deployPlonkVerifierSepoliaFull(): Promise<string> {
  const plonkVerifierSepoliaFull = await ethers.getContractFactory("PlonkVerifierSepoliaFull");
  const verifier = await plonkVerifierSepoliaFull.deploy();
  await verifier.waitForDeployment();
  return await verifier.getAddress();
}

export async function deployTestPlonkVerifierForDataAggregation(): Promise<string> {
  const plonkVerifierSepoliaFull = await ethers.getContractFactory("TestPlonkVerifierForDataAggregation");
  const verifier = await plonkVerifierSepoliaFull.deploy();
  await verifier.waitForDeployment();
  return await verifier.getAddress();
}

export async function deployCallForwardingProxy(target: string): Promise<CallForwardingProxy> {
  const callForwardingProxyFactory = await ethers.getContractFactory("CallForwardingProxy");
  const callForwardingProxy = await callForwardingProxyFactory.deploy(target);
  await callForwardingProxy.waitForDeployment();
  return callForwardingProxy;
}
