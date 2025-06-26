# FileBrowser-WSI: A Whole Slide Image Viewer Integration

This is a fork of the excellent **[FileBrowser-Quantum](https://github.com/gtsteffaniak/filebrowser)** project. The documentation for the original project has been preserved and can be found here: **[README.original.md](./README.original.md)**.

This fork has been extended with a native integration for viewing Whole Slide Images (WSI) directly in the browser. It preserves all the functionality of the original FileBrowser while adding a seamless experience for viewing medical and research images in formats like `.svs`, `.ndpi`, and more, powered by OpenSeadragon.

![WSI Viewer Screenshot](https://user-images.githubusercontent.com/1035832/203498953-27351a9e-062e-4a25-9b34-a034293f350c.png) ## Key Features

* **All Original Features:** Retains 100% of the functionality from the upstream FileBrowser-Quantum.
* **Native WSI Viewer:** Adds a fully integrated viewer for Whole Slide Images based on OpenSeadragon, which loads when a compatible file is selected.
* **Dynamic Scalebar:** A dynamic scalebar is rendered on top of the viewer, providing accurate measurements at any zoom level.
* **Backend Proxy:** Includes a secure backend proxy for the `slideserver`, ensuring that WSI tiles are served only to authenticated users.
* **Customizable Icons:** Viewer controls can be customized and are currently configured to load from a remote repository.

## How It Works

This fork uses a multi-container setup orchestrated by Docker Compose:

1.  **`filebrowser` Service:** The main application, built from this custom fork. It now includes a backend proxy to securely route requests for WSI tiles to the `slideserver`.
2.  **`slideserver` Service:** A lightweight Python/Flask microservice that uses `openslide-python` to read WSI files and serve up image tiles to the OpenSeadragon viewer.

## Configuration

To enable the WSI viewer, you must add the `integrations.wsi` block to your `config.yaml` file.

```yaml
# In ./config/config.yaml

server:
  port: 80
  internalUrl: "http://filebrowser:80"
  sources:
    - path: /srv
      name: main

auth:
  adminUsername: admin
  adminPassword: admin

integrations:
  wsi:
    # This URL is no longer needed as the proxy handles all requests.
    # It is kept for future compatibility but can be an empty string.
    url: ""
    
    # This URL is for internal communication between the containers.
    # It uses the service name 'slideserver' and its internal port.
    internalUrl: "http://slideserver:5000"
```

##Running the Application

This project is designed to be run with Docker Compose
1.  **Clone the repository:**
```bash
git clone [https://192.168.0.184:3100/miho/filebrowser.git](https://192.168.0.184:3100/miho/filebrowser.git)
cd filebrowser
```
2.  **Configure:** Create a `./config/config.yaml` file with the content described above.

3.  Run: Use the provided docker-compose.yml to build and start the services.
```bash
docker-compose up --build -d
```
3.  **Access:** You can now access your FileBrowser instance at `http://<your-ip>:38080`.


**Example** `docker-compose.yml`:

```yaml
version: "3.8"

services:
  filebrowser:
    build:
      context: .
      dockerfile: _docker/Dockerfile
    container_name: filebrowser_osdosintegrated
    user: "0:0"
    ports:
      - "38080:80"
    volumes:
      - /mnt/images:/srv
      - ./config:/config
      - ./database:/database
    environment:
      FILEBROWSER_CONFIG: "/config/config.yaml"
    restart: unless-stopped

  slideserver:
    build:
      context: ./slideserver
    container_name: slideserver_integrated
    volumes:
      - /mnt/images:/srv:ro
      - ./wsi-cache:/cache
    restart: unless-stopped
```

