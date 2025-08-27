// SPDX-License-Identifier: MIT

pragma solidity 0.8.26;

import { StakeManager } from "../../src/StakeManager.sol";
import { IStakeVault } from "../../src/interfaces/IStakeVault.sol";

contract StakeManagerHarness is StakeManager {
    function getVaultLockUntil(address vault) public view returns (uint256) {
        return IStakeVault(vault).lockUntil();
    }
}

