import org.gradle.api.DefaultTask
import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.api.tasks.Input
import org.gradle.api.tasks.TaskAction
import org.web3j.codegen.SolidityFunctionWrapperGenerator

import java.nio.file.Files
import java.nio.file.Path

class GenerateContractWrappersTask extends DefaultTask {
  @Input
  Map<String, String> contracts

  @Input
  String contractsPackage

  @Input
  String generatedClassesDir = project.layout.buildDirectory
    .dir("generated/sources/web3j/main/java")
    .get().asFile.absolutePath

  GenerateContractWrappersTask() {
  }

  @TaskAction
  void action() {
    Path outputDir = Path.of(generatedClassesDir)
    println("Generating contract wrappers in ${outputDir.toAbsolutePath().toString()}")
    if (Files.exists(outputDir)) {
      outputDir.toFile().mkdirs()
    }

    contracts.each { abiFile, contractName ->
      println("Generating contract $contractName $abiFile")
      String[] params = [
        '--abiFile', abiFile,
        '--outputDir', outputDir.toAbsolutePath().toString(),
        '--package', contractsPackage,
        '--contractName', contractName
      ]
      SolidityFunctionWrapperGenerator.main(params)
    }
  }
}

class Web3jCodegenExtension {
  Map<String, String> contracts = [:]
  String contractsPackage = ""
  String generatedClassesDir = ""

  Web3jCodegenExtension(Project project) {
    generatedClassesDir = project.layout.buildDirectory
      .dir("generated/sources/web3j/main/java")
      .get().asFile.absolutePath
  }
}

class Web3jCodegenPlugin implements Plugin<Project> {
  def taskName = "generateWeb3jWrappers"
  @Override
  void apply(Project project) {
    Web3jCodegenExtension extension = project.extensions.create("web3jContractWrappers", Web3jCodegenExtension)

    project.tasks.register(taskName, GenerateContractWrappersTask) {
      group = "Code Generation"
      description = "Creates Web3J contract wrappers from ABIs files."

      // Configure the task using the extension properties
      contracts = extension.contracts
      contractsPackage = extension.contractsPackage
      generatedClassesDir = extension.generatedClassesDir
    }

    project.tasks.named("compileJava").configure {
      dependsOn project.tasks.named(taskName)
    }

    project.tasks.named("jar").configure {
      from(extension.generatedClassesDir)
    }

    project.afterEvaluate {
      project.sourceSets {
        main {
          java {
            srcDir extension.generatedClassesDir
          }
        }
      }
    }
  }
}
