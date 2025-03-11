using Karma as karma;

methods {
    function owner() external returns (address) envfree;
    function totalDistributorAllocation() external returns (uint256) envfree;
    function _.setReward(uint256, uint256) external => HAVOC_ECF;
}

persistent ghost mathint sumOfDistributorAllocations {
    init_state axiom sumOfDistributorAllocations == 0;
}

hook Sstore rewardDistributorAllocations[KEY address addr] uint256 newValue (uint256 oldValue) {
    sumOfDistributorAllocations = sumOfDistributorAllocations - oldValue + newValue;
}

invariant totalDistributorAllocationIsSumOfDistributorAllocations()
    to_mathint(totalDistributorAllocation()) == sumOfDistributorAllocations
    filtered {
        f -> !isUpgradeFunction(f)
    }

// TODO:
// sum of external supply <= total supply

definition isUpgradeFunction(method f) returns bool = (
  f.selector == sig:karma.upgradeToAndCall(address, bytes).selector
);

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

rule totalDistributorAllocationCanOnlyIncrease(method f) filtered { f ->
     !isUpgradeFunction(f)
     && !isERC20TransferFunction(f)
    } {
    env e;
    calldataarg args;


    uint256 totalDistributorAllocationBefore = totalDistributorAllocation();

    f(e, args);

    uint256 totalDistributorAllocationAfter = totalDistributorAllocation();

    assert totalDistributorAllocationAfter >= totalDistributorAllocationBefore;
}
