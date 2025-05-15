import { ethers } from "hardhat";
import { toBeHex } from "ethers";
import { IPlonkVerifier, Mimc, PlonkVerifierDev__factory } from "../../../typechain-types";
import { deployFromFactory } from "../common/deployment";
import { expectEventDirectFromReceiptData, expectRevertWithCustomError } from "../common/helpers";

describe("PlonkVerifierDev", () => {
  let mimc: Mimc;

  async function deployContract(params: IPlonkVerifier.ChainConfigurationParameterStruct[]) {
    const factory = await ethers.getContractFactory("PlonkVerifierDev", {
      libraries: { Mimc: await mimc.getAddress() },
    });
    const verifier = await factory.deploy(params);
    await verifier.waitForDeployment();
    return verifier;
  }

  before(async () => {
    mimc = (await deployFromFactory("Mimc")) as Mimc;
  });

  describe("Deployment", () => {
    it("Should revert when no chain configuration has been provided", async () => {
      await expectRevertWithCustomError(
        new PlonkVerifierDev__factory({ ["src/libraries/Mimc.sol:Mimc"]: await mimc.getAddress() }),
        deployContract([]),
        "ChainConfigurationNotSet",
      );
    });

    it("Should deploy with one configuration value that has a first 0 bit.", async () => {
      const chainId = toBeHex("0x4919", 32); // 0x4919 = 1337

      const params = [
        {
          value: chainId,
          name: "chainId",
        },
      ];
      const verifier = await deployContract(params);
      const receipt = await verifier.deploymentTransaction()?.wait();

      const expectedConfigurationHash = await mimc.hash(chainId);

      expectEventDirectFromReceiptData(verifier, receipt!, "ChainConfigurationSet", [
        expectedConfigurationHash,
        [[chainId, "chainId"]],
      ]);
    });

    it("Should deploy with one configuration value that has a first non 0 bit.", async () => {
      const chainId = ethers.toBeHex("0x8900000000000000000000000000000000000000000000000000000000000089", 32); // 0x4919 = 1337

      const params = [
        {
          value: chainId,
          name: "chainId",
        },
      ];
      const verifier = await deployContract(params);
      const receipt = await verifier.deploymentTransaction()?.wait();

      const leastSignificantBit = BigInt(chainId) >> 128n;
      const mostSignificantBit = BigInt(chainId) & BigInt("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF");

      const mimcPayload = ethers.concat([toBeHex(leastSignificantBit, 32), toBeHex(mostSignificantBit, 32)]);

      const expectedConfigurationHash = await mimc.hash(mimcPayload);

      expectEventDirectFromReceiptData(verifier, receipt!, "ChainConfigurationSet", [
        expectedConfigurationHash,
        [[chainId, "chainId"]],
      ]);
    });

    it("Should deploy with multiple configuration values that have a first 0 bit.", async () => {
      const chainId = toBeHex("0x4919", 32); // 0x4919 = 1337
      const baseFee = toBeHex("0x07", 32);
      const l2MessageServiceAddress = toBeHex("0xe537D669CA013d86EBeF1D64e40fC74CADC91987", 32);

      const params = [
        {
          value: chainId,
          name: "chainId",
        },
        {
          value: baseFee,
          name: "baseFee",
        },
        {
          value: l2MessageServiceAddress,
          name: "l2MessageServiceAddress",
        },
      ];

      const verifier = await deployContract(params);
      const receipt = await verifier.deploymentTransaction()?.wait();

      const mimcPayload = ethers.concat([chainId, baseFee, l2MessageServiceAddress]);
      const expectedConfigurationHash = await mimc.hash(mimcPayload);

      expectEventDirectFromReceiptData(verifier, receipt!, "ChainConfigurationSet", [
        expectedConfigurationHash,
        [
          [chainId, "chainId"],
          [baseFee, "baseFee"],
          [l2MessageServiceAddress, "l2MessageServiceAddress"],
        ],
      ]);
    });

    it("Should deploy with multiple configuration values that have a first non 0 bit.", async () => {
      const chainId = toBeHex("0x8900000000000000000000000000000000000000000000000000000000000089", 32); // 0x4919 = 1337
      const baseFee = toBeHex("0x07", 32);
      const l2MessageServiceAddress = toBeHex("0xe537D669CA013d86EBeF1D64e40fC74CADC91987", 32);

      const params = [
        {
          value: chainId,
          name: "chainId",
        },
        {
          value: baseFee,
          name: "baseFee",
        },
        {
          value: l2MessageServiceAddress,
          name: "l2MessageServiceAddress",
        },
      ];

      const verifier = await deployContract(params);
      const receipt = await verifier.deploymentTransaction()?.wait();

      const leastSignificantBit = BigInt(chainId) >> 128n;
      const mostSignificantBit = BigInt(chainId) & BigInt("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF");

      const mimcPayload = ethers.concat([
        toBeHex(leastSignificantBit, 32),
        toBeHex(mostSignificantBit, 32),
        baseFee,
        l2MessageServiceAddress,
      ]);

      const expectedConfigurationHash = await mimc.hash(mimcPayload);

      expectEventDirectFromReceiptData(verifier, receipt!, "ChainConfigurationSet", [
        expectedConfigurationHash,
        [
          [chainId, "chainId"],
          [baseFee, "baseFee"],
          [l2MessageServiceAddress, "l2MessageServiceAddress"],
        ],
      ]);
    });
  });
});
