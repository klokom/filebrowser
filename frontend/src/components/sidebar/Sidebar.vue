<template>
  <nav
    id="sidebar"
    :class="{ active: active, 'dark-mode': isDarkMode, 'behind-overlay': behindOverlay }"
  >
    <div v-if="shouldShow" class="button release-banner">
      <a :href="releaseUrl">{{ $t("sidebar.updateIsAvailable") }}</a>
      <i @click="setSeenUpdate" aria-label="close-banner" class="material-icons">close</i>
    </div>
    <SidebarSettings v-if="isSettings"></SidebarSettings>
    <SidebarGeneral v-else-if="isLoggedIn"></SidebarGeneral>

    <div class="buffer"></div>
    <div class="credits">
      <span v-for="item in externalLinks" :key="item.title">
        <a :href="item.url" target="_blank" :title="item.title">{{ item.text }}</a>
      </span>
      <span v-if="name != ''">
        <h4 style="margin: 0">{{ name }}</h4>
      </span>
    </div>
  </nav>
</template>

<script>
import { externalLinks, name, updateAvailable } from "@/utils/constants";
import { getters, mutations, state } from "@/store"; // Import your custom store
import SidebarGeneral from "./General.vue";
import SidebarSettings from "./Settings.vue";

export default {
  name: "sidebar",
  components: {
    SidebarGeneral,
    SidebarSettings,
  },
  data() {
    return {
      externalLinks,
      name,
    };
  },
  mounted() {
    // Ensure the sidebar is initialized correctly
    mutations.setSeenUpdate(localStorage.getItem("seenUpdate"));
  },
  computed: {
    releaseUrl: () => updateAvailable,
    isDarkMode: () => getters.isDarkMode(),
    isLoggedIn: () => getters.isLoggedIn(),
    isSettings: () => getters.isSettings(),
    active: () => getters.isSidebarVisible(),
    behindOverlay: () => state.isSearchActive,
    shouldShow() {
      return (
        updateAvailable != "" &&
        state.user.permissions.admin &&
        state.seenUpdate != updateAvailable &&
        !state.user.disableUpdateNotifications
      );
    },
  },
  methods: {
    // Show the help overlay
    help() {
      mutations.showHover("help");
    },
    setSeenUpdate() {
      mutations.setSeenUpdate(updateAvailable);
    },
  },
};
</script>

<style>
.sidebar-scroll-list {
  overflow: auto;
  margin-bottom: 0px !important;
}

#sidebar {
  display: flex;
  flex-direction: column;
  padding: 1em;
  width: 20em;
  position: fixed;
  z-index: 4;
  left: -20em;
  height: 100%;
  box-shadow: 0 0 5px rgba(0, 0, 0, 0.1);
  transition: 0.5s ease;
  top: 4em;
  padding-bottom: 4em;
  background-color: #dddddd;
}

#sidebar.behind-overlay {
  z-index: 3;
}

#sidebar.sticky {
  z-index: 3;
}

@supports (backdrop-filter: none) {
  #sidebar {
    background-color: rgba(237, 237, 237, 0.33) !important;
    backdrop-filter: blur(16px) invert(0.1);
  }
}

body.rtl nav {
  left: unset;
  right: -17em;
}

#sidebar.active {
  left: 0;
}

#sidebar.rtl nav.active {
  left: unset;
  right: 0;
}

#sidebar .action {
  width: 100%;
  display: block;
  white-space: nowrap;
  height: 100%;
  overflow: hidden;
  padding: 0.5em;
  text-overflow: ellipsis;
}

body.rtl .action {
  direction: rtl;
  text-align: right;
}

#sidebar .action > * {
  vertical-align: middle;
}

/* * * * * * * * * * * * * * * *
 *            FOOTER           *
 * * * * * * * * * * * * * * * */

.credits {
  font-size: 1em;
  color: var(--textPrimary);
  padding-left: 1em;
  padding-bottom: 1em;
}

.credits > span {
  display: block;
  margin-top: 0.5em;
  margin-left: 0;
}

.credits a,
.credits a:hover {
  cursor: pointer;
}

.buffer {
  flex-grow: 1;
}

.clickable {
  cursor: pointer;
}

.clickable:hover {
  box-shadow: 0 2px 2px #00000024, 0 1px 5px #0000001f, 0 3px 1px -2px #0003;
}

.release-banner {
  background-color: var(--primarColor);
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1em;
}
</style>
