package net.consensys.linea.jwt

import org.apache.tuweni.bytes.Bytes
import java.nio.file.Files
import java.nio.file.Path
import javax.crypto.spec.SecretKeySpec

fun loadJwtSecretFromFile(jwtSecretFile: Path): SecretKeySpec {
  val bytesFromHex = Bytes.fromHexString(Files.readString(jwtSecretFile).trim())
  return SecretKeySpec(bytesFromHex.toArray(), io.jsonwebtoken.SignatureAlgorithm.HS256.jcaName)
}
