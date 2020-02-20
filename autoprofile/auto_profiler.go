package autoprofile

import (
	"path/filepath"
)

const (
	defaultMaxBufferedProfiles = 100
)

var (
	sensorPath = filepath.Join("github.com", "instana", "go-sensor")
	profiler   = newAutoProfiler()
)

// Enable enables the auto profiling (disabled by default)
func Enable() {
	profiler.Enable()
}

// Disable disables the auto profiling (default)
func Disable() {
	profiler.Disable()
}

// SetGetExternalPIDFunc configures the profiler to use provided function to retrieve the current PID
func SetGetExternalPIDFunc(fn func() string) {
	profiler.GetExternalPID = fn
}

// SetSendProfilesFunc configures the profiler to use provided function to write collected profiles
func SetSendProfilesFunc(fn SendProfilesFunc) {
	if fn == nil {
		fn = noopSendProfiles
	}

	profiler.profileRecorder.SendProfiles = fn
}

// Options contains profiler configuration
type Options struct {
	IncludeSensorFrames bool
	MaxBufferedProfiles int
}

// DefaultOptions returns profiler defaults
func DefaultOptions() Options {
	return Options{
		MaxBufferedProfiles: defaultMaxBufferedProfiles,
	}
}

// SetOptions configures the profiler with provided settings
func SetOptions(opts Options) {
	if opts.MaxBufferedProfiles < 1 {
		opts.MaxBufferedProfiles = defaultMaxBufferedProfiles
	}

	profiler.profileRecorder.MaxBufferedProfiles = opts.MaxBufferedProfiles
	profiler.IncludeSensorFrames = opts.IncludeSensorFrames
}

type autoProfiler struct {
	profileRecorder            *recorder
	cpuSamplerScheduler        *SamplerScheduler
	allocationSamplerScheduler *SamplerScheduler
	blockSamplerScheduler      *SamplerScheduler

	enabled       bool
	samplerActive *flag

	// Options
	IncludeSensorFrames bool

	GetExternalPID func() string
}

func newAutoProfiler() *autoProfiler {
	ap := &autoProfiler{
		samplerActive: &flag{},
	}

	ap.profileRecorder = newRecorder()

	cpuSampler := newCPUSampler(ap)
	cpuSamplerConfig := &SamplerConfig{
		logPrefix:          "CPU sampler:",
		maxProfileDuration: 20,
		maxSpanDuration:    2,
		maxSpanCount:       30,
		samplingInterval:   8,
		reportInterval:     120,
	}
	ap.cpuSamplerScheduler = newSamplerScheduler(ap, cpuSampler, cpuSamplerConfig)

	allocationSampler := newAllocationSampler(ap)
	allocationSamplerConfig := &SamplerConfig{
		logPrefix:      "Allocation sampler:",
		reportOnly:     true,
		reportInterval: 120,
	}
	ap.allocationSamplerScheduler = newSamplerScheduler(ap, allocationSampler, allocationSamplerConfig)

	blockSampler := newBlockSampler(ap)
	blockSamplerConfig := &SamplerConfig{
		logPrefix:          "Block sampler:",
		maxProfileDuration: 20,
		maxSpanDuration:    4,
		maxSpanCount:       30,
		samplingInterval:   16,
		reportInterval:     120,
	}
	ap.blockSamplerScheduler = newSamplerScheduler(ap, blockSampler, blockSamplerConfig)

	return ap
}

func (ap *autoProfiler) Enable() {
	if ap.enabled {
		return
	}

	ap.profileRecorder.start()
	ap.cpuSamplerScheduler.start()
	ap.allocationSamplerScheduler.start()
	ap.blockSamplerScheduler.start()

	log.debug("profiler enabled")
}

func (ap *autoProfiler) Disable() {
	if !ap.enabled {
		return
	}

	ap.profileRecorder.stop()
	ap.cpuSamplerScheduler.stop()
	ap.allocationSamplerScheduler.stop()
	ap.blockSamplerScheduler.stop()

	log.debug("profiler disabled")
}
