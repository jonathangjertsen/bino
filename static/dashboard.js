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

// tag deleter
document.addEventListener("click", e => {
  const btn = e.target.closest('button[type="submit"]');
  if (!btn) return;
  const form = btn.closest("form.tag-delete");
  if (!form) return;
  const td = btn.closest("td");
  e.preventDefault();
  fetch(form.action, { method: "DELETE", credentials: "same-origin" })
    .then(r => {
      if (!r.ok) throw new Error();
      td.remove();
    })
});

// tag create
document.addEventListener("submit", e => {
  const form = e.target.closest("td > form.tag-create");
  if (!form || !form.querySelector("select.form-select")) return;
  e.preventDefault();

  const submitter = e.submitter || form.querySelector('button[type="submit"]');
  const patientId = submitter?.dataset.patientId;
  const tagId = form.querySelector("select.form-select").value;
  if (!patientId || !tagId) {
    console.error("Missing patientId or tagId");
    return;
  }

  const tdForm = form.closest("td");
  const tr = tdForm?.parentElement;
  if (!tr || !tr.parentElement) {
    console.error("Form not inside a table row");
    return;
  }

  fetch(`/patient/${encodeURIComponent(patientId)}/tag/${encodeURIComponent(tagId)}`, {
    method: "POST",
    credentials: "same-origin"
  })
    .then(r => {
      if (!r.ok) throw new Error(`HTTP ${r.status}`);
      return r.text();
    })
    .then(html => {
      const newTr = document.createElement("tr");
      newTr.innerHTML = html.trim();
      const newTd = newTr.querySelector("td");
      if (!newTd) {
        console.error("Response did not contain a TD element");
        return;
      }
      tr.parentElement.insertBefore(newTr, tr);
      form.reset();
    })
    .catch(err => {
      console.error("Failed to create tag:", err);
    });
});


document.querySelectorAll(".dashboard").forEach(setupBoard);
