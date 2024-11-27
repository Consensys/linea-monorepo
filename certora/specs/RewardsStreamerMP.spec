import "./shared.spec";

using ERC20A as staked;

methods {
    function ERC20A.balanceOf(address) external returns (uint256) envfree;
    function ERC20A.allowance(address, address) external returns(uint256) envfree;
    function ERC20A.totalSupply() external returns(uint256) envfree;
    function totalStaked() external returns (uint256) envfree;
    function accounts(address) external returns (uint256, uint256, uint256, uint256, uint256, uint256) envfree;
    function lastMPUpdatedTime() external returns (uint256) envfree;
    function updateGlobalState() external;
    function updateAccountMP(address accountAddress) external;
    function emergencyModeEnabled() external returns (bool) envfree;
    function leave() external;
}

ghost mathint sumOfBalances {
	init_state axiom sumOfBalances == 0;
}

hook Sstore accounts[KEY address account].stakedBalance uint256 newValue (uint256 oldValue) {
    sumOfBalances = sumOfBalances - oldValue + newValue;
}

function getAccountMaxMP(address account) returns uint256 {
    uint256 maxMP;
    _, _, _, maxMP, _, _ = streamer.accounts(account);
    return maxMP;
}

function getAccountMP(address account) returns uint256 {
    uint256 accountMP;
    _, _, accountMP, _, _, _ = streamer.accounts(account);
    return accountMP;
}

function getAccountLockUntil(address account) returns uint256 {
    uint256 lockUntil;
    _, _, _, _, _, lockUntil = streamer.accounts(account);
    return lockUntil;
}

invariant sumOfBalancesIsTotalStaked()
  sumOfBalances == to_mathint(totalStaked())
  filtered {
    f -> f.selector != sig:upgradeToAndCall(address,bytes).selector
  }

invariant accountMPLessEqualAccountMaxMP(address account)
  to_mathint(getAccountMP(account)) <= to_mathint(getAccountMaxMP(account))
  filtered {
    f -> f.selector != sig:upgradeToAndCall(address,bytes).selector
  }

invariant accountMPGreaterEqualAccountStakedBalance(address account)
  to_mathint(getAccountMP(account)) >= to_mathint(getAccountStakedBalance(account))
  filtered {
    f -> f.selector != sig:upgradeToAndCall(address,bytes).selector
  }

rule stakingMintsMultiplierPoints1To1Ratio {

  env e;
  uint256 amount;
  uint256 lockupTime;
  uint256 multiplierPointsBefore;
  uint256 multiplierPointsAfter;

  requireInvariant accountMPGreaterEqualAccountStakedBalance(e.msg.sender);

  require getAccountLockUntil(e.msg.sender) <= e.block.timestamp;

  updateGlobalState(e);
  updateAccountMP(e, e.msg.sender);
  uint256 t = lastMPUpdatedTime();

  multiplierPointsBefore = getAccountMP(e.msg.sender);

  stake(e, amount, lockupTime);

  // we need to ensure time has not progressed because that would accrue MP
  // which makes it harder to proof this rule
  require lastMPUpdatedTime() == t;

  multiplierPointsAfter = getAccountMP(e.msg.sender);

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
  multiplierPointsAfter1 = getAccountMP(e.msg.sender);

  stake(e, amount, lockupTime2) at initalStorage;
  multiplierPointsAfter2 = getAccountMP(e.msg.sender);

  assert lockupTime1 >= lockupTime2 => to_mathint(multiplierPointsAfter1) >= to_mathint(multiplierPointsAfter2);
  satisfy to_mathint(multiplierPointsAfter1) > to_mathint(multiplierPointsAfter2);
}
