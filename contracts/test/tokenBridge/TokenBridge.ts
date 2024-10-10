import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers, upgrades } from "hardhat";
import { deployTokenBridgeWithMockMessaging } from "../../scripts/tokenBridge/test/deployTokenBridges";
import { deployTokens } from "../../scripts/tokenBridge/test/deployTokens";
import { BridgedToken, TestTokenBridge, TokenBridge, ProxyAdmin } from "../../typechain-types";
import { getPermitData } from "./utils/permitHelper";
import { Contract, ZeroAddress } from "ethers";
import {
  ADDRESS_ZERO,
  COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE,
  INITIALIZED_ALREADY_MESSAGE,
  INITIATE_TOKEN_BRIDGING_PAUSE_TYPE,
  PAUSE_INITIATE_TOKEN_BRIDGING_ROLE,
  REMOVE_RESERVED_TOKEN_ROLE,
  SET_CUSTOM_CONTRACT_ROLE,
  SET_MESSAGE_SERVICE_ROLE,
  SET_REMOTE_TOKENBRIDGE_ROLE,
  SET_RESERVED_TOKEN_ROLE,
  UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE,
  pauseTypeRoles,
  unpauseTypeRoles,
  PAUSE_ALL_ROLE,
  UNPAUSE_ALL_ROLE,
  HASH_ZERO,
} from "../common/constants";
import {
  buildAccessErrorMessage,
  expectEvent,
  expectRevertWithCustomError,
  expectRevertWithReason,
} from "../common/helpers";

const initialUserBalance = BigInt(10 ** 9);
const mockName = "L1 DAI";
const mockSymbol = "L1DAI";
const mockDecimals = 18;
const RESERVED_STATUS = ethers.getAddress("0x0000000000000000000000000000000000000111");
const PLACEHOLDER_ADDRESS = ethers.getAddress("0x5555555555555555555555555555555555555555");
const CUSTOM_ADDRESS = ethers.getAddress("0x9999999999999999999999999999999999999999");
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
        bridgeAddress = await deploymentFixture.l1TokenBridge.getAddress();
      }
      if ((await token.name()).includes("L2")) {
        bridgeAddress = await deploymentFixture.l2TokenBridge.getAddress();
      }

      await token.connect(user).approve(bridgeAddress!, ethers.MaxUint256);
    }
    const encodedTokenMetadata = ethers.AbiCoder.defaultAbiCoder().encode(
      ["string", "string", "uint8"],
      [mockName, mockSymbol, mockDecimals],
    );
    return { owner, user, ...deploymentFixture, tokens, encodedTokenMetadata };
  }

  describe("initialize", async function () {
    it("Should revert if it has already been intialized", async function () {
      const { user, l1TokenBridge, chainIds } = await loadFixture(deployContractsFixture);
      await expectRevertWithReason(
        l1TokenBridge.connect(user).initialize({
          defaultAdmin: PLACEHOLDER_ADDRESS,
          messageService: PLACEHOLDER_ADDRESS,
          tokenBeacon: PLACEHOLDER_ADDRESS,
          sourceChainId: chainIds[0],
          targetChainId: chainIds[1],
          reservedTokens: [],
          roleAddresses: [],
          pauseTypeRoles: [],
          unpauseTypeRoles: [],
        }),
        INITIALIZED_ALREADY_MESSAGE,
      );
    });

    it("Should revert if one of the initializing parameters is address 0", async function () {
      const { user, owner, chainIds } = await loadFixture(deployContractsFixture);
      const TokenBridge = await ethers.getContractFactory("TokenBridge");

      await expectRevertWithCustomError(
        TokenBridge,
        upgrades.deployProxy(TokenBridge, [
          {
            defaultAdmin: PLACEHOLDER_ADDRESS,
            messageService: ADDRESS_ZERO,
            tokenBeacon: PLACEHOLDER_ADDRESS,
            sourceChainId: chainIds[0],
            targetChainId: chainIds[1],
            reservedTokens: [],
            roleAddresses: [],
            pauseTypeRoles: [],
            unpauseTypeRoles: [],
          },
        ]),
        "ZeroAddressNotAllowed",
      );

      await expectRevertWithCustomError(
        TokenBridge,
        upgrades.deployProxy(TokenBridge, [
          {
            defaultAdmin: PLACEHOLDER_ADDRESS,
            messageService: PLACEHOLDER_ADDRESS,
            tokenBeacon: ADDRESS_ZERO,
            sourceChainId: chainIds[0],
            targetChainId: chainIds[1],
            reservedTokens: [],
            roleAddresses: [],
            pauseTypeRoles: [],
            unpauseTypeRoles: [],
          },
        ]),
        "ZeroAddressNotAllowed",
      );

      await expectRevertWithCustomError(
        TokenBridge,
        upgrades.deployProxy(TokenBridge, [
          {
            defaultAdmin: PLACEHOLDER_ADDRESS,
            messageService: PLACEHOLDER_ADDRESS,
            tokenBeacon: PLACEHOLDER_ADDRESS,
            sourceChainId: chainIds[0],
            targetChainId: chainIds[1],
            reservedTokens: [PLACEHOLDER_ADDRESS, ADDRESS_ZERO],
            roleAddresses: [
              { addressWithRole: user.address, role: SET_RESERVED_TOKEN_ROLE },
              { addressWithRole: owner.address, role: SET_RESERVED_TOKEN_ROLE },
            ],
            pauseTypeRoles: pauseTypeRoles,
            unpauseTypeRoles: unpauseTypeRoles,
          },
        ]),
        "ZeroAddressNotAllowed",
      );

      await expectRevertWithCustomError(
        TokenBridge,
        upgrades.deployProxy(TokenBridge, [
          {
            defaultAdmin: PLACEHOLDER_ADDRESS,
            messageService: PLACEHOLDER_ADDRESS,
            tokenBeacon: PLACEHOLDER_ADDRESS,
            sourceChainId: chainIds[0],
            targetChainId: chainIds[1],
            reservedTokens: [PLACEHOLDER_ADDRESS],
            roleAddresses: [
              { addressWithRole: ADDRESS_ZERO, role: SET_RESERVED_TOKEN_ROLE },
              { addressWithRole: owner.address, role: SET_RESERVED_TOKEN_ROLE },
            ],
            pauseTypeRoles: pauseTypeRoles,
            unpauseTypeRoles: unpauseTypeRoles,
          },
        ]),
        "ZeroAddressNotAllowed",
      );
    });
    it("Should return 'NOT_VALID_ENCODING' for invalid data in _returnDataToString", async function () {
      const TestTokenBridgeFactory = await ethers.getContractFactory("TestTokenBridge");

      const initData = {
        defaultAdmin: PLACEHOLDER_ADDRESS,
        messageService: PLACEHOLDER_ADDRESS,
        tokenBeacon: PLACEHOLDER_ADDRESS,
        sourceChainId: 5,
        targetChainId: 51940,
        reservedTokens: [],
        roleAddresses: [],
        pauseTypeRoles: [],
        unpauseTypeRoles: [],
      };

      const l1TestTokenBridge = (await upgrades.deployProxy(TestTokenBridgeFactory, [
        initData,
      ])) as unknown as TestTokenBridge;
      await l1TestTokenBridge.waitForDeployment();

      // Test case 1: Data length is not 32 and less than 64
      const invalidData1 = ethers.hexlify(ethers.randomBytes(33)); // 33 bytes
      expect(await l1TestTokenBridge.testReturnDataToString(invalidData1)).to.equal("NOT_VALID_ENCODING");

      // Test case 2: Data length is 32 but starts with a zero byte
      const invalidData2 = ethers.concat([
        ethers.hexlify(new Uint8Array(1)), // One zero byte
        ethers.hexlify(ethers.randomBytes(31)), // 31 random bytes
      ]);
      expect(await l1TestTokenBridge.testReturnDataToString(invalidData2)).to.equal("NOT_VALID_ENCODING");

      // Test case 3: Valid data for comparison
      const validString = "ValidString";
      const encodedValidData = ethers.AbiCoder.defaultAbiCoder().encode(["string"], [validString]);
      expect(await l1TestTokenBridge.testReturnDataToString(encodedValidData)).to.equal(validString);
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

      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge
          .connect(user)
          .completeBridging(await L1DAI.getAddress(), 1, user.address, chainIds[1], encodedTokenMetadata),
        "CallerIsNotMessageService",
      );
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

      const sendCalldata = messageService
        .connect(user)
        .sendMessage(
          await l2TokenBridge.getAddress(),
          0,
          l1TokenBridge.interface.encodeFunctionData("completeBridging", [
            await L1DAI.getAddress(),
            1,
            user.address,
            chainIds[1],
            encodedTokenMetadata,
          ]),
        );
      await expectRevertWithCustomError(l1TokenBridge, sendCalldata, "SenderNotAuthorized");
    });

    describe("setMessageService", function () {
      it("Should revert if trying to set message service to zero address", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await expectRevertWithCustomError(
          l1TokenBridge,
          l1TokenBridge.connect(owner).setMessageService(ADDRESS_ZERO),
          "ZeroAddressNotAllowed",
        );
      });

      it("Should revert if called by non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await expectRevertWithReason(
          l1TokenBridge.connect(user).setMessageService(PLACEHOLDER_ADDRESS),
          buildAccessErrorMessage(user, SET_MESSAGE_SERVICE_ROLE),
        );
      });

      it("Should successfully set new message service address", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);
        const newMessageServiceAddress = ethers.Wallet.createRandom().address;

        await expect(l1TokenBridge.connect(owner).setMessageService(newMessageServiceAddress))
          .to.emit(l1TokenBridge, "MessageServiceUpdated")
          .withArgs(newMessageServiceAddress, await l1TokenBridge.messageService(), owner.address);

        expect(await l1TokenBridge.messageService()).to.equal(newMessageServiceAddress);
      });
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
        await l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), bridgeAmount, user.address);
        const l2TokenAddress = await l2TokenBridge.nativeToBridgedToken(chainIds[0], await l1Token.getAddress());
        const BridgedToken = await ethers.getContractFactory("BridgedToken");
        const l2Token = BridgedToken.attach(l2TokenAddress) as BridgedToken;

        // Check that no allowance exist for l2Token (User => l2TokenBridge)
        expect(await l2Token.allowance(user.address, await l2TokenBridge.getAddress())).to.be.equal(0);

        // Create EIP712-signature
        const deadline = ethers.MaxUint256;
        const { chainId } = await ethers.provider.getNetwork();
        const nonce = await l2Token.nonces(user.address);
        expect(nonce).to.be.equal(0);

        // Try to bridge back without permit data
        await expectRevertWithReason(
          l2TokenBridge.connect(user).bridgeToken(await l2Token.getAddress(), bridgeAmount, user.address),
          "ERC20: insufficient allowance",
        );

        // Capture balances before bridging back
        const l1TokenUserBalanceBefore = await l1Token.balanceOf(user.address);
        const l2TokenUserBalanceBefore = await l2Token.balanceOf(user.address);

        // Prepare data for permit calldata
        const permitData = await getPermitData(
          user,
          l2Token,
          nonce,
          parseInt(chainId.toString()),
          await l2TokenBridge.getAddress(),
          bridgeAmount,
          deadline,
        );

        // Bridge back
        await l2TokenBridge
          .connect(user)
          .bridgeTokenWithPermit(await l2Token.getAddress(), bridgeAmount, user.address, permitData);

        // Capture balances after bridging back
        const l1TokenUserBalanceAfter = await l1Token.balanceOf(user.address);
        const l2TokenUserBalanceAfter = await l2Token.balanceOf(user.address);

        const diffL1UserBalance = l1TokenUserBalanceAfter - l1TokenUserBalanceBefore;
        const diffL2UserBalance = l2TokenUserBalanceBefore - l2TokenUserBalanceAfter;

        expect(diffL1UserBalance).to.be.equal(bridgeAmount);
        expect(diffL2UserBalance).to.be.equal(bridgeAmount);
      });

      it("Should revert if setCustomContract is not called by the owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
        await expectRevertWithReason(
          l1TokenBridge.connect(user).setCustomContract(CUSTOM_ADDRESS, CUSTOM_ADDRESS),
          buildAccessErrorMessage(user, SET_CUSTOM_CONTRACT_ROLE),
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
        const L1DAIAddress = await L1DAI.getAddress();
        // First bridge token (user has L1DAI balance set in the fixture)
        await l1TokenBridge.connect(user).bridgeToken(L1DAIAddress, 1, user.address);
        const l2TokenAddress = await l2TokenBridge.nativeToBridgedToken(chainIds[0], L1DAIAddress);

        await expectRevertWithCustomError(
          l1TokenBridge,
          l1TokenBridge.connect(owner).setCustomContract(L1DAIAddress, CUSTOM_ADDRESS),
          "AlreadyBridgedToken",
          [L1DAIAddress],
        );

        await expectRevertWithCustomError(
          l2TokenBridge,
          l2TokenBridge.connect(owner).setCustomContract(l2TokenAddress, CUSTOM_ADDRESS),
          "AlreadyBridgedToken",
          [l2TokenAddress],
        );
      });

      it("Should revert if _nativeToken is zero address", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await expectRevertWithCustomError(
          l1TokenBridge,
          l1TokenBridge.connect(owner).setCustomContract(ADDRESS_ZERO, CUSTOM_ADDRESS),
          "ZeroAddressNotAllowed",
        );
      });

      it("Should revert if _targetContract is zero address", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);
        const validNativeToken = ethers.Wallet.createRandom().address;

        await expectRevertWithCustomError(
          l1TokenBridge,
          l1TokenBridge.connect(owner).setCustomContract(validNativeToken, ADDRESS_ZERO),
          "ZeroAddressNotAllowed",
        );
      });

      it("Should successfully set custom contract for valid addresses", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);
        const validNativeToken = ethers.Wallet.createRandom().address;
        const validTargetContract = ethers.Wallet.createRandom().address;

        await expect(l1TokenBridge.connect(owner).setCustomContract(validNativeToken, validTargetContract))
          .to.emit(l1TokenBridge, "CustomContractSet")
          .withArgs(validNativeToken, validTargetContract, owner.address);
      });
    });

    describe("Pause / unpause", function () {
      it("Should pause the contract when INITIATE_TOKEN_BRIDGING_PAUSE_TYPE() is called", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);
        expect(await l1TokenBridge.isPaused(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE)).to.equal(true);
      });

      it("Should unpause the contract when unpause() is called", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);

        await l1TokenBridge.connect(owner).unPauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);

        expect(await l1TokenBridge.isPaused(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE)).to.equal(false);
      });
      it("Should revert bridgeToken if paused", async function () {
        const {
          owner,
          l1TokenBridge,
          tokens: { L1DAI },
        } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);

        await expectRevertWithCustomError(
          l1TokenBridge,
          l1TokenBridge.bridgeToken(await L1DAI.getAddress(), 10, owner.address),
          "IsPaused",
          [INITIATE_TOKEN_BRIDGING_PAUSE_TYPE],
        );
      });
      it("Should allow bridgeToken if unpaused", async function () {
        const {
          owner,
          user,
          l1TokenBridge,
          tokens: { L1DAI },
        } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);
        await l1TokenBridge.connect(owner).unPauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);
        await l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), 10, user.address);
      });
      // TODO: COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE tests
      it("Should pause the contract when pause() is called", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pauseByType(COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE);
        expect(await l1TokenBridge.isPaused(COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE)).to.equal(true);
      });

      it("Should unpause the contract when unpause() is called", async function () {
        const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

        await l1TokenBridge.connect(owner).pauseByType(COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE);

        await l1TokenBridge.connect(owner).unPauseByType(COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE);

        expect(await l1TokenBridge.isPaused(COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE)).to.equal(false);
      });

      it("Should emit BridgingInitiatedV2 when bridging", async function () {
        const {
          user,
          l1TokenBridge,
          tokens: { L1DAI },
        } = await loadFixture(deployContractsFixture);

        const L1DAIAddress = L1DAI.getAddress();

        await expectEvent(
          l1TokenBridge,
          l1TokenBridge.connect(user).bridgeToken(L1DAIAddress, BigInt(10), user.address),
          "BridgingInitiatedV2",
          [user.address, user.address, L1DAIAddress, 10],
        );
      });

      it("Should emit BridgingFinalizedV2 event when bridging is complete", async function () {
        const {
          user,
          l1TokenBridge,
          l2TokenBridge,
          tokens: { L1DAI },
          chainIds,
        } = await loadFixture(deployContractsFixture);
        const bridgeAmount = 10;

        const L1DAIAddress = L1DAI.getAddress();

        // const initialAmount = await L1DAI.balanceOf(user.address);
        await l1TokenBridge.connect(user).bridgeToken(L1DAIAddress, bridgeAmount, user.address);
        const L2DAIBridgedAddress = await l2TokenBridge.nativeToBridgedToken(chainIds[0], L1DAIAddress);
        await l2TokenBridge.confirmDeployment([L2DAIBridgedAddress]);

        const bridgedToken = await l2TokenBridge.nativeToBridgedToken(chainIds[0], L1DAIAddress);

        const abi = [
          "event BridgingFinalizedV2(address indexed nativeToken,address indexed bridgedToken,uint256 amount,address indexed recipient)",
        ];

        const contract = new Contract(await l2TokenBridge.getAddress(), abi, ethers.provider);
        // Filtering for indexed fields by default validates they are correct when events are not null
        const filteredEvents = contract.filters.BridgingFinalizedV2(L1DAIAddress, bridgedToken, null, user.address);

        const events = await l2TokenBridge.queryFilter(filteredEvents);
        expect(events).to.not.be.null;
        expect(events).to.not.be.empty;
        expect(events[0].args?.[2]).to.equal(10);
      });
    });

    describe("Owner", function () {
      it("Should revert if setReservedToken is called by a non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
        await expectRevertWithReason(
          l1TokenBridge.connect(user).setReserved(user.address),
          buildAccessErrorMessage(user, SET_RESERVED_TOKEN_ROLE),
        );
      });
      it("Should revert if pause() is called by a non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
        await expectRevertWithReason(
          l1TokenBridge.connect(user).pauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE),
          buildAccessErrorMessage(user, PAUSE_INITIATE_TOKEN_BRIDGING_ROLE),
        );
      });
      it("Should revert if unpause() is called by a non-owner", async function () {
        const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
        await expectRevertWithReason(
          l1TokenBridge.connect(user).unPauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE),
          buildAccessErrorMessage(user, UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE),
        );
      });
      it("Should revert if removeReserved is called by a non-owner", async function () {
        const {
          user,
          l1TokenBridge,
          tokens: { L1DAI },
        } = await loadFixture(deployContractsFixture);
        await expectRevertWithReason(
          l1TokenBridge.connect(user).removeReserved(await L1DAI.getAddress()),
          buildAccessErrorMessage(user, REMOVE_RESERVED_TOKEN_ROLE),
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
      await expect(l1TokenBridge.connect(owner).setReserved(await L1DAI.getAddress())).not.to.be.revertedWith(
        "TokenBridge: token already bridged",
      );
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], await L1DAI.getAddress())).to.be.equal(
        RESERVED_STATUS,
      );
    });

    it("Should be possible for the admin to remove token from the reserved list", async function () {
      // @TODO this test can probably be rewritten, avoiding to set the token as reserved in the first place

      const {
        owner,
        l1TokenBridge,
        tokens: { L1DAI },
        chainIds,
      } = await loadFixture(deployContractsFixture);

      const L1DAIAddress = await L1DAI.getAddress();

      await l1TokenBridge.connect(owner).setReserved(L1DAIAddress);
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], L1DAIAddress)).to.be.equal(RESERVED_STATUS);
      await expect(l1TokenBridge.connect(owner).removeReserved(L1DAIAddress))
        .to.emit(l1TokenBridge, "ReservationRemoved")
        .withArgs(L1DAIAddress);
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], L1DAIAddress)).to.be.equal(ADDRESS_ZERO);
    });

    it("Should not be possible to bridge reserved tokens", async function () {
      const {
        owner,
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await l1TokenBridge.connect(owner).setReserved(await L1DAI.getAddress());

      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), 1, user.address),
        "ReservedToken",
        [L1DAI.getAddress()],
      );
    });

    it("Should only be possible to reserve a token if it has not been bridged before", async function () {
      const {
        owner,
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), 1, user.address);

      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(owner).setReserved(await L1DAI.getAddress()),
        "AlreadyBridgedToken",
        [L1DAI.getAddress()],
      );
    });

    it("Should set reserved tokens in the initializer", async function () {
      const { chainIds } = await loadFixture(deployContractsFixture);
      const TokenBridgeFactory = await ethers.getContractFactory("TokenBridge");
      const l1TokenBridge = await upgrades.deployProxy(TokenBridgeFactory, [
        {
          defaultAdmin: PLACEHOLDER_ADDRESS,
          messageService: PLACEHOLDER_ADDRESS,
          tokenBeacon: PLACEHOLDER_ADDRESS,
          sourceChainId: chainIds[0],
          targetChainId: chainIds[1],
          reservedTokens: [CUSTOM_ADDRESS],
          roleAddresses: [],
          pauseTypeRoles: pauseTypeRoles,
          unpauseTypeRoles: unpauseTypeRoles,
        },
      ]);
      await l1TokenBridge.waitForDeployment();
      expect(await l1TokenBridge.nativeToBridgedToken(chainIds[0], CUSTOM_ADDRESS)).to.be.equal(RESERVED_STATUS);
    });

    it("Should only be possible to call removeReserved if the token is in the reserved list", async function () {
      const {
        owner,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);

      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(owner).removeReserved(await L1DAI.getAddress()),
        "NotReserved",
        [L1DAI.getAddress()],
      );
    });

    it("Should revert if token is the 0 address", async function () {
      const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(owner).setReserved(ADDRESS_ZERO),
        "ZeroAddressNotAllowed",
      );
    });
    it("Should revert if trying to remove reservation for zero address", async function () {
      const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(owner).removeReserved(ADDRESS_ZERO),
        "ZeroAddressNotAllowed",
      );
    });
  });

  describe("bridgeTokenWithPermit", function () {
    it("Should revert if contract is paused", async function () {
      const {
        user,
        tokens: { L1DAI },
        l1TokenBridge,
      } = await loadFixture(deployContractsFixture);
      await l1TokenBridge.pauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);
      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(user).bridgeTokenWithPermit(await L1DAI.getAddress(), 1, user.address, EMPTY_PERMIT_DATA),
        "IsPaused",
        [INITIATE_TOKEN_BRIDGING_PAUSE_TYPE],
      );
    });

    it("Should revert if token is the 0 address", async function () {
      const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(user).bridgeTokenWithPermit(ADDRESS_ZERO, 1, user.address, EMPTY_PERMIT_DATA),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert if token is the amount is 0", async function () {
      const {
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(user).bridgeTokenWithPermit(await L1DAI.getAddress(), 0, user.address, EMPTY_PERMIT_DATA),
        "ZeroAmountNotAllowed",
        [0],
      );
    });

    it("Should not revert if permitData is empty", async function () {
      const {
        user,
        l1TokenBridge,
        tokens: { L1DAI },
      } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge
          .connect(user)
          .bridgeTokenWithPermit(await L1DAI.getAddress(), 10, user.address, EMPTY_PERMIT_DATA),
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
        l1TokenBridge.connect(user).bridgeTokenWithPermit(await L1DAI.getAddress(), 10, user.address, "0x111111111111"),
      ).to.be.revertedWithCustomError(l1TokenBridge, "InvalidPermitData");

      // Test when the spender passed is invalid
      // Prepare data for permit calldata
      const bridgeAmount = 70;

      // Bridge token L1 to L2
      await l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), bridgeAmount, user.address);
      const l2TokenAddress = await l2TokenBridge.nativeToBridgedToken(chainIds[0], await L1DAI.getAddress());
      const BridgedToken = await ethers.getContractFactory("BridgedToken");
      const l2Token = BridgedToken.attach(l2TokenAddress) as BridgedToken;

      // Create EIP712-signature
      const deadline = ethers.MaxUint256;
      const { chainId } = await ethers.provider.getNetwork();
      const nonce = await l2Token.nonces(user.address);

      let permitData = await getPermitData(
        user,
        l2Token,
        nonce,
        parseInt(chainId.toString()),
        user.address,
        bridgeAmount,
        deadline,
      );
      await expect(
        l2TokenBridge
          .connect(user)
          .bridgeTokenWithPermit(await L1DAI.getAddress(), bridgeAmount, user.address, permitData),
      ).to.be.revertedWithCustomError(l1TokenBridge, "PermitNotAllowingBridge");

      // Test when the sender is not the owner of the tokens
      permitData = await getPermitData(
        user,
        l2Token,
        nonce,
        parseInt(chainId.toString()),
        await l2TokenBridge.getAddress(),
        bridgeAmount,
        deadline,
      );
      await expect(
        l1TokenBridge
          .connect(owner)
          .bridgeTokenWithPermit(await L1DAI.getAddress(), bridgeAmount, user.address, permitData),
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
      await expect(l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), 1, user.address)).to.emit(
        l1TokenBridge,
        "NewToken",
      );
      await expect(l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), 1, user.address)).to.not.emit(
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
        l1TokenBridge.connect(user).bridgeToken(await L1DAI.getAddress(), 1, ADDRESS_ZERO),
      ).to.revertedWithCustomError(l1TokenBridge, "ZeroAddressNotAllowed");
    });

    it("Should not be able to call bridgeToken by reentrancy", async function () {
      const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);

      const ReentrancyContract = await ethers.getContractFactory("ReentrancyContract");
      const reentrancyContract = await ReentrancyContract.deploy(await l1TokenBridge.getAddress());

      const MaliciousERC777 = await ethers.getContractFactory("MaliciousERC777");
      const maliciousERC777 = await MaliciousERC777.deploy(await reentrancyContract.getAddress());
      await maliciousERC777.mint(await reentrancyContract.getAddress(), 100);
      await maliciousERC777.mint(owner.address, 100);

      await reentrancyContract.setToken(maliciousERC777.getAddress());

      await expectRevertWithReason(
        l1TokenBridge.bridgeToken(await maliciousERC777.getAddress(), 1, owner.address),
        "ReentrancyGuard: reentrant call",
      );
    });
  });

  describe("setRemoteTokenBridge", function () {
    it("Should revert if remoteTokenBridge has not been initialized", async function () {
      const { owner, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expect(
        l1TokenBridge.connect(owner).setRemoteTokenBridge(await l1TokenBridge.getAddress()),
      ).to.revertedWithCustomError(l1TokenBridge, "RemoteTokenBridgeAlreadySet");
    });

    it("Should revert if called by non-owner", async function () {
      const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);

      await expectRevertWithReason(
        l1TokenBridge.connect(user).setRemoteTokenBridge(await l1TokenBridge.getAddress()),
        buildAccessErrorMessage(user, SET_REMOTE_TOKENBRIDGE_ROLE),
      );
    });
  });

  describe("setDeployed", function () {
    it("Should revert if not called by the messageService", async function () {
      const { user, l1TokenBridge } = await loadFixture(deployContractsFixture);
      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.connect(user).setDeployed([]),
        "CallerIsNotMessageService",
      );
    });

    it("Should revert if message does not come from the remote Token Bridge", async function () {
      const { user, messageService, l1TokenBridge, l2TokenBridge } = await loadFixture(deployContractsFixture);
      await expectRevertWithCustomError(
        l2TokenBridge,
        messageService
          .connect(user)
          .sendMessage(
            await l2TokenBridge.getAddress(),
            0,
            l1TokenBridge.interface.encodeFunctionData("setDeployed", [[]]),
          ),
        "SenderNotAuthorized",
      );
    });
  });

  describe("reinitializePauseTypesAndPermissions", function () {
    it("Should revert with ZeroAddressNotAllowed when addressWithRole is zero address in reinitializePauseTypesAndPermissions", async function () {
      const { l1TokenBridge } = await loadFixture(deployContractsFixture);

      const roleAddresses = [{ addressWithRole: ZeroAddress, role: SET_RESERVED_TOKEN_ROLE }];

      await expectRevertWithCustomError(
        l1TokenBridge,
        l1TokenBridge.reinitializePauseTypesAndPermissions(roleAddresses, pauseTypeRoles, unpauseTypeRoles),
        "ZeroAddressNotAllowed",
      );
    });
  });

  describe("TokenBridge Upgradeable Tests", function () {
    let newRoleAddresses: { addressWithRole: string; role: string }[];

    before(async () => {
      const [owner] = await ethers.getSigners();
      newRoleAddresses = [
        { addressWithRole: owner.address, role: PAUSE_ALL_ROLE },
        { addressWithRole: owner.address, role: UNPAUSE_ALL_ROLE },
        { addressWithRole: owner.address, role: PAUSE_INITIATE_TOKEN_BRIDGING_ROLE },
        { addressWithRole: owner.address, role: UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE },
      ];
    });

    it("Should deploy and manually upgrade the TokenBridge contract", async function () {
      const { owner, messageService, chainIds } = await loadFixture(deployContractsFixture);

      const TokenBridgeV1 = await ethers.getContractFactory(
        "contracts/tokenBridge/mocks/TokenBridgeFlatten.sol:TokenBridgeFlatten",
      );
      const tokenBridgeV1 = (await upgrades.deployProxy(TokenBridgeV1, [
        owner.address,
        await messageService.getAddress(),
        PLACEHOLDER_ADDRESS,
        chainIds[0],
        chainIds[1],
        [],
      ])) as unknown as TokenBridge;
      await tokenBridgeV1.waitForDeployment();

      // Check initial state
      expect(await tokenBridgeV1.sourceChainId()).to.equal(chainIds[0]);
      const initialSourceChainId = await tokenBridgeV1.sourceChainId();
      const initialTargetChainId = await tokenBridgeV1.targetChainId();
      const initialMessageService = await tokenBridgeV1.messageService();
      const initialTokenBeacon = await tokenBridgeV1.tokenBeacon();

      // setCustomContract
      const nativeToken = ethers.Wallet.createRandom().address;
      const targetContract = ethers.Wallet.createRandom().address;
      await tokenBridgeV1.connect(owner).setCustomContract(nativeToken, targetContract);

      // Verify the custom contract was set
      expect(await tokenBridgeV1.nativeToBridgedToken(initialTargetChainId, nativeToken)).to.equal(targetContract);
      expect(await tokenBridgeV1.bridgedToNativeToken(targetContract)).to.equal(nativeToken);

      const proxyAdminAddress = await upgrades.erc1967.getAdminAddress(await tokenBridgeV1.getAddress());
      const ProxyAdminFactory = await ethers.getContractFactory("ProxyAdmin");
      const proxyAdmin = (await ProxyAdminFactory.attach(proxyAdminAddress)) as ProxyAdmin;

      const TokenBridgeV2 = await ethers.getContractFactory("TokenBridge");
      const tokenBridgeV2Implementation = await TokenBridgeV2.deploy();
      await tokenBridgeV2Implementation.waitForDeployment();

      const reinitializeCallData = TokenBridgeV2.interface.encodeFunctionData("reinitializePauseTypesAndPermissions", [
        newRoleAddresses,
        pauseTypeRoles,
        unpauseTypeRoles,
      ]);

      expect(await ethers.provider.getStorage(await tokenBridgeV1.getAddress(), 101)).to.equal(
        ethers.zeroPadValue(owner.address, 32), // _owner
      );
      expect(await ethers.provider.getStorage(await tokenBridgeV1.getAddress(), 213)).to.equal(
        ethers.zeroPadValue("0x01", 32), //_status
      );

      await proxyAdmin.upgradeAndCall(
        await tokenBridgeV1.getAddress(),
        await tokenBridgeV2Implementation.getAddress(),
        reinitializeCallData,
      );

      const upgradedTokenBridge = (await TokenBridgeV2.attach(await tokenBridgeV1.getAddress())) as TokenBridge;

      // Verify that the upgrade was successful
      expect(await upgradedTokenBridge.sourceChainId()).to.equal(initialSourceChainId);
      expect(await upgradedTokenBridge.targetChainId()).to.equal(initialTargetChainId);
      expect(await upgradedTokenBridge.messageService()).to.equal(initialMessageService);
      expect(await upgradedTokenBridge.tokenBeacon()).to.equal(initialTokenBeacon);

      // Verify that the custom contract setting was preserved after upgrade
      expect(await upgradedTokenBridge.nativeToBridgedToken(initialTargetChainId, nativeToken)).to.equal(
        targetContract,
      );
      expect(await upgradedTokenBridge.bridgedToNativeToken(targetContract)).to.equal(nativeToken);

      // Verify that _owner, _status, _paused, and _pendingOwner storage slots were wiped
      expect(await ethers.provider.getStorage(await upgradedTokenBridge.getAddress(), 101)).to.equal(HASH_ZERO); // _owner
      expect(await ethers.provider.getStorage(await upgradedTokenBridge.getAddress(), 213)).to.equal(HASH_ZERO); //_status
      expect(await ethers.provider.getStorage(await upgradedTokenBridge.getAddress(), 51)).to.equal(HASH_ZERO); //_paused
      expect(await ethers.provider.getStorage(await upgradedTokenBridge.getAddress(), 151)).to.equal(HASH_ZERO); // _pendingOwner

      expect(await upgradedTokenBridge.hasRole(PAUSE_ALL_ROLE, owner.address)).to.be.true;
      expect(await upgradedTokenBridge.hasRole(UNPAUSE_ALL_ROLE, owner.address)).to.be.true;
      expect(await upgradedTokenBridge.hasRole(PAUSE_INITIATE_TOKEN_BRIDGING_ROLE, owner.address)).to.be.true;
      expect(await upgradedTokenBridge.hasRole(UNPAUSE_INITIATE_TOKEN_BRIDGING_ROLE, owner.address)).to.be.true;

      expect(await upgradedTokenBridge.isPaused(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE)).to.be.false;
      expect(await upgradedTokenBridge.isPaused(COMPLETE_TOKEN_BRIDGING_PAUSE_TYPE)).to.be.false;

      await upgradedTokenBridge.connect(owner).pauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);
      expect(await upgradedTokenBridge.isPaused(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE)).to.be.true;

      await upgradedTokenBridge.connect(owner).unPauseByType(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE);
      expect(await upgradedTokenBridge.isPaused(INITIATE_TOKEN_BRIDGING_PAUSE_TYPE)).to.be.false;
    });

    it("Should set all permissions", async () => {
      const { owner, messageService, chainIds } = await loadFixture(deployContractsFixture);

      const TokenBridgeV1 = await ethers.getContractFactory(
        "contracts/tokenBridge/mocks/TokenBridgeFlatten.sol:TokenBridgeFlatten",
      );
      const tokenBridgeV1 = (await upgrades.deployProxy(TokenBridgeV1, [
        owner.address,
        await messageService.getAddress(),
        PLACEHOLDER_ADDRESS,
        chainIds[0],
        chainIds[1],
        [],
      ])) as unknown as TokenBridge;
      await tokenBridgeV1.waitForDeployment();

      // Check initial state
      expect(await tokenBridgeV1.sourceChainId()).to.equal(chainIds[0]);
      const initialTargetChainId = await tokenBridgeV1.targetChainId();

      // setCustomContract
      const nativeToken = ethers.Wallet.createRandom().address;
      const targetContract = ethers.Wallet.createRandom().address;
      await tokenBridgeV1.connect(owner).setCustomContract(nativeToken, targetContract);

      // Verify the custom contract was set
      expect(await tokenBridgeV1.nativeToBridgedToken(initialTargetChainId, nativeToken)).to.equal(targetContract);
      expect(await tokenBridgeV1.bridgedToNativeToken(targetContract)).to.equal(nativeToken);

      const proxyAdminAddress = await upgrades.erc1967.getAdminAddress(await tokenBridgeV1.getAddress());
      const ProxyAdminFactory = await ethers.getContractFactory("ProxyAdmin");
      const proxyAdmin = (await ProxyAdminFactory.attach(proxyAdminAddress)) as ProxyAdmin;

      const TokenBridgeV2 = await ethers.getContractFactory("TokenBridge");
      const tokenBridgeV2Implementation = await TokenBridgeV2.deploy();
      await tokenBridgeV2Implementation.waitForDeployment();

      const reinitializeCallData = TokenBridgeV2.interface.encodeFunctionData("reinitializePauseTypesAndPermissions", [
        newRoleAddresses,
        pauseTypeRoles,
        unpauseTypeRoles,
      ]);

      expect(await ethers.provider.getStorage(await tokenBridgeV1.getAddress(), 101)).to.equal(
        ethers.zeroPadValue(owner.address, 32), // _owner
      );
      expect(await ethers.provider.getStorage(await tokenBridgeV1.getAddress(), 213)).to.equal(
        ethers.zeroPadValue("0x01", 32), //_status
      );

      await proxyAdmin.upgradeAndCall(
        await tokenBridgeV1.getAddress(),
        await tokenBridgeV2Implementation.getAddress(),
        reinitializeCallData,
      );

      const upgradedTokenBridge = (await TokenBridgeV2.attach(await tokenBridgeV1.getAddress())) as TokenBridge;

      for (const { role, addressWithRole } of newRoleAddresses) {
        expect(await upgradedTokenBridge.hasRole(role, addressWithRole)).to.be.true;
      }
    });
  });
});
