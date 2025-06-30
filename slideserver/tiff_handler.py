# slideserver/tiff_handler.py
import os
import pyvips
import hashlib

# This is a critical setting for pyvips to handle large images
pyvips.cache_set_max(0)

def generate_tiff_pyramid(image_path, cache_root):
    """
    Uses pyvips to create a DZI pyramid for a standard TIFF file.
    Returns the path to the cache directory for this image.
    """
    try:
        image_hash = hashlib.md5(image_path.encode()).hexdigest()
        image_cache_dir = os.path.join(cache_root, image_hash)
        
        # Only generate if the cache doesn't already exist
        if not os.path.exists(image_cache_dir):
            os.makedirs(image_cache_dir, exist_ok=True)
            print(f"Cache miss for {os.path.basename(image_path)}. Generating TIFF pyramid with pyvips...")
            
            image = pyvips.Image.new_from_file(image_path)
            
            file_basename = os.path.splitext(os.path.basename(image_path))[0]
            
            image.dzsave(
                os.path.join(image_cache_dir, file_basename),
                suffix='.jpeg'
            )
            print("Pyramid generation complete.")
        
        return image_cache_dir
        
    except pyvips.Error as e:
        print(f"ERROR: pyvips failed to process {os.path.basename(image_path)}: {e}")
        return None
    except Exception as e:
        print(f"ERROR: An unexpected error occurred in tiff_handler: {e}")
        return None