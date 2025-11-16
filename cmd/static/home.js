$(function() {
  $("#species-list").sortable({
    items: "tr",
    handle: "td",
    forcePlaceholderSize: true,
    forceHelperSize: true,
    placeholder: "species-placeholder",
    helper: "species-placeholder",
    helper: "clone",
    update: function(e, ui) {
      const home = parseInt(e.target.dataset["home"]);
        var req = {
          ID: home,
          Order: [],
        }

        const id = $(this).data("home");
        req.Order = $(this).children().map(function(){ return $(this).data("id"); }).get();

        fetch(`/home/${home}/species/reorder`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req),
        }).then(() => location.reload());
    }
  });
});
