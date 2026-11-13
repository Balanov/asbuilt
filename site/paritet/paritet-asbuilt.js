(() => {
  "use strict";

  const T = (s) => (window.__t ? window.__t(s) : s);

  const SLIDE_MS = 6500;

  let observer;
  let timer = null;
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

  function showSlide(index) {
    const slides = document.querySelectorAll("#asbuilt-slider .ab-slides > *");
    if (!slides.length) return;
    current = (index + slides.length) % slides.length;
    slides.forEach((slide, position) => slide.classList.toggle("is-active", position === current));
    document.querySelectorAll("#asbuilt-slider .ab-tab").forEach((tab) => {
      tab.classList.toggle("is-active", Number(tab.dataset.index) === current);
    });
  }

  function startTimer() {
    if (timer) clearInterval(timer);
    timer = setInterval(() => showSlide(current + 1), SLIDE_MS);
  }

  /* Один слайдер на два блока: сначала Skein, через несколько секунд
     asbuilt. Переключение автоматическое, вручную — вкладками. */
  function buildSlider(skein, mine) {
    if (document.getElementById("asbuilt-slider")) return true;

    const wrapper = document.createElement("div");
    wrapper.id = "asbuilt-slider";
    wrapper.className = "ab-slider";

    const tabs = document.createElement("div");
    tabs.className = "ab-tabs";
    ["Skein", "asbuilt"].forEach((label, index) => {
      const tab = document.createElement("button");
      tab.type = "button";
      tab.className = "ab-tab" + (index === 0 ? " is-active" : "");
      tab.dataset.index = String(index);
      tab.textContent = label;
      tab.addEventListener("click", () => {
        showSlide(index);
        startTimer();
      });
      tabs.appendChild(tab);
    });

    const slides = document.createElement("div");
    slides.className = "ab-slides";

    skein.parentNode.insertBefore(wrapper, skein);
    slides.appendChild(skein);
    slides.appendChild(mine);
    wrapper.appendChild(tabs);
    wrapper.appendChild(slides);

    showSlide(0);
    startTimer();
    return true;
  }

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
    observer?.disconnect();

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

    const stars = section.querySelector("#asbuilt-main-stars");
    fetch("https://api.github.com/repos/Balanov/asbuilt", { headers: { Accept: "application/vnd.github+json" } })
      .then((response) => {
        if (!response.ok) throw new Error("GitHub statistics unavailable");
        return response.json();
      })
      .then((repository) => {
        if (!stars || !Number.isFinite(repository.stargazers_count)) return;
        stars.textContent = new Intl.NumberFormat(document.documentElement.lang || "ru").format(repository.stargazers_count);
      })
      .catch(() => {
        // The recorded count stays visible if GitHub is unavailable.
      });

    return true;
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
  const timer2 = setInterval(() => {
    attempts += 1;
    if ((insertEvidence() && assembleSlider()) || attempts > 60) clearInterval(timer2);
  }, 600);
})();
