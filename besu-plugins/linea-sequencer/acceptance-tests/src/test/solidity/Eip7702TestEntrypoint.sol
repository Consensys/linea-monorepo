// SPDX-License-Identifier: MIT
pragma solidity >=0.7.0 <0.9.0;

import "./Eip77022Delegated.sol";

contract Eip7702TestEntrypoint {
    event ValueSet(address indexed sender, uint256 value);

    function setValue(uint256 _newValue, address _delegatingAddress, address _nestedContract) external {
        Eip77022Delegated(_delegatingAddress).setValue(_newValue, _nestedContract);
        emit ValueSet(msg.sender, _newValue);
    }
}
