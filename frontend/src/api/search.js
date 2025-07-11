import { fetchURL } from "./utils";
import { notify } from "@/notify";  // Import notify for error handling
import { getApiPath } from "@/utils/url.js";

export default async function search(base, source, query) {
  try {
    query = encodeURIComponent(query);
    if (!base.endsWith("/")) {
      base += "/";
    }
    const apiPath = getApiPath("api/search", { scope: encodeURIComponent(base), query: query, source: encodeURIComponent(source) });
    const res = await fetchURL(apiPath);
    let data = await res.json();

    return data
  } catch (err) {
    notify.showError(err.message || "Error occurred during search");
    throw err;
  }
}
