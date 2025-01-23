// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Ownable, Ownable2Step } from "@openzeppelin/contracts/access/Ownable2Step.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { IRewardProvider } from "./interfaces/IRewardProvider.sol";

contract XPToken is ERC20, Ownable2Step {
    error XPToken__MintAllowanceExceeded();

    string public constant NAME = "XP Token";
    string public constant SYMBOL = "XP";

    IRewardProvider[] public rewardProviders;

    error XPToken__TransfersNotAllowed();
    error RewardProvider__IndexOutOfBounds();

    constructor() ERC20(NAME, SYMBOL) Ownable(msg.sender) { }

    function addRewardProvider(IRewardProvider provider) external onlyOwner {
        rewardProviders.push(provider);
    }

    function removeRewardProvider(uint256 index) external onlyOwner {
        if (index >= rewardProviders.length) {
            revert RewardProvider__IndexOutOfBounds();
        }

        rewardProviders[index] = rewardProviders[rewardProviders.length - 1];
        rewardProviders.pop();
    }

    function getRewardProviders() external view returns (IRewardProvider[] memory) {
        return rewardProviders;
    }

    function _totalSupply() public view returns (uint256) {
        return super.totalSupply() + _externalSupply();
    }

    function totalSupply() public view override returns (uint256) {
        return _totalSupply();
    }

    function mint(address account, uint256 amount) external onlyOwner {
        if (amount > _mintAllowance()) {
            revert XPToken__MintAllowanceExceeded();
        }

        _mint(account, amount);
    }

    function _mintAllowance() internal view returns (uint256) {
        uint256 maxSupply = _externalSupply() * 3;
        uint256 fullTotalSupply = _totalSupply();
        if (maxSupply <= fullTotalSupply) {
            return 0;
        }

        return maxSupply - fullTotalSupply;
    }

    function mintAllowance() public view returns (uint256) {
        return _mintAllowance();
    }

    function _externalSupply() internal view returns (uint256) {
        uint256 externalSupply;

        for (uint256 i = 0; i < rewardProviders.length; i++) {
            externalSupply += rewardProviders[i].totalRewardsSupply();
        }

        return externalSupply;
    }

    function balanceOf(address account) public view override returns (uint256) {
        uint256 externalBalance;

        for (uint256 i = 0; i < rewardProviders.length; i++) {
            IRewardProvider provider = rewardProviders[i];
            externalBalance += provider.rewardsBalanceOfAccount(account);
        }

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
