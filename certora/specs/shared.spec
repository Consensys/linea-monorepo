using RewardsStreamerMP as streamer;

function getVaultStakedBalance(address vault) returns uint256 {
    uint256 stakedBalance;
    stakedBalance, _, _, _, _, _, _, _ = streamer.vaultData(vault);
    return stakedBalance;
}

