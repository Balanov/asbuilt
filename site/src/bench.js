// Benchmark rows for the landing page.
//
// Every runtime figure here was measured on the same machine with the
// harness in bench/measure.py: 20 timed runs after 2 discarded warm-ups,
// binary and fixture on the native Linux filesystem, stdout to DEVNULL.
// Rows that could not be run in that environment carry the reason instead
// of an estimate. The full dataset is in bench/COMPETITORS.md.

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
    note: "C, the only SMBIOS dump here, but no verdict",
    time: "0.69 ms",
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
    ratioCopy: "A full audit of a 112-thread machine with eight memory channels and a degraded PCIe link, in one process, without root.",
    uniq: "the only one",
    uniqLabel: "that runs unprivileged",
    uniqCopy: "Verified byte-identical output as user nobody. The degraded-link verdict comes from world-readable sysfs, so it works in a container.",
  },
  ru: {
    ratio: "43×",
    ratioLabel: "быстрее lshw",
    ratioCopy: "Полный аудит машины на 112 потоков, восемь каналов памяти и деградировавшая PCIe-ссылка — одним процессом, без root.",
    uniq: "единственный",
    uniqLabel: "кто работает без прав",
    uniqCopy: "Проверено: вывод побайтово совпадает под пользователем nobody. Вердикт по ссылке берётся из world-readable sysfs, поэтому работает в контейнере.",
  },
};

export const benchNote = {
  en: 'Runtime figures are the mean of 20 runs of the same commands on one machine, with the binary and the fixture on the native Linux filesystem: 3.16 ms for asbuilt against 137 ms for lshw, measured with <code>bench/measure.py</code> in the repository. Rows marked "—" were not run here, and say why: PerfSpect needs an Intel PMU and root, SAT hwmatch needs an HPE Cray EX stack. Every licence and scope claim links to the project it comes from. The full dataset, including the caveats, is in <a href="https://github.com/Balanov/asbuilt/blob/main/bench/COMPETITORS.md">bench/COMPETITORS.md</a>.',
  ru: 'Время — среднее из 20 прогонов одних и тех же команд на одной машине, бинарник и фикстура на нативной файловой системе Linux: 3,16 мс у asbuilt против 137 мс у lshw, замер скриптом <code>bench/measure.py</code> из репозитория. Строки с «—» здесь не запускались, и рядом указано почему: PerfSpect нужен Intel PMU и root, SAT hwmatch — стек HPE Cray EX. Каждое утверждение о лицензии и области работы ссылается на исходный проект. Полный набор данных вместе с оговорками — в <a href="https://github.com/Balanov/asbuilt/blob/main/bench/COMPETITORS.md">bench/COMPETITORS.md</a>.',
};
