<div align="center">
  <img src="assets/brand/logo-512.png" alt="asbuilt — фактическая конфигурация сервера" width="128">
  <p><strong>Узнать, чем машина является на самом деле, и во сколько это обходится.</strong></p>
  <p>Linux · x86_64 · статичный бинарник · без root · без зависимостей · MIT</p>
  <p>
    <a href="#быстрый-старт">Быстрый старт</a> ·
    <a href="#что-проверяется">Что проверяется</a> ·
    <a href="#профили">Профили</a> ·
    <a href="#манифесты-и-диффы">Манифесты</a> ·
    <a href="#ci-в-gate">CI-ворота</a> ·
    <a href="BENCHMARKS.md">Замеры</a> ·
    <a href="RESEARCH.md">Исследование</a> ·
    <a href="README.md">English</a>
  </p>
</div>

ASBUILT — это фактическая конфигурация сервера. Он читает, чем машина является
на самом деле, судит её по правилам, которые важны для дата-центрового
оборудования, и печатает вердикт с доказательствами и, где арифметика позволяет,
с измеренной потерей производительности.

Сервер бывает собран из деталей, каждая из которых исправна, а вместе они
неверны. Ссылка PCIe скатывается на поколение назад. Память работает на ступень
ниже рейтинга модулей. В один сокет воткнули другой комплект DIMM. Машина
загружается, бенчмарк отрабатывает, а результат просто ниже, чем обещал
даташит. Ни одна операционная система об этом не сообщает.

```console
$ asbuilt check
ASBUILT 0.1.0  (manifest schema 1)
Dell Inc.  PowerEdge R760
SKU 0KJ3T2  serial CN7792
BIOS  Dell Inc. 2.10.1 (03/14/2024)
BOARD Dell Inc. 0KJ3T2

CPU   Intel(R) Xeon(R) Platinum 8480+
      2 socket(s) x 28 cores, 2 threads per core, 112 logical CPUs, SMT on
      2000 MHz base, 3500 MHz max, turbo on, governor performance
      ISA fingerprint 9472f9464c3f44e4 (78 features)
      cache L1d 48K, L1i 32K, L2 2048K, L3 107520K

MEM   8 channels, 16 DIMMs, 1006644 MB total
      4800 MT/s configured, modules rated 5600 MT/s
      peak bandwidth 307.2 GB/s (358.4 GB/s at rated speed)
      ECC on, source smbios

NUMA  2 node(s)

PCIE  2 functions, 1 degraded, AER errors on
      0000:01:00.0 16.0 GT/s x16 of 32.0 GT/s x16

FINDINGS  2 critical, 3 warn, 2 info

CRITICAL PCIe link negotiated below endpoint capability
         pcie.link.degraded
         0000:01:00.0 negotiated 16.0 GT/s x16 while the endpoint supports 32.0 GT/s x16
         impact: pcie_bandwidth -50.0% (31.5 of 63.0 GB/s)
         evidence: current_link_speed=16.0 GT/s current_link_width=16
         evidence: max_link_speed=32.0 GT/s max_link_width=16
         remedy: Seat the card and the riser properly, then check the slot speed setting in firmware.
```

## Почему asbuilt

- **Суждение, а не инвентаризация.** `lshw`, `dmidecode`, `lscpu` и `lspci`
  собирают факты; Intel PerfSpect собирает их прекрасно. Ни один из них не
  решает, правильна ли машина, и ни один не умеет сказать, сколько стоит
  ошибка конфигурации. Этот слой и есть продукт.
- **Потери посчитаны.** Находка сообщает, сколько пропускной способности
  потеряно, вычисляя это по собственным числам машины: 50% линка, 14.3% памяти.
  Если посчитать нельзя — так и написано, число не выдумывается.
- **Работает без root.** Ядро публикует состояние PCIe-ссылки и счётчики AER
  как world-readable атрибуты sysfs, поэтому проверка деградировавшей ссылки
  работает без прав и внутри контейнера, без `pciutils`. Root добавляет таблицу
  прошивки, а в ней живёт раскладка памяти.
- **Честно о том, чего не прочитал.** Недоступный источник становится заметкой в
  манифесте, а не тихим пропуском.
- **Без демона, без базы, без телеметрии, без GUI.** Один статичный бинарник,
  который печатает в stdout и возвращает код.

## Установка

### Linux

```sh
VERSION=0.1.0
ARCH=amd64  # На ARM64 используйте arm64.
curl -fSLo asbuilt "https://github.com/Balanov/asbuilt/releases/download/v${VERSION}/asbuilt-linux-${ARCH}"
mkdir -p "$HOME/.local/bin"
install -m 0755 asbuilt "$HOME/.local/bin/asbuilt"
export PATH="$HOME/.local/bin:$PATH"
asbuilt --version
```

Убедитесь, что `$HOME/.local/bin` в `PATH`. В каждый релиз входит
`checksums.txt`:

```sh
curl -fSLO "https://github.com/Balanov/asbuilt/releases/download/v${VERSION}/checksums.txt"
grep "asbuilt-linux-${ARCH}$" checksums.txt | sha256sum -c -
```

Бинарники пока не подписаны, поэтому SmartScreen или Gatekeeper могут
предупредить. Проверяйте SHA-256 перед запуском скачанного файла. На macOS
после проверки снимите карантин только с этого файла:

```sh
xattr -d com.apple.quarantine "$HOME/.local/bin/asbuilt"
```

### Сборка из исходников

С Go 1.26 или новее:

```sh
go install github.com/Balanov/asbuilt/cmd/asbuilt@latest
```

или в клоне:

```sh
go build -trimpath -ldflags "-s -w" -o asbuilt ./cmd/asbuilt
```

## Быстрый старт

Попробовать можно без сервера и без root: в репозитории лежит генерируемая
фикстура, которая воспроизводит неверно собранный двухсокетный сервер.

```sh
go run ./tools/fixturegen -out tests/fixtures/server
asbuilt --root tests/fixtures/server check
```

Дальше — настоящая машина:

```sh
asbuilt check
asbuilt check --json
asbuilt manifest > node-014.json
```

## Что проверяется

Тридцать три правила в пяти семействах. У каждого правила есть доказательства,
на которых оно основано, и источник, который его обосновывает; всё это
печатает `asbuilt explain`.

| Семейство | Правила |
|---|---|
| PCIe | ссылка ниже возможностей устройства, возможности устройства выше ограничения порта, корректируемые ошибки, фатальные и нефатальные ошибки |
| Память | работа ниже рейтинга модулей, разные рейтинги скорости в канале, разные part numbers в канале, каналы заполнены по-разному, пустые слоты, нет коррекции ошибок |
| NUMA | один узел на многосокетной машине, узлы с разным объёмом памяти |
| CPU | turbo закрыт, governor не производительный, energy performance bias, SMT выключен |
| Прошивка и ОС | transparent huge pages всегда включены, автоматическая балансировка NUMA, политика PCIe ASPM, уязвимость без mitigations, термический троттлинг, троттлинг по лимиту мощности |

Сравнение PCIe — то, чего не делает никто. `lspci` сравнивает устройство с его
собственными возможностями: это не замечает случай, когда слот ограничила
плата или прошивка, и ругается на случай, когда устройство само врёт. asbuilt
сравнивает три числа: что умеет конечное устройство, что разрешает порт выше
по дереву и что реально выторговалось.

```console
$ asbuilt rules
RULES  33

CRITICAL pcie.link.degraded                   PCIe link negotiated below endpoint capability
CRITICAL mem.population.uneven                Memory channels are not populated alike
...
```

## Профили

Профиль — это как флот должен быть собран, в обычном JSON, чтобы hardware-команда
держала его рядом с инвентарём.

```json
{
  "name": "xd-node",
  "expect": {
    "platform": { "product": "PowerEdge R760" },
    "cpu": { "sockets": 2, "threads_per_core": 2, "turbo_enabled": true },
    "memory": { "channels": 8, "dimms_per_channel": 2, "configured_mts": 5600, "ecc_enabled": true },
    "pcie": { "min_link_gts": 32 },
    "os": { "transparent_hugepage": "madvise", "governor": "performance" }
  }
}
```

```console
$ asbuilt --profile profiles/xd-node.json check
...
CRITICAL Memory runs at a speed other than the expected profile
         profile.memory.configured_mts
         memory runs at 4800 MT/s, expected 5600 MT/s
         impact: memory_bandwidth -14.3% (307.2 of 358.4 GB/s)
```

## Манифесты и диффы

`asbuilt manifest` печатает канонический JSON-манифест: идентичность по
прошивке, процессор, раскладку памяти по каналам, PCIe-ссылки, состояние ОС и
отпечаток набора инструкций, который отдаёт ядро.

```sh
asbuilt manifest > node-014.json
asbuilt diff node-014.json node-015.json
```

```console
$ asbuilt diff node-014.json node-015.json
DIFF  3 difference(s)

  memory.configured_mts
    a: 5600
    b: 4800
    affects: memory bandwidth

  pcie.0000:01:00.0.current_speed_gts
    a: 32.0
    b: 16.0
    affects: pcie bandwidth
```

Именно дифф превращает груду JSON в аудит: две ноды, собранные одинаково, должны
давать пустой дифф, а каждая строка в нём — либо ошибка конфигурации, либо
дефект железа.

## CI-ворота

`asbuilt` возвращает ненулевой код, когда находка достигает порога, так что
pipeline снабжения может отказать ноду, которая не соответствует заказу.

```sh
asbuilt check --fail-on critical || exit 1
```

| Код | Значение |
|---|---|
| 0 | ни одна находка не достигла порога |
| 1 | находка достигла порога |
| 2 | аудит не удалось выполнить |

## Устройство проекта

Движок правил намеренно отделён от сборщиков данных. Сборщики читают ядро и
прошивку, движок решает, что эти факты значат. Именно это разделение делает
проект форкабельным: hardware-команда, которая знает свою платформу, может
добавлять правила, не трогая парсеры, а баг в парсере чинится без изменения
вердиктов.

Каждый сборщик только читает и работает с файловой системой относительно
корня, поэтому один и тот же код обслуживает живой сервер, образ контейнера и
записанное дерево фикстур. Декодер SMBIOS написан по спецификации DMTF 3.9.0,
а не по одному-единственному дампу: заполненность памяти, настроенная скорость
и рейтинг модулей — это три разных поля, и именно их различие лежит в основе
всех правил про память.

## Документация

- [RESEARCH.md](RESEARCH.md) — исследование рынка, три рассмотренных кандидата
  и почему построен этот.
- [BENCHMARKS.md](BENCHMARKS.md) — измеренный размер, скорость и покрытие.
- [CHANGELOG.md](CHANGELOG.md) — что менялось.
- `asbuilt explain <rule-id>` — обоснование и источник для любого правила.

## Лицензия

MIT, см. [LICENSE](LICENSE). Форкайте, правьте, используйте.
