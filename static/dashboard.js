// Click-and-drag behaviour similar to Kanban and diagramming software
const captureScroll = (elem) => {
  let dragging = false, lastX = 0;

  const isInteractive = (el) =>
    el.closest('a,button,input,select,textarea,label,summary,[contenteditable],form');

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

document.querySelectorAll(".dashboard").forEach(setupBoard);
