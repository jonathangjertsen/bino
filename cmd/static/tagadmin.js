setupCreateButton((formFields) => {
    let req = {
        DefaultShow: false,
        Languages: {},
    }
    for (const field of formFields) {
        if (field.dataset["field"] == "default-show") {
            req.DefaultShow = field.checked;
        } else if (field.dataset["field"] == "lang") {
            req.Languages[field.dataset["lang"]] = field.value;
        }
    }
    return { url: "/tag", req: req }
});

setupUpdateButtons((id, formFields) => {
    let req = {
        ID: id,
        DefaultShow: false,
        Languages: {},
    }
    for (const field of formFields) {
        if (field.dataset["field"] == "default-show") {
            req.DefaultShow = field.checked
        } else if (field.dataset["field"] == "lang") {
            req.Languages[field.dataset["lang"]] = field.value;
        }
    }
    return { url: "/tag", req: req }
});
