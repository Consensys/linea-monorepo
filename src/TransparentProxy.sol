// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

import { ERC1967Proxy } from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

/**
 * @title TransparentProxy
 * @author 0x-r4bbit
 * @notice ERC1967Proxy contract wrapper to expose implementation address.
 * @dev Transparent Proxy contract
 */
contract TransparentProxy is ERC1967Proxy {
    /**
     * @notice Creates a new TransparentProxy contract.
     * @param _implementation Address of the contract implementation.
     * @param _data Data to be passed to the contract initialization.
     */
    constructor(address _implementation, bytes memory _data) ERC1967Proxy(_implementation, _data) { }

    /**
     * @notice Returns the address of the contract implementation.
     * @dev Default ERC1967Proxy doesn't expose this.
     * @return Address of the contract implementation.
     */
    function implementation() external view returns (address) {
        return _implementation();
    }
}
