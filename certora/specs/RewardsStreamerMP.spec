import "./shared.spec";

using ERC20A as staked;

methods {
    function ERC20A.balanceOf(address) external returns (uint256) envfree;
    function ERC20A.allowance(address, address) external returns(uint256) envfree;
    function ERC20A.totalSupply() external returns(uint256) envfree;
    function totalStaked() external returns (uint256) envfree;
    function vaultData(address) external returns (uint256, uint256, uint256, uint256, uint256, uint256, uint256, uint256, uint256) envfree;
    function lastMPUpdatedTime() external returns (uint256) envfree;
    function updateGlobalState() external;
    function updateVaultMP(address vaultAddress) external;
    function emergencyModeEnabled() external returns (bool) envfree;
    function leave() external;
    function Math.mulDiv(uint256 a, uint256 b, uint256 c) internal returns uint256 => mulDivSummary(a,b,c);
}

function mulDivSummary(uint256 a, uint256 b, uint256 c) returns uint256 {
  require c != 0;
  return require_uint256(a*b/c);
}

ghost mathint sumOfBalances {
	init_state axiom sumOfBalances == 0;
}

hook Sstore vaultData[KEY address vault].stakedBalance uint256 newValue (uint256 oldValue) {
    sumOfBalances = sumOfBalances - oldValue + newValue;
}

function getVaultMaxMP(address vault) returns uint256 {
    uint256 maxMP;
    _, _, _, maxMP, _, _, _, _, _  = streamer.vaultData(vault);
    return maxMP;
}

function getVaultMPAccrued(address vault) returns uint256 {
    uint256 vaultMPAccrued;
    _, _, vaultMPAccrued, _, _, _, _, _, _  = streamer.vaultData(vault);
    return vaultMPAccrued;
}

function getVaultLockUntil(address vault) returns uint256 {
    uint256 lockUntil;
    _, _, _, _, _, lockUntil, _, _, _  = streamer.vaultData(vault);
    return lockUntil;
}

invariant sumOfBalancesIsTotalStaked()
  sumOfBalances == to_mathint(totalStaked())
  filtered {
    f -> f.selector != sig:upgradeToAndCall(address,bytes).selector
  }

invariant vaultMPLessEqualVaultMaxMP(address vault)
  to_mathint(getVaultMPAccrued(vault)) <= to_mathint(getVaultMaxMP(vault))
  filtered {
    f -> f.selector != sig:upgradeToAndCall(address,bytes).selector && f.selector != sig:migrateToVault(address).selector
  }

invariant vaultMPGreaterEqualVaultStakedBalance(address vault)
  to_mathint(getVaultMPAccrued(vault)) >= to_mathint(getVaultStakedBalance(vault))
  filtered {
    f -> f.selector != sig:upgradeToAndCall(address,bytes).selector && f.selector != sig:migrateToVault(address).selector
  }

rule stakingMintsMultiplierPoints1To1Ratio {

  env e;
  uint256 amount;
  uint256 lockupTime;
  uint256 multiplierPointsBefore;
  uint256 multiplierPointsAfter;

  requireInvariant vaultMPGreaterEqualVaultStakedBalance(e.msg.sender);

  require getVaultLockUntil(e.msg.sender) <= e.block.timestamp;

  updateGlobalState(e);
  updateVaultMP(e, e.msg.sender);
  uint256 t = lastMPUpdatedTime();

  multiplierPointsBefore = getVaultMPAccrued(e.msg.sender);

  stake(e, amount, lockupTime);

  // we need to ensure time has not progressed because that would accrue MP
  // which makes it harder to proof this rule
  require lastMPUpdatedTime() == t;

  multiplierPointsAfter = getVaultMPAccrued(e.msg.sender);

  assert lockupTime == 0 => to_mathint(multiplierPointsAfter) == multiplierPointsBefore + amount;
  assert to_mathint(multiplierPointsAfter) >= multiplierPointsBefore + amount;
}

rule stakingGreaterLockupTimeMeansGreaterMPs {

  env e;
  uint256 amount;
  uint256 lockupTime1;
  uint256 lockupTime2;
  uint256 multiplierPointsAfter1;
  uint256 multiplierPointsAfter2;

  storage initalStorage = lastStorage;

  stake(e, amount, lockupTime1);
  multiplierPointsAfter1 = getVaultMPAccrued(e.msg.sender);

  stake(e, amount, lockupTime2) at initalStorage;
  multiplierPointsAfter2 = getVaultMPAccrued(e.msg.sender);

  assert lockupTime1 >= lockupTime2 => to_mathint(multiplierPointsAfter1) >= to_mathint(multiplierPointsAfter2);
  satisfy to_mathint(multiplierPointsAfter1) > to_mathint(multiplierPointsAfter2);
}


rule MPsOnlyDecreaseWhenUnstaking(method f) filtered { f -> f.selector != sig:upgradeToAndCall(address,bytes).selector } {
  env e;
  calldataarg args;

  uint256 totalMPBefore = totalMPAccrued(e);
  f(e, args);
  uint256 totalMPAfter = totalMPAccrued(e);

  assert totalMPAfter < totalMPBefore => f.selector == sig:unstake(uint256).selector || f.selector == sig:leave().selector;
}
