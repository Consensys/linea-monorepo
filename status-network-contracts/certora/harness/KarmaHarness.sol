// SPDX-License-Identifier: MIT

pragma solidity 0.8.26;

import { Karma } from "../../src/Karma.sol";

contract KarmaHarness is Karma {
    function rawBalanceOf(address account) public view returns (uint256) {
        (uint256 rawBalance,) = _rawBalanceAndSlashAmountOf(account);
        return rawBalance;
    }
}


