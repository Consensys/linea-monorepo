package net.consensys.linea.contract.l1

import net.consensys.encodeHex
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.contract.LineaRollup
import net.consensys.toBigInteger
import net.consensys.toULong
import net.consensys.zkevm.coordinator.clients.smartcontract.BlockParameter
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaContractVersion
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaRollupSmartContractClientReadOnly
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.tx.gas.StaticGasProvider
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.atomic.AtomicReference

private val fakeCredentials = Credentials.create(ByteArray(32).encodeHex())

fun BlockParameter.toWeb3j(): DefaultBlockParameter {
  return when (this) {
    is BlockParameter.Tag -> DefaultBlockParameter.valueOf(this.getTag())
    is BlockParameter.BlockNumber -> DefaultBlockParameter.valueOf(this.getNumber().toBigInteger())
  }
}

open class Web3JLineaRollupSmartContractClientReadOnly(
  val web3j: Web3j,
  val contractAddress: String,
  private val log: Logger = LogManager.getLogger(Web3JLineaRollupSmartContractClientReadOnly::class.java)
) : LineaRollupSmartContractClientReadOnly {

  protected fun contractClientAtBlock(blockParameter: BlockParameter): LineaRollup {
    return LineaRollup.load(
      contractAddress,
      web3j,
      fakeCredentials,
      StaticGasProvider(BigInteger.ZERO, BigInteger.ZERO)
    ).apply {
      this.setDefaultBlockParameter(blockParameter.toWeb3j())
    }
  }

  protected val smartContractVersionCache: AtomicReference<LineaContractVersion> =
    AtomicReference(fetchSmartContractVersion().get())

  private fun getSmartContractVersion(): SafeFuture<LineaContractVersion> {
    return if (smartContractVersionCache.get() == LineaContractVersion.V5) {
      // once upgraded, it's not downgraded
      SafeFuture.completedFuture(LineaContractVersion.V5)
    } else {
      fetchSmartContractVersion().thenPeek { contractLatestVersion ->
        if (contractLatestVersion != smartContractVersionCache.get()) {
          log.info(
            "Smart contract upgraded: prevVersion={} upgradedVersion={}",
            smartContractVersionCache.get(),
            contractLatestVersion
          )
        }
        smartContractVersionCache.set(contractLatestVersion)
      }
    }
  }

  private fun fetchSmartContractVersion(): SafeFuture<LineaContractVersion> {
    // FIXME: this is a temporary solution to determine the smart contract version.
    //  It should rely on events
    return SafeFuture.completedFuture(LineaContractVersion.V5)
  }

  override fun getAddress(): String = contractAddress

  override fun getVersion(): SafeFuture<LineaContractVersion> = getSmartContractVersion()

  override fun finalizedL2BlockNumber(blockParameter: BlockParameter): SafeFuture<ULong> {
    return contractClientAtBlock(blockParameter)
      .currentL2BlockNumber().sendAsync()
      .thenApply { it.toULong() }
      .toSafeFuture()
  }

  override fun finalizedL2BlockTimestamp(blockParameter: BlockParameter): SafeFuture<ULong> {
    return contractClientAtBlock(blockParameter).currentTimestamp().sendAsync()
      .thenApply { it.toULong() }
      .toSafeFuture()
  }

  override fun getMessageRollingHash(blockParameter: BlockParameter, messageNumber: Long): SafeFuture<ByteArray> {
    require(messageNumber >= 0) { "messageNumber must be greater than or equal to 0" }

    return contractClientAtBlock(blockParameter).rollingHashes(messageNumber.toBigInteger()).sendAsync().toSafeFuture()
  }

  override fun findBlobFinalBlockNumberByShnarf(blockParameter: BlockParameter, shnarf: ByteArray): SafeFuture<ULong?> {
    return contractClientAtBlock(blockParameter)
      .shnarfFinalBlockNumbers(shnarf).sendAsync()
      .thenApply { if (it == BigInteger.ZERO) null else it.toULong() }
      .toSafeFuture()
  }

  override fun blockStateRootHash(blockParameter: BlockParameter, lineaL2BlockNumber: ULong): SafeFuture<ByteArray> {
    return contractClientAtBlock(blockParameter)
      .stateRootHashes(lineaL2BlockNumber.toBigInteger()).sendAsync()
      .toSafeFuture()
  }
}
