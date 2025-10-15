pragma solidity ^0.8.26;

import { StakeManagerTest, StakeVault, ITrustedCodehashAccess, Clones } from "../../StakeManagerBase.t.sol";

contract TrustedCodehashAccessTest is StakeManagerTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_RevertWhenProxyCloneCodehashNotTrusted() public {
        // create independent (possibly malicious) StakeVault
        address vaultTpl = address(new StakeVault(stakingToken));
        StakeVault proxyClone = StakeVault(Clones.clone(vaultTpl));
        proxyClone.initialize(address(this), address(streamer));

        // registering already fails as codehash is not trusted
        vm.expectRevert(ITrustedCodehashAccess.TrustedCodehashAccess__UnauthorizedCodehash.selector);
        proxyClone.register();

        // staking fails as codehash is not trusted
        vm.expectRevert(ITrustedCodehashAccess.TrustedCodehashAccess__UnauthorizedCodehash.selector);
        proxyClone.stake(10e10, 0);
    }
}