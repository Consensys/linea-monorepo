using RewardsStreamerMP as streamer;
using ERC20A as staked;

methods {
    function emergencyModeEnabled() external returns (bool) envfree;
}

definition isViewFunction(method f) returns bool = (
  f.selector == sig:streamer.STAKING_TOKEN().selector ||
  f.selector == sig:streamer.SCALE_FACTOR().selector ||
  f.selector == sig:streamer.MP_RATE_PER_YEAR().selector ||
  f.selector == sig:streamer.MIN_LOCKUP_PERIOD().selector ||
  f.selector == sig:streamer.MAX_LOCKUP_PERIOD().selector ||
  f.selector == sig:streamer.MAX_MULTIPLIER().selector ||
  f.selector == sig:streamer.rewardIndex().selector ||
  f.selector == sig:streamer.lastMPUpdatedTime().selector ||
  f.selector == sig:streamer.owner().selector ||
  f.selector == sig:streamer.totalStaked().selector ||
  f.selector == sig:streamer.totalMaxMP().selector ||
  f.selector == sig:streamer.totalMP().selector ||
  f.selector == sig:streamer.accounts(address).selector ||
  f.selector == sig:streamer.emergencyModeEnabled().selector ||
  f.selector == sig:streamer.getStakedBalance(address).selector ||
  f.selector == sig:streamer.getAccount(address).selector ||
  f.selector == sig:streamer.rewardsBalanceOf(address).selector ||
  f.selector == sig:streamer.totalRewardsSupply().selector ||
  f.selector == sig:streamer.calculateAccountRewards(address).selector
);

definition isOwnableFunction(method f) returns bool = (
  f.selector == sig:streamer.renounceOwnership().selector ||
  f.selector == sig:streamer.transferOwnership(address).selector
);

definition isTrustedCodehashAccessFunction(method f) returns bool = (
  f.selector == sig:streamer.setTrustedCodehash(bytes32, bool).selector ||
  f.selector == sig:streamer.isTrustedCodehash(bytes32).selector
);

definition isInitializerFunction(method f) returns bool = (
  f.selector == sig:streamer.initialize(address,address).selector
);

definition isUUPSUpgradeableFunction(method f) returns bool = (
  f.selector == sig:streamer.proxiableUUID().selector ||
  f.selector == sig:streamer.UPGRADE_INTERFACE_VERSION().selector ||
  f.selector == sig:streamer.upgradeToAndCall(address, bytes).selector ||
  f.selector == sig:streamer.__TrustedCodehashAccess_init(address).selector
);

rule accountCanOnlyLeaveInEmergencyMode(method f) {
  env e;
  calldataarg args;

  require emergencyModeEnabled() == true;

  f@withrevert(e, args);
  bool isReverted = lastReverted;

  assert !isReverted => f.selector == sig:streamer.leave().selector ||
                        isViewFunction(f) ||
                        isOwnableFunction(f) ||
                        isTrustedCodehashAccessFunction(f) ||
                        isInitializerFunction(f) ||
                        isUUPSUpgradeableFunction(f);
}

