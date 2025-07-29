
# Maru test network topology

- 1 maru validator: `maru-validator` -> `besu-sequencer`
- 1 maru follower: `maru-follower-0-0` -> `besu-follower-0`
- 1 maru follower: `maru-follower-1-0` -> `besu-follower-1`+ `besu-follower-2`
   - `besu-follower-1` is the primary EL client
   - `besu-follower-2` is just EL client replica

# Quick start

### Full provisioning

```bash
make chaos-full-reload
```

- deploys K3S kubernetes cluster
- deploys Chaos-Mesh into `chaos-mesh` namespace
- deploys Maru and Besu network into `default` namespace

### Other helpful commands

- `make helm-redeploy-maru-and-besu` - redeploys Maru + Besu network from genesis
- - `make chaos-experiment-podkill-besu-nodes` - runs Besu downtime experiment of 60s
- `make chaos-redeploy-and-run-experiment` - combines above targets

