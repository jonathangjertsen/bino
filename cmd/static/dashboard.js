// Click-and-drag behaviour similar to Kanban and diagramming software
const captureScroll = (elem) => {
  let dragging = false, lastX = 0;

  const isInteractive = (el) =>
    el.closest('a,button,input,select,textarea,label,summary,[contenteditable],form,.editable,::before,::after,.dashboard-patient-card');

  const down = (ev) => {
    // Don't fuck with right click
    if (ev.button !== 0) {
      return;
    }
    
    // Don't fuck with interactive elements
    if (isInteractive(ev.target)) {
      return;
    }

    // Begin dragging
    dragging = true;
    lastX = ev.clientX;
    elem.setPointerCapture(ev.pointerId);
  };

  const move = (ev) => {
    if (!dragging || !elem.hasPointerCapture(ev.pointerId)) {
      return;
    }

    // Manually move the scroll-x-position
    const dx = ev.clientX - lastX;
    lastX = ev.clientX;
    elem.scrollLeft -= dx;
    
    // Prevents text selection
    ev.preventDefault();
  };

  const up = (ev) => {
    // Stop dragging
    dragging = false;
    elem.releasePointerCapture(ev.pointerId);
  };

  elem.addEventListener("mousedown", down);
  elem.addEventListener("mousemove", move, { passive: false });
  elem.addEventListener("mouseup", up);
  elem.addEventListener("mousecancel", up);
};

// Remember position on the board after a form is submitted
const setupBoard = (elem) => {
  captureScroll(elem);
  const K = "board-scroll-left";
  const F = "board-restore-once";

  const restore = () => {
    if (sessionStorage.getItem(F) === "1") {
      const v = parseInt(sessionStorage.getItem(K) || "0", 10);
      if (!Number.isNaN(v)) elem.scrollLeft = v;
    }
    sessionStorage.removeItem(F);
    sessionStorage.removeItem(K);
  };

  if (document.readyState === "loading") {
    window.addEventListener("DOMContentLoaded", restore, { once: true });
  } else {
    restore();
  }

  document.addEventListener("submit", (e) => {
    if (e.target instanceof HTMLFormElement) {
      sessionStorage.setItem(F, "1");
      sessionStorage.setItem(K, String(elem.scrollLeft));
    }
  }, true);
};

document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".dashboard-patient-list").forEach(initList);
});

function initList(list) {
  let handle, original, ghost, offsetX, offsetY;

  list.querySelectorAll(".card-header-patient").forEach(h => {
    h.addEventListener("pointerdown", e => {
      original = h.closest(".card");
      const r = original.getBoundingClientRect();
      offsetX = e.clientX - r.left;
      offsetY = e.clientY - r.top;

      ghost = original.cloneNode(true);
      ghost.style.position = "fixed";
      ghost.style.left = r.left + "px";
      ghost.style.top = r.top + "px";
      ghost.style.width = r.width + "px";
      ghost.style.pointerEvents = "none";
      ghost.style.opacity = "0.9";
      ghost.classList.add("drag-ghost");
      document.body.appendChild(ghost);

      original.classList.add("drag-origin");

      original.setPointerCapture(e.pointerId);
    });
  });

  document.addEventListener("pointermove", e => {
    if (!ghost) return;

    ghost.style.left = e.clientX - offsetX + "px";
    ghost.style.top = e.clientY - offsetY + "px";

    const lists = [...document.querySelectorAll(".dashboard-patient-list")];
    const overList = lists.find(l => {
      const r = l.getBoundingClientRect();
      return e.clientX >= r.left && e.clientX <= r.right && e.clientY >= r.top && e.clientY <= r.bottom;
    });
    if (!overList) return;

    const items = [...overList.children].filter(i => i !== original);
    const before = items.find(i => e.clientY < i.getBoundingClientRect().top + i.offsetHeight / 2);
    overList.insertBefore(original, before || null);
  });

  document.addEventListener("pointerup", e => {
    if (!ghost || !original) return;

    const home = parseInt(original.closest(".dashboard-patient-list").dataset.home);
    ghost.remove();
    original.classList.remove("drag-origin");
    ghost = null;
    reordered(home);
    original = null;
  });
}

function reordered(home) {
  const list = document.querySelector(`.dashboard-patient-list[data-home='${home}']`);
  const order = [...list.children].map(c => c.dataset.patientId);
  const req = { Id: home, Order: order };

  fetch("/ajaxreorder", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req)
  });
}

