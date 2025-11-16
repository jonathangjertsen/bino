setupCreateButton((formFields) => {
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
    return { url: "/species", req: req }
});

setupUpdateButtons((id, formFields) => {
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
    return { url: "/species", req: req }
});

