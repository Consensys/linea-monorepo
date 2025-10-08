// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { IV3DexSwap } from "./interfaces/IV3DexSwap.sol";
import { ISwapRouterV3 } from "./interfaces/ISwapRouterV3.sol";
import { IWETH9 } from "./interfaces/IWETH9.sol";

/**
 * @title V3DexSwap.
 * @dev A contract for swapping tokens on a decentralized exchange.
 * @author Consensys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
contract V3DexSwap is IV3DexSwap {
  /// @notice Tick spacing of the pool.
  uint24 public immutable POOL_TICK_SPACING;
  /// @notice Address of the Swap Router contract.
  address public immutable ROUTER;
  /// @notice Address of the WETH token contract.
  address public immutable WETH_TOKEN;
  /// @notice Address of the LINEA token contract.
  address public immutable LINEA_TOKEN;
  /// @notice Address of the RollupRevenueVault contract to receive swapped tokens.
  address public immutable ROLLUP_REVENUE_VAULT;

  /**
   * @dev Sets the address of the RollupRevenueVault contract and other immutable values.
   * @param _router Address of the Router contract.
   * @param _wethToken Address of the WETH token contract.
   * @param _lineaToken Address of the LINEA token contract.
   * @param _rollupRevenueVault Address of the RollupRevenueVault contract.
   */
  constructor(
    address _router,
    address _wethToken,
    address _lineaToken,
    address _rollupRevenueVault,
    uint24 _poolTickSpacing
  ) {
    require(_rollupRevenueVault != address(0), ZeroAddressNotAllowed());
    require(_wethToken != address(0), ZeroAddressNotAllowed());
    require(_lineaToken != address(0), ZeroAddressNotAllowed());
    require(_router != address(0), ZeroAddressNotAllowed());
    require(_poolTickSpacing > 0, ZeroTickSpacingNotAllowed());

    ROUTER = _router;
    WETH_TOKEN = _wethToken;
    LINEA_TOKEN = _lineaToken;
    ROLLUP_REVENUE_VAULT = _rollupRevenueVault;
    POOL_TICK_SPACING = _poolTickSpacing;
  }

  /** @notice Swap ETH into LINEA.
   * @param _minLineaOut Minimum number of LINEA tokens to receive (slippage protection).
   * @param _deadline Time after which the transaction will revert if not yet processed.
   * @param _sqrtPriceLimitX96 Price limit of the swap as a Q64.96 value.
   */
  function swap(
    uint256 _minLineaOut,
    uint256 _deadline,
    uint160 _sqrtPriceLimitX96
  ) external payable returns (uint256 amountOut) {
    require(msg.sender == ROLLUP_REVENUE_VAULT, UnauthorizedAccount());
    require(msg.value > 0, NoEthSend());
    require(_minLineaOut > 0, ZeroMinLineaOutNotAllowed());

    IWETH9(WETH_TOKEN).deposit{ value: msg.value }();
    IWETH9(WETH_TOKEN).approve(ROUTER, msg.value);

    amountOut = ISwapRouterV3(ROUTER).exactInputSingle(
      ISwapRouterV3.ExactInputSingleParams({
        tokenIn: WETH_TOKEN,
        tokenOut: LINEA_TOKEN,
        tickSpacing: POOL_TICK_SPACING,
        recipient: ROLLUP_REVENUE_VAULT,
        deadline: _deadline,
        amountIn: msg.value,
        amountOutMinimum: _minLineaOut,
        sqrtPriceLimitX96: _sqrtPriceLimitX96
      })
    );
  }
}
