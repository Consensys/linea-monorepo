pragma solidity ^0.8.26;

import { StakeManagerTest, StakeVault, console, Clones, IStakeManager } from "../../StakeManagerBase.t.sol";

contract MultipleVaultsStakeTest is StakeManagerTest {
    StakeVault public vault1;
    StakeVault public vault2;
    StakeVault public vault3;

    function setUp() public override {
        super.setUp();

        vault1 = _createTestVault(alice);
        vault2 = _createTestVault(alice);
        vault3 = _createTestVault(alice);

        vm.startPrank(alice);
        stakingToken.approve(address(vault1), 10_000e18);
        stakingToken.approve(address(vault2), 10_000e18);
        stakingToken.approve(address(vault3), 10_000e18);
        vm.stopPrank();
    }

    function _stakeWithVault(address account, StakeVault vault, uint256 amount, uint256 lockupTime) public {
        vm.prank(account);
        vault.stake(amount, lockupTime);
    }

    function test_StakeMultipleVaults() public {
        console.log(MAX_BALANCE);

        // Alice vault1 stakes 10 tokens
        _stakeWithVault(alice, vault1, 10e18, 0);

        // Alice vault2 stakes 20 tokens
        _stakeWithVault(alice, vault2, 20e18, 0);

        // Alice vault3 stakes 30 tokens
        _stakeWithVault(alice, vault3, 60e18, 0);

        checkStreamer(
            CheckStreamerParams({
                totalStaked: 90e18,
                totalMPStaked: 90e18,
                totalMPAccrued: 90e18,
                totalMaxMP: 450e18,
                stakingBalance: 90e18,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        checkUserTotals(
            CheckUserTotalsParams({ user: alice, totalStakedBalance: 90e18, totalMPAccrued: 90e18, totalMaxMP: 450e18 })
        );
    }
}

contract StakeVaultMigrationTest is StakeManagerTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_RevertWhenNotOwnerOfMigrationVault() public {
        // alice tries to migrate to a vault she doesn't own
        vm.prank(alice);
        vm.expectRevert(StakeVault.StakeVault__NotAuthorized.selector);
        StakeVault(vaults[alice]).migrateToVault(vaults[bob]);
    }

    function test_RevertWhenMigrationVaultNotEmpty() public {
        // alice creates new vault
        vm.startPrank(alice);
        StakeVault newVault = vaultFactory.createVault();

        // ensure new vault is in use
        stakingToken.approve(address(newVault), 10e18);
        newVault.stake(10e18, 0);

        // alice tries to migrate to a vault that is not empty
        vm.expectRevert(IStakeManager.StakeManager__MigrationTargetHasFunds.selector);
        StakeVault(vaults[alice]).migrateToVault(address(newVault));
    }

    function test_RevertWhenDestinationVaultIsNotRegistered() public {
        // alice creates vaults that's not registered with the stake manager
        vm.startPrank(alice);
        address faultyVault = address(Clones.clone(vaultFactory.vaultImplementation()));
        StakeVault(faultyVault).initialize(alice, address(streamer));

        // alice tries to migrate to a vault that is not registered
        vm.expectRevert(IStakeManager.StakeManager__InvalidVault.selector);
        StakeVault(vaults[alice]).migrateToVault(address(faultyVault));
    }

    function testMigrateToVault() public {
        uint256 stakeAmount = 100e18;

        uint256 initialAccountMP = stakeAmount;
        uint256 initialMaxMP = stakeAmount * streamer.MAX_MULTIPLIER() + stakeAmount;

        // first, ensure alice has a vault with staked funds
        _stake(alice, stakeAmount, 0);

        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP,
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: initialAccountMP,
                totalMPAccrued: initialAccountMP,
                totalMaxMP: initialMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // some time passes
        uint256 currentTime = vm.getBlockTimestamp();
        vm.warp(currentTime + 365 days);

        streamer.updateVault(vaults[alice]);

        // ensure vault has accumulated MPs
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP * 2, // alice now has twice the amount after a year
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: stakeAmount * 2,
                totalMPAccrued: initialAccountMP * 2, // stakemanager has twice the amount after a year
                totalMaxMP: initialMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // alice creates new vault
        vm.prank(alice);
        address newVault = address(vaultFactory.createVault());
        uint256 prevVaultLockUntil = StakeVault(vaults[alice]).lockUntil();

        uint256 prevVaultDepositedBalance = StakeVault(vaults[alice]).depositedBalance();

        // alice migrates to new vault
        vm.prank(alice);
        StakeVault(vaults[alice]).migrateToVault(newVault);

        // ensure stake manager's total stats have not changed
        checkStreamer(
            CheckStreamerParams({
                totalStaked: stakeAmount,
                totalMPStaked: initialAccountMP * 2,
                totalMPAccrued: initialAccountMP * 2,
                totalMaxMP: initialMaxMP,
                stakingBalance: stakeAmount,
                rewardBalance: 0,
                rewardIndex: 0
            })
        );

        // check that alice's funds are now in the new vault
        checkVault(
            CheckVaultParams({
                account: newVault,
                rewardBalance: 0,
                stakedBalance: stakeAmount,
                vaultBalance: stakeAmount,
                rewardIndex: 0,
                mpAccrued: initialAccountMP * 2, // alice now has twice the amount after a year
                maxMP: initialMaxMP,
                rewardsAccrued: 0
            })
        );

        assertEq(
            StakeVault(newVault).depositedBalance(), prevVaultDepositedBalance, "deposited balance should be preserved"
        );
        assertEq(StakeVault(newVault).lockUntil(), prevVaultLockUntil, "lock time should be preserved");

        // check that alice's old vault is empty
        checkVault(
            CheckVaultParams({
                account: vaults[alice],
                rewardBalance: 0,
                stakedBalance: 0,
                vaultBalance: 0,
                rewardIndex: 0,
                mpAccrued: 0,
                maxMP: 0,
                rewardsAccrued: 0
            })
        );

        assertEq(StakeVault(vaults[alice]).depositedBalance(), 0, "old vault deposited balance should be 0");
        assertEq(StakeVault(vaults[alice]).lockUntil(), 0, "old vault lock time should be reset");
    }
}