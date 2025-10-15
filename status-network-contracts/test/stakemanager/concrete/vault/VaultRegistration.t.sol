pragma solidity ^0.8.26;

import { StakeManagerTest } from "../../StakeManagerBase.t.sol";

contract VaultRegistrationTest is StakeManagerTest {
    function setUp() public virtual override {
        super.setUp();
    }

    function test_VaultRegistration() public view {
        address[4] memory accounts = [alice, bob, charlie, dave];
        for (uint256 i = 0; i < accounts.length; i++) {
            address[] memory userVaults = streamer.getAccountVaults(accounts[i]);
            assertEq(userVaults.length, 1, "wrong number of vaults");
            assertEq(userVaults[0], vaults[accounts[i]], "wrong vault address");
        }
    }
}