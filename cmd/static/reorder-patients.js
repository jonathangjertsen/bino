
document.addEventListener("DOMContentLoaded", function() {
  document.querySelectorAll(".dashboard-patient-list").forEach(function(list) {
    new Sortable(list, {
      handle: ".card-gripper",
      animation: 0,
      forceFallback: true,
      fallbackClass: "sortable-fallback",
      ghostClass: "drop-target",
      chosenClass: "chosen",
      dragClass: "dragging",
      filter: "a,button",
      onUpdate: function(evt) {
        reordered(parseInt(evt.to.dataset.home));
      }
    });
  });
});

function reordered(home) {
  var req = {
    Id: home,
    Order: []
  };

  document.querySelectorAll(".dashboard-patient-list").forEach(function(list) {
    if (parseInt(list.dataset.home) === home) {
      req.Order = Array.from(list.children).map(function(item) {
        return parseInt(item.dataset.patientId);
      });
    }
  });

  fetch("/ajaxreorder", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(req)
  });
}
