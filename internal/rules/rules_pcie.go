package rules

import (
	"fmt"

	"github.com/Balanov/asbuilt/internal/collect"
)

// pcieLinkDegraded catches a link that negotiated below what the endpoint
// can do.
//
// lspci reports "downgraded" by comparing a device with its own capability,
// which misses the case where the upstream port or the firmware pinned the
// slot to a lower generation. The kernel publishes the negotiated state as
// world-readable sysfs, so this check needs no root and no pciutils.
func pcieLinkDegraded() Rule { return Rule{
	ID:       "pcie.link.degraded",
	Title:    "PCIe link negotiated below endpoint capability",
	Severity: Critical,
	Source:   "https://www.dell.com/support/kbdoc/en-us/000412311/",
	Remedy:   "Seat the card and the riser properly, then check the slot speed setting in firmware.",
	Check: func(c *Context) []Finding {
		var out []Finding
		for _, d := range c.Manifest.PCIe.Devices {
			if !linkDegraded(d) {
				continue
			}
			potential := collect.LinkBandwidthGBs(d.MaxSpeedGTs, d.MaxWidth)
			actual := collect.LinkBandwidthGBs(d.CurrentSpeedGTs, d.CurrentWidth)
			f := finding(pcieLinkDegraded(),
				fmt.Sprintf("%s negotiated %.1f GT/s x%d while the endpoint supports %.1f GT/s x%d",
					d.Address, d.CurrentSpeedGTs, d.CurrentWidth, d.MaxSpeedGTs, d.MaxWidth),
				fmt.Sprintf("current_link_speed=%s current_link_width=%d", speedText(d.CurrentSpeedGTs), d.CurrentWidth),
				fmt.Sprintf("max_link_speed=%s max_link_width=%d", speedText(d.MaxSpeedGTs), d.MaxWidth),
			)
			f.Impact = &Impact{
				Kind:     "pcie_bandwidth",
				LossPct:  lossPct(potential, actual),
				Potential: potential,
				Actual:   actual,
				Unit:     "GB/s",
			}
			out = append(out, f)
		}
		return out
	},
}
}

// pciePortCapped catches a link limited by the upstream port rather than
// by the device. This is the case a board vendor pins in firmware, and the
// case no existing tool explains.
func pciePortCapped() Rule { return Rule{
	ID:       "pcie.link.port_capped",
	Title:    "PCIe endpoint capability exceeds the upstream port limit",
	Severity: Warn,
	Source:   "https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/_sources/troubleshooting/networking_troubleshooting.rst.txt",
	Remedy:   "Align the firmware slot speed setting with the platform specification.",
	Check: func(c *Context) []Finding {
		var out []Finding
		for _, d := range c.Manifest.PCIe.Devices {
			if d.PortMaxSpeedGTs == 0 || d.MaxSpeedGTs == 0 {
				continue
			}
			if d.MaxSpeedGTs <= d.PortMaxSpeedGTs {
				continue
			}
			potential := collect.LinkBandwidthGBs(d.MaxSpeedGTs, d.MaxWidth)
			actual := collect.LinkBandwidthGBs(d.PortMaxSpeedGTs, d.PortMaxWidth)
			f := finding(pciePortCapped(),
				fmt.Sprintf("%s supports %.1f GT/s but the upstream port allows %.1f GT/s x%d",
					d.Address, d.MaxSpeedGTs, d.PortMaxSpeedGTs, d.PortMaxWidth),
				fmt.Sprintf("endpoint max_link_speed=%s max_link_width=%d", speedText(d.MaxSpeedGTs), d.MaxWidth),
				fmt.Sprintf("upstream port max_link_speed=%s max_link_width=%d", speedText(d.PortMaxSpeedGTs), d.PortMaxWidth),
			)
			f.Impact = &Impact{
				Kind:     "pcie_bandwidth",
				LossPct:  lossPct(potential, actual),
				Potential: potential,
				Actual:   actual,
				Unit:     "GB/s",
			}
			out = append(out, f)
		}
		return out
	},
}
}

// pcieAERCorrectable reports correctable PCIe errors, the early signal of
// a marginal link that will eventually train down.
func pcieAERCorrectable() Rule { return Rule{
	ID:       "pcie.aer.correctable",
	Title:    "PCIe correctable errors were reported",
	Severity: Warn,
	Source:   "https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-pci-devices-aer",
	Remedy:   "Watch the error rate under load: a rising rate means the link is losing margin.",
	Check: func(c *Context) []Finding {
		var out []Finding
		for _, d := range c.Manifest.PCIe.Devices {
			if d.AERCorrectable == 0 {
				continue
			}
			out = append(out, finding(pcieAERCorrectable(),
				fmt.Sprintf("%s reported %d correctable PCIe errors", d.Address, d.AERCorrectable),
				fmt.Sprintf("aer_dev_correctable=%d", d.AERCorrectable)))
		}
		return out
	},
}
}

// pcieAERFatal reports unrecoverable PCIe errors.
func pcieAERFatal() Rule { return Rule{
	ID:       "pcie.aer.fatal",
	Title:    "PCIe non-fatal or fatal errors were reported",
	Severity: Critical,
	Source:   "https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-pci-devices-aer",
	Remedy:   "Replace the cable, riser or card: the link is failing, not merely noisy.",
	Check: func(c *Context) []Finding {
		var out []Finding
		for _, d := range c.Manifest.PCIe.Devices {
			if d.AERFatal == 0 {
				continue
			}
			out = append(out, finding(pcieAERFatal(),
				fmt.Sprintf("%s reported %d fatal or non-fatal PCIe errors", d.Address, d.AERFatal),
				fmt.Sprintf("aer_dev_fatal+aer_dev_nonfatal=%d", d.AERFatal)))
		}
		return out
	},
}
}

func speedText(gts float64) string {
	if gts <= 0 {
		return "unknown"
	}
	return fmt.Sprintf("%.1f GT/s", gts)
}

// linkDegraded reports whether a link negotiated below what the endpoint
// can do.
//
// The verdict is computed from the numbers rather than read from a flag, so
// that it holds for a manifest that was loaded from a file as well.
func linkDegraded(d collect.PCIeDevice) bool {
	if d.MaxSpeedGTs > 0 && d.CurrentSpeedGTs > 0 && d.CurrentSpeedGTs < d.MaxSpeedGTs {
		return true
	}
	if d.MaxWidth > 0 && d.CurrentWidth > 0 && d.CurrentWidth < d.MaxWidth {
		return true
	}
	return false
}

func lossPct(potential, actual float64) float64 {
	if potential <= 0 {
		return 0
	}
	l := (potential - actual) / potential * 100
	if l < 0 {
		return 0
	}
	return l
}
