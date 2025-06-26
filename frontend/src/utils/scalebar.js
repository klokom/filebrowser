// in frontend/src/utils/scalebar.js

export function addScalebar(viewer, viewerDiv, metadata) {
  const micronsPerPixel = parseFloat(metadata?.['openslide.mpp-x']) || 0;

  if (micronsPerPixel === 0) {
    console.warn("Scalebar not added: 'openslide.mpp-x' not found in metadata.");
    return;
  }

  const scalebar = document.createElement('div');
  scalebar.id = 'custom-scalebar';
  
  const bar = document.createElement('div');
  bar.className = 'bar';
  
  const label = document.createElement('div');
  label.className = 'label';
  
  scalebar.appendChild(bar);
  scalebar.appendChild(label);
  viewerDiv.appendChild(scalebar);

  const updateScaleBar = () => {
    const zoom = viewer.viewport.getZoom(true);
    const imageZoom = viewer.viewport.viewportToImageZoom(zoom);
    const µmPerScreenPixel = micronsPerPixel / imageZoom;

    const potentialLengths = [1, 2, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000];
    let targetLength;
    let barPx;

    for (const length of potentialLengths) {
      const pixels = length / µmPerScreenPixel;
      if (pixels > 80) {
        targetLength = length;
        barPx = pixels;
        break;
      }
    }

    if (!targetLength) {
        targetLength = potentialLengths[potentialLengths.length - 1];
        barPx = targetLength / µmPerScreenPixel;
    }

    bar.style.width = `${barPx}px`;
    
    if (targetLength >= 1000) {
      label.textContent = `${targetLength / 1000} mm`;
    } else {
      label.textContent = `${targetLength} µm`;
    }
  };

  viewer.addHandler('open', updateScaleBar);
  viewer.addHandler('animation', updateScaleBar);
  viewer.addHandler('zoom', updateScaleBar);
}