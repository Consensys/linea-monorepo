// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

import { BeaconProxy } from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";

contract LidoStVaultYieldProviderFactory {
    event LidoStVaultYieldProviderCreated(
        address stakingVault,
        address providerAddress
    );

    address public immutable BEACON;

    constructor(address _beacon) {
        BEACON = _beacon;
    }

    function createLidoStVaultYieldProvider(address _stakingVault) external returns (address yieldProviderAddress) {
        yieldProviderAddress = address(
            new BeaconProxy{ salt: keccak256(abi.encode(_stakingVault)) }(BEACON, "")
        );
        emit LidoStVaultYieldProviderCreated(_stakingVault, yieldProviderAddress);
    }
}
