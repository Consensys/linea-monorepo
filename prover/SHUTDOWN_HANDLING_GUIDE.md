# Controller Shutdown Handling Guide

## Overview

The controller supports intelligent shutdown handling that adapts to different deployment scenarios:
- **Spot Instances**: Immediate job requeuing (no wasted work on 2-minute warnings)
- **Normal K8s**: Graceful shutdown with full grace period (3550s)
- **Local Development**: Configurable via file marker

## Signal Handling

### Supported Signals

| Signal | Purpose | Target Process | Behavior |
|--------|---------|----------------|----------|
| `SIGUSR1` | Spot reclamation notification | **Controller** (PID 1) | Immediate shutdown + requeue |
| `SIGTERM` | Standard shutdown request | **Controller** (PID 1) | Checks `/tmp/termination-type` file |

**Critical**: Both signals must be sent to the **controller process**, NOT the prover child process!

**Why?**
- Controller manages job queue and requeuing logic
- Prover is a temporary child process with no queue knowledge
- Controller orchestrates graceful shutdown

**Note**: Both signals listen on the same context (`signal.NotifyContext`), so **first signal wins**. Subsequent signals are ignored.

### File-Based Decision: Termination Type File

The controller reads a file (default: `/tmp/termination-type`) on SIGTERM to determine shutdown strategy:

**File path is configurable** via `termination_type_file` in the config:
```toml
[controller]
# Optional: custom path (defaults to /tmp/termination-type if not set)
termination_type_file = "/var/run/controller/shutdown-mode"
```

**File contents determine behavior:**

| File State | Content | Behavior | Use Case |
|------------|---------|----------|----------|
| ❌ **Missing** | N/A | **⚠️ SPOT RECLAIM** (immediate requeue) | Spot instances (safe default) |
| ✅ Exists | `"NORMAL_SHUTDOWN"` | Graceful shutdown (3550s grace period) | Normal K8s, local dev |
| ✅ Exists | `"SPOT_RECLAIM"` | Immediate requeue | Explicit spot termination |
| ✅ Exists | Any other value | **⚠️ SPOT RECLAIM** (safe default) | Fallback to safe mode |

**Critical Design Decision**: File missing = Spot reclaim mode
- **Rationale**: Safer to requeue immediately than waste 59 minutes on a doomed spot instance
- **Impact**: Normal K8s deployments **MUST** create the file with `"NORMAL_SHUTDOWN"`

## Process Targeting

### ⚠️ Always Target the Controller Process

```
┌──────────────────────────────────────────────┐
│ Controller Process (PID 1 in container)      │  ← Send signals HERE
│ • Listens for SIGUSR1/SIGTERM                │
│ • Manages job queue                          │
│ • Handles requeuing                          │
│                                              │
│   ┌────────────────────────────────────┐    │
│   │ Prover Child Process (ephemeral)   │    │  ← DON'T send signals here
│   │ • No signal handlers                │    │
│   │ • No queue knowledge                │    │
│   │ • Spawned per job                   │    │
│   └────────────────────────────────────┘    │
│                                              │
└──────────────────────────────────────────────┘
```

### How to Target Controller

**In Kubernetes Container:**
```bash
# Controller is always PID 1 in its container
kill -USR1 1
kill -TERM 1
```

**From Host or Sidecar:**
```bash
# Find by process name
CONTROLLER_PID=$(pgrep -f "bin/controller")

# Send signal
kill -USR1 $CONTROLLER_PID
```

**Common Mistakes:**
```bash
# ❌ WRONG: Sending to prover child
kill -USR1 $(pgrep -f "prover")  # This finds the wrong process!

# ✅ CORRECT: Sending to controller
kill -USR1 $(pgrep -f "controller")
```

## Use Cases & Handling

### Use Case 1: Spot Instance with PreStop Hook (Recommended)

**Scenario**: Kubernetes pod on spot instance with termination detection

**Deployment Pattern**:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: prover-controller
spec:
  containers:
  - name: controller
    image: prover:latest
    command: ["./bin/controller"]
    volumeMounts:
    - name: tmp
      mountPath: /tmp
    
    lifecycle:
      preStop:
        exec:
          command: ["/bin/bash", "-c", "
            # 1. Write marker file FIRST
            echo 'SPOT_RECLAIM' > /tmp/termination-type
            
            # 2. Signal the CONTROLLER (PID 1 in container)
            # DO NOT send to prover child process!
            kill -USR1 1
            
            # 3. Give it time to process
            sleep 1
          "]
  
  volumes:
  - name: tmp
    emptyDir: {}
```

**Timeline**:
```
T=0s    → Cloud provider spot interruption detected
T=0s    → K8s triggers preStop hook
T=0s    → preStop writes "SPOT_RECLAIM" to /tmp/termination-type
T=0s    → preStop sends SIGUSR1 to controller (PID 1)
T=0s    → Controller shutdown handler wakes up (<-ctx.Done())
T=0s    → Controller reads /tmp/termination-type → "SPOT_RECLAIM"
T=0s    → Controller immediately sends SIGKILL to running job
T=0s    → Job requeued to request folder
T=0.5s  → preStop exits
T=1s    → Controller exits cleanly
T=1s    → K8s sends SIGTERM (redundant, ctx already Done)
T=1s    → K8s pod terminated successfully
T=120s  → Cloud provider SIGKILL (pod already gone ✅)
```

**Logs**:
```
level=warning msg="Spot instance reclamation detected (SIGUSR1 or /tmp/termination-type=SPOT_RECLAIM)"
level=warning msg="Initiating immediate shutdown to requeue job before VM termination"
level=warning msg="Spot reclaim: Immediately killing process execution-X to requeue job"
level=info msg="Job X was killed by us. Reputting it in the request folder"
level=info msg="Graceful shutdown complete"
```

**Race Condition Handling**:

If SIGTERM arrives before SIGUSR1:
```
T=0s    → K8s sends SIGTERM (happens first)
T=0s    → Controller reads /tmp/termination-type → MAY NOT EXIST YET
T=0s    → Controller defaults to spot reclaim mode (safe!)
T=0.1s  → preStop writes file (too late, but OK)
T=0.1s  → preStop sends SIGUSR1 (ignored, ctx already Done)
```
**Result**: Still requeues immediately ✅

---

### Use Case 2: Spot Instance WITHOUT PreStop Hook (Fallback)

**Scenario**: Spot instance without detection script

**Deployment Pattern**:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: prover-controller
spec:
  containers:
  - name: controller
    image: prover:latest
    command: ["./bin/controller"]
    # NO preStop hook
    # NO /tmp/termination-type file created
```

**Timeline**:
```
T=0s    → Cloud provider spot interruption (2-minute warning)
T=0s    → No preStop hook, no file written
T=120s  → Cloud provider sends SIGTERM to pod
T=120s  → K8s propagates SIGTERM to controller
T=120s  → Controller reads /tmp/termination-type → FILE NOT FOUND
T=120s  → Controller defaults to spot reclaim mode (safe fallback!)
T=120s  → Immediately sends SIGKILL to running job
T=120s  → Job requeued to request folder
T=120s  → Controller exits cleanly
T=122s  → Cloud provider SIGKILL (controller already gone ✅)
```

**Logs**:
```
level=warning msg="File /tmp/termination-type not found, defaulting to spot instance reclaim mode"
level=warning msg="Spot instance reclamation detected"
level=warning msg="Spot reclaim: Immediately killing process execution-X to requeue job"
```

**Why This Works**: 
- File missing = Safe assumption of spot reclaim
- Job requeues immediately
- No 59 minutes wasted trying to finish

---

### Use Case 3: Normal Kubernetes Deployment (Rolling Update)

**Scenario**: Regular K8s deployment with rolling updates

**Deployment Pattern**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prover-controller
spec:
  template:
    spec:
      # CRITICAL: Init container to mark as normal shutdown
      initContainers:
      - name: init-shutdown-mode
        image: prover:latest
        command: ["/bin/bash", "-c", "
          echo 'NORMAL_SHUTDOWN' > /tmp/termination-type && 
          chmod 644 /tmp/termination-type
        "]
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      
      containers:
      - name: controller
        image: prover:latest
        command: ["./bin/controller"]
        volumeMounts:
        - name: tmp
          mountPath: /tmp
      
      volumes:
      - name: tmp
        emptyDir: {}
      
      terminationGracePeriodSeconds: 3600
```

**Timeline**:
```
T=-5s   → Init container runs
T=-5s   → Init writes "NORMAL_SHUTDOWN" to /tmp/termination-type
T=0s    → Controller starts
T=0s    → Controller picks up job (takes ~1 hour)
...
T=1800s → K8s rolling update triggered
T=1800s → K8s sends SIGTERM to pod
T=1800s → Controller reads /tmp/termination-type → "NORMAL_SHUTDOWN"
T=1800s → Controller calculates deadline: now + 3550s
T=1800s → Controller monitors job, allows it to continue
T=3000s → Job completes successfully
T=3000s → Controller moves job to done folder
T=3000s → Controller exits cleanly
T=3000s → K8s pod terminated (used 1200s of 3600s grace period)
```

**Logs**:
```
level=info msg="File /tmp/termination-type contains NORMAL_SHUTDOWN, using graceful shutdown"
level=info msg="Received SIGTERM. Grace period: 3550s (until 2025-10-13T23:26:00Z)"
level=info msg="Job execution-X is currently processing. Will wait for completion or grace period expiry."
level=info msg="Job completed before grace period, cancelling context for cleanup"
level=info msg="Shutting down metrics server"
level=info msg="Graceful shutdown complete"
```

**If Job Takes Too Long**:
```
T=1800s → SIGTERM received, start monitoring
T=5340s → Approaching deadline (1800 + 3540s)
T=5340s → Send SIGTERM to child process
T=5350s → Child doesn't respond, send SIGKILL
T=5350s → Job requeued to request folder
T=5350s → Controller exits
T=5400s → K8s SIGKILL (not needed, controller already gone)
```

---

### Use Case 4: Local Development

**Scenario**: Developer running controller locally

**Setup Option A - Normal Shutdown Mode** (for testing graceful shutdown):
```bash
# Create marker file for normal shutdown
echo "NORMAL_SHUTDOWN" > /tmp/termination-type

# Start controller
./bin/controller --config config.toml

# Test graceful shutdown
kill -TERM $(pgrep controller)
# Expected: 3550s grace period, job allowed to complete
```

**Setup Option B - Spot Reclaim Mode** (for testing immediate requeue):
```bash
# Don't create the file (or delete it)
rm -f /tmp/termination-type

# Start controller
./bin/controller --config config.toml

# Test spot reclaim
kill -TERM $(pgrep controller)
# Expected: Immediate job requeue
```

**Setup Option C - Test SIGUSR1** (for testing spot signal):
```bash
# Start controller (file state doesn't matter)
./bin/controller --config config.toml

# Send spot reclaim signal
kill -USR1 $(pgrep controller)
# Expected: Immediate job requeue, same as spot mode
```

---

## Decision Matrix

### When Controller Receives Signal

```
┌─────────────────────────────────────────────────────────────────────┐
│ Signal Received: SIGTERM or SIGUSR1                                 │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ Read file:       │
                    │ /tmp/termination │
                    │ -type            │
                    └──────────────────┘
                              │
                ┌─────────────┼─────────────┐
                │             │             │
                ▼             ▼             ▼
          ┌─────────┐   ┌──────────┐  ┌──────────┐
          │ Missing │   │ "NORMAL  │  │ Other    │
          │         │   │ SHUTDOWN"│  │ value    │
          └─────────┘   └──────────┘  └──────────┘
                │             │             │
                ▼             ▼             ▼
       ┌────────────┐  ┌────────────┐ ┌────────────┐
       │ SPOT       │  │ GRACEFUL   │ │ SPOT       │
       │ RECLAIM    │  │ SHUTDOWN   │ │ RECLAIM    │
       │ MODE       │  │ MODE       │ │ MODE       │
       └────────────┘  └────────────┘ └────────────┘
                │             │             │
                ▼             ▼             ▼
       ┌────────────┐  ┌────────────┐ ┌────────────┐
       │ Immediate  │  │ Wait up to │ │ Immediate  │
       │ SIGKILL    │  │ 3550s for  │ │ SIGKILL    │
       │ to job     │  │ completion │ │ to job     │
       └────────────┘  └────────────┘ └────────────┘
                │             │             │
                ▼             ▼             ▼
       ┌────────────┐  ┌────────────┐ ┌────────────┐
       │ Requeue    │  │ Job done   │ │ Requeue    │
       │ job        │  │ or requeue │ │ job        │
       └────────────┘  └────────────┘ └────────────┘
                │             │             │
                └─────────────┼─────────────┘
                              ▼
                    ┌──────────────────┐
                    │ Exit cleanly     │
                    └──────────────────┘
```

---

## Configuration

### Controller Config (Same for All Use Cases)

```toml
[controller]
retry_delays = [0, 1]

# Graceful shutdown settings (only used when file = "NORMAL_SHUTDOWN")
termination_grace_period = "3550s"           # 59 minutes
child_process_shutdown_timeout = "10s"       # SIGTERM → SIGKILL timeout

# Optional: custom path for termination type file
# Defaults to /tmp/termination-type if not specified
# termination_type_file = "/tmp/termination-type"

# Spot instance mode flag (DEPRECATED, not checked by code)
# Behavior now controlled by termination type file
spot_instance_mode = false
```

**Note**: No config changes needed between spot and normal deployments!

**Custom File Path Example**:
```toml
[controller]
# Use a custom path (useful for read-only /tmp or security requirements)
termination_type_file = "/var/run/controller/shutdown-mode"
```

When using a custom path, mount your volume accordingly:
```yaml
volumeMounts:
- name: controller-state
  mountPath: /var/run/controller
```

---

## Comparison Table

| Aspect | Spot Instance | Normal K8s | Local Dev (Normal) | Local Dev (Spot) |
|--------|---------------|------------|-------------------|------------------|
| **File Required?** | ❌ No (defaults to spot) | ✅ **YES** (init container) | ✅ Manual | ❌ Delete file |
| **File Content** | N/A or "SPOT_RECLAIM" | "NORMAL_SHUTDOWN" | "NORMAL_SHUTDOWN" | Missing |
| **PreStop Hook?** | ✅ Recommended | ❌ No | N/A | N/A |
| **Grace Period** | ~10s (cleanup only) | 3550s | 3550s | ~10s |
| **Job Behavior** | Requeue immediately | Complete or requeue | Complete or requeue | Requeue immediately |
| **Best For** | Cost optimization | Completing jobs | Testing graceful shutdown | Testing spot behavior |

---

## Testing Scenarios

### Test 1: Spot Reclaim with File
```bash
echo "SPOT_RECLAIM" > /tmp/termination-type
./bin/controller &
sleep 2
kill -TERM $(pgrep controller)

# Expected:
# - Log: "File ... contains 'SPOT_RECLAIM', treating as spot instance reclaim"
# - Immediate job requeue
# - Fast exit (<1s)
```

### Test 2: Spot Reclaim without File (Default)
```bash
rm -f /tmp/termination-type
./bin/controller &
sleep 2
kill -TERM $(pgrep controller)

# Expected:
# - Log: "File /tmp/termination-type not found, defaulting to spot instance reclaim mode"
# - Immediate job requeue
# - Fast exit (<1s)
```

### Test 3: Normal Shutdown
```bash
echo "NORMAL_SHUTDOWN" > /tmp/termination-type
./bin/controller &
sleep 2
kill -TERM $(pgrep controller)

# Expected:
# - Log: "File ... contains NORMAL_SHUTDOWN, using graceful shutdown"
# - Log: "Grace period: 3550s"
# - Waits for job completion
```

### Test 4: SIGUSR1 Signal
```bash
# File state doesn't matter for SIGUSR1
./bin/controller &
sleep 2
kill -USR1 $(pgrep controller)

# Expected:
# - Immediate requeue regardless of file state
# - Fast exit (<1s)
```

### Test 5: Race Condition (SIGTERM before preStop)
```bash
# Simulate race: SIGTERM arrives before file written
./bin/controller &
CONTROLLER_PID=$!
sleep 2

# Send SIGTERM first (simulating K8s)
kill -TERM $CONTROLLER_PID &

# Write file slightly after (simulating slow preStop)
sleep 0.1
echo "SPOT_RECLAIM" > /tmp/termination-type

wait

# Expected:
# - File missing when SIGTERM processed
# - Defaults to spot reclaim (safe!)
# - Job requeued
```

---

## Troubleshooting

### Issue: Normal K8s pods requeuing jobs immediately

**Symptom**:
```
level=warning msg="File /tmp/termination-type not found, defaulting to spot instance reclaim mode"
```

**Cause**: Missing init container to create `/tmp/termination-type`

**Fix**: Add init container to deployment:
```yaml
initContainers:
- name: init-shutdown-mode
  image: prover:latest
  command: ["/bin/bash", "-c", "echo NORMAL_SHUTDOWN > /tmp/termination-type"]
  volumeMounts:
  - name: tmp
    mountPath: /tmp
```

### Issue: Spot instances not requeuing fast enough

**Symptom**: Jobs killed by cloud provider before requeue

**Cause**: No preStop hook for early detection

**Fix**: Add preStop lifecycle hook:
```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/bash", "-c", "
        echo SPOT_RECLAIM > /tmp/termination-type && 
        kill -USR1 1 && 
        sleep 1
      "]
```

### Issue: PreStop hook not working

**Symptom**: SIGUSR1 sent but controller doesn't respond

**Possible Causes**:

**1. Wrong Process Targeted** ⚠️ Most Common
```bash
# ❌ WRONG: Sending to prover child
kill -USR1 $(pgrep -f "prover")

# ✅ CORRECT: Sending to controller
kill -USR1 1                         # In container
kill -USR1 $(pgrep -f "controller")  # From host
```

**2. Wrong PID**: Use `kill -USR1 1` (PID 1 in container)

**3. No shared /tmp**: Ensure volume mounted to both containers

**4. Permission issues**: Ensure file writable

**Debug**:
```bash
# Check what processes are running
kubectl exec <pod> -- ps aux

# You should see:
# PID 1: ./bin/controller          ← Send signals HERE
# PID X: /bin/sh <prover-script>   ← NOT here

# Check if file exists
kubectl exec <pod> -- cat /tmp/termination-type

# Check controller PID (should be 1)
kubectl exec <pod> -- ps aux | grep controller

# Test signal to controller (PID 1)
kubectl exec <pod> -- kill -USR1 1
```

**Verify Signal is Reaching Controller**:
```bash
# Send signal and check logs immediately
kubectl exec <pod> -- kill -USR1 1
kubectl logs <pod> --tail=20

# Expected log:
# level=warning msg="Spot instance reclamation detected"
```

---

## Best Practices

### 1. Always Use Init Container for Normal K8s
```yaml
# DO THIS for production K8s deployments
initContainers:
- name: init-shutdown-mode
  command: ["sh", "-c", "echo NORMAL_SHUTDOWN > /tmp/termination-type"]
```

### 2. Use PreStop for Spot Instances
```yaml
# DO THIS for spot/preemptible instances
lifecycle:
  preStop:
    exec:
      command: ["sh", "-c", "echo SPOT_RECLAIM > /tmp/termination-type && kill -USR1 1"]
```

### 3. Write File BEFORE Signaling
```bash
# CORRECT order in preStop:
echo SPOT_RECLAIM > /tmp/termination-type  # 1. File first
kill -USR1 1                                # 2. Signal second
sleep 1                                     # 3. Wait for processing
```

### 4. Share /tmp via Volume
```yaml
# Ensure /tmp is shared between init and main container
volumes:
- name: tmp
  emptyDir: {}
```

### 5. Set Appropriate Grace Period
```yaml
# Match terminationGracePeriodSeconds to config
terminationGracePeriodSeconds: 3600  # K8s side
# termination_grace_period = "3550s"  # Config side (slightly less)
```

---

## Migration Checklist

### For Existing Normal K8s Deployments

- [ ] Review all deployment manifests
- [ ] Add init container to create `/tmp/termination-type`
- [ ] Add `tmp` emptyDir volume
- [ ] Mount volume to init container and main container
- [ ] Test in staging environment
- [ ] Verify file exists: `kubectl exec <pod> -- cat /tmp/termination-type`
- [ ] Trigger rolling update and monitor logs
- [ ] Confirm graceful shutdown behavior

### For New Spot Instance Deployments

- [ ] Add preStop lifecycle hook
- [ ] Add `tmp` emptyDir volume (for file sharing)
- [ ] Deploy and test spot interruption simulation
- [ ] Monitor requeue behavior
- [ ] Verify jobs complete on other instances
- [ ] (Optional) Deploy metadata monitoring sidecar

---

## Summary

**Key Points**:
1. ✅ **File missing = Spot reclaim mode** (safe default for cost-sensitive environments)
2. ✅ **Single signal triggers shutdown** (first signal wins, subsequent ignored)
3. ✅ **Race conditions handled** by safe defaults
4. ✅ **Same config works everywhere** (behavior controlled by file, not config)
5. ⚠️ **Normal K8s REQUIRES init container** (must explicitly opt-in to graceful shutdown)

**When in doubt**: Controller defaults to immediate requeue, which is the safest behavior for spot instances.


