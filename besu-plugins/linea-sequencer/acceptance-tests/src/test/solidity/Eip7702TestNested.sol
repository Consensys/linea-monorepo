// SPDX-License-Identifier: MIT
pragma solidity >=0.7.0 <0.9.0;

contract Eip7702TestNested {
    mapping(address => uint256) public storedValue;

    event ValueSet(address indexed sender, uint256 value);

    function setValue(uint256 newValue) external payable {
        storedValue[msg.sender] = newValue;
        emit ValueSet(msg.sender, newValue);
    }

    function getValue(address user) external view returns (uint256) {
        return storedValue[user];
    }
}
