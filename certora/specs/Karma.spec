using Karma as karma;

// TODO:
// totalDistributorAllocation can only increase
// totalDistributorAllocation == sum of all distributor allocations
// sum of external supply <= total supply
// addRewardDistributor only owner
// removeRewardDistributor only owner
// setReward only owner
// mint only owner

definition isERC20TransferFunction(method f) returns bool = (
  f.selector == sig:karma.transfer(address, uint256).selector
                || f.selector == sig:karma.transferFrom(address, address, uint256).selector
                || f.selector == sig:karma.approve(address, uint256).selector
);

rule erc20TransferIsDisabled(method f) {
    env e;
    calldataarg args;

    f@withrevert(e, args);
    bool isReverted = lastReverted;

    assert isERC20TransferFunction(f) => isReverted;
}
