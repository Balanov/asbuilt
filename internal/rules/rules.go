// Package rules is the semantic layer of asbuilt.
//
// Collecting facts is a commodity: lshw, dmidecode, lspci and friends do it.
// Judging them is not. Every rule here turns a machine fact into a verdict
// with evidence, a severity and, where the arithmetic allows it, a measured
// performance loss. That judgement layer is the product.
package rules

import (
	"fmt"
	"sort"

	"github.com/Balanov/asbuilt/internal/collect"
)

// Severity ranks a finding.
type Severity string

const (
	Info     Severity = "info"
	Warn     Severity = "warn"
	Critical Severity = "critical"
)

// Rank returns the numeric weight of a severity.
func (s Severity) Rank() int {
	switch s {
	case Critical:
		return 3
	case Warn:
		return 2
	case Info:
		return 1
	}
	return 0
}

// Impact quantifies what a finding costs.
//
// LossPct is the share of the achievable performance that is missing, not a
// guess: it is computed from the machine's own numbers.
type Impact struct {
	// Kind names the resource: "memory_bandwidth" or "pcie_bandwidth".
	Kind string `json:"kind"`
	// LossPct is the percentage of the achievable bandwidth that is lost.
	LossPct float64 `json:"loss_pct"`
	// Potential is the achievable bandwidth in GB/s per direction.
	Potential float64 `json:"potential_gbs,omitempty"`
	// Actual is the bandwidth the machine actually gets.
	Actual float64 `json:"actual_gbs,omitempty"`
	// Unit is the unit of Potential and Actual.
	Unit string `json:"unit,omitempty"`
}

// Finding is one verdict about the machine.
type Finding struct {
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Title    string   `json:"title"`
	Detail   string   `json:"detail"`
	Evidence []string `json:"evidence,omitempty"`
	Impact   *Impact  `json:"impact,omitempty"`
	// Remedy says what to change, when there is something to change.
	Remedy string `json:"remedy,omitempty"`
	// Source is the reference that justifies the rule.
	Source string `json:"source,omitempty"`
}

// Context is everything a rule may look at.
type Context struct {
	Manifest *collect.Manifest
	Profile  *Profile
}

// Rule is one check with its metadata.
type Rule struct {
	ID       string
	Title    string
	Severity Severity
	Source   string
	Remedy   string
	Check    func(c *Context) []Finding
}

// Catalog returns every rule, in a stable order.
func Catalog() []Rule {
	all := []Rule{
		pcieLinkDegraded(),
		pciePortCapped(),
		pcieAERCorrectable(),
		pcieAERFatal(),
		memSpeedDownclocked(),
		memMixedSpeedWithinChannel(),
		memMixedPartsWithinChannel(),
		memPopulationUneven(),
		memSlotsEmpty(),
		memECCDisabled(),
		numaSingleNodeMultiSocket(),
		numaAsymmetric(),
		cpuTurboDisabled(),
		cpuGovernorNotPerformance(),
		cpuEnergyPerfBiasNotPerformance(),
		cpuSMTDisabled(),
		osTHPAlways(),
		osNUMABalancingEnabled(),
		osASPMPolicyNotPerformance(),
		osMitigationVulnerable(),
		powerThermalThrottled(),
		powerLimitThrottled(),
		profilePlatformMismatch(),
		profileCPUSockets(),
		profileCPUThreads(),
		profileCPUTurbo(),
		profileMemoryChannels(),
		profileMemoryDPC(),
		profileMemorySlots(),
		profileMemorySpeed(),
		profileMemoryECC(),
		profilePCIeLink(),
		profileOSConfig(),
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	return all
}

// Run evaluates every rule against the context.
func Run(c *Context) []Finding {
	var findings []Finding
	for _, r := range Catalog() {
		findings = append(findings, r.Check(c)...)
	}
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return findings[i].Severity.Rank() > findings[j].Severity.Rank()
		}
		return findings[i].ID < findings[j].ID
	})
	return findings
}

// finding is a small helper that fills the metadata every rule repeats.
func finding(r Rule, detail string, evidence ...string) Finding {
	return Finding{
		ID:       r.ID,
		Severity: r.Severity,
		Title:    r.Title,
		Detail:   detail,
		Evidence: evidence,
		Remedy:   r.Remedy,
		Source:   r.Source,
	}
}

func pct(parts ...float64) string {
	return fmt.Sprintf("%.0f%%", parts[0])
}
