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
