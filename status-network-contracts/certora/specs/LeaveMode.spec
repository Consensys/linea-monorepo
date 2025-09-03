methods {
  function hasLeft() external returns (bool) envfree;
  function getInitializedVersion() external returns (uint8) envfree;
}

definition isOwnableFunction(method f) returns bool = (
  f.selector == sig:renounceOwnership().selector ||
  f.selector == sig:transferOwnership(address).selector
);

rule allowedActionsAfterLeaving(method f) {
  env e;
  calldataarg args;

  // Assuming vault has been initialized and marked as left
  require hasLeft() == true && getInitializedVersion() > 0;

  f@withrevert(e, args);
  bool reverted = lastReverted;

  assert !reverted => f.isView ||
    isOwnableFunction(f) ||
    f.selector == sig:initialize(address,address).selector ||
    f.selector == sig:leave(address).selector || // calling leave() is not going to do anything harmful
    f.selector == sig:withdraw(address,uint256).selector ||
    f.selector == sig:withdraw(address,uint256,address).selector ||
    f.selector == sig:unstake(uint256).selector ||
    f.selector == sig:unstake(uint256,address).selector ||
    // Calling register() is possible in cases where someone creates a vault without
    // registering and *then* calling leave() directly.
    // Unlikely scenario, but also not harmful, so no need to prevent it.
    f.selector == sig:register().selector ||
    // In practice, `migrateFromVault()` will only be called from stake manager when `migrateToVault()`
    // is called on the stake vault, which is prohibited when hasLeft() is true.
    f.selector == sig:migrateFromVault(IStakeVault.MigrationData).selector ||
    f.selector == sig:emergencyExit(address).selector;
}


