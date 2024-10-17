// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Ownable, Ownable2Step } from "@openzeppelin/contracts/access/Ownable2Step.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { IXPProvider } from "./interfaces/IXPProvider.sol";

contract XPToken is ERC20, Ownable2Step {
    string public constant NAME = "XP Token";
    string public constant SYMBOL = "XP";

    uint256 public externalSupply;

    IXPProvider[] public xpProviders;

    error XPToken__TransfersNotAllowed();
    error XPProvider__IndexOutOfBounds();

    constructor(uint256 _totalSupply) ERC20(NAME, SYMBOL) Ownable(msg.sender) {
        externalSupply = _totalSupply;
    }

    function setExternalSupply(uint256 _externalSupply) external onlyOwner {
        externalSupply = _externalSupply;
    }

    function addXPProvider(IXPProvider provider) external onlyOwner {
        xpProviders.push(provider);
    }

    function removeXPProvider(uint256 index) external onlyOwner {
        if (index >= xpProviders.length) {
            revert XPProvider__IndexOutOfBounds();
        }

        xpProviders[index] = xpProviders[xpProviders.length - 1];
        xpProviders.pop();
    }

    function getXPProviders() external view returns (IXPProvider[] memory) {
        return xpProviders;
    }

    function totalSupply() public view override returns (uint256) {
        return super.totalSupply() + externalSupply;
    }

    function balanceOf(address account) public view override returns (uint256) {
        uint256 userTotalXPShare = 0;
        uint256 totalXPShares = 0;

        for (uint256 i = 0; i < xpProviders.length; i++) {
            IXPProvider provider = xpProviders[i];
            userTotalXPShare += provider.getUserXPShare(account);
            totalXPShares += provider.getTotalXPShares();
        }

        if (totalXPShares == 0) {
            return 0;
        }

        uint256 externalBalance = (externalSupply * userTotalXPShare) / totalXPShares;
        return super.balanceOf(account) + externalBalance;
    }

    function transfer(address, uint256) public pure override returns (bool) {
        revert XPToken__TransfersNotAllowed();
    }

    function approve(address, uint256) public pure override returns (bool) {
        revert XPToken__TransfersNotAllowed();
    }

    function transferFrom(address, address, uint256) public pure override returns (bool) {
        revert XPToken__TransfersNotAllowed();
    }

    function allowance(address, address) public pure override returns (uint256) {
        return 0;
    }
}
