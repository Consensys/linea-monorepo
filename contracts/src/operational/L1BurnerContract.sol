// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity 0.8.30;

import { ERC20Burnable } from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import { IL1BurnerContract } from "./interfaces/IL1BurnerContract.sol";
import { IL1MessageService } from "../messaging/l1/interfaces/IL1MessageService.sol";
import { IL1MessageManager } from "../messaging/l1/interfaces/IL1MessageManager.sol";

contract L1BurnerContract is IL1BurnerContract {
  address public immutable LINEA_TOKEN;
  address public immutable MESSAGE_SERVICE;

  constructor(address _messageService, address _lineaToken) {
    MESSAGE_SERVICE = _messageService;
    LINEA_TOKEN = _lineaToken;
  }

  function completeBurn(IL1MessageService.ClaimMessageWithProofParams calldata _params) external {
    if (!IL1MessageManager(MESSAGE_SERVICE).isMessageClaimed(_params.messageNumber)) {
      IL1MessageService(MESSAGE_SERVICE).claimMessageWithProof(_params);
    }
    uint256 balance = ERC20Burnable(LINEA_TOKEN).balanceOf(address(this));
    if (balance > 0) {
      ERC20Burnable(LINEA_TOKEN).burn(balance);
    }
  }

  /**
   * @notice Receive function - Receives Funds.
   */
  receive() external payable {}
}
