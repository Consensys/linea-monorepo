import org.gradle.api.DefaultTask
import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.api.artifacts.type.ArtifactTypeDefinition
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

    project.afterEvaluate {
      // Compile generated sources in a dedicated source set so that -Werror can be
      // suppressed for generated code without affecting hand-written sources in main.
      project.sourceSets {
        generatedWeb3j {
          java {
            srcDir extension.generatedClassesDir
          }
          // inherit main's compile classpath (compileOnly, api, implementation)
          compileClasspath += project.sourceSets.main.compileClasspath
        }
      }

      def generatedCompileTask = project.tasks.named("compileGeneratedWeb3jJava")
      def generatedClassesDirs = project.sourceSets.generatedWeb3j.output.classesDirs

      generatedCompileTask.configure {
        dependsOn project.tasks.named(taskName)
        // Generated sources are not held to -Werror; suppress it explicitly so that
        // the root build.gradle doFirst heuristic is no longer needed for this task.
        doFirst {
          options.compilerArgs.remove('-Werror')
        }
      }

      // Register the generated classes directory in the 'classes' secondary variant of
      // apiElements and runtimeElements so that consuming project() dependencies resolve
      // them directly (compile-avoidance path) without sharing compileJava's output dir.
      ["apiElements", "runtimeElements"].each { configName ->
        project.configurations.named(configName).configure { config ->
          config.outgoing.variants.named("classes") { variant ->
            generatedClassesDirs.each { dir ->
              variant.artifact(dir) {
                type = ArtifactTypeDefinition.JVM_CLASS_DIRECTORY
                builtBy generatedCompileTask
              }
            }
          }
        }
      }

      // Include generated compiled classes in the jar artifact.
      // dependsOn is inferred: classesDirs carries builtBy(compileGeneratedWeb3jJava).
      project.tasks.named("jar").configure {
        from generatedClassesDirs
      }
    }
  }
}
