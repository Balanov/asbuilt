# Rule reference

These are the rule ids printed by `asbuilt rules`. Each row describes the condition a finding reports, not a benchmark or a guarantee of performance. Missing or unreadable evidence does not prove a fault. Profile rules apply only when a profile supplies the corresponding expectation.

| Rule id | Condition reported | Evidence | Reference |
| --- | --- | --- | --- |
| `cpu.energy_perf_bias.not_performance` | CPU energy performance bias is not set to performance | CPU topology and frequency/power settings exposed by the OS | [Source](https://www.kernel.org/doc/html/latest/admin-guide/pm/amd-pstate.html) |
| `cpu.governor.not_performance` | CPU frequency governor is not a performance governor | CPU topology and frequency/power settings exposed by the OS | [Source](https://documentation.suse.com/sles/15-SP7/single-html/SLES-tuning/index.html) |
| `cpu.smt.disabled` | Simultaneous multithreading is disabled | CPU topology and frequency/power settings exposed by the OS | [Source](https://www.intel.com/content/www/us/en/developer/articles/technical/measuring-impact-of-numa-migrations-on-performance.html) |
| `cpu.turbo.disabled` | CPU turbo is not available | CPU topology and frequency/power settings exposed by the OS | [Source](https://www.intel.com/content/www/us/en/support/articles/000057277.html) |
| `mem.ecc.disabled` | No memory error correction is exposed | SMBIOS/DMI memory devices and channel inventory | [Source](https://www.vmware.com/docs/vsphere-esxi-vcenter-server-9-1-performance-best-practices) |
| `mem.parts.mixed_within_channel` | A memory channel mixes module part numbers | SMBIOS/DMI memory devices and channel inventory | [Source](https://www.intel.com/content/www/us/en/support/articles/000088926.html) |
| `mem.population.uneven` | Memory channels are not populated alike | SMBIOS/DMI memory devices and channel inventory | [Source](https://lenovopress.lenovo.com/lp2000-balanced-memory-configurations-4th-5th-gen-2s-intel-processors) |
| `mem.slots.empty` | Memory slots are empty | SMBIOS/DMI memory devices and channel inventory | [Source](https://server-configurator.com/learn/server-memory-channels-ranks) |
| `mem.speed.below_module_rating` | Memory runs below the rating of the installed modules | SMBIOS/DMI memory devices and channel inventory | [Source](https://server-configurator.com/learn/ddr5-server-memory-population-rules) |
| `mem.speed.mixed_within_channel` | A memory channel mixes modules of different rated speed | SMBIOS/DMI memory devices and channel inventory | [Source](https://www.intel.com/content/www/us/en/support/articles/000088926.html) |
| `numa.asymmetric` | NUMA nodes hold different amounts of memory | OS NUMA node topology and memory sizes | [Source](https://lenovopress.lenovo.com/lp2000-balanced-memory-configurations-4th-5th-gen-2s-intel-processors) |
| `numa.single_node_multi_socket` | Multi-socket machine presents a single NUMA node | OS NUMA node topology and memory sizes | [Source](https://www.intel.com/content/www/us/en/developer/articles/technical/measuring-impact-of-numa-migrations-on-performance.html) |
| `os.mitigation.vulnerable` | The kernel reports a vulnerability without mitigation | Kernel and OS settings | [Source](https://www.kernel.org/doc/html/latest/admin-guide/hw-vuln/index.html) |
| `os.numa_balancing.enabled` | Automatic NUMA balancing is active | Kernel and OS settings | [Source](https://documentation.suse.com/sles/15-SP7/single-html/SLES-tuning/index.html) |
| `os.pcie_aspm.not_performance` | PCIe link power management is not set to performance | Kernel and OS settings | [Source](https://onestopsystems.com/blogs/one-stop-systems-blog/beyond-lspci-how-to-debug-maintain-high-density-pcie-gen5-gen6-fabrics-at-the-edge) |
| `os.thp.always` | Transparent huge pages are enabled for all allocations | Kernel and OS settings | [Source](https://documentation.suse.com/sles/15-SP7/single-html/SLES-tuning/index.html) |
| `pcie.aer.correctable` | PCIe correctable errors were reported | PCIe endpoint/upstream link capabilities and AER counters | [Source](https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-pci-devices-aer) |
| `pcie.aer.fatal` | PCIe non-fatal or fatal errors were reported | PCIe endpoint/upstream link capabilities and AER counters | [Source](https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-bus-pci-devices-aer) |
| `pcie.link.degraded` | PCIe link negotiated below endpoint capability | PCIe endpoint/upstream link capabilities and AER counters | [Source](https://www.dell.com/support/kbdoc/en-us/000412311/) |
| `pcie.link.port_capped` | PCIe endpoint capability exceeds the upstream port limit | PCIe endpoint/upstream link capabilities and AER counters | [Source](https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/_sources/troubleshooting/networking_troubleshooting.rst.txt) |
| `power.throttle.limit` | The processor has been power limit throttled | Processor throttling counters | [Source](https://www.intel.com/content/www/us/en/support/articles/000057277.html) |
| `power.throttle.thermal` | The processor has been thermally throttled | Processor throttling counters | [Source](https://www.suse.com/support/update/announcement/2024/suse-su-20240224-1/) |
| `profile.cpu.sockets` | Socket count does not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://cray-hpe.github.io/docs-sat/en-25/introduction/) |
| `profile.cpu.threads_per_core` | Threads per core do not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://cray-hpe.github.io/docs-sat/en-25/introduction/) |
| `profile.cpu.turbo` | Turbo state does not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://cray-hpe.github.io/docs-sat/en-25/introduction/) |
| `profile.memory.channels` | Populated memory channels do not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://cray-hpe.github.io/docs-sat/en-25/introduction/) |
| `profile.memory.configured_mts` | Memory runs at a speed other than the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://server-configurator.com/learn/ddr5-server-memory-population-rules) |
| `profile.memory.dimms_per_channel` | Modules per channel do not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://lenovopress.lenovo.com/lp2000-balanced-memory-configurations-4th-5th-gen-2s-intel-processors) |
| `profile.memory.ecc` | Memory error correction does not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://www.vmware.com/docs/vsphere-esxi-vcenter-server-9-1-performance-best-practices) |
| `profile.memory.slots_filled` | Populated memory slots do not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://cray-hpe.github.io/docs-sat/en-25/introduction/) |
| `profile.os.mismatch` | Operating system settings do not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://github.com/intel/PerfSpect) |
| `profile.pcie.min_link` | A PCIe link is slower than the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/_sources/troubleshooting/networking_troubleshooting.rst.txt) |
| `profile.platform.mismatch` | The platform does not match the expected profile | Collected manifest compared with the explicit expected profile | [Source](https://github.com/intel/PerfSpect) |

Use `asbuilt explain <rule-id>` for the remedy and reference, or `asbuilt rules --json` for the machine-readable catalogue.
