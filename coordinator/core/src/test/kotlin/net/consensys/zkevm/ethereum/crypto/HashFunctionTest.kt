package net.consensys.zkevm.ethereum.crypto

import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Test
import java.security.MessageDigest
import kotlin.random.Random

class HashFunctionTest {
  @Test
  fun `test single MessageDigest instance does not produce rolling hashes`() {
    val digester = MessageDigest.getInstance("SHA-256")
    for (i in 1..10) {
      val randomBytes = Random.nextBytes(32)
      assertArrayEquals(
        digester.digest(randomBytes),
        MessageDigest.getInstance("SHA-256").digest(randomBytes),
      )
    }
  }

  @Test
  fun `test Sha256HashFunction produces correct hash`() {
    val hashFunction = Sha256HashFunction()
    for (i in 1..10) {
      val randomBytes = Random.nextBytes(32)
      assertArrayEquals(
        hashFunction.hash(randomBytes),
        MessageDigest.getInstance("SHA-256").digest(randomBytes),
      )
    }
  }

  @Test
  fun `test Sha256HashFunction under concurrent calls produces correct hash`() {
    val hashFunction = Sha256HashFunction()
    val threads = List(10) {
      Thread {
        for (i in 1..1000) {
          val randomBytes = Random.nextBytes(32)
          assertArrayEquals(
            hashFunction.hash(randomBytes),
            MessageDigest.getInstance("SHA-256").digest(randomBytes),
          )
        }
      }
    }
    threads.forEach { it.start() }
    threads.forEach { it.join() }
  }
}
