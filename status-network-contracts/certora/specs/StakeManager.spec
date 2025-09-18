import "./shared.spec";

using ERC20A as staked;

methods {
    function ERC20A.balanceOf(address) external returns (uint256) envfree;
    function ERC20A.allowance(address, address) external returns(uint256) envfree;
    function ERC20A.totalSupply() external returns(uint256) envfree;
    function totalStaked() external returns (uint256) envfree;
    function vaultData(address) external returns (uint256, uint256, uint256, uint256, uint256, uint256) envfree;
    function lastMPUpdatedTime() external returns (uint256) envfree;
    function updateGlobalState() external;
    function updateVault(address vaultAddress) external;
    function getVaultLockUntil(address) external returns (uint256) envfree;
    function emergencyModeEnabled() external returns (bool) envfree;
    function leave() external;
    function Math.mulDiv(uint256 a, uint256 b, uint256 c) internal returns uint256 => mulDivSummary(a,b,c);
    function paused() external returns (bool) envfree;

    function _.migrateFromVault(IStakeVault.MigrationData) external => DISPATCHER(true);
    function _.lockUntil() external => DISPATCHER(true);
}

function mulDivSummary(uint256 a, uint256 b, uint256 c) returns uint256 {
  require c != 0;
  return require_uint256(a*b/c);
}

ghost mapping(address => uint256) mirrorStaked
{
    init_state axiom (usum address vault. mirrorStaked[vault]) == 0;
}

hook Sstore vaultData[KEY address vault].stakedBalance uint256 newValue (uint256 oldValue) {
    mirrorStaked[vault] = newValue;
}

hook Sload uint256 val vaultData[KEY address vault].stakedBalance {
    require mirrorStaked[vault] == val, "staked is mirrored";
}

ghost mathint sumOfAccruedRewards {
	init_state axiom sumOfAccruedRewards == 0;
}

hook Sstore vaultData[KEY address vault].rewardsAccrued uint256 newValue (uint256 oldValue) {
    sumOfAccruedRewards = sumOfAccruedRewards - oldValue + newValue;
}

invariant sumOfBalancesIsTotalStaked()
  totalStaked() == (usum address vault. mirrorStaked[vault])
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

  updateVault(e, e.msg.sender);
  uint256 t = lastMPUpdatedTime();

  multiplierPointsBefore = getVaultMPAccrued(e.msg.sender);

  stake(e, amount, lockupTime, 0);

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

  stake(e, amount, lockupTime1, 0);
  multiplierPointsAfter1 = getVaultMPAccrued(e.msg.sender);

  stake(e, amount, lockupTime2, 0) at initalStorage;
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

rule allowedActionsWhenPaused(method f) {
  env e;
  calldataarg args;

  require paused();

  f@withrevert(e, args);
  bool reverted = lastReverted;

  assert !reverted => f.isView ||
    f.selector == sig:streamer.initialize(address,address,address).selector ||
    f.selector == sig:streamer.upgradeTo(address).selector ||
    f.selector == sig:streamer.upgradeToAndCall(address, bytes).selector ||
    f.selector == sig:streamer.grantRole(bytes32, address).selector ||
    f.selector == sig:streamer.revokeRole(bytes32, address).selector ||
    f.selector == sig:streamer.renounceRole(bytes32, address).selector ||
    f.selector == sig:streamer.setTrustedCodehash(bytes32, bool).selector ||
    f.selector == sig:streamer.__TrustedCodehashAccess_init(address).selector ||
    f.selector == sig:streamer.enableEmergencyMode().selector ||
    f.selector == sig:streamer.unpause().selector;
}

