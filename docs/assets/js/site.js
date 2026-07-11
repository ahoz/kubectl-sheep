document.querySelectorAll("[data-copy]").forEach((btn) => {
  btn.addEventListener("click", async () => {
    const target = document.querySelector(btn.dataset.copy);
    if (!target) return;
    const text = target.textContent.trim();
    try {
      await navigator.clipboard.writeText(text);
      const original = btn.textContent;
      btn.textContent = "Copied!";
      setTimeout(() => {
        btn.textContent = original;
      }, 1500);
    } catch {
      btn.textContent = "Failed";
    }
  });
});

document.querySelectorAll(".install-tab").forEach((tab) => {
  tab.addEventListener("click", () => {
    const panel = tab.closest(".install-panel");
    if (!panel) return;
    const target = tab.dataset.panel;
    panel.querySelectorAll(".install-tab").forEach((t) => t.classList.remove("active"));
    panel.querySelectorAll(".install-content").forEach((c) => c.hidden = true);
    tab.classList.add("active");
    const content = panel.querySelector(`[data-panel="${target}"]`);
    if (content) content.hidden = false;
  });
});
