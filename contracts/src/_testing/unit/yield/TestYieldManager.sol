// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { YieldManager } from "../../../yield/YieldManager.sol";

/// @custom:oz-upgrades-unsafe-allow missing-initializer
contract TestYieldManager is YieldManager {
    constructor(address _l1MessageService) YieldManager(_l1MessageService) {}
}