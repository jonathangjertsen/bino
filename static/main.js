document.getElementById("language-select").addEventListener('change', (event) => {
  event.target.form.submit();
});

document.querySelector('.closer').addEventListener('click', function() {
  this.parentElement.style.display = 'none';
});

$(function () {
  $('[data-toggle="tooltip"]').tooltip()
})
