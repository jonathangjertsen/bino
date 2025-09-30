function setupCreateButton(listener) {
    document.getElementById("new-submit").addEventListener('click', (event) => {
        const formFields = event.target.parentElement.parentElement.getElementsByClassName("form-control");
        const { url, req } = listener(formFields);
        fetch(url, {
            method: "POSt",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req),
        }).then(() => location.reload());
    });
}

function setupUpdateButtons(listener) {
    document.addEventListener('click', event => {
        if (!event.target.classList.contains('update-submit')) return;
        const id = parseInt(event.target.dataset["id"]);
        const formFields = event.target.parentElement.parentElement.getElementsByClassName("form-control");
        const { url, req } = listener(id, formFields);
        fetch(url, {
            method: "PUT",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(req),
        }).then(() => location.reload());
    });
}
