<template>
  <div id="openseadragon-container" ref="osdContainer"></div>
</template>

<script>
// in frontend/src/views/files/WSIViewer.vue

import OpenSeadragon from 'openseadragon';
// IMPORT 'mutations' and 'state' from the store
import { state, mutations } from "@/store";
// IMPORT the addScalebar function from our new utility file.
import { addScalebar } from '@/utils/scalebar.js';

// function addScalebar(viewer, viewerDiv, metadata) {
//   console.log("Scalebar placeholder with metadata:", metadata);
// }

export default {
  name: 'WSIViewer',
  data: () => ({
    viewer: null
  }),
  computed: {
    req() {
      return state.req;
    }
  },
  // THIS IS THE NEW MOUNTED HOOK
  mounted() {
    // Tell the store to clear any previous selections
    mutations.resetSelected();
    // Tell the store that this file is the currently selected item
    mutations.addSelected(this.req);
    // Now initialize the viewer
    this.initializeViewer();
  },
  watch: {
    req: 'initializeViewer'
  },
  beforeUnmount() {
    if (this.viewer) {
      this.viewer.destroy();
    }
  },
  methods: {
    // The initializeViewer() method remains unchanged.
    initializeViewer() {
      if (this.viewer) {
        this.viewer.destroy();
      }
      const { WSIUrl, BaseURL: baseURL } = window.FileBrowser;
      if (!WSIUrl) {
        console.error("WSI integration URL is not defined in the backend configuration.");
        if (this.$refs.osdContainer) {
          this.$refs.osdContainer.innerHTML = '<p style="text-align: center; color: white;">⚠️ WSI Viewer not configured.</p>';
        }
        return;
      }
      const giteaPrefixUrl = 'https://192.168.0.184:3100/miho/openseadragon-icons/raw/branch/main/images/';
      const dziUrl = `${baseURL}api/wsi${this.req.path}.dzi`;
      const metadataUrl = `${baseURL}api/wsi${this.req.path}.metadata`;
      console.log(`Initializing WSI viewer for: ${this.req.path}`);
      console.log(`Requesting DZI from: ${dziUrl}`);
      this.viewer = OpenSeadragon({
        element: this.$refs.osdContainer,
        prefixUrl: giteaPrefixUrl,
        showNavigator: true,
        tileSources: dziUrl,
      });
      this.viewer.addHandler("open-failed", (event) => {
        console.error("OpenSeadragon failed to open:", event);
        if (this.$refs.osdContainer) {
          this.$refs.osdContainer.innerHTML = '<p style="text-align: center; color: white;">⚠️ Could not load slide image.</p>';
        }
      });
      fetch(metadataUrl)
        .then(res => {
          if (!res.ok) {
            return Promise.reject(`Metadata fetch failed: ${res.status} ${res.statusText}`);
          }
          return res.json();
        })
        .then(metadata => {
          console.log("Metadata:", metadata);
          addScalebar(this.viewer, this.$refs.osdContainer, metadata);
        })
        .catch(err => {
          console.warn("❌ Metadata fetch error:", err);
          addScalebar(this.viewer, this.$refs.osdContainer, {});
        });
    }
  }
}
</script>

<style scoped>
#openseadragon-container {
  width: 100%;
  height: 100%;
  background-color: #000;
}
</style>
<style>
/* By removing 'scoped', these styles become global and can correctly
   target the dynamically created scalebar element. */

#custom-scalebar {
    position: absolute;
    bottom: 10px;
    left: 10px;
    color: white;
    font: 12px sans-serif;
    background: rgba(0,0,0,0.5);
    padding: 4px 8px;
    border-radius: 4px;
    pointer-events: none;
}

#custom-scalebar .bar {
    height: 2px;
    background: white;
    margin-bottom: 4px;
}
</style>