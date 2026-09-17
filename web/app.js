const input = document.querySelector('#image-input');
const fileName = document.querySelector('#file-name');
const processButton = document.querySelector('#process');
const status = document.querySelector('#status');
const original = document.querySelector('#original');
const result = document.querySelector('#result');
const download = document.querySelector('#download');
const paletteInput = document.querySelector('#palette-input');
const paletteName = document.querySelector('#palette-name');
const mosaicInput = document.querySelector('#mosaic');
const paletteSizeField = document.querySelector('#palette-size-field');
const bundledPaletteSelect = document.querySelector('#bundled-palette');

let selectedFile;
let originalURL;
let resultURL;
let paletteFiles = [];
let bundledPaletteManifest;
let bundledPalettePromise;

bundledPaletteSelect.addEventListener('change', () => {
  bundledPaletteManifest = undefined;
  bundledPalettePromise = undefined;
});

function updateOptionVisibility() {
  paletteSizeField.hidden = mosaicInput.checked;
}

async function startWasm() {
  const go = new Go();
  const response = await fetch('app.wasm');
  const instance = await WebAssembly.instantiateStreaming(response, go.importObject);
  void go.run(instance.instance);
}

startWasm()
  .then(() => {
    processButton.disabled = !selectedFile;
    status.textContent = 'Ready';
  })
  .catch((error) => {
    status.textContent = `WASM failed to load: ${error.message}`;
  });

input.addEventListener('change', () => {
  selectedFile = input.files[0];
  if (!selectedFile) return;
  fileName.textContent = selectedFile.name;
  if (originalURL) URL.revokeObjectURL(originalURL);
  originalURL = URL.createObjectURL(selectedFile);
  original.src = originalURL;
  processButton.disabled = false;
  download.hidden = true;
  status.textContent = 'Image ready';
});

paletteInput.addEventListener('change', () => {
  paletteFiles = [...paletteInput.files];
  paletteName.textContent = paletteFiles.length
    ? `${paletteFiles.length} tile${paletteFiles.length === 1 ? '' : 's'} selected`
    : 'Required for mosaic mode';
});

mosaicInput.addEventListener('change', updateOptionVisibility);
updateOptionVisibility();

async function loadBundledPalette(manifest) {
  if (bundledPaletteManifest !== manifest) {
    bundledPaletteManifest = manifest;
    bundledPalettePromise = fetch(manifest)
      .then((response) => response.json())
      .then((paths) => fetchWithConcurrencyLimit(paths, 24, async (path) => {
        const response = await fetch(path);
        return new Uint8Array(await response.arrayBuffer());
      }));
  }
  return bundledPalettePromise;
}

// Runs `task` over `items` with at most `limit` fetches in flight at once,
// since firing thousands of concurrent requests (e.g. the full dota-icons
// set) can exhaust the browser's connection pool (net::ERR_INSUFFICIENT_RESOURCES).
async function fetchWithConcurrencyLimit(items, limit, task) {
  const results = new Array(items.length);
  let nextIndex = 0;

  async function worker() {
    while (nextIndex < items.length) {
      const index = nextIndex++;
      results[index] = await task(items[index]);
    }
  }

  const workers = Array.from({ length: Math.min(limit, items.length) }, worker);
  await Promise.all(workers);
  return results;
}

processButton.addEventListener('click', async () => {
  if (!selectedFile || typeof processImage !== 'function') return;
  processButton.disabled = true;
  status.textContent = 'Processing...';
  try {
    const bytes = new Uint8Array(await selectedFile.arrayBuffer());
    const options = {
      pixelWidth: Number(document.querySelector('#pixel-width').value),
      scale: Number(document.querySelector('#scale').value),
      tileSize: Number(document.querySelector('#tile-size').value),
      mosaic: mosaicInput.checked,
      paletteSize: Number(document.querySelector('#palette-size').value),
    };
    let paletteSources = paletteFiles;
    if (options.mosaic && paletteSources.length === 0) {
      paletteSources = await loadBundledPalette(bundledPaletteSelect.value);
      paletteName.textContent = `${paletteSources.length} bundled tiles loaded`;
    }
    const paletteBytes = await Promise.all(paletteSources.map(async (file) => (
      file instanceof Uint8Array ? file :
      new Uint8Array(await file.arrayBuffer())
    )));
    const response = processImage(
      bytes,
      options.pixelWidth,
      options.scale,
      options.tileSize,
      options.mosaic,
      paletteBytes,
      options.paletteSize,
    );
    if (response.error) throw new Error(response.error);
    if (resultURL) URL.revokeObjectURL(resultURL);
    resultURL = URL.createObjectURL(new Blob([response.data], { type: 'image/png' }));
    result.src = resultURL;
    download.href = resultURL;
    download.hidden = false;
    status.textContent = 'Done';
  } catch (error) {
    status.textContent = `Could not process image: ${error.message}`;
  } finally {
    processButton.disabled = false;
  }
});