import "./shared.spec";

using ERC20A as staked;
using StakeManager as stakeManager;

methods {
  function ERC20A.balanceOf(address) external returns (uint256) envfree;
  function ERC20A.allowance(address, address) external returns(uint256) envfree;
  function ERC20A.totalSupply() external returns(uint256) envfree;
  function StakeManager.accounts(address) external returns(uint256, uint256, uint256, uint256, uint256, uint256) envfree;
  function _.owner() external => DISPATCHER(true);
  function _.transfer(address, uint256) external => DISPATCHER(true);
}

// check that the ERC20.balanceOf(vault) is >= to StakeManager.accounts[a].balance
invariant accountBalanceVsERC20Balance()
  staked.balanceOf(currentContract) >= getAccountStakedBalance(currentContract)
  { preserved with (env e) {
      // the sender can't be the vault otherwise it can transfer tokens
      require e.msg.sender != currentContract;
      // nobody has allowance to spend the tokens of the vault
      require staked.allowance(currentContract, e.msg.sender) == 0;
      // if it's a generic transfer to the vault, it can't overflow
      require staked.balanceOf(currentContract) + staked.balanceOf(e.msg.sender) <= to_mathint(staked.totalSupply());
      // if it's a transfer from the StakeManager to the vault as reward address, it can't overflow
      require staked.balanceOf(currentContract) + staked.balanceOf(stakeManager) <= to_mathint(staked.totalSupply());
    }

    // the next blocked is run instead of the general one if the current function is staked.transferFrom.
    // if it's a transferFrom, we don't have the from in the first preserved block to check for an overflow
    preserved staked.transferFrom(address from, address to, uint256 amount) with (env e) {
      // if the msg.sender is the vault than it would be able to move tokens.
      // it would be possible only if the Vault contract called the ERC20.transferFrom.
      require e.msg.sender != currentContract;
      // no one has allowance to move tokens owned by the vault
      require staked.allowance(currentContract, e.msg.sender) == 0;
      require staked.balanceOf(from) + staked.balanceOf(to) <= to_mathint(staked.totalSupply());
    }

    preserved stake(uint256 amount, uint256 duration) with (env e) {

      require e.msg.sender != currentContract;

      require staked.balanceOf(currentContract) + staked.balanceOf(e.msg.sender) + staked.balanceOf(stakeManager) <= to_mathint(staked.totalSupply());
    }

    preserved stake(uint256 amount, uint256 duration, address from) with (env e) {
      require e.msg.sender != currentContract;
      require from != currentContract;
      require staked.balanceOf(currentContract) + staked.balanceOf(from) + staked.balanceOf(stakeManager) <= to_mathint(staked.totalSupply());
    }
  }


rule reachability(method f) {
  calldataarg args;
  env e;
  f(e,args);
  satisfy true;
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
