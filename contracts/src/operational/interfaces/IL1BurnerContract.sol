// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { IL1MessageService } from "../../messaging/l1/interfaces/IL1MessageService.sol";

interface IL1BurnerContract {
  function completeBurn(IL1MessageService.ClaimMessageWithProofParams calldata _params) external;
}
