function setupSearchForm(target) {
    let isAdvancedChecked = false; 
    target.querySelectorAll("[name='mode']").forEach(elem => {
        if (elem.checked && elem.value === "advanced") {
            isAdvancedChecked = true;
        }
    });
    if (isAdvancedChecked) {
        target.querySelector(".search-advanced-options").dataset.show = "true";
    } else {
        target.querySelector(".search-advanced-options").dataset.show = "false";
    }
}
document.body.addEventListener('htmx:beforeSend', function(evt) {
    const target = evt.target;
    setupSearchForm(target);

    const params = new URLSearchParams(new FormData(target))
    const url = location.pathname + '?' + params.toString()
    history.replaceState(null, '', url)
});

let searchForm = document.getElementById("search-form");
setupSearchForm(searchForm);
searchForm.querySelector(".search-filter-clear-created").addEventListener("click", e => {
    searchForm.querySelectorAll(".search-created-filter").forEach(el => {
        el.value = "";
    });
    htmx.trigger(searchForm, "submit");
}, { capture: true });
