# slideserver/app.py

from flask import Flask, Response, abort, make_response, send_file, jsonify
from openslide import OpenSlide, OpenSlideError
from openslide.deepzoom import DeepZoomGenerator
from PIL import Image
import io
import os
from threading import Lock

app = Flask(__name__)

# --- Configuration ---
IMAGE_ROOT = "/srv/"
DEEPZOOM_FORMAT = 'jpeg'
DEEPZOOM_TILE_SIZE = 256
DEEPZOOM_OVERLAP = 0
DEEPZOOM_LIMIT_BOUNDS = False
DEEPZOOM_TILE_QUALITY = 90

generator_cache = {}
openslide_cache = {}
lock = Lock()

def get_openslide(path):
    """Returns an OpenSlide object from cache or creates a new one."""
    with lock:
        if path in openslide_cache:
            return openslide_cache[path]
    try:
        slide = OpenSlide(path)
        with lock:
            openslide_cache[path] = slide
        return slide
    except OpenSlideError:
        return None

def get_generator(path):
    """Returns a DeepZoomGenerator object from cache or creates a new one."""
    with lock:
        if path in generator_cache:
            return generator_cache[path]
    try:
        slide = get_openslide(path)
        if slide is None:
            return None
        generator = DeepZoomGenerator(
            slide,
            tile_size=DEEPZOOM_TILE_SIZE,
            overlap=DEEPZOOM_OVERLAP,
            limit_bounds=DEEPZOOM_LIMIT_BOUNDS
        )
        with lock:
            generator_cache[path] = generator
        return generator
    except Exception:
        return None

@app.route('/wsi/<path:filepath>.dzi')
def dzi(filepath):
    """Serves the DZI metadata."""
    full_path = os.path.join(IMAGE_ROOT, filepath)
    generator = get_generator(full_path)
    if generator is None:
        abort(404)
    resp = make_response(generator.get_dzi(DEEPZOOM_FORMAT))
    resp.mimetype = 'application/xml'
    return resp

@app.route('/wsi/<path:filepath>_files/<int:level>/<int:col>_<int:row>.<format>')
def tile(filepath, level, col, row, format):
    """Serves individual tiles for DeepZoom."""
    full_path = os.path.join(IMAGE_ROOT, filepath)
    generator = get_generator(full_path)
    if generator is None:
        abort(404)
    try:
        tile = generator.get_tile(level, (col, row))
    except ValueError:
        abort(404)
    if tile.mode != 'RGB':
        tile = tile.convert('RGB')
    buf = io.BytesIO()
    tile.save(buf, DEEPZOOM_FORMAT, quality=DEEPZOOM_TILE_QUALITY)
    resp = make_response(buf.getvalue())
    resp.mimetype = f'image/{format}'
    return resp

@app.route('/wsi/<path:filepath>.metadata')
def metadata(filepath):
    """Returns OpenSlide metadata as JSON."""
    full_path = os.path.join(IMAGE_ROOT, filepath)
    slide = get_openslide(full_path)
    if slide is None:
        abort(404)
    try:
        props = dict(slide.properties)
        props['dimensions'] = slide.dimensions
        props['level_count'] = slide.level_count
        props['level_dimensions'] = slide.level_dimensions
        return jsonify(props)
    except Exception:
        abort(500)
