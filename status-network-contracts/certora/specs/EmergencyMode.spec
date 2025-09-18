using StakeManager as streamer;
using ERC20A as staked;

methods {
    function emergencyModeEnabled() external returns (bool) envfree;
}

definition isAccessControlFunction(method f) returns bool = (
  f.selector == sig:streamer.grantRole(bytes32, address).selector ||
  f.selector == sig:streamer.revokeRole(bytes32, address).selector ||
  f.selector == sig:renounceRole(bytes32, address).selector
);

definition isTrustedCodehashAccessFunction(method f) returns bool = (
  f.selector == sig:streamer.setTrustedCodehash(bytes32, bool).selector
);

definition isPausableFunction(method f) returns bool = (
  f.selector == sig:streamer.pause().selector ||
  f.selector == sig:streamer.unpause().selector
);

definition isInitializerFunction(method f) returns bool = (
  f.selector == sig:streamer.initialize(address,address,address).selector
);

definition isUUPSUpgradeableFunction(method f) returns bool = (
  f.selector == sig:streamer.upgradeTo(address).selector ||
  f.selector == sig:streamer.upgradeToAndCall(address, bytes).selector ||
  f.selector == sig:streamer.__TrustedCodehashAccess_init(address).selector
);

definition noCallDuringEmergency(method f) returns bool = (
  f.selector == sig:streamer.updateGlobalState().selector
                || f.selector == sig:streamer.setRewardsSupplier(address).selector
                || f.selector == sig:streamer.registerVault().selector
                || f.selector == sig:streamer.migrateToVault(address).selector
                || f.selector == sig:streamer.updateAccount(address).selector
                || f.selector == sig:streamer.updateVault(address).selector
                || f.selector == sig:streamer.unstake(uint256).selector
                || f.selector == sig:streamer.stake(uint256, uint256, uint256).selector
                || f.selector == sig:streamer.lock(uint256, uint256).selector
                || f.selector == sig:streamer.setReward(uint256, uint256).selector
                || f.selector == sig:enableEmergencyMode().selector
                || f.selector == sig:pause().selector
                || f.selector == sig:unpause().selector
                || f.selector == sig:redeemRewards(address).selector
);

rule allowedActionsInEmergencyMode(method f) {
  env e;
  calldataarg args;

  require emergencyModeEnabled() == true;

  f@withrevert(e, args);
  bool isReverted = lastReverted;

  assert !isReverted => f.selector == sig:streamer.leave().selector ||
                        f.isView ||
                        isAccessControlFunction(f) ||
                        isPausableFunction(f) ||
                        isTrustedCodehashAccessFunction(f) ||
                        isInitializerFunction(f) ||
                        isUUPSUpgradeableFunction(f);
}

rule cantBeCalledInEmergency(method f)
{
    env e;
    calldataarg args;

    bool inEmergencyMode = emergencyModeEnabled();

    f@withrevert(e, args);
    bool isReverted = lastReverted;

    assert inEmergencyMode && noCallDuringEmergency(f) => isReverted;

    satisfy !noCallDuringEmergency(f) => !isReverted && inEmergencyMode;
}


