document.getElementById("language-select").addEventListener('change', (event) => {
    event.target.form.submit();
});

$(function () {
  $('[data-toggle="tooltip"]').tooltip()
})
