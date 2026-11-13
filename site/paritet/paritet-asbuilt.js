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
    if (value >= 100) return `×${value.toFixed(1).replace(".", ",")}`;
    if (value >= 10) return `×${Math.round(value)}`;
    return `×${value.toFixed(1).replace(".", ",")}`;
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

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", insertEvidence);
  } else {
    insertEvidence();
  }

  let attempts = 0;
  const watcher = setInterval(() => {
    attempts += 1;
    if ((insertEvidence() && assembleSlider()) || attempts > 60) {
      clearInterval(watcher);
      markAllGaps();
    }
  }, 600);
})();
