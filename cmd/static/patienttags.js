
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
