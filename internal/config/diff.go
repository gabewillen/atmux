package config

import "reflect"

// DiffConfig returns a list of changes between two configs for reloadable keys.
func DiffConfig(oldCfg Config, newCfg Config) []ConfigChange {
	var changes []ConfigChange
	compare(&changes, "timeouts.idle", oldCfg.Timeouts.Idle, newCfg.Timeouts.Idle)
	compare(&changes, "timeouts.stuck", oldCfg.Timeouts.Stuck, newCfg.Timeouts.Stuck)
	compare(&changes, "telemetry.enabled", oldCfg.Telemetry.Enabled, newCfg.Telemetry.Enabled)
	compare(&changes, "telemetry.service_name", oldCfg.Telemetry.ServiceName, newCfg.Telemetry.ServiceName)
	compare(&changes, "telemetry.exporter.endpoint", oldCfg.Telemetry.Exporter.Endpoint, newCfg.Telemetry.Exporter.Endpoint)
	compare(&changes, "telemetry.exporter.protocol", oldCfg.Telemetry.Exporter.Protocol, newCfg.Telemetry.Exporter.Protocol)
	compare(&changes, "telemetry.traces.enabled", oldCfg.Telemetry.Traces.Enabled, newCfg.Telemetry.Traces.Enabled)
	compare(&changes, "telemetry.traces.sampler", oldCfg.Telemetry.Traces.Sampler, newCfg.Telemetry.Traces.Sampler)
	compare(&changes, "telemetry.traces.sampler_arg", oldCfg.Telemetry.Traces.SamplerArg, newCfg.Telemetry.Traces.SamplerArg)
	compare(&changes, "telemetry.metrics.enabled", oldCfg.Telemetry.Metrics.Enabled, newCfg.Telemetry.Metrics.Enabled)
	compare(&changes, "telemetry.metrics.interval", oldCfg.Telemetry.Metrics.Interval, newCfg.Telemetry.Metrics.Interval)
	compare(&changes, "telemetry.logs.enabled", oldCfg.Telemetry.Logs.Enabled, newCfg.Telemetry.Logs.Enabled)
	compare(&changes, "telemetry.logs.level", oldCfg.Telemetry.Logs.Level, newCfg.Telemetry.Logs.Level)
	changes = append(changes, diffAdapterPatterns(oldCfg.Adapters, newCfg.Adapters)...)
	return changes
}

func compare(changes *[]ConfigChange, path string, oldVal any, newVal any) {
	if reflect.DeepEqual(oldVal, newVal) {
		return
	}
	*changes = append(*changes, ConfigChange{Path: path, OldValue: oldVal, NewValue: newVal})
}

func diffAdapterPatterns(oldAdapters map[string]AdapterConfig, newAdapters map[string]AdapterConfig) []ConfigChange {
	var changes []ConfigChange
	seen := make(map[string]struct{})
	for name := range oldAdapters {
		seen[name] = struct{}{}
	}
	for name := range newAdapters {
		seen[name] = struct{}{}
	}
	for name := range seen {
		oldPatterns := adapterPatterns(oldAdapters[name])
		newPatterns := adapterPatterns(newAdapters[name])
		keys := make(map[string]struct{})
		for key := range oldPatterns {
			keys[key] = struct{}{}
		}
		for key := range newPatterns {
			keys[key] = struct{}{}
		}
		for key := range keys {
			path := "adapters." + name + ".patterns." + key
			compare(&changes, path, oldPatterns[key], newPatterns[key])
		}
	}
	return changes
}

func adapterPatterns(cfg AdapterConfig) map[string]any {
	if cfg == nil {
		return map[string]any{}
	}
	patternsRaw, ok := cfg["patterns"].(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return patternsRaw
}
