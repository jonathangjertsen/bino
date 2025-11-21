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
  document.querySelectorAll(".dashboard").forEach(setupBoard);

  document.querySelectorAll(".dashboard-patient-list").forEach(list => {
    let dragged;

    list.querySelectorAll(".card-header-patient").forEach(handle => {
      handle.addEventListener("pointerdown", e => {
        dragged = handle.closest(".card");
        dragged.setPointerCapture(e.pointerId);
        dragged.classList.add("dragging");
      });
    });

    document.addEventListener("pointermove", e => {
      if (!dragged) return;
      const lists = [...document.querySelectorAll(".dashboard-patient-list")];
      const target = lists.find(l => {
        const r = l.getBoundingClientRect();
        return e.clientX >= r.left && e.clientX <= r.right && e.clientY >= r.top && e.clientY <= r.bottom;
      });
      if (target) {
        target.classList.add("drop-target");
        const items = [...target.children];
        const before = items.find(i => e.clientY < i.getBoundingClientRect().top + i.offsetHeight / 2);
        target.classList.remove("drop-target");
        target.insertBefore(dragged, before || null);
      }
    });

    document.addEventListener("pointerup", e => {
      if (!dragged) return;
      const home = parseInt(dragged.closest(".dashboard-patient-list").dataset.home);
      reordered(home);
      dragged.classList.remove("dragging");
      dragged.releasePointerCapture(e.pointerId);
      dragged = null;
    });
  });
});

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