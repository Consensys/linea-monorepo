// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.19;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { AccessControl } from "@openzeppelin/contracts/access/AccessControl.sol";

contract MyToken is ERC20, AccessControl {
  error TokenIsSoulBound();

  bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");

  constructor(address minter) ERC20("MyToken", "MTK") {
    _grantRole(DEFAULT_ADMIN_ROLE, minter);
    _grantRole(MINTER_ROLE, minter);
  }

  function mint(address _to, uint256 _amount) external onlyRole(MINTER_ROLE) {
    _mint(_to, _amount);
  }

  function batchMint(address[] calldata _to, uint256 _amount) external onlyRole(MINTER_ROLE) {
    uint256 addressLength = _to.length;

    for (uint256 i; i < addressLength; ) {
      unchecked {
        _mint(_to[i], _amount);
        ++i;
      }
    }
  }

  function batchMintMultiple(address[] calldata _to, uint256[] calldata _amounts) external onlyRole(MINTER_ROLE) {
    require(_to.length == _amounts.length, "Array lengths do not match");

    uint256 addressLength = _to.length;
    for (uint256 i; i < addressLength; ) {
      unchecked {
        _mint(_to[i], _amounts[i]);
        ++i;
      }
    }
  }

  function transfer(address, uint256) public virtual override returns (bool) {
    revert TokenIsSoulBound();
  }

  function transferFrom(address, address, uint256) public virtual override returns (bool) {
    revert TokenIsSoulBound();
  }

  function approve(address, uint256) public virtual override returns (bool) {
    revert TokenIsSoulBound();
  }

  function increaseAllowance(address, uint256) public virtual override returns (bool) {
    revert TokenIsSoulBound();
  }

  function decreaseAllowance(address, uint256) public virtual override returns (bool) {
    revert TokenIsSoulBound();
  }
}
