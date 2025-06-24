package linea.coordinator.config.v2

import java.io.File
import java.net.URL

data class SignerConfig(
  val type: SignerType,
  val web3j: Web3jConfig?,
  val web3signer: Web3SignerConfig?,
) {
  init {
    when {
      type == SignerType.WEB3J && web3j == null -> {
        throw IllegalArgumentException("signetType=$type requires web3j config")
      }

      type == SignerType.WEB3SIGNER && web3signer == null -> {
        throw IllegalArgumentException("signetType=$type requires web3signer config")
      }

      type == SignerType.WEB3SIGNER &&
        web3signer?.tls != null &&
        web3signer.endpoint.protocol != "https" -> {
        throw IllegalArgumentException("signetType=$type with TLS configs requires endpoint URL in https")
      }
    }
  }

  enum class SignerType() {
    WEB3J,
    WEB3SIGNER,
  }

  data class Web3jConfig(
    val privateKey: ByteArray,
  ) {
    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false

      other as Web3jConfig

      return privateKey.contentEquals(other.privateKey)
    }

    override fun hashCode(): Int {
      return privateKey.hashCode()
    }

    override fun toString(): String {
      return "Web3jConfig(privateKey=***${privateKey.size}bytes***)"
    }
  }

  data class Web3SignerConfig(
    val endpoint: URL,
    val publicKey: ByteArray,
    val maxPoolSize: Int = 10,
    val keepAlive: Boolean = true,
    val tls: TlsConfig?,
  ) {
    data class TlsConfig(
      val keyStorePath: String,
      val keyStorePassword: String,
      val trustStorePath: String,
      val trustStorePassword: String,
    ) {
      init {
        require(!keyStorePassword.isEmpty()) { "keyStorePassword must not be empty" }
        require(!trustStorePassword.isEmpty()) { "trustStorePassword must not be empty" }
        require(File(keyStorePath).exists()) { "keyStorePath=$keyStorePath must point to an existing file" }
        require(File(trustStorePath).exists()) { "trustStorePath=$trustStorePath must point to an existing file" }
      }

      override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false

        other as TlsConfig

        if (keyStorePath != other.keyStorePath) return false
        if (keyStorePassword != other.keyStorePassword) return false
        if (trustStorePath != other.trustStorePath) return false
        if (trustStorePassword != other.trustStorePassword) return false

        return true
      }

      override fun hashCode(): Int {
        var result = keyStorePath.hashCode()
        result = 31 * result + keyStorePassword.hashCode()
        result = 31 * result + trustStorePath.hashCode()
        result = 31 * result + trustStorePassword.hashCode()
        return result
      }
    }

    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false

      other as Web3SignerConfig

      if (maxPoolSize != other.maxPoolSize) return false
      if (keepAlive != other.keepAlive) return false
      if (endpoint != other.endpoint) return false
      if (!publicKey.contentEquals(other.publicKey)) return false

      return true
    }

    override fun hashCode(): Int {
      var result = maxPoolSize
      result = 31 * result + keepAlive.hashCode()
      result = 31 * result + endpoint.hashCode()
      result = 31 * result + publicKey.contentHashCode()
      return result
    }
  }
}
