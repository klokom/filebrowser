<template>
  <div class="card floating">
    <div class="card-title">
      <h2>WSI Metadata</h2>
    </div>

    <div class="card-content">
      <p v-if="!hasMetadata">
        No WSI metadata available for this file.
      </p>
      <template v-else>
        <p v-for="([key, value]) in metadataForDisplay" :key="key">
          <strong>{{ key.replace('aperio.', '').replace('openslide.', '') }}:</strong>
          <span>{{ value }}</span>
        </p>
      </template>
    </div>

    <div class="card-action">
      <button
        type="submit"
        @click="closeHovers"
        class="button button--flat"
        aria-label="OK"
        title="OK"
      >
        OK
      </button>
    </div>
  </div>
</template>

<script>
import { state, mutations } from "@/store";

export default {
  name: "WSImetadata",
  computed: {
    closeHovers() {
      return mutations.closeHovers;
    },
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
        .slice(0, 25);
    }
  },
};
</script>

<style scoped>
.card-content {
  max-height: 70vh;
  overflow-y: auto;
}
.card-content p {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>