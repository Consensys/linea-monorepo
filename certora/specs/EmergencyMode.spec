using RewardsStreamerMP as streamer;
using ERC20A as staked;

methods {
    function emergencyModeEnabled() external returns (bool) envfree;
}

definition isOwnableFunction(method f) returns bool = (
  f.selector == sig:streamer.renounceOwnership().selector ||
  f.selector == sig:streamer.transferOwnership(address).selector ||
  f.selector == sig:streamer.setReward(uint256, uint256).selector
);

definition isTrustedCodehashAccessFunction(method f) returns bool = (
  f.selector == sig:streamer.setTrustedCodehash(bytes32, bool).selector
);

definition isInitializerFunction(method f) returns bool = (
  f.selector == sig:streamer.initialize(address,address).selector
);

definition isUUPSUpgradeableFunction(method f) returns bool = (
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
                        f.isView ||
                        isOwnableFunction(f) ||
                        isTrustedCodehashAccessFunction(f) ||
                        isInitializerFunction(f) ||
                        isUUPSUpgradeableFunction(f);
}

