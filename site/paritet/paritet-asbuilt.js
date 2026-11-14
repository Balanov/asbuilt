(() => {
  "use strict";

  const T = (s) => (window.__t ? window.__t(s) : s);

  /* How long one slide stays on screen. The line under the tabs empties
     during this time and the slide changes when it runs out. */
  const SLIDE_MS = 9000;

  let raf = null;
  let deadline = 0;
  let remaining = SLIDE_MS;
  let current = 0;

  /* ---------- the locale of the page ---------- */

  function localeOf() {
    const raw = window.__paritetLang || document.documentElement.lang || "ru";
    return String(raw).toLowerCase().split("-")[0];
  }

  function decimalOf(lang) {
    return lang === "ru" || lang === "be" || lang === "es" ? "," : ".";
  }

  /* Переводит текст уже вставленного блока тем же словарём, что и остальная
     страница: неизвестные строки остаются как есть. */
  function translateTree(root) {
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    const nodes = [];
    while (walker.nextNode()) nodes.push(walker.currentNode);
    for (const node of nodes) {
      const raw = node.nodeValue;
      const key = raw.trim();
      if (!key) continue;
      const tr = T(key);
      if (tr !== key) {
        const lead = raw.match(/^\s*/)[0];
        const trail = raw.match(/\s*$/)[0];
        node.nodeValue = lead + tr + trail;
      }
    }
    root.querySelectorAll("[aria-label]").forEach((el) => {
      const v = el.getAttribute("aria-label");
      const tr = T(v);
      if (tr !== v) el.setAttribute("aria-label", tr);
    });
  }

  /* Якорь: рядом с блоком Skein, иначе после раздела с продуктами, иначе
     перед подвалом. */
  function findAnchor() {
    const skein = document.getElementById("skein-evidence");
    if (skein) return { node: skein, position: "afterend" };

    const titles = ["Наши продукты.", "Our products."];
    for (const title of titles) {
      const needle = T(title).toLocaleUpperCase();
      const headings = document.querySelectorAll("#root section h2, #root h2");
      const found = [...headings].find((heading) => heading.textContent.toLocaleUpperCase().includes(needle));
      const section = found?.closest("section");
      if (section) return { node: section, position: "afterend" };
    }

    const root = document.getElementById("root");
    if (!root) return null;
    const footer = root.querySelector("footer");
    if (footer) return { node: footer, position: "beforebegin" };
    return { node: root, position: "beforeend" };
  }

  /* ---------- слайды ---------- */

  function slideList() {
    return [...document.querySelectorAll("#asbuilt-slider .ab-slides > *")];
  }

  function showSlide(index) {
    const slides = slideList();
    if (!slides.length) return;
    current = (index + slides.length) % slides.length;
    slides.forEach((slide, position) => slide.classList.toggle("is-active", position === current));
    document.querySelectorAll("#asbuilt-slider .ab-tab").forEach((tab) => {
      const active = Number(tab.dataset.index) === current;
      tab.classList.toggle("is-active", active);
      tab.setAttribute("aria-selected", active ? "true" : "false");
    });
    restartTimer();
  }

  /* ---------- линия-таймер ---------- */

  function timerFill() {
    return document.querySelector("#asbuilt-slider .ab-timer-fill");
  }

  function restartTimer() {
    remaining = SLIDE_MS;
    deadline = Date.now() + remaining;
    if (raf) cancelAnimationFrame(raf);
    raf = requestAnimationFrame(tick);
  }

  function tick() {
    const left = Math.max(0, deadline - Date.now());
    const fill = timerFill();
    if (fill) fill.style.transform = `scaleX(${(left / SLIDE_MS).toFixed(4)})`;
    if (left <= 0) {
      showSlide(current + 1);
      return;
    }
    raf = requestAnimationFrame(tick);
  }

  /* Скрытая вкладка не должна съесть время показа. */
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
      remaining = Math.max(0, deadline - Date.now());
      if (raf) cancelAnimationFrame(raf);
      raf = null;
      return;
    }
    if (remaining > 0) {
      deadline = Date.now() + remaining;
      if (!raf) raf = requestAnimationFrame(tick);
    }
  });

  /* ---------- множители отрыва ---------- */

  function parseNumber(text) {
    const match = String(text ?? "").match(/\d+(?:[.,]\d+)?/);
    if (!match) return null;
    const value = parseFloat(match[0].replace(",", "."));
    return Number.isFinite(value) ? value : null;
  }

  function formatGap(value) {
    if (!Number.isFinite(value) || value <= 0) return "";
    const separator = decimalOf(localeOf());
    if (value >= 100) return `×${value.toFixed(1).replace(".", separator)}`;
    if (value >= 10) return `×${Math.round(value)}`;
    return `×${value.toFixed(1).replace(".", separator)}`;
  }

  /* Ставит множитель рядом с каждой более медленной строкой: отрыв должен
     читаться сразу, без деления в уме. */
  function markGaps(rows, ours, msClass, gapClass, ratioClass) {
    if (!rows || !ours) return;
    const base = parseNumber(ours.querySelector(`.${msClass}`)?.textContent);
    if (!base) return;

    const section = rows.closest("section");
    const headline = parseNumber(section?.querySelector(`.${ratioClass}`)?.textContent);

    const items = [...rows.querySelectorAll(`.${msClass}`)].map((cell) => ({
      cell,
      value: parseNumber(cell.textContent) || 0,
    }));
    const max = items.reduce((acc, item) => Math.max(acc, item.value), 0);

    for (const item of items) {
      if (item.cell === ours.querySelector(`.${msClass}`)) continue;
      if (item.cell.querySelector(`.${gapClass}`)) continue;
      let ratio = item.value / base;
      if (headline && item.value === max && ratio > 1) ratio = headline;
      const label = formatGap(ratio);
      if (!label) continue;
      const mark = document.createElement("em");
      mark.className = gapClass;
      mark.textContent = label;
      item.cell.appendChild(mark);
    }
  }

  function markAllGaps() {
    const skein = document.getElementById("skein-evidence");
    if (skein) {
      const rows = skein.querySelector(".sk-rows");
      markGaps(rows, rows?.querySelector(".sk-row"), "sk-ms", "sk-gap", "sk-ratio");
    }
    const mine = document.getElementById("asbuilt-evidence");
    if (mine) {
      const rows = mine.querySelector(".ab-rows");
      markGaps(rows, rows?.querySelector(".ab-row.is-ours"), "ab-ms", "ab-gap", "ab-ratio");
    }
  }

  /* ---------- блок Skein: тот же язык, что у нас ---------- */

  /* Блок Skein принадлежит основной странице. Его формулировки задаются
     здесь, чтобы оба продукта описывали себя одинаково: сначала что это,
     потом что это даёт. Переписываются ru и en — остальные языки сохраняют
     перевод основной страницы. */
  const SKEIN_COPY = [
    {
      keys: ["Открытая инженерия / Skein", "Open engineering / Skein"],
      ru: "Оркестрация агентов / Skein",
      en: "Agent orchestration / Skein",
    },
    {
      keys: ["Проверяемо.", "Verifiable."],
      ru: "Каждая задача — узел графа.",
      en: "Every task is a node of a graph.",
    },
    {
      keys: ["В коде и в замерах.", "In code and in measurements."],
      ru: "1,58 мс на 1 000 задач.",
      en: "1.58 ms per 1,000 tasks.",
    },
    {
      keys: ["Явные графы задач агентов.", "Explicit agent task graphs."],
      prefix: true,
      ru: "Запускает рой агентов по явному графу: видно, кто кого ждёт и кто отменён. Эпоха отсекает поздние результаты, stop завершает дерево процессов. Go · MIT · Linux / macOS / Windows.",
      en: "Runs an agent swarm on an explicit graph: who waits for whom and who was cancelled is visible. The epoch cuts late results, stop terminates the process tree. Go · MIT · Linux / macOS / Windows.",
    },
    {
      keys: ["Страница всех проверок →", "All checks page →"],
      ru: "Страница продукта →",
      en: "Product page →",
    },
  ];

  /* Шкала в блоке Skein теперь логарифмическая, и примечание говорит об
     этом во всех языках, где блок переведён. */
  const SKEIN_SCALE_KEYS = [
    "Срез шкалы от 1,50 мс",
    "Scale slice from 1.50 ms",
  ];

  const SKEIN_SCALE = {
    ru: "Шкала логарифмическая по времени. ×2,2 и ×554,9 — реальные отношения.",
    be: "Лагарыфічная шкала часу. ×2,2 і ×554,9 — сапраўдныя адносіны.",
    es: "Escala logarítmica del tiempo. ×2,2 y ×554,9 son las proporciones reales.",
    en: "Logarithmic time scale. ×2.2 and ×554.9 are the real ratios.",
    zh: "时间对数刻度。×2.2 和 ×554.9 是真实比值。",
    ja: "時間の対数スケール。×2.2 と ×554.9 は実際の比率です。",
    ko: "시간 로그 스케일. ×2.2와 ×554.9는 실제 비율입니다.",
    ar: "مقياس زمني لوغاريتمي. ×2.2 و ×554.9 هما النسبان الحقيقيان.",
    hi: "समय का लघुगणकीय पैमाना। ×2.2 और ×554.9 असली अनुपात हैं।",
  };

  function patchSkeinText(root, lang) {
    if (!root) return;
    const rules = SKEIN_COPY.filter((rule) => rule[lang]);
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
    const nodes = [];
    while (walker.nextNode()) nodes.push(walker.currentNode);

    for (const node of nodes) {
      const raw = node.nodeValue;
      const key = raw.trim();
      if (!key) continue;

      if (SKEIN_SCALE_KEYS.some((start) => key.startsWith(start))) {
        const note = SKEIN_SCALE[lang];
        if (note && key !== note) node.nodeValue = raw.replace(key, note);
        continue;
      }

      for (const rule of rules) {
        const hit = rule.keys.some((candidate) => (rule.prefix ? key.startsWith(candidate) : key === candidate));
        if (!hit) continue;
        const value = rule[lang];
        if (value && key !== value) node.nodeValue = raw.replace(key, value);
        break;
      }
    }
  }

  /* Формулировки меняются в шаблоне до вставки: тогда страница переводит их
     своим же словарём, как и всё остальное. */
  function patchSkeinTemplate() {
    const template = document.getElementById("skein-evidence-template");
    if (!template) return;
    patchSkeinText(template.content, localeOf());
  }

  /* Полосы блока Skein: разброс 555×, на линейной шкале видно только самое
     медленное. Логарифмическая показывает все три, а реальные отношения
     стоят в подписях. */
  function alignSkeinBars() {
    const skein = document.getElementById("skein-evidence");
    if (!skein) return;
    const rows = [...skein.querySelectorAll(".sk-row")];
    if (!rows.length) return;

    const values = rows.map((row) => ({
      row,
      value: parseNumber(row.querySelector(".sk-ms")?.textContent) || 0,
    }));
    const max = values.reduce((acc, item) => Math.max(acc, item.value), 0);
    if (!max) return;

    for (const item of values) {
      const fill = item.row.querySelector(".sk-fill");
      if (!fill || !item.value) continue;
      fill.style.width = `${((Math.log(item.value) / Math.log(max)) * 100).toFixed(2)}%`;
    }
  }

  /* Кто это сделал: знак PARITET в углу карточки. */
  const PARITET_MARK =
    '<span class="pp-by"><svg viewBox="0 0 48 48" fill="none" aria-hidden="true">' +
    '<circle cx="24" cy="24" r="16" stroke="currentColor" stroke-width="3.2"/>' +
    '<circle cx="24" cy="8" r="3.2" fill="currentColor"/>' +
    '<circle cx="24" cy="24" r="5.4" fill="var(--signal)"/></svg>' +
    "<span>by PARITET</span></span>";

  function decorateSkeinCard() {
    const card = document.querySelector("#skein-evidence .sk-home");
    if (!card || card.querySelector(".pp-by")) return;
    card.insertAdjacentHTML("beforeend", PARITET_MARK);
  }

  /* ---------- сборка слайдера ---------- */

  function buildSlider(skein, mine) {
    if (document.getElementById("asbuilt-slider")) return true;

    const wrapper = document.createElement("div");
    wrapper.id = "asbuilt-slider";
    wrapper.className = "ab-slider";

    const bar = document.createElement("div");
    bar.className = "ab-bar";

    const tabs = document.createElement("div");
    tabs.className = "ab-tabs";
    tabs.setAttribute("role", "tablist");

    ["Skein", "ASBUILT"].forEach((label, index) => {
      const tab = document.createElement("button");
      tab.type = "button";
      tab.className = "ab-tab" + (index === 0 ? " is-active" : "");
      tab.dataset.index = String(index);
      tab.textContent = label;
      tab.setAttribute("role", "tab");
      tab.setAttribute("aria-selected", index === 0 ? "true" : "false");
      tab.addEventListener("click", () => showSlide(index));
      tabs.appendChild(tab);
    });

    tabs.addEventListener("keydown", (event) => {
      if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
      const slides = slideList();
      if (!slides.length) return;
      event.preventDefault();
      const step = event.key === "ArrowRight" ? 1 : -1;
      const next = (current + step + slides.length) % slides.length;
      showSlide(next);
      document.querySelector(`#asbuilt-slider .ab-tab[data-index="${next}"]`)?.focus();
    });

    const timer = document.createElement("div");
    timer.className = "ab-timer";
    timer.setAttribute("role", "button");
    timer.setAttribute("tabindex", "0");
    timer.setAttribute("aria-label", "Следующий слайд");
    const fill = document.createElement("span");
    fill.className = "ab-timer-fill";
    timer.appendChild(fill);

    const advance = () => showSlide(current + 1);
    timer.addEventListener("click", advance);
    timer.addEventListener("keydown", (event) => {
      if (event.key !== "Enter" && event.key !== " ") return;
      event.preventDefault();
      advance();
    });

    const slides = document.createElement("div");
    slides.className = "ab-slides";

    skein.parentNode.insertBefore(wrapper, skein);
    slides.appendChild(skein);
    slides.appendChild(mine);
    bar.appendChild(tabs);
    bar.appendChild(timer);
    wrapper.appendChild(bar);
    wrapper.appendChild(slides);

    skein.classList.add("ab-slide");
    mine.classList.add("ab-slide");

    showSlide(0);
    markAllGaps();
    return true;
  }

  /* ---------- вставка блока ---------- */

  function insertEvidence() {
    if (document.getElementById("asbuilt-evidence")) return true;
    const template = document.getElementById("asbuilt-evidence-template");
    if (!template) return false;

    /* Блок Skein ставится тем же способом и рядом. Если его ещё нет, даём
       ему секунду иначе, чтобы не оказаться под подвалом. */
    if (!document.getElementById("skein-evidence") && Date.now() - startedAt < 2500) return false;

    const anchor = findAnchor();
    if (!anchor) return false;

    const section = template.content.firstElementChild.cloneNode(true);
    translateTree(section);
    anchor.node.insertAdjacentElement(anchor.position, section);

    if ("IntersectionObserver" in window && !window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      const reveal = new IntersectionObserver((entries, instance) => {
        if (!entries.some((entry) => entry.isIntersecting)) return;
        section.classList.add("is-visible");
        instance.disconnect();
      }, { threshold: 0.2 });
      reveal.observe(section);
    } else {
      section.classList.add("is-visible");
    }

    loadStars(section);
    return true;
  }

  /* Счётчик звёзд: сначала shields.io, затем API GitHub. */
  function loadStars(section) {
    const stars = section.querySelector("#asbuilt-main-stars");
    if (!stars) return;
    const show = (value) => {
      const count = Number(String(value).replace(/[^\d]/g, ""));
      if (Number.isFinite(count) && count > 0) {
        stars.textContent = new Intl.NumberFormat(document.documentElement.lang || "ru").format(count);
      }
    };
    fetch("https://img.shields.io/github/stars/Balanov/asbuilt.json")
      .then((response) => {
        if (!response.ok) throw new Error("shields.io unavailable");
        return response.json();
      })
      .then((data) => show(data.value))
      .catch(() => {
        fetch("https://api.github.com/repos/Balanov/asbuilt", { headers: { Accept: "application/vnd.github+json" } })
          .then((response) => {
            if (!response.ok) throw new Error("GitHub statistics unavailable");
            return response.json();
          })
          .then((repository) => show(repository.stargazers_count))
          .catch(() => {
            /* Показанное значение остаётся, если GitHub недоступен. */
          });
      });
  }

  function assembleSlider() {
    const skein = document.getElementById("skein-evidence");
    const mine = document.getElementById("asbuilt-evidence");
    if (!skein || !mine) return false;
    return buildSlider(skein, mine);
  }

  const startedAt = Date.now();

  /* Формулировки Skein меняются до того, как страница соберёт блок. */
  patchSkeinTemplate();

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", insertEvidence);
  } else {
    insertEvidence();
  }

  let attempts = 0;
  const watcher = setInterval(() => {
    attempts += 1;
    patchSkeinTemplate();
    patchSkeinText(document.getElementById("skein-evidence"), localeOf());
    alignSkeinBars();
    decorateSkeinCard();
    if ((insertEvidence() && assembleSlider()) || attempts > 60) {
      clearInterval(watcher);
      markAllGaps();
      alignSkeinBars();
      decorateSkeinCard();
    }
  }, 600);
})();
