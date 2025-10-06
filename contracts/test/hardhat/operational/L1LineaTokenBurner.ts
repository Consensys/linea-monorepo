import { ethers, network } from "hardhat";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { deployFromFactory, deployUpgradableFromFactory } from "../common/deployment";
import { L1LineaTokenBurner, TestERC20, TestL1MessageServiceMerkleProof } from "../../../typechain-types";
import { expectEvent, expectRevertWithCustomError } from "../common/helpers";
import {
  ADDRESS_ZERO,
  EMPTY_CALLDATA,
  INITIAL_WITHDRAW_LIMIT,
  MESSAGE_FEE,
  MESSAGE_VALUE_1ETH,
  ONE_DAY_IN_SECONDS,
  pauseTypeRoles,
  unpauseTypeRoles,
  VALID_MERKLE_PROOF,
} from "../common/constants";

describe("L1LineaTokenBurner", () => {
  let l1LineaTokenBurner: L1LineaTokenBurner;
  let admin: SignerWithAddress;
  let lineaToken: TestERC20;
  let messageService: TestL1MessageServiceMerkleProof;

  async function deployTestL1MessageServiceFixture(): Promise<TestL1MessageServiceMerkleProof> {
    return deployUpgradableFromFactory(
      "TestL1MessageServiceMerkleProof",
      [ONE_DAY_IN_SECONDS, INITIAL_WITHDRAW_LIMIT, pauseTypeRoles, unpauseTypeRoles],
      { unsafeAllow: ["constructor", "incorrect-initializer-order"] },
    ) as unknown as Promise<TestL1MessageServiceMerkleProof>;
  }

  async function deployL1LineaTokenBurnerContractFixture() {
    const lineaTokenFn = () => deployFromFactory("TestERC20", "TestERC20", "TEST", ethers.parseUnits("1000000000", 18));
    const lineaToken = (await loadFixture(lineaTokenFn)) as TestERC20;

    const messageService = await loadFixture(deployTestL1MessageServiceFixture);
    const l1LineaTokenBurnerFn = async () =>
      deployFromFactory("L1LineaTokenBurner", await messageService.getAddress(), await lineaToken.getAddress());
    const l1LineaTokenBurner = (await loadFixture(l1LineaTokenBurnerFn)) as L1LineaTokenBurner;

    return { l1LineaTokenBurner, lineaToken, messageService };
  }

  before(async () => {
    await network.provider.send("hardhat_reset");
    [admin] = await ethers.getSigners();
  });

  beforeEach(async () => {
    ({ l1LineaTokenBurner, lineaToken, messageService } = await loadFixture(deployL1LineaTokenBurnerContractFixture));
  });

  describe("construtor", () => {
    it("Should revert when messageService address is zero", async () => {
      await expectRevertWithCustomError(
        l1LineaTokenBurner,
        deployFromFactory("L1LineaTokenBurner", ADDRESS_ZERO, await lineaToken.getAddress()),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert when LINEA token address is zero", async () => {
      await expectRevertWithCustomError(
        l1LineaTokenBurner,
        deployFromFactory("L1LineaTokenBurner", await messageService.getAddress(), ADDRESS_ZERO),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should set the correct addresses", async () => {
      const lineaTokenAddress = await l1LineaTokenBurner.LINEA_TOKEN();
      const messageServiceAddress = await l1LineaTokenBurner.MESSAGE_SERVICE();

      expect(lineaTokenAddress).to.equal(await lineaToken.getAddress());
      expect(messageServiceAddress).to.equal(await messageService.getAddress());
    });
  });

  describe("claimMessageWithProof", () => {
    it("Should revert when message is not anchored on L1", async () => {
      await expectRevertWithCustomError(
        messageService,
        l1LineaTokenBurner.claimMessageWithProof({
          proof: VALID_MERKLE_PROOF.proof,
          messageNumber: 1,
          leafIndex: VALID_MERKLE_PROOF.index,
          from: admin.address,
          to: admin.address,
          fee: MESSAGE_FEE,
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
          feeRecipient: ADDRESS_ZERO,
          merkleRoot: VALID_MERKLE_PROOF.merkleRoot,
          data: EMPTY_CALLDATA,
        }),
        "L2MerkleRootDoesNotExist",
      );
    });

    it("Should do nothing when message is already claimed and burner contract balance == 0", async () => {
      await messageService.setL2L1MessageToClaimed(1);

      const burnerBalanceBefore = await lineaToken.balanceOf(await l1LineaTokenBurner.getAddress());
      expect(burnerBalanceBefore).to.equal(0);

      await expect(
        l1LineaTokenBurner.claimMessageWithProof({
          proof: VALID_MERKLE_PROOF.proof,
          messageNumber: 1,
          leafIndex: VALID_MERKLE_PROOF.index,
          from: admin.address,
          to: admin.address,
          fee: MESSAGE_FEE,
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
          feeRecipient: ADDRESS_ZERO,
          merkleRoot: VALID_MERKLE_PROOF.merkleRoot,
          data: EMPTY_CALLDATA,
        }),
      ).to.not.emit(lineaToken, "Transfer");
    });

    it("Should only burn LINEA tokens when message is already claimed and burner contract balance > 0", async () => {
      await messageService.setL2L1MessageToClaimed(1);

      const lineaTokensBalance = ethers.parseUnits("100", 18);
      await lineaToken.mint(await l1LineaTokenBurner.getAddress(), lineaTokensBalance);

      const burnerBalanceBefore = await lineaToken.balanceOf(await l1LineaTokenBurner.getAddress());
      expect(burnerBalanceBefore).to.equal(lineaTokensBalance);

      await expectEvent(
        lineaToken,
        l1LineaTokenBurner.claimMessageWithProof({
          proof: VALID_MERKLE_PROOF.proof,
          messageNumber: 1,
          leafIndex: VALID_MERKLE_PROOF.index,
          from: admin.address,
          to: admin.address,
          fee: MESSAGE_FEE,
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
          feeRecipient: ADDRESS_ZERO,
          merkleRoot: VALID_MERKLE_PROOF.merkleRoot,
          data: EMPTY_CALLDATA,
        }),
        "Transfer",
        [await l1LineaTokenBurner.getAddress(), ADDRESS_ZERO, lineaTokensBalance],
      );

      const burnerBalanceAfter = await lineaToken.balanceOf(await l1LineaTokenBurner.getAddress());
      expect(burnerBalanceAfter).to.equal(0);
    });

    it("Should claim message and burn LINEA tokens when message is not yet claimed and burner contract balance > 0", async () => {
      await messageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT * 2n });

      const lineaTokensBalance = ethers.parseUnits("100", 18);
      await lineaToken.mint(await l1LineaTokenBurner.getAddress(), lineaTokensBalance);

      const burnerBalanceBefore = await lineaToken.balanceOf(await l1LineaTokenBurner.getAddress());
      expect(burnerBalanceBefore).to.equal(lineaTokensBalance);

      await messageService.addL2MerkleRoots([VALID_MERKLE_PROOF.merkleRoot], VALID_MERKLE_PROOF.proof.length);

      const txPromise = l1LineaTokenBurner.claimMessageWithProof({
        proof: VALID_MERKLE_PROOF.proof,
        messageNumber: 1,
        leafIndex: VALID_MERKLE_PROOF.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: VALID_MERKLE_PROOF.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      const messageLeafHash = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["address", "address", "uint256", "uint256", "uint256", "bytes"],
          [admin.address, admin.address, MESSAGE_FEE, MESSAGE_FEE + MESSAGE_VALUE_1ETH, "1", EMPTY_CALLDATA],
        ),
      );

      await expectEvent(messageService, txPromise, "MessageClaimed", [messageLeafHash]);

      await expectEvent(lineaToken, txPromise, "Transfer", [
        await l1LineaTokenBurner.getAddress(),
        ADDRESS_ZERO,
        lineaTokensBalance,
      ]);
    });

    it("Should only claim message when message is not yet claimed and burner contract balance == 0", async () => {
      await messageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT * 2n });

      const burnerBalanceBefore = await lineaToken.balanceOf(await l1LineaTokenBurner.getAddress());
      expect(burnerBalanceBefore).to.equal(0);

      await messageService.addL2MerkleRoots([VALID_MERKLE_PROOF.merkleRoot], VALID_MERKLE_PROOF.proof.length);

      const txPromise = l1LineaTokenBurner.claimMessageWithProof({
        proof: VALID_MERKLE_PROOF.proof,
        messageNumber: 1,
        leafIndex: VALID_MERKLE_PROOF.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: VALID_MERKLE_PROOF.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      const messageLeafHash = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["address", "address", "uint256", "uint256", "uint256", "bytes"],
          [admin.address, admin.address, MESSAGE_FEE, MESSAGE_FEE + MESSAGE_VALUE_1ETH, "1", EMPTY_CALLDATA],
        ),
      );

      await expectEvent(messageService, txPromise, "MessageClaimed", [messageLeafHash]);
      await expect(txPromise).to.not.emit(lineaToken, "Transfer");
    });
  });
});
