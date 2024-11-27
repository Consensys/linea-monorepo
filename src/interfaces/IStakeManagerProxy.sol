// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IStakeManager } from "./IStakeManager.sol";

interface IStakeManagerProxy is IStakeManager {
    function implementation() external view returns (address);
}
