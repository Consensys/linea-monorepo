package net.consensys.zkevm.load.model.swagger

import net.consensys.zkevm.load.model.JSON
import net.consensys.zkevm.load.model.Wallet
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test
import org.web3j.crypto.Keys
import java.math.BigInteger

class TestJson {

  @Test
  fun testToJsonWallet() {
    val keyPair = Keys.createEcKeyPair()
    val wallet = Wallet(keyPair.privateKey.toString(16), 0, BigInteger.ZERO)
    val walletJson = JSON.createGson().create().toJson(wallet)

    assertEquals("""{"id":0,"initialNonce":0,"address":"${wallet.address}"}""", walletJson)
    assertFalse(walletJson.contains(wallet.privateKey))
    assertFalse(walletJson.contains("credentials"))
  }

  @Test
  fun testMapSerializationDoesNotLeakWalletSecrets() {
    val keyPair = Keys.createEcKeyPair()
    val wallet = Wallet(keyPair.privateKey.toString(16), 0, BigInteger.ZERO)

    val reportJson = JSON.createGson().create().toJson(mapOf(wallet to listOf(1)))

    assertFalse(reportJson.contains(wallet.privateKey))
    assertFalse(reportJson.contains("credentials"))
    assertTrue(reportJson.contains(wallet.address))
  }
}
