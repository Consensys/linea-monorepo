// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

import { PauseManager } from "./PauseManager.sol";

/**
 * @title Contract to manage pausing roles for the YieldManager.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
abstract contract YieldManagerPauseManager is PauseManager {
}
