// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { DeployKarmaScript } from "../../script/DeployKarma.s.sol";
import { Karma } from "../../src/Karma.sol";
import { KarmaTest } from "./Karma.t.sol";

contract UpgradeKarmaTest is KarmaTest {
    function test_RevertWhen_NonAdminTriesToUpgrade() public {
        address deployer = new DeployKarmaScript().getDeployer();
        // deploy first version of Karma
        bytes memory initializeData = abi.encodeCall(Karma.initialize, deployer);
        address impl = address(new Karma());
        // Create upgradeable proxy
        address proxy = address(new ERC1967Proxy(impl, initializeData));

        // Deploy new implementation
        Karma newImplementation = new Karma();

        // Try to upgrade as non-admin
        vm.prank(alice);
        vm.expectRevert();
        UUPSUpgradeable(proxy).upgradeTo(address(newImplementation));
    }
}
