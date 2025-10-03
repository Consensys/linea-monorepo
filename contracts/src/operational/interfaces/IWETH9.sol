// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/**
 * @title WETH9 interface.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
interface IWETH9 is IERC20 {
  function deposit() external payable;
  function withdraw(uint256) external;
}
