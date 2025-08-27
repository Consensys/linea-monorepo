using KarmaHarness as karma;

methods {
    function balanceOf(address) external returns (uint256) envfree;
    function rawBalanceOf(address) external returns (uint256) envfree;
    function totalDistributorAllocation() external returns (uint256) envfree;
    function totalSupply() external returns (uint256) envfree;
    function externalSupply() external returns (uint256) envfree;
    function accountSlashAmount(address) external returns (uint256) envfree;
    function _.setReward(uint256, uint256) external => DISPATCHER(true);
    function _.rewardsBalanceOfAccount(address) external => DISPATCHER(true);
    function hasRole(bytes32, address) external returns (bool) envfree;
    function DEFAULT_ADMIN_ROLE() external returns (bytes32) envfree;
    function OPERATOR_ROLE() external returns (bytes32) envfree;
    function Math.mulDiv(uint256 a, uint256 b, uint256 c) internal returns uint256 => mulDivSummary(a,b,c);
}

function mulDivSummary(uint256 a, uint256 b, uint256 c) returns uint256 {
  require c != 0;
  return require_uint256(a*b/c);
}

persistent ghost mathint sumOfDistributorAllocations {
    init_state axiom sumOfDistributorAllocations == 0;
}

hook Sstore rewardDistributorAllocations[KEY address addr] uint256 newValue (uint256 oldValue) {
    sumOfDistributorAllocations = sumOfDistributorAllocations - oldValue + newValue;
}

definition isUpgradeFunction(method f) returns bool = (
  f.selector == sig:karma.upgradeToAndCall(address, bytes).selector ||
  f.selector == sig:karma.upgradeTo(address).selector
);

definition isERC20TransferFunction(method f) returns bool = (
  f.selector == sig:karma.transfer(address, uint256).selector
                || f.selector == sig:karma.transferFrom(address, address, uint256).selector
                || f.selector == sig:karma.approve(address, uint256).selector
);

definition isAdminFunction(method f) returns bool = (
  f.selector == sig:karma.addRewardDistributor(address).selector
                || f.selector == sig:karma.removeRewardDistributor(address).selector

);

definition isOperatorFunction(method f) returns bool = (
    f.selector == sig:karma.setReward(address, uint256, uint256).selector
                || f.selector == sig:karma.mint(address, uint256).selector
);

invariant totalDistributorAllocationIsSumOfDistributorAllocations()
    to_mathint(totalDistributorAllocation()) == sumOfDistributorAllocations
    filtered {
        f -> !isUpgradeFunction(f)
    }

// invariant slashAmountIsLessEqualToKarmaBalance(address account)
//     rawBalanceOf(account) >= getSlashAmountForAccount(account)
//     filtered {
//         f -> !isUpgradeFunction(f) && !isERC20TransferFunction(f)
//     }

// rule externalSupplyIsLessOrEqThanTotalDistributorAllocation() {
//     assert externalSupply() <= totalDistributorAllocation();
// }

rule erc20TransferIsDisabled(method f) {
    env e;
    calldataarg args;

    f@withrevert(e, args);
    bool isReverted = lastReverted;

    assert isERC20TransferFunction(f) => isReverted;
}

rule adminFuncsOnlyCallableByAdmin(method f) {
    env e;
    calldataarg args;

    bool isOwner = hasRole(DEFAULT_ADMIN_ROLE(), e.msg.sender);

    f@withrevert(e, args);
    bool isReverted = lastReverted;

    assert isAdminFunction(f) && !isOwner => isReverted;
}

rule operatorFuncsCallableByAdminAndOperators(method f) {
    env e;
    calldataarg args;

    bool isOwner = hasRole(DEFAULT_ADMIN_ROLE(), e.msg.sender);
    bool isOperator = hasRole(OPERATOR_ROLE(), e.msg.sender);

    f@withrevert(e, args);
    bool isReverted = lastReverted;

    assert isOperatorFunction(f) && !isOwner && !isOperator => isReverted;
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
