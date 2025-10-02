// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { IV3DexSwap } from "../../../operational/interfaces/IV3DexSwap.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IWETH9 is IERC20 {
  function deposit() external payable;
  function withdraw(uint256) external;
}

interface ISwapRouterV3 {
  struct ExactInputSingleParams {
    address tokenIn;
    address tokenOut;
    uint24 tickSpacing;
    address recipient;
    uint256 deadline;
    uint256 amountIn;
    uint256 amountOutMinimum;
    uint160 sqrtPriceLimitX96;
  }

  function exactInputSingle(ExactInputSingleParams calldata params) external payable returns (uint256 amountOut);
}

contract TestDexSwap is IV3DexSwap {
  /// @dev Etherex uses a fixed tick spacing of 50 for WETH/LINEA pool
  uint24 public constant POOL_TICK_SPACING = 50;
  /// @dev Address of the Etherex SwapRouter contract
  address public immutable ROUTER;
  address public immutable WETH_TOKEN;
  address public immutable LINEA_TOKEN;
  /// @dev Address of the RollupFeeVault contract authorized to call the swap function
  address public immutable ROLLUP_FEE_VAULT;

  /**
   * @dev Sets the address of the RollupFeeVault contract.
   * @param _rollupFeeVault Address of the RollupFeeVault contract.
   */
  constructor(address _rollupFeeVault, address _lineaToken, address _wethToken, address _router) {
    require(_rollupFeeVault != address(0), ZeroAddressNotAllowed());
    ROLLUP_FEE_VAULT = _rollupFeeVault;
    LINEA_TOKEN = _lineaToken;
    WETH_TOKEN = _wethToken;
    ROUTER = _router;
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
