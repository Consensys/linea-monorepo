// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { KarmaTest } from "./Karma.t.sol";

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
