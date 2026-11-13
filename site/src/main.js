import "@fontsource/space-grotesk/latin-500.css";
import "@fontsource/space-grotesk/latin-700.css";
import "@fontsource/manrope/cyrillic-400.css";
import "@fontsource/manrope/cyrillic-500.css";
import "@fontsource/manrope/cyrillic-600.css";
import "@fontsource/ibm-plex-mono/latin-400.css";
import "@fontsource/ibm-plex-mono/latin-500.css";
import "@fontsource/ibm-plex-mono/cyrillic-400.css";
import "@fontsource/ibm-plex-mono/cyrillic-500.css";
import "./style.css";
import { messages, familyText } from "./i18n.js";
import { auditLines, families } from "./data.js";
import { benchRows, benchNote, benchHeadline } from "./bench.js";

// Mark the document as scripted before the first paint of the styles that
// depend on it, so content never stays hidden when the script fails.
document.documentElement.classList.add("js");

const stored = localStorage.getItem("asbuilt-language");
const browserLanguages = navigator.languages || [navigator.language];
let language = stored || (browserLanguages.some((item) => item.toLowerCase().startsWith("ru")) ? "ru" : "en");

const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

/* ---------- theme ---------- */
const themeButtons = document.querySelectorAll("[data-theme-set]");

function applyTheme(name) {
  document.documentElement.dataset.theme = name;
  localStorage.setItem("asbuilt-theme", name);
  themeButtons.forEach((button) => {
    const active = button.dataset.themeSet === name;
    button.classList.toggle("is-active", active);
    button.setAttribute("aria-pressed", String(active));
  });
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) meta.setAttribute("content", name === "draft" ? "#f2f4f6" : "#0d1117");
}

const storedTheme = localStorage.getItem("asbuilt-theme");
applyTheme(storedTheme || document.documentElement.dataset.theme || "night");

themeButtons.forEach((button) => {
  button.addEventListener("click", () => applyTheme(button.dataset.themeSet));
});

/* ---------- language ---------- */
function text(key) {
  return messages[language][key] ?? key;
}

function renderLanguage() {
  document.documentElement.lang = language;
  document.querySelectorAll("[data-i18n]").forEach((element) => {
    const key = element.dataset.i18n;
    if (messages[language][key]) element.textContent = messages[language][key];
  });
  document.querySelectorAll("[data-lang]").forEach((button) => {
    const active = button.dataset.lang === language;
    button.classList.toggle("is-active", active);
    button.setAttribute("aria-pressed", String(active));
  });
  renderFamilies();
  renderBenchmark();
}

document.querySelectorAll("[data-lang]").forEach((button) => {
  button.addEventListener("click", () => {
    language = button.dataset.lang;
    localStorage.setItem("asbuilt-language", language);
    renderLanguage();
  });
});

/* ---------- families ---------- */
function renderFamilies() {
  const host = document.querySelector("#families");
  if (!host) return;
  const copy = familyText[language];
  host.innerHTML = families
    .map((family) => {
      const rules = copy[family.id]?.rules ?? [];
      const sub = copy[family.id]?.sub ? `<p class="family-sub">${copy[family.id].sub}</p>` : "";
      return `<article class="family reveal">
        <div class="family-head"><span class="family-name">${copy[family.id]?.name ?? family.id}</span><span class="family-count">${family.count}</span></div>
        ${sub}
        <ul class="family-list">${rules.map((rule) => `<li>${rule}</li>`).join("")}</ul>
      </article>`;
    })
    .join("");
  observeReveals(host);
}

/* ---------- benchmark table ---------- */
// The scale of the little bars: the slowest tool that was measured here.
const BENCH_SCALE = 136.69;

function renderBenchmark() {
  const body = document.querySelector("#bench-body");
  if (!body) return;
  const badge = (value) => {
    if (value === "yes") return `<span class="badge is-yes">${text("bench.yes")}</span>`;
    if (value === "no") return `<span class="badge is-no">${text("bench.no")}</span>`;
    return `<span class="badge is-part">${text("bench.part")}</span>`;
  };
  body.innerHTML = benchRows
    .map((row) => {
      const loss = row.loss ? `<span class="cell-loss">${row.loss}</span>` : `<span class="cell-loss is-none">${text("bench.none")}</span>`;
      const verdict = row.verdict === "full" ? text("bench.full") : row.verdict === "compare" ? text("bench.compare") : text("bench.collect");
      const bar = row.ms ? `<span class="cell-bar"><i style="width:${((row.ms / BENCH_SCALE) * 100).toFixed(2)}%"></i></span>` : "";
      return `<tr class="${row.asbuilt ? "is-asbuilt" : ""}">
        <td class="tool">${row.tool}${row.note ? `<small>${row.note}</small>` : ""}</td>
        <td class="cell-time">${row.time}${bar}</td>
        <td>${badge(row.root)}</td>
        <td>${verdict}</td>
        <td>${loss}</td>
        <td>${row.license}</td>
      </tr>`;
    })
    .join("");

  const note = document.querySelector("#bench-note");
  if (note) note.innerHTML = benchNote[language];

  const headline = benchHeadline[language];
  const set = (id, value) => {
    const node = document.querySelector(id);
    if (node && value) node.textContent = value;
  };
  set("#bench-ratio", headline.ratio);
  set("#bench-ratio-label", headline.ratioLabel);
  set("#bench-ratio-copy", headline.ratioCopy);
  set("#bench-uniq", headline.uniq);
  set("#bench-uniq-label", headline.uniqLabel);
  set("#bench-uniq-copy", headline.uniqCopy);
}

/* ---------- terminal demo ---------- */
const terminal = document.querySelector("#terminal");

function paintLine(line) {
  const span = document.createElement("span");
  span.className = line.cls;
  span.textContent = line.text;
  return span;
}

function runDemo() {
  if (!terminal) return;
  terminal.innerHTML = "";
  let index = 0;

  const step = () => {
    if (index >= auditLines.length) {
      window.setTimeout(() => {
        index = 0;
        terminal.innerHTML = "";
        step();
      }, 2600);
      return;
    }
    const line = auditLines[index];
    const node = paintLine(line);
    terminal.appendChild(node);
    terminal.appendChild(document.createTextNode("\n"));
    if (line.text.startsWith("CRITICAL") || line.text.startsWith("WARN")) {
      node.classList.add("t-signal");
    }
    index += 1;
    const delay = line.text.startsWith("$") ? 260 : line.text === "" ? 90 : 46;
    window.setTimeout(step, reduced ? 0 : delay);
  };

  step();
}

if (reduced) {
  if (terminal) {
    terminal.innerHTML = auditLines.map((line) => `<span class="${line.cls}">${line.text}</span>`).join("\n");
  }
} else {
  runDemo();
}

/* ---------- reveal ---------- */
let observer = null;

function observeReveals(root = document) {
  const targets = root.querySelectorAll(".reveal:not(.is-visible)");
  if (!targets.length) return;
  if (!("IntersectionObserver" in window) || reduced) {
    targets.forEach((target) => target.classList.add("is-visible"));
    return;
  }
  observer = observer ?? new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (!entry.isIntersecting) return;
      entry.target.classList.add("is-visible");
      observer.unobserve(entry.target);
    });
  }, { threshold: 0.15 });
  targets.forEach((target) => observer.observe(target));
}

document.querySelectorAll(".section-head, .linkcheck, .table-wrap, .bench-headline, .usage-grid, .install-line").forEach((element) => {
  element.classList.add("reveal");
});

/* ---------- link bars ---------- */
const linkRows = document.querySelectorAll(".link-row .link-fill");
function fillBars() {
  linkRows.forEach((fill, index) => {
    window.setTimeout(() => {
      fill.style.width = fill.style.getPropertyValue("--w");
    }, reduced ? 0 : 120 + index * 110);
  });
}

if ("IntersectionObserver" in window && !reduced) {
  const barObserver = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (!entry.isIntersecting) return;
      fillBars();
      barObserver.disconnect();
    });
  }, { threshold: 0.3 });
  const track = document.querySelector(".link-rows");
  if (track) barObserver.observe(track);
} else {
  fillBars();
}

/* ---------- github stats ---------- */
function showStars(target, value) {
  const count = Number(String(value).replace(/[^\d]/g, ""));
  if (!Number.isFinite(count) || !target) return;
  target.textContent = new Intl.NumberFormat(language === "ru" ? "ru-RU" : "en-US").format(count);
}

fetch("https://img.shields.io/github/stars/Balanov/asbuilt.json")
  .then((response) => {
    if (!response.ok) throw new Error("shields.io unavailable");
    return response.json();
  })
  .then((data) => {
    showStars(document.querySelector("#gh-stars"), data.value);
    showStars(document.querySelector("#gh-forks"), data.value);
  })
  .catch(() => {
    fetch("https://api.github.com/repos/Balanov/asbuilt", { headers: { Accept: "application/vnd.github+json" } })
      .then((response) => {
        if (!response.ok) throw new Error("GitHub unavailable");
        return response.json();
      })
      .then((repository) => {
        showStars(document.querySelector("#gh-stars"), repository.stargazers_count);
        showStars(document.querySelector("#gh-forks"), repository.forks_count);
      })
      .catch(() => {
        /* the placeholder stays */
      });
  });

renderLanguage();
observeReveals();
