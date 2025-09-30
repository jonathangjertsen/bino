// Submit a new species
document.getElementById("new-species-submit").addEventListener('click', (event) => {
    const formFields = event.target.parentElement.parentElement.getElementsByClassName("form-control");
    let req = {
        Latin: "",
        Languages: {},
    }
    for (const field of formFields) {
        if (field.dataset["field"] == "latin") {
            req.Latin = field.value;
        } else if (field.dataset["field"] == "lang") {
            req.Languages[field.dataset["lang"]] = field.value;
        }
    }
    fetch("/species", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(req),
    }).then(() => location.reload());
});

// Update an existing species
document.addEventListener('click', event => {
if (!event.target.classList.contains('update-species-submit')) return;
    const id = parseInt(event.target.dataset["id"]);

    const formFields = event.target.parentElement.parentElement.getElementsByClassName("form-control");
    let req = {
        ID: id,
        Latin: "",
        Languages: {},
    }
    for (const field of formFields) {
        if (field.dataset["field"] == "latin") {
            req.Latin = field.value;
        } else if (field.dataset["field"] == "lang") {
            req.Languages[field.dataset["lang"]] = field.value;
        }
    }
    console.log(req);
    fetch("/species", {
        method: "PUT",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(req),
    }).then(() => location.reload());
});
