import { Uppy, XHRUpload, StatusBar } from "/assets/js/libs/uppy.min.mjs"

export function setupUploadButton() {

  let client = new Uppy().use(XHRUpload, {
    endpoint: '/browser/upload',
    allowMultipleUploadBatches: true
  })

  const folderPath = document.getElementById("folder-path-meta")
  const spaceID = document.getElementById("space-id-meta")

  client.setMeta({ rootPath: folderPath.value })
  client.setMeta({ spaceID: spaceID.value })
  client.use(StatusBar, { target: '#status-bar' });

  client.on('complete', (result) => {
    htmx.trigger("body", "refreshFolder");
  });

  document.getElementById('upload-file-btn').
    addEventListener("click", (e) => {
      var input = document.createElement('input');
      input.type = 'file';
      input.multiple = true

      input.onchange = e => {
        for (const file of e.target.files) {
          client.addFile(file)
        }

        client.upload()
      }

      input.click();
    })

  document.getElementById('upload-folder-btn').
    addEventListener("click", (e) => {
      var input = document.createElement('input');
      input.type = 'file';
      input.multiple = true
      input.webkitdirectory = true
      input.mozdirectory = true
      input.directory = true

      input.onchange = e => {
        for (const file of e.target.files) {
          client.addFile({
            name: file.webkitRelativePath,
            data: file,
            type: file.type,
            source: 'Local',
            isRemote: false,
          })
        }

        client.upload()
      }

      input.click();
    })
}
