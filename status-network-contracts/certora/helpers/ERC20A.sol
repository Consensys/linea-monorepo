// SPDX-License-Identifier: MIT

pragma solidity ^0.8.26;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract ERC20A is ERC20 {
   constructor(string memory name_, string memory symbol_) ERC20(name_, symbol_) {}
}

