/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.api.logging.Logger
import org.gradle.api.provider.ProviderFactory

/**
 * Besu checkout/version resolution uses {@link org.gradle.api.provider.ProviderFactory#exec} so the configuration cache
 * can record the external process.
 * <p>
 * Task actions must not call {@link org.gradle.api.Task#getProject()} at execution time (configuration cache).
 * Captured values are bound at configuration time ({@link ProviderFactory}, paths, flags).
 */
class LineaBesuPlugin implements Plugin<Project> {

  @Override
  void apply(Project project) {
    def rootProject = project.rootProject
    def providers = rootProject.providers
    def rootDirFile = rootProject.layout.projectDirectory.asFile

    // Must run during configuration so rootProject.ext.resolvedBesuVer is visible
    // before besu-plugin-library (tracer/sequencer) configures besuPlugin { besuVersion = ... }.
    // linea-besu is declared first in settings.gradle to guarantee this runs before any
    // dependent subproject configures.    
    resolveBesuVersionFromCommit(project)
    buildAndPublishBesu(project)

    def resolvedVerForTasks = rootProject.ext.has('resolvedBesuVer') ? (rootProject.ext.resolvedBesuVer as String) : null
    def publishToMaven = project.hasProperty('publishToMaven') ? project.publishToMaven.toBoolean() : false
    def skipDownloadBesuDist = project.hasProperty('skipDownloadBesuDist') ? project.skipDownloadBesuDist.toBoolean() : true
    def cloudsmithUser = project.hasProperty('cloudsmithUser') ? project.cloudsmithUser.toString() : ''
    def cloudsmithApiKey = project.hasProperty('cloudsmithApiKey') ? project.cloudsmithApiKey.toString() : ''
    def mavenLocalBase = System.getProperty('maven.repo.local') ?: "${System.getProperty('user.home')}/.m2/repository"
    def rootPath = rootDirFile.absolutePath

    project.tasks.register('buildAndPublishBesu') { t ->
      group = 'Build'
      description = 'Build Besu at the resolved version (distTar publish/publishToMavenLocal)'
      doLast {
        LineaBesuPlugin.buildAndPublishBesuFromCaptured(
            t.logger,
            providers,
            rootDirFile,
            publishToMaven,
            skipDownloadBesuDist,
            resolvedVerForTasks,
            cloudsmithUser,
            cloudsmithApiKey
            )
      }
    }

    // Wire clean logic into :linea-besu:clean such that running 
    // "./gradlew clean" at the root also executes subproject clean tasks
    project.tasks.named('clean').configure { t ->
      doLast {
        LineaBesuPlugin.cleanBesuAndMavenLocal(t.logger, rootPath, mavenLocalBase)
      }
    }
  }

  /**
   * Resolve the besu version from besuCommit during configuration phase so that
   * every build script and external plugin (e.g. besu-plugin-library's afterEvaluate)
   * sees the correct version before any task runs.
   *
   * Uses {@code providers.exec} (not {@code ExecOperations}) so configuration cache can track the process.
   */
  private static String resolveBesuVersionFromCommit(Project project) {
    def rootProject = project.rootProject
    if (rootProject.ext.has('resolvedBesuVer')) {
      return rootProject.ext.resolvedBesuVer as String
    }

    def besuCommit = rootProject.libs.versions.besuCommit.get()
    def rootDir = rootProject.layout.projectDirectory.asFile
    def outputText = rootProject.providers.exec { spec ->
      spec.workingDir(rootDir)
      spec.environment('BESU_DIR', "${rootDir.absolutePath}/tmp/besu-eth")
      spec.environment('BESU_COMMIT', besuCommit)
      spec.environment('VERSION_LABEL', "")
      spec.commandLine('bash', "${rootDir.absolutePath}/linea-besu/scripts/checkout-and-resolve-version.sh")
    }.standardOutput.asText.get()

    def resolvedBesuVerOutput = outputText.trim().readLines().last()
    rootProject.ext.resolvedBesuVer = resolvedBesuVerOutput
    project.logger.lifecycle("Resolved besu version (configuration phase): ${resolvedBesuVerOutput}")

    return resolvedBesuVerOutput
  }

  private static void buildAndPublishBesu(Project project) {
    buildAndPublishBesuFromCaptured(
        project.logger,
        project.rootProject.providers,
        project.rootProject.layout.projectDirectory.asFile,
        project.hasProperty('publishToMaven') ? project.publishToMaven.toBoolean() : false,
        project.hasProperty('skipDownloadBesuDist') ? project.skipDownloadBesuDist.toBoolean() : true,
        project.rootProject.ext.resolvedBesuVer as String,
        project.hasProperty('cloudsmithUser') ? project.cloudsmithUser.toString() : '',
        project.hasProperty('cloudsmithApiKey') ? project.cloudsmithApiKey.toString() : ''
        )
  }

  static void buildAndPublishBesuFromCaptured(
      Logger logger,
      ProviderFactory providers,
      File rootDir,
      boolean publishToMaven,
      boolean skipDownloadBesuDist,
      String resolvedBesuVer,
      String cloudsmithUser,
      String cloudsmithApiKey
  ) {
    if (!resolvedBesuVer) {
      logger.lifecycle('Skipping buildAndPublishBesu: no resolved Besu version (run a non-clean-only build first)')
      return
    }
    def publishGradleTaskName = publishToMaven ? "publish" : "publishToMavenLocal"
    def shouldSkip = false
    if (publishToMaven) {
      shouldSkip = isBesuAvailableInMaven(logger, resolvedBesuVer) &&
          canDownloadBesuDistributionFromMaven(logger, resolvedBesuVer)
    } else {
      if (isBesuAndDistributionAvailableInMavenLocal(logger, rootDir, resolvedBesuVer)) {
        shouldSkip = true
      } else if (isBesuAvailableInMaven(logger, resolvedBesuVer) &&
          (skipDownloadBesuDist || downloadBesuDistributionFromMaven(logger, rootDir, resolvedBesuVer))) {
        if (skipDownloadBesuDist) {
          logger.lifecycle("Skipping download besu distribution from maven as skipDownloadBesuDist=${skipDownloadBesuDist}")
        }
        shouldSkip = true
      }
    }
    if (shouldSkip) {
      logger.lifecycle("Skipping buildAndPublishBesu: Besu ${resolvedBesuVer} already available")
      return
    }
    def rootPath = rootDir.absolutePath
    providers.exec { spec ->
      spec.workingDir(rootDir)
      spec.environment('BESU_DIR', "${rootPath}/tmp/besu-eth")
      spec.environment('RESOLVED_BESU_VERSION', resolvedBesuVer)
      spec.environment('CLOUDSMITH_USER', cloudsmithUser)
      spec.environment('CLOUDSMITH_API_KEY', cloudsmithApiKey)
      spec.commandLine('bash', "${rootPath}/linea-besu/scripts/build-dist-and-publish.sh", publishGradleTaskName)
    }.result.get()
  }

  static void cleanBesuAndMavenLocal(Logger logger, String rootPath, String mavenLocalBase) {
    deletePath(logger, new File(rootPath, 'tmp/besu-eth'))
    deletePath(logger, new File(mavenLocalBase, 'besu'))
    deletePath(logger, new File(mavenLocalBase, 'org/hyperledger/besu/bom'))
  }

  private static void deletePath(Logger logger, File f) {
    if (f.exists()) {
      if (f.isDirectory()) {
        f.deleteDir()
      } else {
        f.delete()
      }
      logger.lifecycle("Deleted: ${f.absolutePath}")
    }
  }

  private static boolean isBesuAndDistributionAvailableInMavenLocal(Logger logger, File rootDir, String version) {
    if (!version) return false
    def mavenLocalBase = System.getProperty('maven.repo.local') ?: "${System.getProperty('user.home')}/.m2/repository"
    def mavenLocalPom = new File(mavenLocalBase, "org/hyperledger/besu/bom/${version}/bom-${version}.pom")
    def pomExists = mavenLocalPom.exists()
    if (!pomExists) {
      logger.lifecycle("isBesuAndDistributionAvailableInMavenLocal: besu:${version} is not in maven local")
      return false
    } else {
      logger.lifecycle("isBesuAndDistributionAvailableInMavenLocal: Besu ${version} was found in maven local")
    }
    def besuDistTar = new File(rootDir, "tmp/besu-eth/build/distributions/besu-${version}.tar.gz")
    def distExists = besuDistTar.exists()
    if (!distExists) {
      logger.lifecycle("isBesuAndDistributionAvailableInMavenLocal: besu-${version}.tar.gz distribution doesn't exist")
    } else {
      logger.lifecycle("isBesuAndDistributionAvailableInMavenLocal: besu-${version}.tar.gz distribution was found under \"tmp/besu-eth/build/distributions\"")
    }
    return pomExists && distExists
  }

  private static boolean isBesuAvailableInMaven(Logger logger, String version) {
    if (!version) return false
    def candidates = [
        "https://artifacts.consensys.net/public/linea-besu/maven/org/hyperledger/besu/bom/${version}/bom-${version}.pom",
        "https://repo.maven.apache.org/maven2/org/hyperledger/besu/bom/${version}/bom-${version}.pom",
        "https://hyperledger.jfrog.io/hyperledger/besu-maven/org/hyperledger/besu/bom/${version}/bom-${version}.pom",
    ]
    def connectTimeoutMs = 5000
    def readTimeoutMs = 5000
    for (def pomUrl : candidates) {
      def conn = null
      try {
        conn = (java.net.HttpURLConnection) new URL(pomUrl).openConnection()
        conn.setConnectTimeout(connectTimeoutMs)
        conn.setReadTimeout(readTimeoutMs)
        conn.setRequestMethod('HEAD')
        if (conn.getResponseCode() == 200) {
          logger.lifecycle("isBesuAvailableInMaven: Besu ${version} found at maven repo: ${pomUrl}")
          return true
        }
      } catch (Exception ignored) {
      } finally {
        if (conn != null) conn.disconnect()
      }
    }
    logger.lifecycle("isBesuAvailableInMaven: Besu ${version} not found in any maven repo")
    return false
  }

  private static boolean canDownloadBesuDistributionFromMaven(Logger logger, String version) {
    if (!version) return false
    def baseUrl = "https://artifacts.consensys.net/public/linea-besu/raw/names/linea-besu.tar.gz/versions/"
    def url = "${baseUrl}${version}/besu-${version}.tar.gz"
    def conn = null
    try {
      conn = (java.net.HttpURLConnection) new URL(url).openConnection()
      conn.setConnectTimeout(10000)
      conn.setReadTimeout(60000)
      conn.setRequestMethod('HEAD')
      if (conn.getResponseCode() == 200) {
        logger.lifecycle("canDownloadBesuDistributionFromMaven: Found besu distribution from Maven (${url})")
        return true
      }
      return false
    } finally {
      if (conn != null) conn.disconnect()
    }
  }

  private static boolean downloadBesuDistributionFromMaven(Logger logger, File rootDir, String version) {
    if (!version) return false
    def destDir = new File(rootDir, "tmp/besu-eth/build/distributions")
    def destFile = new File(destDir, "besu-${version}.tar.gz")
    if (destFile.exists()) {
      logger.lifecycle("downloadBesuDistributionFromMaven: Found existing besu distribution at ${destFile}, skipping download")
      return true
    }
    def baseUrl = "https://artifacts.consensys.net/public/linea-besu/raw/names/linea-besu.tar.gz/versions/"
    def url = "${baseUrl}${version}/besu-${version}.tar.gz"
    def conn = null
    try {
      conn = (java.net.HttpURLConnection) new URL(url).openConnection()
      conn.setConnectTimeout(10000)
      conn.setReadTimeout(60000)
      conn.setRequestMethod('GET')
      destDir.mkdirs()

      if (conn.getResponseCode() != 200) {
        logger.lifecycle("downloadBesuDistributionFromMaven: Could not find and download besu distribution from Maven (${url})")
        return false
      }
      conn.getInputStream().withStream { input ->
        destFile.withOutputStream { it << input }
      }
      logger.lifecycle("downloadBesuDistributionFromMaven: Downloaded besu-${version}.tar.gz from Maven to ${destFile}")
      return true
    } catch (Exception e) {
      logger.lifecycle("downloadBesuDistributionFromMaven: Failed to download besu distribution from Maven (${url}): ${e.message}")
      if (destFile.exists()) {
        destFile.delete()
        logger.lifecycle("downloadBesuDistributionFromMaven: Removed partial/corrupt file so next run can retry: ${destFile}")
      }
      return false
    } finally {
      if (conn != null) conn.disconnect()
    }
  }
}
