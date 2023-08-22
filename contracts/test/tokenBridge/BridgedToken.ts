import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers, upgrades } from "hardhat";

const initialUserBalance = 10000;

async function createTokenBeaconProxy() {
  const [admin, unknown] = await ethers.getSigners();

  const BridgedToken = await ethers.getContractFactory("BridgedToken");

  // Deploy token beacon
  const l1TokenBeacon = await upgrades.deployBeacon(BridgedToken);
  await l1TokenBeacon.deployed();

  const l2TokenBeacon = await upgrades.deployBeacon(BridgedToken);
  await l2TokenBeacon.deployed();

  // Create tokens
  const abcToken = await upgrades.deployBeaconProxy(l1TokenBeacon.address, BridgedToken, ["AbcToken", "ABC", 18]);

  const sixDecimalsToken = await upgrades.deployBeaconProxy(l1TokenBeacon.address, BridgedToken, [
    "sixDecimalsToken",
    "SIX",
    6,
  ]);

  // Create a new token implementation
  const UpgradedBridgedToken = await ethers.getContractFactory("UpgradedBridgedToken");
  const newImplementation = await UpgradedBridgedToken.deploy();
  await newImplementation.deployed();

  // Update l2TokenBeacon with new implementation
  await l2TokenBeacon.connect(admin).upgradeTo(newImplementation.address);

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
    expect(abcToken.address).to.be.not.null;
    expect(sixDecimalsToken.address).to.be.not.null;
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
    const { admin, l1TokenBeacon, abcToken, newImplementation, UpgradedBridgedToken } = await loadFixture(
      createTokenBeaconProxy,
    );
    await l1TokenBeacon.connect(admin).upgradeTo(newImplementation.address);
    expect(await l1TokenBeacon.implementation()).to.be.equal(newImplementation.address);
    expect(await UpgradedBridgedToken.attach(abcToken.address).isUpgraded()).to.be.equal(true);
  });

  it("Should deploy new beacon proxy with the updated implementation", async function () {
    const { l2TokenBeacon, UpgradedBridgedToken } = await loadFixture(createTokenBeaconProxy);
    const newTokenBeaconProxy = await upgrades.deployBeaconProxy(l2TokenBeacon.address, UpgradedBridgedToken, [
      "NAME",
      "SYMBOL",
      18, // Decimals
    ]);
    expect(await newTokenBeaconProxy.isUpgraded()).to.be.equal(true);
  });

  it("Beacon upgrade should only be done by the owner", async function () {
    const { unknown, l1TokenBeacon, newImplementation } = await loadFixture(createTokenBeaconProxy);
    await expect(l1TokenBeacon.connect(unknown).upgradeTo(newImplementation.address)).to.be.revertedWith(
      "Ownable: caller is not the owner",
    );
  });
});
