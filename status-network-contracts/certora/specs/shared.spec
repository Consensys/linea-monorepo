using StakeManagerHarness as streamer;

function getVaultStakedBalance(address vault) returns uint256 {
    uint256 stakedBalance;
    stakedBalance, _, _, _, _, _ = streamer.vaultData(vault);
    return stakedBalance;
}

function getVaultMPAccrued(address vault) returns uint256 {
    uint256 vaultMPAccrued;
    _, _, vaultMPAccrued, _, _, _ = streamer.vaultData(vault);
    return vaultMPAccrued;
}

function getVaultMaxMP(address vault) returns uint256 {
    uint256 maxMP;
    _, _, _, maxMP, _, _= streamer.vaultData(vault);
    return maxMP;
}


