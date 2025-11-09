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

document.querySelectorAll(".dashboard").forEach(setupBoard);

document.getElementById("dashboard-tags-show-all").addEventListener("click", ev => {
  document.querySelectorAll(".dashboard-tag").forEach(e => {
    e.style.display = "inline-block";
  });
  ev.target.style.display = "none";
})

$(function() {
  $(".dashboard-patient-list").sortable({
    connectWith: ".dashboard-patient-list",
    cancel: "a,button",
    handle: ".card-header-patient",
    forcePlaceholderSize: true,
    forceHelperSize: true,
    zIndex: 9999,
    appendTo: document.body,
    update: function(e, ui) {
      if (TRANSFERSTARTED) {
        return;
      }
      const sender = parseInt(e.target.dataset["home"]);
      const receiver = ui.item.parent().data("home");
      if (sender == receiver) {
        reordered(sender);
      } else {
        patientTransferred(ui.item.data("patient-id"), sender, receiver);
      }
    },
    over: function(e, ui) {
      e.target.classList.add("drop-target");
    },
    out: function(e, ui) {
      e.target.classList.remove("drop-target");
    }
  }).disableSelection();
});

function reordered(home) {
  var req = {
    Id: home,
    Order: [],
  }

  $(".dashboard-patient-list").each(function(){
    const id = $(this).data("home");
    if (id == home) {
      req.Order = $(this).children().map(function(){ return $(this).data("patient-id"); }).get();
    }
  });

  fetch("/ajaxreorder", {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
      body: JSON.stringify(req),
  });
}

var TRANSFERSTARTED = false;

function patientTransferred(patient, sender, receiver) {
  TRANSFERSTARTED = true;
  $( ".dashboard-patient-list" ).sortable({
    disabled: true
  });

  var req = {
    Sender: {
      ID: sender,
      Order: [],
    },
    Receiver: {
      ID: receiver,
      Order: [],
    },
    Patient: patient,
  }
  $(".dashboard-patient-list").each(function(){
    const id = $(this).data("home");
    if (id == receiver) {
      req.Receiver.Order = $(this).children().map(function(){ return $(this).data("patient-id"); }).get();
    }
    if (id == sender) {
      req.Sender.Order = $(this).children().map(function(){ return $(this).data("patient-id"); }).get();
    }
  });

  fetch("/ajaxtransfer", {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify(req),
  }).then(() => location.reload());
}
