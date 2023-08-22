# End to end tests
## Setup
Run `npm install` to setup typechain

Run `make fresh-start-all` from root directory to spin up local environment


## Run tests
| ENV | Command | Description |
|----|---|---|
| Local | `npm run test:e2e:local` | Uses already running docker environment and deployed smart contracts |
| DEV | `npm run test:e2e:dev` | Uses DEV env, may need to update constants in `constants.dev.ts`  |
| UAT | `npm run test:e2e:uat` | Uses UAT env, may need to update constants in `constants.uat.ts` |

## Remote workflows
Workflow options:
- `e2e-tests-with-ssh` - Enable to run `Setup upterm session` step, manually ssh into the github actions workflow using
the steps output, can be used to debug containers.
  - The step will output a string used to connect to the workflow.
  - Example: `ssh XTpun7OCRZMgaCZkiHqU:MWNlNmQ0OGEudm0udXB0ZXJtLmludGVybmFsOjIyMjI=@uptermd.upterm.dev`
  - After connecting create a new file called `continue` in the root directory: `touch continue`
- `e2e-tests-logs-dump` - Enable to print logs after e2e tests have ran
