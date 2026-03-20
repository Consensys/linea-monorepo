// SPDX-License-Identifier: MIT
pragma solidity >=0.7.0 <0.9.0;

import "./Eip7702TestNested.sol";

contract Eip77022Delegated {
    mapping(address => uint256) public storedValue;

    event ValueSet(address indexed sender, uint256 value);

    function setValue(uint256 _newValue, address _nestedContract) external {
        storedValue[msg.sender] = _newValue;
        emit ValueSet(msg.sender, _newValue);
        Eip7702TestNested(_nestedContract).setValue{value: 1000000000000000}(_newValue);
    }

    function getValue(address user) external view returns (uint256) {
        return storedValue[user];
    }
}
