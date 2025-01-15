import org.gradle.api.provider.ListProperty
import org.gradle.api.provider.Property
import org.gradle.api.tasks.Exec
import org.gradle.api.tasks.Input
import org.gradle.api.tasks.Optional

abstract class TraceFilesTask extends Exec {

  @Input
  @Optional
  abstract Property<String> getClassName()

  @Input
  abstract Property<String> getModule()

  @Input
  abstract ListProperty<String> getFiles()

  @Input
  @Optional
  abstract Property<String> getModuleDir()

  @Override
  protected void exec() {
    def arguments = ["besu",
                     "-P", "${moduleDir.getOrElse(module.get()).replaceAll('/','.')}",
                     "-o", "${project.rootDir}/arithmetization/src/main/java/net/consensys/linea/zktracer/module/${moduleDir.getOrElse(module.get())}",
                     "-c", "${className.getOrElse('Trace')}"
    ]
    arguments.addAll(files.get().collect({"linea-constraints/${it}"}))

    workingDir project.rootDir
    executable"which corset"
    executable "corset"
    args arguments

    println "Generating traces for ${module.get()} from ${arguments}"
    super.exec()
  }
}
