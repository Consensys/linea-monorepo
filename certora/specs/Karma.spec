using Karma as karma;

methods {
    function owner() external returns (address) envfree;
}

// TODO:
// totalDistributorAllocation can only increase
// totalDistributorAllocation == sum of all distributor allocations
// sum of external supply <= total supply

definition isERC20TransferFunction(method f) returns bool = (
  f.selector == sig:karma.transfer(address, uint256).selector
                || f.selector == sig:karma.transferFrom(address, address, uint256).selector
                || f.selector == sig:karma.approve(address, uint256).selector
);

definition isOwnableFunction(method f) returns bool = (
  f.selector == sig:karma.addRewardDistributor(address).selector
                || f.selector == sig:karma.removeRewardDistributor(address).selector
                || f.selector == sig:karma.setReward(address, uint256, uint256).selector
                || f.selector == sig:karma.mint(address, uint256).selector

);

rule erc20TransferIsDisabled(method f) {
    env e;
    calldataarg args;

    f@withrevert(e, args);
    bool isReverted = lastReverted;

    assert isERC20TransferFunction(f) => isReverted;
}

rule ownableFuncsOnlyCallableByOwner(method f) {
    env e;
    calldataarg args;

    bool isOwner = owner() == e.msg.sender;

    f@withrevert(e, args);
    bool isReverted = lastReverted;

    assert isOwnableFunction(f) && !isOwner => isReverted;
}
