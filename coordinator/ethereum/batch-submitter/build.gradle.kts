@file:Suppress("UNCHECKED_CAST")

import com.avast.gradle.dockercompose.ComposeExtension
import java.io.InputStreamReader
import java.time.Duration

plugins {
  id("net.consensys.zkevm.kotlin-library-conventions")
  id("com.avast.gradle.docker-compose").version("0.14.2")
}

dependencies {
  val versions: Map<String, String> = project.ext["versions"] as Map<String, String>

  api(project(":coordinator:core"))
  implementation(project(":coordinator:clients:prover-client"))
  implementation("io.vertx:vertx-core:${versions["vertx"]}")
  implementation("tech.pegasys.teku.internal:executionclient:${versions["teku"]}")
  implementation("org.web3j:crypto:${versions["web3j"]}") {
    exclude(group = "org.slf4j", module = "slf4j-nop")
  }
  implementation("org.web3j:core:${versions["web3j"]}") {
    exclude(group = "org.slf4j", module = "slf4j-nop")
  }
  implementation(project(":coordinator:clients:smart-contract-client"))
  implementation(project(":jvm-libs:future-extensions"))
  api("io.vertx:vertx-pg-client:${versions["vertx"]}")
  implementation("com.ongres.scram:common:2.1") {
    because("Vertx pg client fails without it")
  }
  implementation("com.ongres.scram:client:2.1") {
    because("Vertx pg client fails without it")
  }
  implementation("org.postgresql:postgresql:42.6.0")
  implementation("org.flywaydb:flyway-core:8.4.3")
  implementation("org.slf4j:slf4j-api:1.7.30") {
    because("Flyway DB and other dependencies use SLF4J")
  }

  testImplementation("io.vertx:vertx-junit5:${versions["vertx"]}")
}

tasks.test {
  useJUnitPlatform()
}

configure<ComposeExtension> {
  useComposeFiles = listOf("${project.rootDir.path}/docker/compose.yml")
  waitForHealthyStateTimeout = Duration.ofMinutes(1)
  startedServices = listOf(
    "postgres",
    "l1-node",
    "l1-validator",
    // For debug
//    "l1-blockscout"
  )
  environment.putAll(
    mapOf(
      "POSTGRES_USER" to "coordinator",
      "POSTGRES_DB" to "coordinator_tests",
      "POSTGRES_PASSWORD" to "coordinator_tests"
    )
  )
  waitForTcpPorts = false
  removeOrphans = true

  projectName = "docker"
}

sourceSets {
  val integrationTest by creating {
    val mainOutput = sourceSets["main"].output
    compileClasspath += mainOutput
    runtimeClasspath += mainOutput
  }
}

configurations {
  val integrationTestImplementation by getting {
    extendsFrom(configurations.testImplementation.get())
  }
}

tasks.register<Test>("integrationTest") {
  testLogging.showStandardStreams = true
  val test = this
  description = "Runs integration tests."
  group = "verification"

  useJUnitPlatform()
  mustRunAfter(":coordinator:app:integrationTest")
  doFirst {
    val contractAddress = deployContracts()
    test.systemProperty("ContractAddress", contractAddress)
  }
  finalizedBy(tasks.composeDown)

  testClassesDirs = sourceSets.getByName("integrationTest").output.classesDirs
  classpath = sourceSets["integrationTest"].runtimeClasspath

  dependsOn(tasks.composeUp)
}

fun deployContracts(): String {
  // Container can start afresh and this file becomes outdated and causes issues
  delete(project.rootProject.file("./contracts/.openzeppelin/unknown-31648428.json"))
  println("Running contract deployment scripts")
  val outputFile = file("output.txt")
  val deploymentProcessBuilder = ProcessBuilder("make", "-C", project.rootDir.path, "deploy-zkevm2-to-local")
  deploymentProcessBuilder.environment()["VERIFIER_CONTRACT_NAME"] = "IntegrationTestTrueVerifier"
  deploymentProcessBuilder.redirectOutput(outputFile)
  deploymentProcessBuilder.redirectError(outputFile)
  val deploymentProcess = deploymentProcessBuilder.start()
  val deploymentResult = deploymentProcess.waitFor(4, TimeUnit.MINUTES)
  val output = printInputStreamReader(outputFile.reader())
  outputFile.delete()
  if (!deploymentResult) {
    throw GradleException("Deployment timed out")
  }
  val deploymentAddressLine = output.split("\n").find { it.startsWith("ZkEvmV2 deployed at ") }!!
  val contractAddressMatch = Regex("^ZkEvmV2 deployed at (.*)$").matchEntire(deploymentAddressLine)
  return if (contractAddressMatch == null) {
    throw IllegalStateException("Couldn't extract contract address from the output: $output")
  } else {
    contractAddressMatch.groupValues[1]
  }
}

fun printInputStreamReader(reader: InputStreamReader): String {
  val output = reader.use {
    it.readText()
  }
  println(output)
  return output
}

tasks.check {
  finalizedBy("integrationTest")
}
