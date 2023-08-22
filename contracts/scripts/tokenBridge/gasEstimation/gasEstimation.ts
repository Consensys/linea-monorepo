import { ethers, upgrades } from "hardhat";
import { deployBridgedTokenBeacon } from "../test/deployBridgedTokenBeacon";
import { deployTokens } from "../test/deployTokens";
import { BigNumber } from "ethers";
import { getPermitData } from "../../../test/tokenBridge/utils/permitHelper";

const initialUserBalance = BigNumber.from(10 ** 9);
const bridgeAmount = 70;
const DEPLOYED_STATUS = ethers.utils.getAddress("0x0000000000000000000000000000000000000333");
const deadline = ethers.constants.MaxUint256;

/**
 * Simple script to test the gas cost of the method bridgeToken
 */
async function main() {
  const [user] = await ethers.getSigners();
  const { chainId } = await ethers.provider.getNetwork();

  // Deploy beacon for bridged tokens
  const tokenBeacons = await deployBridgedTokenBeacon();

  // Deploy the messageService
  const MockMessageServiceV2 = await ethers.getContractFactory("MockMessageServiceV2");
  const mockMessageServiceV2 = await MockMessageServiceV2.deploy();
  await mockMessageServiceV2.deployed();

  // Deploy tokenBridges
  const TokenBridgeFactory = await ethers.getContractFactory("MockTokenBridge");
  const l1TokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
    user.address,
    mockMessageServiceV2.address,
    tokenBeacons.l1TokenBeacon.address,
    [],
  ]);
  await l1TokenBridge.deployed();

  const l2TokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
    user.address,
    mockMessageServiceV2.address,
    tokenBeacons.l2TokenBeacon.address,
    [],
  ]);
  await l2TokenBridge.deployed();

  // Setting reciprocal addresses of TokenBridges
  await l1TokenBridge.setRemoteTokenBridge(l2TokenBridge.address);
  await l2TokenBridge.setRemoteTokenBridge(l1TokenBridge.address);

  // Deploy test tokens
  const tokens = await deployTokens(false);

  // Mint tokens for user and approve bridge
  for (const name in tokens) {
    const token = tokens[name];
    await token.mint(user.address, initialUserBalance);
    await token.connect(user).approve(l1TokenBridge.address, ethers.constants.MaxUint256);
  }

  // Create a bridgedToken to test to bridge with the metadata
  const BridgedToken = await ethers.getContractFactory("BridgedToken");
  const abcToken = await upgrades.deployBeaconProxy(tokenBeacons.l1TokenBeacon.address, BridgedToken, [
    "AbcTokendfgrdgredt",
    "ABC",
    18,
  ]);

  // Estimate gas cost without permitData
  let gasCost = await l1TokenBridge.connect(user).estimateGas.bridgeToken(tokens.L1DAI.address, 10, user.address);
  console.log("basic bridgeToken:                           ", gasCost.toString());

  // Prepare data for permit calldata
  await abcToken.mint(user.address, initialUserBalance);
  const nonce = await abcToken.nonces(user.address);
  const permitData = await getPermitData(user, abcToken, nonce, chainId, l1TokenBridge.address, bridgeAmount, deadline);

  gasCost = await l1TokenBridge
    .connect(user)
    .estimateGas.bridgeTokenWithPermit(abcToken.address, bridgeAmount, user.address, permitData);
  console.log("bridgeToken with permit:                     ", gasCost.toString());

  await l1TokenBridge.setNativeMappingValue(tokens.L1DAI.address, DEPLOYED_STATUS);

  // console.log(await l1TokenBridge.nativeToBridgedToken(tokens.L1DAI.address));

  gasCost = await l1TokenBridge.connect(user).estimateGas.bridgeToken(tokens.L1DAI.address, 10, user.address);
  console.log("bridgeToken after confirmDeploy:             ", gasCost.toString());

  await l1TokenBridge.setNativeMappingValue(abcToken.address, DEPLOYED_STATUS);

  gasCost = await l1TokenBridge
    .connect(user)
    .estimateGas.bridgeTokenWithPermit(abcToken.address, bridgeAmount, user.address, permitData);
  console.log("bridgeToken with permit after confirmDeploy: ", gasCost.toString());
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
