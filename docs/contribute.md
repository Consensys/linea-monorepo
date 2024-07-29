# Contributing

## Introduction

Thank you for your interest in contributing to Linea's zk-EVM monorepo. This document provides guidelines and instructions on how to contribute effectively. We use trunk-based development, and our main languages are Kotlin, Golang, Solidity, and TypeScript. Our releases are done using GitOps, ArgoCD, Helm, and Kubernetes.

## Submit a Contribution

## Prerequisites

Before contributing, please ensure you're familiar with:

- Our [development guidelines](development-guidelines.md) on how to write code, tests, and documentation.
- Our [code of conduct](code-of-conduct.md).

### Create an issue

* Check beforehand that the issue you want to raise isn't present (even with other keywords) to avoid creating duplicates,
* Use an issue template. Using the correct one will aid other contributors in responding to your issue. Use `Feature request`, `Bug report` or `Operational Task`, depending on the context.
* Follow any directions described in the issue template itself. Be descriptive with the issue you are raising.
* Don't forget to add any impact on:
  * Internal changes, i.e. features that impact `Metrics`, `Runbooks`, `Operations`
  * External changes, i.e. features that impact `Documentation`, `Support`, `Community`, `Release notes`
* Assign appropriate labels, including priority and team involved.
* Be selective when assigning issues using /assign @<username> or /cc @<username>. Your issue will be triaged more effectively applying correct labels over assigning more people to the issue.
* Leave issues in `Backlog`, prioritization in further columns are performed as part of sprint preparation.
* Issues should be closed only by the original author of the ticket after making sure it's fully completed and can be closed.

### Respond to an issue

* When tackling an issue, comment on it letting others know you are working on it to avoid duplicate work.
* When you have resolved something on your own at any future time, comment on the issue letting people know before closing it.
* Include references to other PRs or issues (or any accessible materials), example: "ref: #1234". It is useful to identify that related work has been addressed somewhere else.

### Development process

We follow a [trunk-based](https://trunkbaseddevelopment.com/) development process. This means all developers work on a single branch, 'main', and features are developed in short-lived feature branches. Issues are your mandate, ensure you have an up-to-date issue when working on a feature or bug fix by editing the description and status. Once you start working on an issue, you become the owner of it, you are responsible to make it progress through the different steps and pass any blockers. Please escalate any problem you cannot overcome to the Linea team. Proper testing and documentation are always part of the definition of done.

1. **Create a New Branch**: For each new feature or bug fix, create a new branch. Branches should be named descriptively, following the pattern `type/issue#-short-description`, e.g., `feature/123-add-login-button` or `bugfix/456-fix-login-error`.

1. **Write Your Code**: Develop your feature or fix the bug in your branch. Make sure your code is clean, well-commented, and adheres to our coding standards.
<!--- TODO: Section on coding standards -->

1. **Test Your Changes**: Write appropriate tests for your changes and ensure all tests pass before submitting a pull request. You are responsible for implementing robust but short-lived automated tests, to avoid feature breaking because of a lack of automated test, and to avoid CI growing to an infinite time.
<!--- TODO: Section on unit & integration testing -->

1. **Commit Your Changes**: Commit your changes regularly and write clear, concise commit messages that explain your changes. Follow the pattern `type(issue#): short description of change`, e.g., `fix(456): corrected login error`.

1. **Push Your Changes**: Push your changes to your forked repository on GitHub.

1. **Submit a Pull Request**: Create a new pull request from your branch to the main repository's trunk/main branch. Include a detailed description of your changes and link the related issue number, in the template created.

1. **CI**: Ensure all elements in the CI are passing in a reliable fashion, don't retry to fix a flaky test. Ensure quality gates (tests, coverage, static analysis, and security analysis) are passing.


## Code Reviews

Once you submit a pull request, it will be reviewed by the maintainers. They may ask for changes or improvements. Please be patient and responsive to their feedback.

## Release Process

We use GitOps, ArgoCD, Helm, and Kubernetes for our release process. When changes are merged into the main branch, they can be deployed to a testnet environment using ArgoCD and Helm. After successful testing in the testnet environment, the changes can be promoted to the mainnet environment.

Software Engineers are responsible for pushing their changes to testnet and mainnet, and monitoring the impact of their changes.

The Release Manager is responsible for the technical soundness of each release. This includes:
* Reviewing each PR. Will it break something? Were necessary integration and regression tests done? What are dependencies? Are we following the correct order/timeline for testnet/mainnet?
* Ensuring communication has been shared with the relevant channels (Ops, SRE, support, community & documentation). Both with internal and external stakeholders (partners, community). This doesn't have to be done directly by the release manager, however the release manager is responsible for making sure it was done.
* Ensuring people are available to react in case the release has an issue (engineer, SRE).
* Controlling that proper monitoring has been done by engineers when releasing.
* Make sure the release notes are accurate and up to date. [Release Notes](https://github.com/Consensys/zkevm-monorepo/wiki/Linea-Releases).
* Be the first contact point for operations/SRE if something happens on the cluster.
<!-- * Make summaries of what's being released and their statuses on slack #linea-testnet and #linea-mainnet -->
* Do a proper handover to the next release manager.

### Release types
* Hotfix:
  * Content: **critical** fixes that should be deployed ASAP. Fixes that are not critical will be part of the normal release calendar
  * Frequency: ad hoc, anytime
* Continuous release:
  * Content: Intermediary deliverables of major releases and non critical fixes.
  * Frequency: Every week. Developers are encouraged to release as frequently as possible their features and break them down using minor releases.

Configuration changes that are part of the [Ops Runbook](https://www.notion.so/consensys/Linea-Runbooks-55e170e0db8f4d71add01c2ef1611cb6) are not considered releases. This includes for example changing the # of provers or updating price thresholds. These configurations changes can be performed at any time.

### Standard deployment timeline
* Internal Consensys engineers share their PRs with the release manager before each Monday mid-day, ideally before Friday of the previous week. Details are added to the [Release Notes](https://github.com/Consensys/zkevm-monorepo/wiki/Linea-Releases).
* The release manager greenlights the release, proposes a day and time for deploying the release to the engineer. The release manager ensure communication has been shared with internal stakeholders (SRE) as well as external stakeholders when relevant (community/partners).
* The deployment on testnet is performed between Monday and Wednesday before 15:00 local Software Engineer time. With the help of DevOps and SRE teams.
* The deployment on mainnet is performed following the exact same process as testnet, after sufficient time to be confident the release can be promoted to mainnet.

### Release to Testnet
We release on a weekly basis. The standard day-by-day process is as follows:
1. **Create a Testnet PR**: Create a new pull request to deploy the changes, once your PR has been merged. Only release changes that were merged in Consensys/zkevm-monorepo/main. Include a detailed description of your changes and link the related issue number, in the template created.
1. **Prepare deployment with Release Manager**: Reach out to the Release Manager with your PR to prepare its inclusion in the next batch of releases, ideally before Friday (cut-off day is Monday mid-day). Use the next Testnet release, and add a tentative deployment to Mainnet at least 1 week after. Release notes can be found [here](https://github.com/Consensys/zkevm-monorepo/wiki/Linea-Releases).
1. **Open a maintenance window for releases having an impact on user experience**: Reach out to NOC to share the expected maintenance window so it can be added to the [status page](https://linea.statuspage.io/). Maintenance windows can be requested by adding a ticket to [this board](https://consensyssoftware.atlassian.net/jira/core/projects/NOC/board?groupBy=status). You can make sure the request was seen by tagggign @noc on the testnet/mainnet slack channels.
1. **Merge your PR**: Release your changes and follow instructions on the Testnet PR template. Deployments are performed between Monday and Wednesday before 15:00 local Software Engineer time. DevOps and SRE teams should be involved in the deployment activities. Major releases should include a detailed release plan that also includes a study of risks, mitigations and revert plans.

### Release to Mainnet

The deployment on mainnet is performed following the exact same process as testnet after sufficient time and testing to be confident the release can be promoted to mainnet.

## Conclusion

Your contributions are greatly appreciated. By following these guidelines, we can ensure a smooth and effective collaboration. Thank you for your contribution!
