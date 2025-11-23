FilePond.registerPlugin(
    FilePondPluginFileValidateSize,
    FilePondPluginImageExifOrientation,
    FilePondPluginImageValidateSize,
    FilePondPluginImageCrop,
    FilePondPluginImageEdit,
    FilePondPluginImageTransform,
    FilePondPluginImagePreview,
);

const fileInput = document.getElementById('general-file-uploader');
const fileSubmit = document.getElementById('general-file-submit');
FilePond.create(fileInput, {
    server: '/file/filepond',
    instantUpload: true,
    maxFileSize: '50MB',
    onaddfilestart: _file => {
        fileSubmit.disabled = true;
        fileSubmit.textContent = LN.FilesPleaseWait;
    },
    onprocessfiles: (err, _file) => {
        fileSubmit.disabled = false;
        fileSubmit.textContent = LN.GenericSave;
    }
});
