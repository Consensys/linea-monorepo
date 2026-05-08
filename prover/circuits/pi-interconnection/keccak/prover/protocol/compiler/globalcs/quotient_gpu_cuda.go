//go:build cuda

package globalcs

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/gpu"
	gpubls12377 "github.com/consensys/linea-monorepo/prover/gpu/plonk2/bls12377"
	"github.com/sirupsen/logrus"
)

const (
	envPIQuotientGPUReeval                 = "LINEA_PROVER_GPU_PI_QUOTIENT_REEVAL"
	envPIQuotientGPUSecondaryDeviceID      = "LINEA_PROVER_GPU_PI_QUOTIENT_SECONDARY_DEVICE_ID"
	envPIQuotientGPUDisableSecondaryDevice = "LINEA_PROVER_GPU_PI_QUOTIENT_DISABLE_SECONDARY_DEVICE"

	gpuQuotientReevalMinDomain = 1 << 18
)

func tryGPUQuotientReevalCoset(
	domainSize int,
	shift field.Element,
	inputs [][]field.Element,
	outputs [][]field.Element,
) bool {
	if os.Getenv(envPIQuotientGPUReeval) != "1" {
		return false
	}
	if domainSize < gpuQuotientReevalMinDomain {
		return false
	}
	if len(inputs) == 0 {
		return true
	}
	if len(inputs) != len(outputs) {
		logrus.Warn("PI quotient GPU reeval input/output length mismatch; falling back to CPU")
		return false
	}
	devices, err := quotientReevalDevices()
	if err != nil {
		logrus.WithError(err).Warn("PI quotient GPU reeval device selection failed; falling back to CPU")
		return false
	}
	if len(devices) == 0 {
		return false
	}

	for i := range inputs {
		if len(inputs[i]) != domainSize || len(outputs[i]) != domainSize {
			logrus.Warnf("PI quotient GPU reeval vector size mismatch at %d; falling back to CPU", i)
			return false
		}
	}

	chunk := (len(inputs) + len(devices) - 1) / len(devices)
	var wg sync.WaitGroup
	errCh := make(chan error, len(devices))
	for deviceIndex, dev := range devices {
		start := deviceIndex * chunk
		stop := min(start+chunk, len(inputs))
		if start >= stop {
			continue
		}
		wg.Add(1)
		go func(dev *gpu.Device, start, stop int) {
			defer wg.Done()
			if err := runGPUQuotientReevalCoset(dev, domainSize, shift, inputs[start:stop], outputs[start:stop]); err != nil {
				errCh <- err
			}
		}(dev, start, stop)
	}
	wg.Wait()
	close(errCh)
	if err := <-errCh; err != nil {
		logrus.WithError(err).Warn("PI quotient GPU reeval failed; falling back to CPU")
		return false
	}

	logrus.Infof(
		"PI quotient GPU reeval completed roots=%d domain=%d devices=%d",
		len(inputs), domainSize, len(devices),
	)
	return true
}

func runGPUQuotientReevalCoset(
	dev *gpu.Device,
	domainSize int,
	shift field.Element,
	inputs [][]field.Element,
	outputs [][]field.Element,
) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := dev.Bind(); err != nil {
		return fmt.Errorf("bind GPU device %d: %w", dev.DeviceID(), err)
	}
	domain, err := gpubls12377.NewFFTDomain(dev, domainSize)
	if err != nil {
		return fmt.Errorf("create GPU FFT domain on device %d: %w", dev.DeviceID(), err)
	}
	defer domain.Close()

	vec, err := gpubls12377.NewFrVector(dev, domainSize)
	if err != nil {
		return fmt.Errorf("allocate quotient reeval vector on device %d: %w", dev.DeviceID(), err)
	}
	defer vec.Free()

	for i := range inputs {
		vec.CopyFromHost(field.Vector(inputs[i]))
		domain.BitReverse(vec)
		domain.CosetFFT(vec, shift)
		vec.CopyToHost(field.Vector(outputs[i]))
	}
	if err := dev.Sync(); err != nil {
		return fmt.Errorf("sync GPU device %d: %w", dev.DeviceID(), err)
	}
	return nil
}

func quotientReevalDevices() ([]*gpu.Device, error) {
	primary, primaryID, err := gpu.DeviceFromEnvOrCurrent()
	if err != nil {
		return nil, err
	}
	if primary == nil {
		primary = gpu.GetDevice()
		if primary == nil {
			return nil, nil
		}
		primaryID = primary.DeviceID()
	}

	devices := []*gpu.Device{primary}
	if os.Getenv(envPIQuotientGPUDisableSecondaryDevice) != "" {
		return devices, nil
	}

	secondaryID, ok, err := quotientReevalSecondaryDeviceID(primaryID)
	if err != nil || !ok {
		return devices, err
	}
	secondary := gpu.GetDeviceN(secondaryID)
	if secondary == nil {
		return devices, fmt.Errorf("secondary GPU device %d is unavailable", secondaryID)
	}
	return append(devices, secondary), nil
}

func quotientReevalSecondaryDeviceID(primaryID int) (int, bool, error) {
	raw := os.Getenv(envPIQuotientGPUSecondaryDeviceID)
	if raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil {
			return 0, false, fmt.Errorf("invalid %s %q: %w", envPIQuotientGPUSecondaryDeviceID, raw, err)
		}
		if id < 0 {
			return 0, false, fmt.Errorf("%s must be non-negative, got %d", envPIQuotientGPUSecondaryDeviceID, id)
		}
		if id == primaryID {
			return 0, false, fmt.Errorf("PI quotient secondary device matches primary device %d", primaryID)
		}
		return id, true, nil
	}

	n := gpu.PhysicalDeviceCount()
	if n < 2 {
		return 0, false, nil
	}
	return (primaryID + 1) % n, true, nil
}
