// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

/**
 * @title ITransparentProxy
 * @notice Interface for the TransparentProxy contract.
 * @dev ERC1967Proxy contract wrapper to expose implementation address.
 */
interface ITransparentProxy {
    /**
     * @notice Returns the address of the contract implementation.
     * @dev Default ERC1967Proxy doesn't expose this.
     * @return Address of the contract implementation.
     */
    function implementation() external view returns (address);
}
