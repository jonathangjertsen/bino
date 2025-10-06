
const originals = new WeakMap();

document.addEventListener('click', e => {
  const el = e.target.closest('.editable');
  if (!el) return;

  const range = document.createRange();
  range.selectNodeContents(el);
  const action = el.dataset.action;

  const form = document.createElement('form');
  form.action = action;
  form.method = 'POST';
  form.classList.add('d-flex', 'form-control-sm', 'form-control-plaintext');

  const input = document.createElement('input');
  input.name = 'value';
  input.value = el.textContent.trim();
  input.type = 'text';
  input.classList.add('form-control', 'w-75');

  const submit = document.createElement('button');
  submit.type = 'submit';
  submit.textContent = LN.GenericUpdate;
  submit.classList.add('btn', 'btn-primary');

  form.append(input, submit);
  originals.set(form, el.cloneNode(true));
  el.replaceWith(form);
  input.focus();
});

document.addEventListener('mousedown', e => {
  const form = document.querySelector('form[action][style]');
  if (!form) return;
  if (form.contains(e.target)) return;
  const original = originals.get(form);
  if (!original) return;
  form.replaceWith(original);
});

document.addEventListener('submit', e => {
  const form = e.target.closest('form[action]');
  if (!form) return;
  originals.delete(form);
});
