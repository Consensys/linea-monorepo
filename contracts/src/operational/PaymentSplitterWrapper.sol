// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { PaymentSplitter } from "@openzeppelin/contracts/finance/PaymentSplitter.sol";

/**
 * @title PaymentSplitterWrapper
 * @dev Wrapper contract for OpenZeppelin's PaymentSplitter to enable Hardhat compilation.
 * This contract simply inherits from PaymentSplitter and passes constructor arguments through.
 * The deployed contract is functionally identical to deploying PaymentSplitter directly.
 */
contract PaymentSplitterWrapper is PaymentSplitter {
    constructor(address[] memory payees, uint256[] memory shares_) payable PaymentSplitter(payees, shares_) {}
}
