import "./shared.spec";

using ERC20A as staked;

methods {
  function owner() external returns (address) envfree;
  function isInitializing() external returns (bool) envfree;
  function getExcessBalance() external returns (uint256) envfree;
  function depositedBalance() external returns (uint256) envfree;
  function hasLeft() external returns (bool) envfree;
  function lockUntil() external returns (uint256) envfree;
  function ERC20A.balanceOf(address) external returns (uint256) envfree;
  function ERC20A.allowance(address, address) external returns(uint256) envfree;
  function ERC20A.totalSupply() external returns(uint256) envfree;
  function StakeManagerHarness.vaultData(address) external returns(uint256, uint256, uint256, uint256, uint256, uint256) envfree;
  function StakeManagerHarness.isVaultTrusted(address) external returns (bool) envfree;
  function StakeManagerHarness.lastMPUpdatedTime() external returns (uint256) envfree;
  function StakeManagerHarness.totalStaked() external returns (uint256) envfree;
  function StakeManagerHarness.totalMPAccrued() external returns (uint256) envfree;
  function StakeManagerHarness.totalMaxMP() external returns (uint256) envfree;
  function _.owner() external => DISPATCHER(true);
  function _.transfer(address, uint256) external => DISPATCHER(true);
  function _.balanceOf(address) external => DISPATCHER(true);
  function _.lockUntil() external => DISPATCHER(true);
  function _.depositedBalance() external => DISPATCHER(true);
  function _.unstake(uint256) external => DISPATCHER(true);
  function _.migrateFromVault(IStakeVault.MigrationData) external => DISPATCHER(true);
}

ghost mapping(address => uint256) mirrorBalances
{
    init_state axiom (usum address a. mirrorBalances[a]) == 0;
}

hook Sstore staked._balances[KEY address a] uint256 newVal {
    mirrorBalances[a] = newVal;
}

hook Sload uint256 val staked._balances[KEY address a]  {
    require mirrorBalances[a] == val, "balance is mirrored";
}

invariant totalSupplyIsSumOfBalances()
  staked._totalSupply == (usum address a. mirrorBalances[a])
  filtered {
      f -> f.contract == staked
  }


invariant depositedBalanceLessEqualToERC20Balance()
  depositedBalance() <= staked.balanceOf(currentContract)
  filtered {
      f -> f.contract == currentContract &&
        f.selector != sig:migrateFromVault(IStakeVault.MigrationData).selector
  }
  { preserved with (env e) {
      require e.msg.sender != currentContract;
      require staked.allowance(currentContract, currentContract) == 0;
      requireInvariant totalSupplyIsSumOfBalances();
    }
  }

invariant depositedBalanceAndExcessBalanceEqualsERC20Balance()
  depositedBalance() + getExcessBalance() == staked.balanceOf(currentContract)
  filtered {
      f -> f.contract == currentContract &&
        f.selector != sig:migrateFromVault(IStakeVault.MigrationData).selector
  }
  {
      preserved with (env e) {
        require e.msg.sender != currentContract;
        require staked.allowance(currentContract, currentContract) == 0;
        requireInvariant totalSupplyIsSumOfBalances();
      }
  }

invariant depositedBalanceEqualsStakedBalance()
    depositedBalance() == getVaultStakedBalance(currentContract)
    filtered {
        f -> f.contract == currentContract &&
            f.selector != sig:emergencyExit(address).selector &&
            // leave() is hard to proof because the prover will find all sort
            // of revert cases inside of `leave()` that will make it find counter examples
            f.selector != sig:leave(address).selector  &&
            // this is only called from StakeManager and is intended to reset `depositedBalacne()`
            f.selector != sig:migrateFromVault(IStakeVault.MigrationData).selector
    }
    {
        // for the cases where `withdraw()` is called with an amount
        // that decreases `depositedBalance`, we have to assume that
        // the vault has called `leave()` first and therefore its staked
        // balance is 0
        preserved withdraw(address token, uint256 amount) with (env e) {
            if (hasLeft()) {
                require getVaultStakedBalance(currentContract) == 0;
            }
        }
        preserved withdraw(address token, uint256 amount, address destination) with (env e) {
            if (hasLeft()) {
                require getVaultStakedBalance(currentContract) == 0;
            }
        }
    }


invariant stakedBalanceZeroIfDepositedBalanceZero()
    getVaultStakedBalance(currentContract) == 0 => depositedBalance() == 0
    filtered {
        f -> f.contract == currentContract &&
          f.selector != sig:migrateFromVault(IStakeVault.MigrationData).selector &&
          f.selector != sig:leave(address).selector
    }
    {
        preserved with (env e) {
            requireInvariant depositedBalanceEqualsStakedBalance();
        }
    }


invariant depositedBalanceZeroIfERC20BalanceZero()
  staked.balanceOf(currentContract) == 0 => depositedBalance() == 0
  filtered {
      f -> f.contract == currentContract &&
        f.selector != sig:migrateFromVault(IStakeVault.MigrationData).selector
  }
  {
    preserved with (env e) {
      requireInvariant depositedBalanceLessEqualToERC20Balance();
      require e.msg.sender != currentContract;
      require staked.allowance(currentContract, currentContract) == 0;
      requireInvariant totalSupplyIsSumOfBalances();
    }
  }

rule ownerCannotChange(method f) {
  address owner = owner();
  env e;
  calldataarg args;

  // in case the starting value of initialing is true
  bool initializing = isInitializing();
  bool isInitializerFunction = f.selector == sig:initialize(address,address).selector;

  f(e, args);

  assert !initializing && !isInitializerFunction => owner() == owner;
}

rule ownerCanOnlyWithdrawStakedFundsWhenNotLocked(method f) filtered {
    f -> f.selector != sig:renounceOwnership().selector &&
    f.selector != sig:initialize(address,address).selector &&
    f.selector != sig:emergencyExit(address).selector &&
    f.contract == currentContract
} {
    env e;
    calldataarg args;

    require e.block.timestamp > 0;
    require e.block.timestamp < max_uint64;
    require lockUntil() < max_uint64;

    uint256 ownerBalanceBefore = staked.balanceOf(owner());
    uint256 depositedBefore = depositedBalance();

    f(e, args);

    uint256 ownerBalanceAfter = staked.balanceOf(owner());
    uint256 depositedAfter = depositedBalance();

    assert ownerBalanceAfter > ownerBalanceBefore && depositedAfter < depositedBefore => lockUntil() <= e.block.timestamp;
}


rule reachability(method f) {
  calldataarg args;
  env e;
  f(e,args);
  satisfy true;
}


// check that the ERC20.balanceOf(vault) is >= to StakeManager.vaultData[a].stakedBalance
invariant vaultBalanceVsERC20Balance()
  staked.balanceOf(currentContract) >= getVaultStakedBalance(currentContract)
  filtered {
    f -> f.contract == currentContract &&
      f.selector != sig:emergencyExit(address).selector &&
      f.selector != sig:leave(address).selector
  }
  { preserved with (env e) {
      // the sender can't be the vault otherwise it can transfer tokens
      require e.msg.sender != currentContract;
      // nobody has allowance to spend the tokens of the vault
      require staked.allowance(currentContract, e.msg.sender) == 0;
      require staked.allowance(currentContract, currentContract) == 0;

      requireInvariant totalSupplyIsSumOfBalances();
      requireInvariant depositedBalanceEqualsStakedBalance();
      requireInvariant stakedBalanceZeroIfDepositedBalanceZero();
      requireInvariant depositedBalanceLessEqualToERC20Balance();
    }

    preserved withdraw(address token, uint256 amount) with (env e) {
        if (hasLeft()) {
            require getVaultStakedBalance(currentContract) == 0;
        }
        requireInvariant totalSupplyIsSumOfBalances();
        requireInvariant depositedBalanceEqualsStakedBalance();
        requireInvariant depositedBalanceLessEqualToERC20Balance();
    }

    preserved withdraw(address token, uint256 amount, address destination) with (env e) {
        if (hasLeft()) {
            require getVaultStakedBalance(currentContract) == 0;
        }
        requireInvariant totalSupplyIsSumOfBalances();
        requireInvariant depositedBalanceEqualsStakedBalance();
        requireInvariant depositedBalanceLessEqualToERC20Balance();
    }
  }

/*
  The rule below is commented out as it's merely used to easily have the
  prover find all functions that change balances.

rule whoChangeERC20Balance(method f)
{
  simplification();
  address user;
  uint256 before = staked.balanceOf(user);
  calldataarg args;
  env e;
  f(e,args);
  assert before == staked.balanceOf(user);
}
*/
