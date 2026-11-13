// Benchmark rows for the landing page.
//
// Every runtime figure was measured on the same machine with the harness in
// bench/measure.py: 20 timed runs after 2 discarded warm-ups, binary and
// fixture on the native Linux filesystem, stdout to DEVNULL. Rows that could
// not be run in that environment carry the reason instead of an estimate.
// The full dataset is in bench/COMPETITORS.md.

export const benchRows = [
  {
    tool: "asbuilt",
    note: "3.35 MB, static, Go, zero dependencies",
    time: "3.16 ms",
    root: "no",
    verdict: "full",
    loss: "−50.0% PCIe · −14.3% RAM",
    license: "MIT",
    asbuilt: true,
  },
  {
    tool: "lshw",
    note: "C, distribution package, tree output only",
    time: "137 ms",
    root: "yes",
    verdict: "collect",
    loss: null,
    license: "GPL-2.0",
  },
  {
    tool: "pcicrawler",
    note: "Python package, PCIe only, README requires root",
    time: "50.9 ms",
    root: "yes",
    verdict: "collect",
    loss: null,
    license: "MIT",
  },
  {
    tool: "lspci",
    note: "pciutils, link state needs root, compares two numbers",
    time: "12.6 ms",
    root: "part",
    verdict: "collect",
    loss: null,
    license: "GPL-2.0",
  },
  {
    tool: "lscpu",
    note: "util-linux, no DIMM population, no link state",
    time: "8.24 ms",
    root: "no",
    verdict: "collect",
    loss: null,
    license: "GPL-2.0",
  },
  {
    tool: "dmidecode",
    note: "C, SMBIOS dump, no verdict",
    time: "0.69 ms*",
    root: "yes",
    verdict: "collect",
    loss: null,
    license: "GPL-2.0",
  },
  {
    tool: "Intel PerfSpect",
    note: "45.7 MB tarball, BSD-3, needs an Intel PMU and root",
    time: "—",
    root: "yes",
    verdict: "collect",
    loss: null,
    license: "BSD-3-Clause",
  },
  {
    tool: "Cray SAT hwmatch",
    note: "HPE Cray EX + CSM only, CPU and memory only",
    time: "—",
    root: "no",
    verdict: "compare",
    loss: null,
    license: "MIT",
  },
];

export const benchHeadline = {
  en: {
    ratio: "43×",
    ratioLabel: "faster than lshw",
    ratioCopy: "Full audit of a 112-thread machine: 3.16 ms against 136.69 ms. 20 runs, one host.",
    uniq: "0",
    uniqLabel: "root required",
    uniqCopy: "The link verdict is read from world-readable sysfs. Verified as user nobody: byte-identical output.",
  },
  ru: {
    ratio: "43×",
    ratioLabel: "быстрее lshw",
    ratioCopy: "Полный аудит машины на 112 потоков: 3,16 мс против 136,69 мс. 20 прогонов, один хост.",
    uniq: "0",
    uniqLabel: "нужен root",
    uniqCopy: "Вердикт по ссылке читается из world-readable sysfs. Проверено под пользователем nobody: вывод побайтово тот же.",
  },
};

export const benchNote = {
  en: 'Runtime figures are the mean of 20 runs of the same commands on one host, binary and fixture on the native Linux filesystem: <code>bench/measure.py</code> in the repository. Rows marked "—" were not run here, and say why: PerfSpect needs an Intel PMU and root, SAT hwmatch needs an HPE Cray EX stack. * dmidecode has no SMBIOS table on this host, so its timing measures the absence of one. Full dataset with the caveats: <a href="https://github.com/Balanov/asbuilt/blob/main/bench/COMPETITORS.md">bench/COMPETITORS.md</a>.',
  ru: 'Время — среднее из 20 прогонов одних и тех же команд на одном хосте, бинарник и фикстура на нативной файловой системе Linux: <code>bench/measure.py</code> из репозитория. Строки с «—» здесь не запускались, и рядом указано почему: PerfSpect нужен Intel PMU и root, SAT hwmatch — стек HPE Cray EX. * У dmidecode на этом хосте нет таблицы SMBIOS, поэтому его замер меряет отсутствие таблицы. Полный набор данных с оговорками: <a href="https://github.com/Balanov/asbuilt/blob/main/bench/COMPETITORS.md">bench/COMPETITORS.md</a>.',
};
