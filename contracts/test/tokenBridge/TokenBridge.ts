import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers, upgrades } from "hardhat";
import { deployTokenBridgeWithMockMessaging } from "../../scripts/tokenBridge/test/deployTokenBridges";
import { deployTokens } from "../../scripts/tokenBridge/test/deployTokens";
import { getPermitData } from "./utils/permitHelper";

const initialUserBalance = BigNumber.from(10 ** 9);
const mockName = "L1 DAI";
const mockSymbol = "L1DAI";
const mockDecimals = 18;
const RESERVED_STATUS = ethers.utils.getAddress("0x0000000000000000000000000000000000000111");
const PLACEHOLDER_ADDRESS = ethers.utils.getAddress("0x5555555555555555555555555555555555555555");
const CUSTOM_ADDRESS = ethers.utils.getAddress("0x9999999999999999999999999999999999999999");
const EMPTY_PERMIT_DATA = "0x";

describe("TokenBridge", function () {
  async function deployContractsFixture() {
    const [owner, user] = await ethers.getSigners();

    // Deploy and configure bridges
    const deploymentFixture = await deployTokenBridgeWithMockMessaging();

    // Deploy tokens
    const tokens = await deployTokens();

    // Mint tokens for user and approve bridge
    for (const name in tokens) {
      const token = tokens[name];
      await token.mint(user.address, initialUserBalance);

      let bridgeAddress;
      if ((await token.name()).includes("L1")) {
        bridgeAddress = deploymentFixture.l1TokenBridge.address;
      }
      if ((await token.name()).includes("L2")) {
        bridgeAddress = deploymentFixture.l2TokenBridge.address;
      }

      await token.connect(user).approve(bridgeAddress, ethers.constants.MaxUint256);
    }
    const encodedTokenMetadata = ethers.utils.defaultAbiCoder.encode(
      ["string", "string", "uint8"],
      [mockName, mockSymbol, mockDecimals],
    );
    return { owner, user, ...deploymentFixture, tokens, encodedTokenMetadata };
  }

  describe("initialize", function () {
    it("Should revert if it has already been intialized", async function () {
      const { user, l1TokenBridge, chainIds } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge
          .connect(user)
          .initialize(PLACEHOLDER_ADDRESS, PLACEHOLDER_ADDRESS, PLACEHOLDER_ADDRESS, chainIds[0], chainIds[1], []),
      ).to.be.revertedWith("Initializable: contract is already initialized");
    });

    it("Should revert if one of the initializing parameters is address 0", async function () {
      const { l1TokenBridge, chainIds } = await loadFixture(deployContractsFixture);
      const TokenBridge = await ethers.getContractFactory("TokenBridge");
      await expect(
        upgrades.deployProxy(TokenBridge, [
          ethers.constants.AddressZero,
          PLACEHOLDER_ADDRESS,
          PLACEHOLDER_ADDRESS,
          chainIds[0],
          chainIds[1],
          [],
        ]),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");

      await expect(
        upgrades.deployProxy(TokenBridge, [
          PLACEHOLDER_ADDRESS,
          ethers.constants.AddressZero,
          PLACEHOLDER_ADDRESS,
          chainIds[0],
          chainIds[1],
          [],
        ]),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");

      await expect(
        upgrades.deployProxy(TokenBridge, [
          PLACEHOLDER_ADDRESS,
          PLACEHOLDER_ADDRESS,
          ethers.constants.AddressZero,
          chainIds[0],
          chainIds[1],
          [],
        ]),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");

      await expect(
        upgrades.deployProxy(TokenBridge, [
          PLACEHOLDER_ADDRESS,
          PLACEHOLDER_ADDRESS,
          PLACEHOLDER_ADDRESS,
          chainIds[0],
          chainIds[1],
          [PLACEHOLDER_ADDRESS, ethers.constants.AddressZero],
        ]),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");
    });
  });

  describe("Permissions", function () {
    it("Should revert if completeBridging  is not called by the messageService", async function () {
      const {
        user,
        l1TokenBridge,
        tokens: { L1DAI },
        encodedTokenMetadata,
        chainIds,
      } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge.connect(user).completeBridging(L1DAI.address, 1, user.address, chainIds[1], encodedTokenMetadata),
      ).to.be.revertedWithCustomError(l1TokenBridge, "CallerIsNotMessageService");
    });

    it("Should revert if completeBridging  message does not come from the remote Token Bridge", async function () {
      const {
        user,
        messageService,
        l1TokenBridge,
        l2TokenBridge,
        tokens: { L1DAI },
        encodedTokenMetadata,
        chainIds,
      } = await loadFixture(deployContractsFixture);

      await expect(
        messageService.connect(user).sendMessage(
          l2TokenBridge.address,
          0, // fee
          l1TokenBridge.interface.encodeFunctionData(
            // calldata
            "completeBridging ",
            [L1DAI.address, 1, user.address, chainIds[1], encodedTokenMetadata],
          ),
        ),
      ).to.be.revertedWithCustomError(l1TokenBridge, "SenderNotAuthorized");
    });

    describe("setCustomContract", function () {
      it("Should bridge EIP712-compliant-token with permit", async function () {
        const {
          user,
          l1TokenBridge,
          l2TokenBridge,
          tokens: { L1DAI },
          chainIds,
        } = await loadFixture(deployContractsFixture);

        const l1Token = L1DAI;
        const bridgeAmount = 70;

        // Bridge token L1 to L2
        await l1TokenBridge.connect(user).bridgeToken(L1DAI.address, bridgeAmount, user.address);
        const l2TokenAddress = await l2TokenBridge.nativeToBridgedToken(chainIds[0], l1Token.address);
        const BridgedToken = await ethers.getContractFactory("BridgedToken");
        const l2Token = BridgedToken.attach(l2TokenAddress);

        // Check that no allowance exist for l2Token (User => l2TokenBridge)
        expect(await l2Token.allowance(user.address, l2TokenBridge.address)).to.be.equal(0);

        // Create EIP712-signature
        const deadline = ethers.constants.MaxUint256;
        const { chainId } = await ethers.provider.getNetwork();
        const nonce = await l2Token.nonces(user.address);
        expect(nonce).to.be.equal(0);

        // Try to bridge back without permit data
        await expect(
          l2TokenBridge.connect(user).bridgeToken(l2Token.address, bridgeAmount, user.address),
        ).to.be.revertedWith("ERC20: insufficient allowance");

        // Capture balances before bridging back
        const l1TokenUserBalanceBefore = await l1Token.balanceOf(user.address);
        const l2TokenUserBalanceBefore = await l2Token.balanceOf(user.address);

        // Prepare data for permit calldata
        const permitData = await getPermitData(
          user,
          l2Token,
          nonce,
          chainId,
          l2TokenBridge.address,
          bridgeAmount,
          deadline,
        );

        // Bridge back
        await l2TokenBridge
          .connect(user)
          .bridgeTokenWithPermit(l2Token.address, bridgeAmount, user.address, permitData);

        // Capture balances after bridging back
        const l1TokenUserBalanceAfter = await l1Token.balanceOf(user.address);
        const l2TokenUserBalanceAfter = await l2Token.balanceOf(user.address);

        const diffL1UserBalance = l1TokenUserBalanceAfter.sub(l1TokenUserBalanceBefore);
        const diffL2UserBalance = l2TokenUserBalanceBefore.sub(l2TokenUserBalanceAfter);

        expect(diffL1UserBalance).to.be.equal(bridgeAmount);
        expect(diffL2UserBalance).to.be.equal(bridgeAmount);
      });
    });

    describe("setCustomContract", function () {
      it("Should revert if setCustomContract is not called by the owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
        await expect(l1TokenBridge.connect(user).setCustomContract(CUSTOM_ADDRESS, CUSTOM_ADDRESS)).to.be.revertedWith(
          "Ownable: caller is not the owner",
        );
      });

      it("Should revert if a native token has already been bridged", async function () {
        const {
          user,
          owner,
          l1TokenBridge,
          l2TokenBridge,
          tokens: { L1DAI },
          chainIds,
        } = await loadFixture(deployContractsFixture);
        // First bridge token (user has L1DAI balance set in the fixture)
        await l1TokenBridge.connect(user).bridgeToken(L1DAI.address, 1, user.address);
        const l2TokenAddress = await l2TokenBridge.nativeToBridgedToken(chainIds[0], L1DAI.address);
        await expect(
          l1TokenBridge.connect(owner).setCustomContract(L1DAI.address, CUSTOM_ADDRESS),
        ).to.be.revertedWithCustomError(l1TokenBridge, "AlreadyBridgedToken");
        await expect(
          l2TokenBridge.connect(owner).setCustomContract(l2TokenAddress, CUSTOM_ADDRESS),
        ).to.be.revertedWithCustomError(l1TokenBridge, "AlreadyBridgedToken");
      });
    });

    describe("Pause / unpause", function () {
      it("Should pause the contract when pause() is called", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pause();
        expect(await l1TokenBridge.paused()).to.equal(true);
      });

      it("Should unpause the contract when unpause() is called", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pause();

        await l1TokenBridge.connect(owner).unpause();

        expect(await l1TokenBridge.paused()).to.equal(false);
      });
      it("Should revert bridgeToken if paused", async function () {
        const {
          owner,
          l1TokenBridge,
          tokens: { L1DAI },
        } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pause();
        await expect(l1TokenBridge.bridgeToken(L1DAI.address, 10, owner.address, [])).to.be.revertedWith(
          "Pausable: paused",
        );
      });
      it("Should allow bridgeToken if unpaused", async function () {
        const {
          owner,
          user,
          l1TokenBridge,
          tokens: { L1DAI },
        } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pause();
        await l1TokenBridge.connect(owner).unpause();
        await l1TokenBridge.connect(user).bridgeToken(L1DAI.address, 10, user.address, []);
      });
    });

    describe("Owner", function () {
      it("Should revert if setReservedToken is called by a non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
        await expect(l1TokenBridge.connect(user).setReserved(user.address)).to.be.revertedWith(
          "Ownable: caller is not the owner",
        );
      });
      it("Should revert if pause() is called by a non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await expect(l1TokenBridge.connect(user).pause()).to.be.revertedWith("Ownable: caller is not the owner");
      });
      it("Should revert if unpause() is called by a non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await expect(l1TokenBridge.connect(user).unpause()).to.be.revertedWith("Ownable: caller is not the owner");
      });
      it("Should revert if transferOwnership is called by a non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
        await expect(l1TokenBridge.connect(user).transferOwnership(user.address)).to.be.revertedWith(
          "Ownable: caller is not the owner",
        );
      });
      it("Should revert if removeReserved is called by a non-owner", async function () {
        const {
          user,
          l1TokenBridge,
          tokens: { L1DAI },
        } = await loadFixture(deployContractsFixture);
        await expect(l1TokenBridge.connect(user).removeReserved(L1DAI.address)).to.be.revertedWith(
          "Ownable: caller is not the owner",
        );
      });
    });
  });

  describe("Reserved tokens", function () {
    it("Should be possible for the admin to reserve a token", async function () {
      const {
        owner,
        l1TokenBridge,
        tokens: { L1DAI },
        chainIds,
      } = await loadFixture(deployContractsFixture);
      await expect(l1TokenBridge.connect(owner).setReserved(L1DAI.address)).not.to.be.revertedWith(
        "TokenBridge: token already bridged",
      );
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], L1DAI.address)).to.be.equal(RESERVED_STATUS);
    });

    it("Should be possible for the admin to remove token from the reserved list", async function () {
      // @TODO this test can probably be rewritten, avoiding to set the token as reserved in the first place
      const {
        owner,
        l1TokenBridge,
        tokens: { L1DAI },
        chainIds,
      } = await loadFixture(deployContractsFixture);
      await l1TokenBridge.connect(owner).setReserved(L1DAI.address);
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], L1DAI.address)).to.be.equal(RESERVED_STATUS);
      await l1TokenBridge.connect(owner).removeReserved(L1DAI.address);
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], L1DAI.address)).to.be.equal(
        ethers.constants.AddressZero,
      );
    });

    it("Should not be possible to bridge reserved tokens", async function () {
      const {
        owner,
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await l1TokenBridge.connect(owner).setReserved(L1DAI.address);
      await expect(
        l1TokenBridge.connect(user).bridgeToken(L1DAI.address, 1, user.address),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ReservedToken");
    });

    it("Should only be possible to reserve a token if it has not been bridged before", async function () {
      const {
        owner,
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      l1TokenBridge.connect(user).bridgeToken(L1DAI.address, 1, user.address);
      await expect(l1TokenBridge.connect(owner).setReserved(L1DAI.address)).to.be.revertedWithCustomError(
        l1TokenBridge,
        "AlreadyBridgedToken",
      );
    });

    it("Should set reserved tokens in the initializer", async function () {
      const { chainIds } = await loadFixture(deployContractsFixture);
      const TokenBridgeFactory = await ethers.getContractFactory("TokenBridge");
      const l1TokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
        PLACEHOLDER_ADDRESS, // owner
        PLACEHOLDER_ADDRESS, // messageService
        PLACEHOLDER_ADDRESS, // tokenBeacon
        chainIds[0],
        chainIds[1],
        [CUSTOM_ADDRESS], // reservedTokens
      ]);
      await l1TokenBridge.deployed();
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], CUSTOM_ADDRESS)).to.be.equal(RESERVED_STATUS);
    });

    it("Should only be possible to call removeReserved if the token is in the reserved list", async function () {
      const {
        owner,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await expect(l1TokenBridge.connect(owner).removeReserved(L1DAI.address)).to.be.revertedWithCustomError(
        l1TokenBridge,
        "NotReserved",
      );
    });

    it("Should revert if token is the 0 address", async function () {
      const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge.connect(owner).setReserved(ethers.constants.AddressZero),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");
    });
  });

  describe("bridgeTokenWithPermit", function () {
    it("Should revert if contract is paused", async function () {
      const {
        user,
        tokens: { L1DAI },
        l1TokenBridge,
      } = await loadFixture(deployContractsFixture);
      await l1TokenBridge.pause();
      await expect(
        l1TokenBridge.connect(user).bridgeTokenWithPermit(L1DAI.address, 1, user.address, EMPTY_PERMIT_DATA),
      ).to.be.revertedWith("Pausable: paused");
    });

    it("Should revert if token is the 0 address", async function () {
      const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge
          .connect(user)
          .bridgeTokenWithPermit(ethers.constants.AddressZero, 1, user.address, EMPTY_PERMIT_DATA),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");
    });

    it("Should revert if token is the amount is 0", async function () {
      const {
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge.connect(user).bridgeTokenWithPermit(L1DAI.address, 0, user.address, EMPTY_PERMIT_DATA),
      ).to.be.revertedWithCustomError(l1TokenBridge, "ZeroAmountNotAllowed");
    });

    it("Should not revert if permitData is empty", async function () {
      const {
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge.connect(user).bridgeTokenWithPermit(L1DAI.address, 10, user.address, EMPTY_PERMIT_DATA),
      ).to.be.not.reverted;
    });

    it("Should revert if permitData is invalid", async function () {
      const {
        owner,
        user,
        l1TokenBridge,
        l2TokenBridge,
        tokens: { L1DAI },
        chainIds,
      } = await loadFixture(deployContractsFixture);
      // Test when the permitData has an invalid format
      await expect(
        l1TokenBridge.connect(user).bridgeTokenWithPermit(L1DAI.address, 10, user.address, "0x111111111111"),
      ).to.be.revertedWithCustomError(l1TokenBridge, "InvalidPermitData");

      // Test when the spender passed is invalid
      // Prepare data for permit calldata
      const bridgeAmount = 70;

      // Bridge token L1 to L2
      await l1TokenBridge.connect(user).bridgeToken(L1DAI.address, bridgeAmount, user.address);
      const l2TokenAddress = await l2TokenBridge.nativeToBridgedToken(chainIds[0], L1DAI.address);
      const BridgedToken = await ethers.getContractFactory("BridgedToken");
      const l2Token = BridgedToken.attach(l2TokenAddress);

      // Create EIP712-signature
      const deadline = ethers.constants.MaxUint256;
      const { chainId } = await ethers.provider.getNetwork();
      const nonce = await l2Token.nonces(user.address);

      let permitData = await getPermitData(user, l2Token, nonce, chainId, user.address, bridgeAmount, deadline);
      await expect(
        l2TokenBridge.connect(user).bridgeTokenWithPermit(L1DAI.address, bridgeAmount, user.address, permitData),
      ).to.be.revertedWithCustomError(l1TokenBridge, "PermitNotAllowingBridge");

      // Test when the sender is not the owner of the tokens
      permitData = await getPermitData(user, l2Token, nonce, chainId, l2TokenBridge.address, bridgeAmount, deadline);
      await expect(
        l1TokenBridge.connect(owner).bridgeTokenWithPermit(L1DAI.address, bridgeAmount, user.address, permitData),
      ).to.be.revertedWithCustomError(l1TokenBridge, "PermitNotFromSender");
    });
  });

  describe("bridgeToken", function () {
    it("Should not emit event NewToken if the token has already been bridged once", async function () {
      const {
        user,
        tokens: { L1DAI },
        l1TokenBridge,
      } = await loadFixture(deployContractsFixture);
      await expect(l1TokenBridge.connect(user).bridgeToken(L1DAI.address, 1, user.address)).to.emit(
        l1TokenBridge,
        "NewToken",
      );
      await expect(l1TokenBridge.connect(user).bridgeToken(L1DAI.address, 1, user.address)).to.not.emit(
        l1TokenBridge,
        "NewToken",
      );
    });

    it("Should revert if recipient is set at 0 address", async function () {
      const {
        user,
        tokens: { L1DAI },
        l1TokenBridge,
      } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge.connect(user).bridgeToken(L1DAI.address, 1, ethers.constants.AddressZero),
      ).to.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");
    });

    it("Should not be able to call bridgeToken by reentrancy", async function () {
      const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

      const ReentrancyContract = await ethers.getContractFactory("ReentrancyContract");
      const reentrancyContract = await ReentrancyContract.deploy(l1TokenBridge.address);

      const MaliciousERC777 = await ethers.getContractFactory("MaliciousERC777");
      const maliciousERC777 = await MaliciousERC777.deploy(reentrancyContract.address);
      await maliciousERC777.mint(reentrancyContract.address, 100);
      await maliciousERC777.mint(owner.address, 100);

      await reentrancyContract.setToken(maliciousERC777.address);

      await expect(l1TokenBridge.bridgeToken(maliciousERC777.address, 1, owner.address)).to.be.revertedWith(
        "ReentrancyGuard: reentrant call",
      );
    });
  });

  describe("setRemoteTokenBridge", function () {
    it("Should revert if remoteTokenBridge has not been initialized", async function () {
      const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expect(l1TokenBridge.connect(owner).setRemoteTokenBridge(l1TokenBridge.address)).to.revertedWithCustomError(
        l1TokenBridge,
        "RemoteTokenBridgeAlreadySet",
      );
    });

    it("Should revert if called by non-owner", async function () {
      const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expect(l1TokenBridge.connect(user).setRemoteTokenBridge(l1TokenBridge.address)).to.revertedWith(
        "Ownable: caller is not the owner",
      );
    });
  });

  describe("setDeployed", function () {
    it("Should revert if not called by the messageService", async function () {
      const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expect(l1TokenBridge.connect(user).setDeployed([])).to.be.revertedWithCustomError(
        l1TokenBridge,
        "CallerIsNotMessageService",
      );
    });

    it("Should revert if message does not come from the remote Token Bridge", async function () {
      const { user, messageService, l1TokenBridge, l2TokenBridge } = await loadFixture(deployContractsFixture);

      await expect(
        messageService.connect(user).sendMessage(
          l2TokenBridge.address,
          0, // fee
          l1TokenBridge.interface.encodeFunctionData(
            // calldata
            "setDeployed",
            [[]],
          ),
        ),
      ).to.be.revertedWithCustomError(l1TokenBridge, "SenderNotAuthorized");
    });
  });
});
