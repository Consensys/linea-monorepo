// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Strings } from "@openzeppelin/contracts/utils/Strings.sol";

import { Test } from "forge-std/Test.sol";
import { DeployKarmaScript } from "../script/DeployKarma.s.sol";
import { DeploymentConfig } from "../script/DeploymentConfig.s.sol";
import { Karma } from "../src/Karma.sol";
import { KarmaDistributorMock } from "./mocks/KarmaDistributorMock.sol";

contract KarmaTest is Test {
    Karma public karma;

    address public owner;
    address public alice = makeAddr("alice");
    address public bob = makeAddr("bob");

    KarmaDistributorMock public distributor1;
    KarmaDistributorMock public distributor2;

    address public operator = makeAddr("operator");

    function setUp() public virtual {
        DeployKarmaScript karmaDeployment = new DeployKarmaScript();
        (Karma _karma, DeploymentConfig deploymentConfig) = karmaDeployment.run();
        karma = _karma;
        (address deployer,) = deploymentConfig.activeNetworkConfig();
        owner = deployer;

        distributor1 = new KarmaDistributorMock();
        distributor2 = new KarmaDistributorMock();

        vm.startBroadcast(owner);
        karma.addRewardDistributor(address(distributor1));
        karma.addRewardDistributor(address(distributor2));
        vm.stopBroadcast();
    }

    function _accessControlError(address account, bytes32 role) internal pure returns (bytes memory) {
        string memory expectedError = string(
            abi.encodePacked(
                "AccessControl: account ",
                Strings.toHexString(uint160(account)),
                " is missing role ",
                Strings.toHexString(uint256(role), 32)
            )
        );
        return bytes(expectedError);
    }

    function testAddKarmaDistributorOnlyAdmin() public {
        KarmaDistributorMock distributor3 = new KarmaDistributorMock();

        bytes memory expectedError = _accessControlError(alice, karma.DEFAULT_ADMIN_ROLE());
        vm.prank(alice);
        vm.expectRevert(expectedError);
        karma.addRewardDistributor(address(distributor3));

        vm.prank(owner);
        karma.addRewardDistributor(address(distributor3));

        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[0], address(distributor1));
        assertEq(distributors[1], address(distributor2));
        assertEq(distributors[2], address(distributor3));
    }

    function testRemoveKarmaDistributorOnlyOwner() public {
        bytes memory expectedError = _accessControlError(alice, karma.DEFAULT_ADMIN_ROLE());
        vm.prank(alice);
        vm.expectRevert(expectedError);
        karma.removeRewardDistributor(address(distributor1));

        vm.prank(owner);
        karma.removeRewardDistributor(address(distributor1));

        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 1);
        assertEq(distributors[0], address(distributor2));
    }

    function testRemoveUnknownKarmaDistributor() public {
        vm.prank(owner);
        vm.expectRevert(Karma.Karma__UnknownDistributor.selector);
        karma.removeRewardDistributor(address(1));
    }

    function testTotalSupply() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();

        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        vm.prank(owner);
        karma.mint(owner, 500 ether);

        uint256 totalSupply = karma.totalSupply();
        assertEq(totalSupply, 3500 ether);
    }

    function testBalanceOfWithNoSystemTotalKarma() public view {
        uint256 aliceBalance = karma.balanceOf(alice);
        assertEq(aliceBalance, 0);

        uint256 bobBalance = karma.balanceOf(bob);
        assertEq(bobBalance, 0);
    }

    function testBalanceOf() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();

        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);

        distributor1.setUserKarmaShare(alice, 1000e18);
        distributor2.setUserKarmaShare(alice, 2000e18);

        vm.prank(owner);
        karma.mint(alice, 500e18);

        uint256 expectedBalance = 3500e18;

        uint256 balance = karma.balanceOf(alice);
        assertEq(balance, expectedBalance);
    }

    function testMintOnlyAdmin() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), 1000 ether, 1000);
        karma.setReward(address(distributor2), 2000 ether, 2000);
        vm.stopBroadcast();

        distributor1.setTotalKarmaShares(1000 ether);
        distributor2.setTotalKarmaShares(2000 ether);
        assertEq(karma.totalSupply(), 3000 ether);

        vm.prank(alice);
        vm.expectRevert(Karma.Karma__Unauthorized.selector);
        karma.mint(alice, 1000e18);

        vm.prank(owner);
        karma.mint(alice, 1000e18);
        assertEq(karma.totalSupply(), 4000e18);
    }

    function testTransfersNotAllowed() public {
        vm.expectRevert(Karma.Karma__TransfersNotAllowed.selector);
        karma.transfer(alice, 100e18);

        vm.expectRevert(Karma.Karma__TransfersNotAllowed.selector);
        karma.approve(alice, 100e18);

        vm.expectRevert(Karma.Karma__TransfersNotAllowed.selector);
        karma.transferFrom(alice, bob, 100e18);

        uint256 allowance = karma.allowance(alice, bob);
        assertEq(allowance, 0);
    }
}

contract KarmaOwnershipTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function testInitialOwner() public view {
        assert(karma.hasRole(karma.DEFAULT_ADMIN_ROLE(), owner));
    }

    function testOwnershipTransfer() public {
        vm.startPrank(owner);
        karma.grantRole(karma.DEFAULT_ADMIN_ROLE(), alice);
        vm.stopPrank();
        assert(karma.hasRole(karma.DEFAULT_ADMIN_ROLE(), alice));
    }
}

contract AddRewardDistributorTest is KarmaTest {
    address public distributor;

    function setUp() public virtual override {
        super.setUp();
        distributor = address(new KarmaDistributorMock());
    }

    function test_RevertWhen_SenderIsNotDefaultAdmin() public {
        vm.prank(makeAddr("someone"));
        vm.expectRevert();
        karma.addRewardDistributor(distributor);
    }

    function testAddRewardDistributorAsOtherAdmin() public {
        address otherAdmin = makeAddr("otherAdmin");
        vm.startPrank(owner);
        karma.grantRole(karma.DEFAULT_ADMIN_ROLE(), otherAdmin);
        vm.stopPrank();

        vm.startPrank(otherAdmin);
        karma.addRewardDistributor(distributor);
        vm.stopPrank();
        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[2], distributor);
    }
}

contract RemoveRewardDistributorTest is KarmaTest {
    address public distributor;

    function setUp() public virtual override {
        super.setUp();
        distributor = address(new KarmaDistributorMock());
    }

    function test_RevertWhen_SenderIsNotDefaultAdmin() public {
        vm.expectRevert();
        karma.removeRewardDistributor(distributor);
    }

    function testRemoveRewardDistributor() public {
        // add a distributor
        vm.prank(owner);
        karma.addRewardDistributor(distributor);
        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[2], distributor);

        // remove the distributor
        vm.prank(owner);
        karma.removeRewardDistributor(distributor);
        distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 2);
    }

    function testRemoveRewardDistributorAsOtherAdmin() public {
        // add a distributor
        vm.prank(owner);
        karma.addRewardDistributor(distributor);
        address[] memory distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 3);
        assertEq(distributors[2], distributor);

        // grant admin role
        address otherAdmin = makeAddr("otherAdmin");
        vm.startPrank(owner);
        karma.grantRole(karma.DEFAULT_ADMIN_ROLE(), otherAdmin);
        vm.stopPrank();

        // remove the distributor
        vm.prank(otherAdmin);
        karma.removeRewardDistributor(address(distributor1));
        distributors = karma.getRewardDistributors();
        assertEq(distributors.length, 2);
    }
}

contract SetRewardTest is KarmaTest {
    address public distributor;

    function setUp() public virtual override {
        super.setUp();
        distributor = address(new KarmaDistributorMock());
    }

    function test_RevertWhen_SenderIsNotDefaultAdmin() public {
        vm.prank(makeAddr("someone"));
        vm.expectRevert();
        karma.setReward(distributor, 0, 0);
    }

    function test_RevertWhen_SenderIsNotOperator() public {
        assert(karma.hasRole(karma.OPERATOR_ROLE(), operator) == false);

        vm.prank(operator);
        vm.expectRevert();
        karma.setReward(distributor, 0, 0);
    }

    function testSetRewardAsAdmin() public {
        vm.startPrank(owner);
        karma.addRewardDistributor(distributor);
        karma.setReward(distributor, 0, 0);
        vm.stopPrank();
    }

    function testSetRewardAsOtherAdmin() public {
        vm.startPrank(owner);
        karma.grantRole(karma.DEFAULT_ADMIN_ROLE(), operator);
        karma.addRewardDistributor(distributor);
        vm.stopPrank();

        vm.prank(operator);
        karma.setReward(distributor, 0, 0);
    }

    function testSetRewardAsOperator() public {
        // grant operator role
        assert(karma.hasRole(karma.DEFAULT_ADMIN_ROLE(), owner));

        // actually `vm.prank()` should be used here, but for some reason
        // foundry seems to mess up the context for what `owner` is
        vm.startPrank(owner);
        karma.grantRole(karma.OPERATOR_ROLE(), operator);
        vm.stopPrank();

        // set reward as operator
        vm.prank(operator);
        karma.setReward(address(distributor1), 0, 0);
    }
}

contract OverflowTest is KarmaTest {
    function setUp() public override {
        super.setUp();
    }

    function test_RevertWhen_MintingCausesOverflow() public {
        vm.startBroadcast(owner);
        karma.setReward(address(distributor1), type(uint256).max, 1000);
        vm.stopBroadcast();

        vm.prank(owner);
        vm.expectRevert();
        karma.mint(owner, 1e18);
    }

    function test_RevertWhen_SettingRewardCausesOverflow() public {
        vm.prank(owner);
        karma.mint(owner, type(uint256).max);

        vm.prank(owner);
        vm.expectRevert();
        karma.setReward(address(distributor1), 1e18, 1000);
    }
}

contract SlashAmountOfTest is KarmaTest {
    address public slasher = makeAddr("slasher");

    function _mintKarmaToAccount(address account, uint256 amount) internal {
        vm.startPrank(owner);
        karma.mint(account, amount);
        vm.stopPrank();
    }

    function setUp() public override {
        super.setUp();

        vm.startPrank(owner);
        karma.grantRole(karma.SLASHER_ROLE(), slasher);
        vm.stopPrank();
    }

    function test_SlashAmountOf() public {
        uint256 accountBalance = 100 ether;
        uint256 distributorBalance = 50 ether;
        _mintKarmaToAccount(alice, accountBalance);
        vm.prank(owner);
        distributor1.setUserKarmaShare(alice, distributorBalance);

        vm.prank(owner);
        karma.slash(alice);

        uint256 slashedAmount = karma.slashedAmountOf(alice);
        assertEq(
            slashedAmount, karma.calculateSlashAmount(accountBalance) + karma.calculateSlashAmount(distributorBalance)
        );
    }

    function testFuzz_SlashAmountOf(
        uint256 accountBalance,
        uint256 distributor1Balance,
        uint256 distributor2Balance
    )
        public
    {
        // adding some bounds here to ensure we don't overflow
        vm.assume(accountBalance < 1e30);
        vm.assume(distributor1Balance < 1e30);
        vm.assume(distributor2Balance < 1e30);

        // Ensure Alice has at least some balance to slash
        vm.assume(accountBalance + distributor1Balance + distributor2Balance > 0);

        _mintKarmaToAccount(alice, accountBalance);
        vm.startPrank(owner);
        distributor1.setUserKarmaShare(alice, distributor1Balance);
        distributor2.setUserKarmaShare(alice, distributor2Balance);
        vm.stopPrank();

        vm.prank(owner);
        karma.slash(alice);

        uint256 slashedAmount = karma.slashedAmountOf(alice);
        uint256 expectedSlashAmount = karma.calculateSlashAmount(accountBalance)
            + karma.calculateSlashAmount(distributor1Balance) + karma.calculateSlashAmount(distributor2Balance);
        assertEq(slashedAmount, expectedSlashAmount);
    }
}

contract SlashTest is KarmaTest {
    address public slasher = makeAddr("slasher");

    function _mintKarmaToAccount(address account, uint256 amount) internal {
        vm.startPrank(owner);
        karma.mint(account, amount);
        vm.stopPrank();
    }

    function setUp() public override {
        super.setUp();

        vm.startPrank(owner);
        karma.grantRole(karma.SLASHER_ROLE(), slasher);
        vm.stopPrank();
    }

    function test_RevertWhen_SenderIsNotDefaultAdminOrSlasher() public {
        vm.prank(makeAddr("someone"));
        vm.expectRevert(Karma.Karma__Unauthorized.selector);
        karma.slash(alice);
    }

    function test_RevertWhen_KarmaBalanceIsInvalid() public {
        vm.prank(slasher);
        vm.expectRevert(Karma.Karma__CannotSlashZeroBalance.selector);
        karma.slash(alice);
    }

    function test_SlashRemainingBalanceIfBalanceIsLow() public {
        _mintKarmaToAccount(alice, karma.MIN_SLASH_AMOUNT() - 1);

        vm.prank(slasher);
        karma.slash(alice);

        assertEq(karma.balanceOf(alice), 0);
    }

    function test_Slash() public {
        // ensure rewards
        uint256 currentBalance = 100 ether;
        _mintKarmaToAccount(alice, currentBalance);
        uint256 slashedAmount = karma.calculateSlashAmount(currentBalance);

        // slash the account
        vm.prank(slasher);
        karma.slash(alice);
        assertEq(karma.balanceOf(alice), currentBalance - slashedAmount);

        currentBalance = karma.balanceOf(alice);
        slashedAmount = karma.calculateSlashAmount(currentBalance);

        // slash again
        vm.prank(slasher);
        karma.slash(alice);

        assertEq(karma.balanceOf(alice), currentBalance - slashedAmount);
    }

    function testFuzz_Slash(uint256 rewardsAmount) public {
        vm.assume(rewardsAmount > 0);
        _mintKarmaToAccount(alice, rewardsAmount);

        vm.prank(slasher);
        karma.slash(alice);
        uint256 slashedAmount = karma.slashedAmountOf(alice);

        assertEq(karma.balanceOf(alice), rewardsAmount - slashedAmount);
    }

    function testRemoveRewardDistributorShouldReduceSlashAmount() public {
        uint256 distributorRewards = 1000 ether;
        uint256 mintedRewards = 1000 ether;
        uint256 totalRewards = distributorRewards + mintedRewards;

        // set up rewards for alice
        vm.prank(owner);
        distributor1.setUserKarmaShare(alice, distributorRewards);
        _mintKarmaToAccount(alice, mintedRewards);
        assertEq(distributor1.rewardsBalanceOfAccount(alice), distributorRewards);
        assertEq(karma.balanceOf(alice), totalRewards);

        // slash alice
        uint256 accountSlashAmount = karma.calculateSlashAmount(mintedRewards);
        uint256 distributorSlashAmount = karma.calculateSlashAmount(distributorRewards);
        vm.prank(owner);
        karma.slash(alice);

        uint256 totalSlashAmount = karma.slashedAmountOf(alice);

        assertEq(karma.accountSlashAmount(alice), accountSlashAmount);
        assertEq(karma.rewardDistributorSlashAmount(address(distributor1), alice), distributorSlashAmount);
        assertEq(karma.slashedAmountOf(alice), totalSlashAmount);
        assertEq(karma.balanceOf(alice), totalRewards - totalSlashAmount);

        // remove the distributor
        vm.prank(owner);
        karma.removeRewardDistributor(address(distributor1));
        totalSlashAmount = karma.slashedAmountOf(alice);
        assertEq(karma.accountSlashAmount(alice), accountSlashAmount);
        assertEq(karma.balanceOf(alice), totalRewards - distributorRewards - totalSlashAmount);
    }
}
