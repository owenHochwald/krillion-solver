const $ = (id) => document.getElementById(id);
const state = { result: null, selected: 0, picks: [], loading: false };

function clean(value) {
  return String(value ?? "").replace(/[\u0000-\u001f\u007f-\u009f]/g, " ").trim();
}

function notice(message, error = false) {
  const el = $("notice");
  el.textContent = message;
  el.classList.toggle("error", error);
  el.hidden = !message;
}

function validate(data) {
  if (!data || !Array.isArray(data.prompts) || data.prompts.length !== 7 || typeof data.date !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(data.date)) throw new Error("The daily data is incomplete.");
  for (const round of data.prompts) {
    if (typeof round.text !== "string" || !round.text.trim() || !Array.isArray(round.winners) || !round.winners.length || round.winners.some((answer) => typeof answer.answer !== "string" || !answer.answer.trim() || !Number.isFinite(answer.score))) throw new Error("The answer data is incomplete.");
  }
  return data;
}

function plainText() {
  const data = state.result;
  const intro = data.demo ? "DEMO — illustrative answers, not today's puzzle" : `KRILLION · Dive #${data.dayNumber} · ${data.date}`;
  return [intro, ...data.prompts.map((round, index) => `${index + 1}. ${clean(round.text)}\n${round.fallback ? "   No One in a Krillion answer — best available:\n" : ""}${round.winners.map((answer) => `   ${clean(answer.answer)}`).join("\n")}`)].join("\n\n");
}

async function copy(text) {
  try {
    await navigator.clipboard.writeText(text);
    notice("Copied to clipboard.");
  } catch {
    notice("Copy failed. Your browser may require a secure connection.", true);
  }
}

function renderList() {
  const list = $("round-list");
  list.replaceChildren();
  state.result.prompts.forEach((round, index) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "round-item" + (index === state.selected ? " active" : "");
    button.setAttribute("aria-current", index === state.selected ? "true" : "false");
    const number = document.createElement("span");
    number.className = "round-number";
    number.textContent = String(index + 1).padStart(2, "0");
    const description = document.createElement("span");
    const answer = document.createElement("strong");
    answer.textContent = clean(round.winners[0].answer) + (round.winners.length > 1 ? ` +${round.winners.length - 1}` : "");
    const prompt = document.createElement("small");
    prompt.textContent = clean(round.text);
    description.append(answer, prompt);
    const arrow = document.createElement("span");
    arrow.className = "round-arrow";
    arrow.setAttribute("aria-hidden", "true");
    arrow.textContent = "›";
    button.append(number, description, arrow);
    button.setAttribute("aria-label", `Round ${index + 1}: ${clean(round.text)}`);
    button.addEventListener("click", () => { state.selected = index; render(); });
    list.append(button);
  });
}

function element(tag, className, content) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  node.textContent = content;
  return node;
}

function renderDetail() {
  const round = state.result.prompts[state.selected];
  const pick = state.picks[state.selected] || 0;
  const answer = round.winners[pick];
  const panel = $("answer-panel");
  panel.replaceChildren();
  const top = element("div", "detail-top", "");
  top.append(element("p", "eyebrow", `ROUND ${String(state.selected + 1).padStart(2, "0")} / 07`));
  panel.append(top, element("p", "prompt", clean(round.text)), element("div", "separator", ""));
  panel.append(element("p", "answer-label", round.fallback ? "◆ BEST AVAILABLE" : "◆ ONE IN A KRILLION"));
  panel.append(element("h3", "answer-name", clean(answer.answer)));
  if (answer.quip) panel.append(element("p", "quip", clean(answer.quip)));
  if (round.fallback) panel.append(element("p", "fallback", "No One in a Krillion answer was available for this round."));
  const actions = element("div", "answer-actions", "");
  const copyButton = element("button", "button", "Copy answer");
  copyButton.type = "button";
  copyButton.addEventListener("click", () => copy(clean(answer.answer)));
  actions.append(copyButton);
  if (round.winners.length > 1) {
    const alternate = element("div", "alternate", "");
    const previous = element("button", "", "←");
    const next = element("button", "", "→");
    previous.type = next.type = "button";
    previous.setAttribute("aria-label", "Previous equal answer");
    next.setAttribute("aria-label", "Next equal answer");
    previous.addEventListener("click", () => changePick(-1));
    next.addEventListener("click", () => changePick(1));
    alternate.append(previous, element("span", "", `${pick + 1} of ${round.winners.length} equal picks`), next);
    actions.append(alternate);
  }
  panel.append(actions);
  panel.append(element("p", "detail-note", `Selected from ${Number(round.accepted || 0).toLocaleString()} accepted answers.`));
}

function changePick(direction) {
  const count = state.result.prompts[state.selected].winners.length;
  state.picks[state.selected] = (state.picks[state.selected] + count + direction) % count;
  renderDetail();
}

function updateReset() {
  const data = state.result;
  if (!data || data.demo || !data.resetsAt) { $("reset-time").textContent = ""; return; }
  const remaining = new Date(data.resetsAt).getTime() - Date.now();
  if (!Number.isFinite(remaining)) { $("reset-time").textContent = ""; return; }
  $("reset-time").textContent = remaining <= 0 ? "A new dive may be available · refresh" : `Resets in ${String(Math.floor(remaining / 3600000)).padStart(2, "0")}h ${String(Math.floor(remaining / 60000) % 60).padStart(2, "0")}m`;
}

function render() {
  const data = state.result;
  $("dive-title").textContent = data.demo ? "Illustrative dive" : `DIVE ${String(data.dayNumber).padStart(3, "0")}`;
  $("dive-meta").textContent = data.demo ? "Demo only · these are not today's answers" : `${data.date} · ${data.prompts.reduce((sum, round) => sum + round.winners.length, 0)} top answers`;
  $("round-count").textContent = `${String(state.selected + 1).padStart(2, "0")} / 07`;
  $("copy-all").disabled = false;
  renderList();
  renderDetail();
  updateReset();
}

async function load() {
  if (state.loading) return;
  state.loading = true;
  $("refresh").disabled = true;
  notice("");
  try {
    const response = await fetch(`./data.json?v=${Date.now()}`, { cache: "no-store" });
    if (!response.ok) throw new Error(response.status === 404 ? "Today's answers have not been published yet. Try refreshing shortly." : `Could not load answers (HTTP ${response.status}).`);
    const data = validate(await response.json());
    state.result = data;
    state.selected = 0;
    state.picks = data.prompts.map(() => 0);
    render();
    if (data.demo) notice("Demo preview: these answers are illustrative.");
  } catch (error) {
    notice(error.message || "Could not load today's answers.", true);
    if (!state.result) {
      $("dive-title").textContent = "The current got away.";
      $("dive-meta").textContent = "Use refresh to try again.";
      $("answer-panel").replaceChildren(element("div", "empty-state", "Today's answers are unavailable right now."));
    }
  } finally {
    state.loading = false;
    $("refresh").disabled = false;
  }
}

$("refresh").addEventListener("click", load);
$("copy-all").addEventListener("click", () => { if (state.result) copy(plainText()); });
document.addEventListener("keydown", (event) => {
  if (!state.result || /^(INPUT|TEXTAREA|BUTTON)$/.test(document.activeElement?.tagName || "")) return;
  if (event.key >= "1" && event.key <= "7") { state.selected = Number(event.key) - 1; render(); }
  if (event.key === "ArrowLeft" || event.key === "ArrowRight") changePick(event.key === "ArrowLeft" ? -1 : 1);
});
setInterval(updateReset, 60000);
load();
