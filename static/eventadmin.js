setupCreateButton((formFields) => {
    let req = {
        Languages: {},
    }
    for (const field of formFields) {
        if (field.dataset["field"] == "lang") {
            req.Languages[field.dataset["lang"]] = field.value;
        }
    }
    return { url: "/event", req: req }
});

setupUpdateButtons((id, formFields) => {
    let req = {
        ID: id,
        Languages: {},
    }
    for (const field of formFields) {
        if (field.dataset["field"] == "lang") {
            req.Languages[field.dataset["lang"]] = field.value;
        }
    }
    return { url: "/event", req: req }
});
