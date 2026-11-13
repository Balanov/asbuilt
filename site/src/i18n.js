// Bilingual copy for the asbuilt landing page.
//
// The text is deliberately dry: what the tool reads, what it finds, what it
// costs, how long it takes. No adjectives that are not measurements.

export const messages = {
  en: {
    "nav.check": "Link check",
    "nav.numbers": "Numbers",
    "nav.rules": "Rules",
    "nav.usage": "Run",
    "nav.download": "Download",
    "brand.parent": "by PARITET",

    "hero.pill": "Open source · MIT · 3.35 MB · no root",
    "hero.title": "The server as built.",
    "hero.lead": "asbuilt reads the state of the machine: how the memory channels are populated, how fast the PCIe links negotiated, whether the package throttled, what the CPU is doing. Every deviation from the expected configuration is priced in lost bandwidth.",
    "hero.download": "Download asbuilt",
    "hero.numbers": "Numbers",
    "hero.meta": "v0.1.0 · Linux, macOS, Windows · static binary · 3.16 ms for a full audit",

    "panel.title": "asbuilt check",
    "panel.findings": "findings",
    "panel.runtime": "per audit",
    "panel.deps": "dependencies",

    "check.kicker": "Link check",
    "check.title": "Three numbers, not two.",
    "check.body": "lspci compares a device with its own capability. It stays silent where the board or the firmware capped the slot. asbuilt reads the endpoint capability, the limit of the port above it and what was actually negotiated, and compares all three.",
    "check.capability": "Endpoint capability",
    "check.port": "Upstream port limit",
    "check.negotiated": "Negotiated",
    "check.verdictKicker": "Verdict",
    "check.verdict": "Link below its capability",
    "check.impact": "of link bandwidth · 31.5 of 63.0 GB/s per direction",
    "check.copy": "The operating system says nothing: the link is up, the device works, and the cluster runs at half the speed on the invoice.",

    "bench.kicker": "Measurements",
    "bench.title": "3.16 ms against 137 ms.",
    "bench.body": "A full audit of one machine: 112 threads, 8 memory channels, one degraded link. 20 runs, one host, the same commands.",
    "bench.tool": "Tool",
    "bench.audit": "Full audit",
    "bench.root": "Root",
    "bench.verdict": "Verdict",
    "bench.loss": "Loss",
    "bench.license": "License",
    "bench.yes": "needed",
    "bench.no": "not needed",
    "bench.part": "partial",
    "bench.full": "full audit",
    "bench.collect": "collects only",
    "bench.compare": "compares nodes",
    "bench.none": "not computed",

    "rules.kicker": "33 rules",
    "rules.title": "Five families.",
    "rules.body": "Every rule carries the measurements it used, the source that justifies it and the remedy. Findings that cannot be priced say so: no number is invented.",

    "usage.kicker": "Run",
    "usage.title": "Three commands.",
    "usage.body": "No daemon, no database, no telemetry. The answer goes to stdout, the exit code goes to CI.",
    "usage.one": "Audit the machine: findings, losses, evidence.",
    "usage.two": "The canonical manifest, for comparing nodes.",
    "usage.three": "Check against an expected profile and fail a pipeline.",
    "usage.install": "Install",

    "footer.tagline": "The server as built · MIT",
    "footer.repo": "Repository",
    "footer.bench": "Measurements",
    "footer.research": "Research",
    "footer.parentBy": "by",
  },

  ru: {
    "nav.check": "Проверка",
    "nav.numbers": "Цифры",
    "nav.rules": "Правила",
    "nav.usage": "Запуск",
    "nav.download": "Скачать",
    "brand.parent": "by PARITET",

    "hero.pill": "Открытый код · MIT · 3,35 МБ · без root",
    "hero.title": "Фактическая конфигурация сервера.",
    "hero.lead": "asbuilt читает состояние машины: заполненность каналов памяти, скорость PCIe-ссылок, троттлинг питания, режимы CPU. Каждое расхождение с ожиданием считается в потерянной пропускной способности.",
    "hero.download": "Скачать asbuilt",
    "hero.numbers": "Цифры",
    "hero.meta": "v0.1.0 · Linux, macOS, Windows · статичный бинарник · 3,16 мс на полный аудит",

    "panel.title": "asbuilt check",
    "panel.findings": "находок",
    "panel.runtime": "на аудит",
    "panel.deps": "зависимостей",

    "check.kicker": "Проверка ссылки",
    "check.title": "Три числа, а не два.",
    "check.body": "lspci сравнивает устройство с его собственными возможностями. Там, где слот ограничила плата или прошивка, он молчит. asbuilt читает capability устройства, лимит порта выше по дереву и фактически выторгованное — и сравнивает все три.",
    "check.capability": "Возможности устройства",
    "check.port": "Лимит порта выше",
    "check.negotiated": "Выторговано",
    "check.verdictKicker": "Вердикт",
    "check.verdict": "Ссылка ниже возможностей",
    "check.impact": "пропускной способности линка · 31,5 из 63,0 ГБ/с",
    "check.copy": "ОС об этом не сообщает: ссылка поднята, устройство работает, а кластер вдвое медленнее, чем в счёте.",

    "bench.kicker": "Замеры",
    "bench.title": "3,16 мс против 137 мс.",
    "bench.body": "Полный аудит одной машины: 112 потоков, 8 каналов памяти, один деградировавший линк. 20 прогонов, один хост, одинаковые команды.",
    "bench.tool": "Инструмент",
    "bench.audit": "Полный аудит",
    "bench.root": "Root",
    "bench.verdict": "Вердикт",
    "bench.loss": "Потеря",
    "bench.license": "Лицензия",
    "bench.yes": "нужен",
    "bench.no": "не нужен",
    "bench.part": "частично",
    "bench.full": "полный аудит",
    "bench.collect": "только сбор",
    "bench.compare": "сравнивает ноды",
    "bench.none": "не считается",

    "rules.kicker": "33 правила",
    "rules.title": "Пять семейств.",
    "rules.body": "У каждого правила — доказательства, источник и лекарство. Находки, которые нельзя посчитать, так и помечены: число не выдумывается.",

    "usage.kicker": "Запуск",
    "usage.title": "Три команды.",
    "usage.body": "Без демона, без базы, без телеметрии. Ответ — в stdout, код возврата — для CI.",
    "usage.one": "Аудит машины: находки, потери, доказательства.",
    "usage.two": "Канонический манифест, для сравнения нод.",
    "usage.three": "Проверка по ожидаемому профилю и код возврата для pipeline.",
    "usage.install": "Установка",

    "footer.tagline": "Фактическая конфигурация сервера · MIT",
    "footer.repo": "Репозиторий",
    "footer.bench": "Замеры",
    "footer.research": "Исследование",
    "footer.parentBy": "by",
  },
};

export const familyText = {
  en: {
    pcie: {
      name: "PCIe",
      rules: [
        "Link below endpoint capability",
        "Endpoint capability above the port limit",
        "Correctable errors",
        "Fatal and non-fatal errors",
      ],
    },
    memory: {
      name: "Memory",
      rules: [
        "Below the rating of the modules",
        "Mixed rated speeds in a channel",
        "Mixed part numbers in a channel",
        "Channels populated differently",
        "Empty slots",
        "No error correction",
      ],
    },
    numa: {
      name: "NUMA",
      rules: [
        "One node on a multi-socket machine",
        "Nodes with different memory sizes",
      ],
    },
    cpu: {
      name: "CPU",
      rules: [
        "Turbo closed",
        "Governor is not performance",
        "Energy performance bias",
        "Hardware threads off",
      ],
    },
    firmware: {
      name: "Firmware and OS",
      rules: [
        "Transparent huge pages always on",
        "Automatic NUMA balancing",
        "PCIe ASPM policy",
        "Vulnerability without mitigation",
        "Thermal throttling",
        "Power limit throttling",
      ],
    },
  },
  ru: {
    pcie: {
      name: "PCIe",
      rules: [
        "Ссылка ниже возможностей устройства",
        "Возможности выше лимита порта",
        "Корректируемые ошибки",
        "Фатальные и нефатальные ошибки",
      ],
    },
    memory: {
      name: "Память",
      rules: [
        "Работа ниже рейтинга модулей",
        "Разные рейтинги скорости в канале",
        "Разные part numbers в канале",
        "Каналы заполнены по-разному",
        "Пустые слоты",
        "Нет коррекции ошибок",
      ],
    },
    numa: {
      name: "NUMA",
      rules: [
        "Один узел на многосокетной машине",
        "Узлы с разным объёмом памяти",
      ],
    },
    cpu: {
      name: "CPU",
      rules: [
        "Turbo закрыт",
        "Governor не производительный",
        "Energy performance bias",
        "Аппаратные потоки выключены",
      ],
    },
    firmware: {
      name: "Прошивка и ОС",
      rules: [
        "Transparent huge pages всегда включены",
        "Автобалансировка NUMA",
        "Политика PCIe ASPM",
        "Уязвимость без mitigation",
        "Термический троттлинг",
        "Троттлинг по лимиту мощности",
      ],
    },
  },
};
