const { loadFixture } = networkHelpers;
import { expect } from "chai";
import hre from "hardhat";
const { ethers, networkHelpers } = await hre.network.connect();
import { BridgedToken, UpgradedBridgedToken } from "../../../../typechain-types";

const initialUserBalance = 10000;

async function createTokenBeaconProxy() {
  const [admin, unknown] = await ethers.getSigners();

  const BridgedToken = await ethers.getContractFactory("BridgedToken");

  // Deploy token beacon for L1
  const l1Impl = await BridgedToken.deploy();
  await l1Impl.waitForDeployment();
  const l1TokenBeacon = await (await ethers.getContractFactory("UpgradeableBeacon")).deploy(await l1Impl.getAddress());
  await l1TokenBeacon.waitForDeployment();

  // Deploy token beacon for L2
  const l2Impl = await BridgedToken.deploy();
  await l2Impl.waitForDeployment();
  const l2TokenBeacon = await (await ethers.getContractFactory("UpgradeableBeacon")).deploy(await l2Impl.getAddress());
  await l2TokenBeacon.waitForDeployment();

  // Create tokens via BeaconProxy
  const beaconProxyFactory = await ethers.getContractFactory("BeaconProxy");

  const abcInitData = BridgedToken.interface.encodeFunctionData("initialize", ["AbcToken", "ABC", 18]);
  const abcProxy = await beaconProxyFactory.deploy(await l1TokenBeacon.getAddress(), abcInitData);
  await abcProxy.waitForDeployment();
  const abcToken = BridgedToken.attach(await abcProxy.getAddress()) as unknown as BridgedToken;

  const sixInitData = BridgedToken.interface.encodeFunctionData("initialize", ["sixDecimalsToken", "SIX", 6]);
  const sixProxy = await beaconProxyFactory.deploy(await l1TokenBeacon.getAddress(), sixInitData);
  await sixProxy.waitForDeployment();
  const sixDecimalsToken = BridgedToken.attach(await sixProxy.getAddress()) as unknown as BridgedToken;

  // Create a new token implementation
  const UpgradedBridgedToken = await ethers.getContractFactory("UpgradedBridgedToken");
  const newImplementation = await UpgradedBridgedToken.deploy();
  await newImplementation.waitForDeployment();

  // Update l2TokenBeacon with new implementation
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  await l2TokenBeacon.connect(admin).upgradeTo(newImplementation.getAddress());

  // Set initial balance
  await sixDecimalsToken.connect(admin).mint(unknown.address, initialUserBalance);

  return {
    admin,
    unknown,
    l1TokenBeacon,
    l2TokenBeacon,
    newImplementation,
    UpgradedBridgedToken,
    abcToken,
    sixDecimalsToken,
  };
}

describe("BridgedToken", function () {
  it("Should deploy BridgedToken", async function () {
    const { abcToken, sixDecimalsToken } = await loadFixture(createTokenBeaconProxy);
    expect(await abcToken.getAddress()).to.be.not.null;
    expect(await sixDecimalsToken.getAddress()).to.be.not.null;
  });

  it("Should set the right metadata", async function () {
    const { abcToken, sixDecimalsToken } = await loadFixture(createTokenBeaconProxy);
    expect(await abcToken.name()).to.be.equal("AbcToken");
    expect(await abcToken.symbol()).to.be.equal("ABC");
    expect(await abcToken.decimals()).to.be.equal(18);
    expect(await sixDecimalsToken.name()).to.be.equal("sixDecimalsToken");
    expect(await sixDecimalsToken.symbol()).to.be.equal("SIX");
    expect(await sixDecimalsToken.decimals()).to.be.equal(6);
  });

  it("Should mint tokens", async function () {
    const { admin, unknown, abcToken } = await loadFixture(createTokenBeaconProxy);
    const amount = 100;
    await abcToken.connect(admin).mint(unknown.address, amount);
    expect(await abcToken.balanceOf(unknown.address)).to.be.equal(amount);
  });

  it("Should burn tokens", async function () {
    const { admin, unknown, sixDecimalsToken } = await loadFixture(createTokenBeaconProxy);
    const amount = 100;
    await sixDecimalsToken.connect(unknown).approve(admin.address, amount);
    await sixDecimalsToken.connect(admin).burn(unknown.address, amount);
    expect(await sixDecimalsToken.balanceOf(unknown.address)).to.be.equal(initialUserBalance - amount);
  });

  it("Should revert if mint/burn are called by an unknown address", async function () {
    const { unknown, abcToken } = await loadFixture(createTokenBeaconProxy);
    const amount = 100;
    await expect(abcToken.connect(unknown).mint(unknown.address, amount)).to.be.revertedWithCustomError(
      abcToken,
      "OnlyBridge",
    );
    await expect(abcToken.connect(unknown).burn(unknown.address, amount)).to.be.revertedWithCustomError(
      abcToken,
      "OnlyBridge",
    );
  });
});

describe("BeaconProxy", function () {
  it("Should enable upgrade of existing beacon proxy", async function () {
    const { admin, l1TokenBeacon, abcToken, newImplementation, UpgradedBridgedToken } =
      await loadFixture(createTokenBeaconProxy);
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    await l1TokenBeacon.connect(admin).upgradeTo(await newImplementation.getAddress());
    expect(await l1TokenBeacon.implementation()).to.be.equal(await newImplementation.getAddress());
    expect(
      await (UpgradedBridgedToken.attach(await abcToken.getAddress()) as UpgradedBridgedToken).isUpgraded(),
    ).to.be.equal(true);
  });

  it("Should deploy new beacon proxy with the updated implementation", async function () {
    const { l2TokenBeacon, UpgradedBridgedToken } = await loadFixture(createTokenBeaconProxy);
    const beaconProxyFactory = await ethers.getContractFactory("BeaconProxy");
    const initData = UpgradedBridgedToken.interface.encodeFunctionData("initialize", ["NAME", "SYMBOL", 18]);
    const proxy = await beaconProxyFactory.deploy(await l2TokenBeacon.getAddress(), initData);
    await proxy.waitForDeployment();
    const newTokenBeaconProxy = UpgradedBridgedToken.attach(
      await proxy.getAddress(),
    ) as unknown as import("../../../../typechain-types").UpgradedBridgedToken;
    expect(await newTokenBeaconProxy.isUpgraded()).to.be.equal(true);
  });

  it("Beacon upgrade should only be done by the owner", async function () {
    const { unknown, l1TokenBeacon, newImplementation } = await loadFixture(createTokenBeaconProxy);
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    await expect(l1TokenBeacon.connect(unknown).upgradeTo(await newImplementation.getAddress())).to.be.revertedWith(
      "Ownable: caller is not the owner",
    );
  });
});
