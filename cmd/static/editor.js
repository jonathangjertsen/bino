
document.body.setAttribute('spellcheck', false);
document.body.setAttribute('autocomplete', false);
document.body.setAttribute('autocorrect', false);
const editor = new EditorJS({
    holder: "editorjs",
    placeholder: "Write something...",
    readOnly: true,
    tools: { 
        header: {
            class: Header,
        },
        list: {
            class: EditorjsList,
        },
        image: {
            class: SimpleImage,
        }
    },
});
document.getElementById('editor-edit-toggler').addEventListener("click", () => {
    editor.readOnly.toggle();
})
