// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { IStakeManagerProxy } from "./IStakeManagerProxy.sol";

interface IStakeVault {
    function owner() external view returns (address);
    function stakeManager() external view returns (IStakeManagerProxy);
    function register() external;
}
