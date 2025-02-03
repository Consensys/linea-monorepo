// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

interface ITransparentProxy {
    function implementation() external view returns (address);
}
