<template>
  <div id="openseadragon-container" ref="osdContainer">
    <div v-if="showMetadata && hasMetadata" class="card floating wsi-metadata-card">
      <div class="card-title">
        <h2>Slide Metadata</h2>
        <button @click="toggleMetadataDisplay" class="action" aria-label="Close" title="Close">
          <i class="material-icons">close</i>
        </button>
      </div>

      <div class="card-content">
        <p v-for="([key, value]) in metadataForDisplay" :key="key">
          <strong>{{ key.replace('aperio.', '').replace('openslide.', '') }}:</strong>
          <span>{{ value }}</span>
        </p>
      </div>
    </div>
  </div>
</template>

<script>
import OpenSeadragon from 'openseadragon';
import { state, mutations } from "@/store";
import { addScalebar } from '@/utils/scalebar.js';
import { eventBus } from "@/store/eventBus";

export default {
  name: 'WSIViewer',
  data: () => ({
    viewer: null,
    showMetadata: false,
  }),
  computed: {
    req() {
      return state.req;
    },
    hasMetadata() {
      return this.req.wsiMetadata && Object.keys(this.req.wsiMetadata).length > 0;
    },
    metadataForDisplay() {
      if (!this.hasMetadata) return [];
      return Object.entries(this.req.wsiMetadata)
        .filter(([key]) => key.startsWith('aperio.') || key === 'openslide.mpp-x' || key === 'openslide.objective-power')
        .slice(0, 20);
    }
  },
  mounted() {
    mutations.resetSelected();
    mutations.addSelected(this.req);
    this.initializeViewer();
    eventBus.on('toggle-wsi-metadata', this.toggleMetadataDisplay);
  },
  watch: {
    req: 'initializeViewer'
  },
  beforeUnmount() {
    if (this.viewer) {
      this.viewer.destroy();
    }
    eventBus.off('toggle-wsi-metadata', this.toggleMetadataDisplay);
  },
  methods: {
    toggleMetadataDisplay() {
      this.showMetadata = !this.showMetadata;
    },
    initializeViewer() {
      if (this.viewer) { this.viewer.destroy(); }
      const { BaseURL: baseURL } = window.FileBrowser;
      const dziUrl = `${baseURL}api/wsi${this.req.path}.dzi`;
      const metadata = this.req.wsiMetadata || {};
      const giteaPrefixUrl = 'https://192.168.0.184:3100/miho/openseadragon-icons/raw/branch/main/images/';
      this.viewer = OpenSeadragon({
        element: this.$refs.osdContainer,
        showNavigator: true,
        tileSources: dziUrl,
        prefixUrl: giteaPrefixUrl
      });
      this.viewer.addHandler("open-failed", (event) => {
        if (this.$refs.osdContainer) { this.$refs.osdContainer.innerHTML = '<p style="text-align: center; color: white;">⚠️ Could not load slide image.</p>'; }
      });
      addScalebar(this.viewer, this.$refs.osdContainer, metadata);
    }
  }
}
</script>

<style scoped>
#openseadragon-container {
  width: 100%;
  height: 100%;
  background-color: #000;
  position: relative;
}

/* New styles for the metadata card, inspired by Info.vue */
.wsi-metadata-card {
  position: absolute;
  top: 70px;
  left: 15px;
  max-width: 380px;
  max-height: calc(100% - 140px); /* Adjust to not overlap with other controls */
  z-index: 10; /* Ensure it's above the viewer canvas */
  display: flex;
  flex-direction: column;
}

.wsi-metadata-card .card-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.wsi-metadata-card .card-content {
  overflow-y: auto;
  padding-right: 1.5em; /* Add some padding for the scrollbar */
}

.wsi-metadata-card .card-content p {
  margin: 0 0 0.5em 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.wsi-metadata-card .card-content strong {
  text-transform: capitalize;
}

.wsi-metadata-card .card-title .action {
  padding: 0;
  width: 2em;
  height: 2em;
}
</style>