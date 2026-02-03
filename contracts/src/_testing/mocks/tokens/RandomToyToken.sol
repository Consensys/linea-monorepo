// SPDX-License-Identifier: MIT

pragma solidity ^0.8.33;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract RandomToyToken is ERC20 {
  constructor() ERC20("RandomToyToken", "RTOY") {
    _mint(msg.sender, 1_000_000 ether);
  }

  function mint(address to, uint256 amount) external {
    _mint(to, amount);
  }
}
