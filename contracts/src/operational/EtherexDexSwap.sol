// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { IV3DexSwap } from "./interfaces/IV3DexSwap.sol";
import { ISwapRouterV3 } from "./interfaces/ISwapRouterV3.sol";
import { IWETH9 } from "./interfaces/IWETH9.sol";

/**
 * @title EtherexDexSwap.
 * @dev A contract for swapping tokens on the Etherex decentralized exchange.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract EtherexDexSwap is IV3DexSwap {
  /// @dev Address of the Etherex SwapRouter contract
  address public constant ROUTER = 0x8BE024b5c546B5d45CbB23163e1a4dca8fA5052A;
  address public constant WETH_TOKEN = 0xe5D7C2a44FfDDf6b295A15c148167daaAf5Cf34f;
  address public constant LINEA_TOKEN = 0x1789e0043623282D5DCc7F213d703C6D8BAfBB04;
  /// @dev Etherex uses a fixed tick spacing of 50 for WETH/LINEA pool
  uint24 public constant POOL_TICK_SPACING = 50;
  /// @dev Address of the RollupFeeVault contract authorized to call the swap function
  address public immutable ROLLUP_FEE_VAULT;

  /**
   * @dev Sets the address of the RollupFeeVault contract.
   * @param _rollupFeeVault Address of the RollupFeeVault contract.
   */
  constructor(address _rollupFeeVault) {
    require(_rollupFeeVault != address(0), ZeroAddressNotAllowed());
    ROLLUP_FEE_VAULT = _rollupFeeVault;
  }

  /** @notice Swap ETH into LINEA.
   * @param _minLineaOut Number of LINEA tokens to receive (slippage protection).
   * @param _deadline Time after which the transaction will revert if not yet processed.
   * @param _sqrtPriceLimitX96 Price limit of the swap as a Q64.96 value.
   */
  function swap(
    uint256 _minLineaOut,
    uint256 _deadline,
    uint160 _sqrtPriceLimitX96
  ) external payable returns (uint256 amountOut) {
    require(msg.value > 0, NoEthSend());

    IWETH9(WETH_TOKEN).deposit{ value: msg.value }();
    IERC20(WETH_TOKEN).approve(ROUTER, msg.value);

    amountOut = ISwapRouterV3(ROUTER).exactInputSingle(
      ISwapRouterV3.ExactInputSingleParams({
        tokenIn: WETH_TOKEN,
        tokenOut: LINEA_TOKEN,
        tickSpacing: POOL_TICK_SPACING,
        recipient: ROLLUP_FEE_VAULT,
        deadline: _deadline,
        amountIn: msg.value,
        amountOutMinimum: _minLineaOut,
        sqrtPriceLimitX96: _sqrtPriceLimitX96
      })
    );

    require(amountOut >= _minLineaOut, MinOutputAmountNotMet());
  }
}
