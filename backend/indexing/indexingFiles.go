package indexing

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gtsteffaniak/filebrowser/backend/common/errors"
	"github.com/gtsteffaniak/filebrowser/backend/common/settings"
	"github.com/gtsteffaniak/filebrowser/backend/common/utils"
	"github.com/gtsteffaniak/filebrowser/backend/indexing/iteminfo"
	"github.com/gtsteffaniak/go-cache/cache"
	"github.com/gtsteffaniak/go-logger/logger"
)

var RealPathCache = cache.NewCache(48*time.Hour, 72*time.Hour)

// reduced index is json exposed to the client
type ReducedIndex struct {
	IdxName         string      `json:"name"`
	DiskUsed        int64       `json:"used"`
	DiskTotal       int64       `json:"total"`
	Status          IndexStatus `json:"status"`
	NumDirs         uint64      `json:"numDirs"`
	NumFiles        uint64      `json:"numFiles"`
	NumDeleted      uint64      `json:"numDeleted"`
	LastIndexed     time.Time   `json:"-"`
	LastIndexedUnix int64       `json:"lastIndexedUnixTime"`
	QuickScanTime   int         `json:"quickScanDurationSeconds"`
	FullScanTime    int         `json:"fullScanDurationSeconds"`
	Assessment      string      `json:"assessment"`
}
type Index struct {
	ReducedIndex
	CurrentSchedule            int `json:"-"`
	settings.Source            `json:"-"`
	Directories                map[string]*iteminfo.FileInfo `json:"-"`
	DirectoriesLedger          map[string]bool               `json:"-"`
	runningScannerCount        int                           `json:"-"`
	SmartModifier              time.Duration                 `json:"-"`
	FilesChangedDuringIndexing bool                          `json:"-"`
	mock                       bool
	mu                         sync.RWMutex
}

var (
	indexes      map[string]*Index
	indexesMutex sync.RWMutex
)

type IndexStatus string

const (
	READY       IndexStatus = "ready"
	INDEXING    IndexStatus = "indexing"
	UNAVAILABLE IndexStatus = "unavailable"
)

func init() {
	indexes = make(map[string]*Index)
}

func Initialize(source settings.Source, mock bool) {
	indexesMutex.Lock()
	newIndex := Index{
		mock:              mock,
		Source:            source,
		Directories:       make(map[string]*iteminfo.FileInfo),
		DirectoriesLedger: make(map[string]bool),
	}
	newIndex.ReducedIndex = ReducedIndex{
		Status:     "indexing",
		IdxName:    source.Name,
		Assessment: "unknown",
	}
	indexes[newIndex.Name] = &newIndex
	indexesMutex.Unlock()
	if !newIndex.Config.DisableIndexing {
		time.Sleep(time.Second)
		logger.Infof("initializing index: [%v]", newIndex.Name)
		newIndex.RunIndexing("/", false)
		go newIndex.setupIndexingScanners()
	} else {
		newIndex.Status = "ready"
		logger.Debug("indexing disabled for source: " + newIndex.Name)
	}
}

// Define a function to recursively index files and directories
func (idx *Index) indexDirectory(adjustedPath string, quick, recursive bool) error {
	realPath := strings.TrimRight(idx.Path, "/") + adjustedPath
	// Open the directory
	dir, err := os.Open(realPath)
	if err != nil {
		// must have been deleted
		return err
	}
	defer dir.Close()

	dirInfo, err := dir.Stat()
	if err != nil {
		return err
	}

	// check if excluded from indexing
	hidden := isHidden(dirInfo, idx.Path+adjustedPath)
	if idx.shouldSkip(dirInfo.IsDir(), hidden, adjustedPath) {
		return errors.ErrNotIndexed
	}

	// if indexing, mark the directory as valid and indexed.
	if recursive {
		// Prevent race conditions if scanning becomes concurrent in the future.
		idx.mu.Lock()
		idx.DirectoriesLedger[adjustedPath] = true
		idx.mu.Unlock()
	}
	combinedPath := adjustedPath + "/"
	if adjustedPath == "/" {
		combinedPath = "/"
	}
	// get whats currently in cache
	idx.mu.RLock()
	cacheDirItems := []iteminfo.ItemInfo{}
	modChange := false
	cachedDir, exists := idx.Directories[adjustedPath]
	if exists {
		modChange = dirInfo.ModTime() != cachedDir.ModTime
		cacheDirItems = cachedDir.Folders
	}
	idx.mu.RUnlock()

	// If the directory has not been modified since the last index, skip expensive readdir
	// recursively check cached dirs for mod time changes as well
	if recursive {
		if modChange {
			idx.mu.Lock()
			idx.FilesChangedDuringIndexing = true
			idx.mu.Unlock()
		} else if quick {
			for _, item := range cacheDirItems {
				err = idx.indexDirectory(combinedPath+item.Name, quick, true)
				if err != nil && err != errors.ErrNotIndexed {
					logger.Errorf("error indexing directory %v : %v", combinedPath+item.Name, err)
				}
			}
			return nil
		}
	}
	dirFileInfo, err2 := idx.GetDirInfo(dir, dirInfo, realPath, adjustedPath, combinedPath, quick, recursive,true)
	if err2 != nil {
		return err2
	}
	// Update the current directory metadata in the index
	idx.UpdateMetadata(dirFileInfo)
	return nil
}

func (idx *Index) GetFsDirInfo(adjustedPath string) (*iteminfo.FileInfo, error) {
	realPath, isDir, err := idx.GetRealPath(adjustedPath)
	if err != nil {
		return nil, err
	}
	originalPath := realPath
	if !isDir {
		// get parent directory info
		realPath = filepath.Dir(realPath)
	}
	dir, err := os.Open(realPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	dirInfo, err := dir.Stat()
	if err != nil {
		return nil, err
	}
	combinedPath := adjustedPath + "/"
	if adjustedPath == "/" {
		combinedPath = "/"
	}
	var response *iteminfo.FileInfo
	response, err = idx.GetDirInfo(dir, dirInfo, realPath, adjustedPath, combinedPath, false, false,false)
	if err != nil {
		return nil, err
	}
	if !isDir {
		baseName := filepath.Base(originalPath)
		idx.MakeIndexPath(realPath)
		found := false
		for _, item := range response.Files {
			if item.Name == baseName {
				response = &iteminfo.FileInfo{
					Path:     adjustedPath + "/" + item.Name,
					ItemInfo: item,
				}
				found = true
				continue
			}
		}
		if !found {
			return nil, fmt.Errorf("file not found in directory: %s", adjustedPath)
		}

	}
	return response, nil

}

func (idx *Index) GetDirInfo(dirInfo *os.File, stat os.FileInfo, realPath, adjustedPath, combinedPath string, quick, recursive, checkSkip bool) (*iteminfo.FileInfo, error) {
	// Read directory contents
	files, err := dirInfo.Readdir(-1)
	if err != nil {
		return nil, err
	}
	var totalSize int64
	fileInfos := []iteminfo.ItemInfo{}
	dirInfos := []iteminfo.ItemInfo{}

	// Process each file and directory in the current directory
	for _, file := range files {
		hidden := isHidden(file, idx.Path+combinedPath)
		isDir := iteminfo.IsDirectory(file)
		fullCombined := combinedPath + file.Name()
		if checkSkip && idx.shouldSkip(isDir, hidden, fullCombined) {
			continue
		}
		itemInfo := &iteminfo.ItemInfo{
			Name:    file.Name(),
			ModTime: file.ModTime(),
			Hidden:  hidden,
		}

		if isDir {
			// skip non-indexable dirs.
			if file.Name() == "$RECYCLE.BIN" || file.Name() == "System Volume Information" {
				continue
			}

			dirPath := combinedPath + file.Name()
			if recursive {
				// clear for garbage collection
				file = nil
				// Recursively index the subdirectory
				err = idx.indexDirectory(dirPath, quick, recursive)
				if err != nil {
					logger.Errorf("Failed to index directory %s: %v", dirPath, err)
					continue
				}
			}
			realDirInfo, exists := idx.GetMetadataInfo(dirPath, true)
			if exists {
				itemInfo.Size = realDirInfo.Size
			}
			totalSize += itemInfo.Size
			itemInfo.Type = "directory"
			dirInfos = append(dirInfos, *itemInfo)
			idx.NumDirs++
		} else {
			itemInfo.DetectType(fullCombined, false)
			itemInfo.Size = file.Size()
			fileInfos = append(fileInfos, *itemInfo)
			totalSize += itemInfo.Size
			idx.NumFiles++
		}
	}

	if totalSize == 0 && idx.Config.IgnoreZeroSizeFolders {
		return nil, errors.ErrNotIndexed
	}

	// Create FileInfo for the current directory
	dirFileInfo := &iteminfo.FileInfo{
		Path:    adjustedPath,
		Files:   fileInfos,
		Folders: dirInfos,
	}
	dirFileInfo.ItemInfo = iteminfo.ItemInfo{
		Name:    filepath.Base(dirInfo.Name()),
		Type:    "directory",
		Size:    totalSize,
		ModTime: stat.ModTime(),
	}
	dirFileInfo.SortItems()
	return dirFileInfo, nil
}

// input should be non-index path.
func (idx *Index) MakeIndexPath(subPath string) string {
	if strings.HasPrefix(subPath, "./") {
		subPath = strings.TrimPrefix(subPath, ".")
	}
	if idx.Path == subPath || subPath == "." {
		return "/"
	}
	// clean path
	subPath = strings.TrimSuffix(subPath, "/")
	adjustedPath := strings.TrimPrefix(subPath, idx.Path)
	// remove index prefix
	adjustedPath = strings.ReplaceAll(adjustedPath, "\\", "/")
	// remove trailing slash
	adjustedPath = strings.TrimSuffix(adjustedPath, "/")
	if !strings.HasPrefix(adjustedPath, "/") {
		adjustedPath = "/" + adjustedPath
	}
	return adjustedPath
}

func (idx *Index) recursiveUpdateDirSizes(childInfo *iteminfo.FileInfo, previousSize int64) {
	parentDir := utils.GetParentDirectoryPath(childInfo.Path)
	parentInfo, exists := idx.GetMetadataInfo(parentDir, true)
	if !exists || parentDir == "" {
		return
	}
	newSize := parentInfo.Size - previousSize + childInfo.Size
	parentInfo.Size += newSize
	idx.UpdateMetadata(parentInfo)
	idx.recursiveUpdateDirSizes(parentInfo, newSize)
}

func (idx *Index) GetRealPath(relativePath ...string) (string, bool, error) {
	combined := append([]string{idx.Path}, relativePath...)
	joinedPath := filepath.Join(combined...)
	isDir, _ := RealPathCache.Get(joinedPath + ":isdir").(bool)
	cached, ok := RealPathCache.Get(joinedPath).(string)
	if ok && cached != "" {
		return cached, isDir, nil
	}
	// Convert relative path to absolute path
	absolutePath, err := filepath.Abs(joinedPath)
	if err != nil {
		return absolutePath, false, fmt.Errorf("could not get real path: %v, %s", joinedPath, err)
	}
	// Resolve symlinks and get the real path
	realPath, isDir, err := utils.ResolveSymlinks(absolutePath)
	if err == nil {
		RealPathCache.Set(joinedPath, realPath)
		RealPathCache.Set(joinedPath+":isdir", isDir)
	}
	return realPath, isDir, err
}

func (idx *Index) RefreshFileInfo(opts iteminfo.FileOptions) error {
	refreshOptions := iteminfo.FileOptions{
		Path:  opts.Path,
		IsDir: opts.IsDir,
	}
	if !refreshOptions.IsDir {
		refreshOptions.Path = idx.MakeIndexPath(filepath.Dir(refreshOptions.Path))
		refreshOptions.IsDir = true
	}
	err := idx.indexDirectory(refreshOptions.Path, false, false)
	if err != nil {
		return err
	}
	file, exists := idx.GetMetadataInfo(refreshOptions.Path, true)
	if !exists {
		return fmt.Errorf("file/folder does not exist in metadata: %s", refreshOptions.Path)
	}

	current, firstExisted := idx.GetMetadataInfo(refreshOptions.Path, true)
	refreshParentInfo := firstExisted && current.Size != file.Size
	//utils.PrintStructFields(*file)
	result := idx.UpdateMetadata(file)
	if !result {
		return fmt.Errorf("file/folder does not exist in metadata: %s", refreshOptions.Path)
	}
	if !exists {
		return nil
	}
	if refreshParentInfo {
		idx.recursiveUpdateDirSizes(file, current.Size)
	}
	return nil
}

func isHidden(file os.FileInfo, srcPath string) bool {
	// Check if the file starts with a dot (common on Unix systems)
	if file.Name()[0] == '.' {
		return true
	}

	if runtime.GOOS == "windows" {
		return CheckWindowsHidden(filepath.Join(srcPath, file.Name()))
	}

	// Default behavior for non-Windows systems
	return false
}

func (idx *Index) shouldSkip(isDir bool, isHidden bool, fullCombined string) bool {
	if idx.Config.DisableIndexing {
		return true
	}
	// check inclusions first
	if isDir && len(idx.Config.Include.Folders) > 0 {
		if !slices.Contains(idx.Config.Include.Folders, fullCombined) {
			return true
		}
	}
	if !isDir && len(idx.Config.Include.Files) > 0 {
		if !slices.Contains(idx.Config.Include.Files, fullCombined) {
			return true
		}
	}

	if !isDir && len(idx.Config.Include.FileEndsWith) > 0 {
		shouldSkip := true
		for _, end := range idx.Config.Include.FileEndsWith {
			if strings.HasSuffix(fullCombined, end) {
				shouldSkip = false
				break
			}
		}
		if shouldSkip {
			return true
		}
	}
	// check exclusions
	if isDir && slices.Contains(idx.Config.Exclude.Folders, fullCombined) {
		return true
	}
	if !isDir && slices.Contains(idx.Config.Exclude.Files, fullCombined) {
		return true
	}
	if idx.Config.IgnoreHidden && isHidden {
		return true
	}

	if !isDir && len(idx.Config.Exclude.FileEndsWith) > 0 {
		shouldSkip := false
		for _, end := range idx.Config.Exclude.FileEndsWith {
			if strings.HasSuffix(fullCombined, end) {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			return true
		}
	}

	return false
}

type DiskUsage struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

func (idx *Index) SetUsage(usage DiskUsage) {
	if settings.Config.Frontend.DisableUsedPercentage {
		return
	}
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.DiskUsed = int64(usage.Used)
	idx.DiskTotal = int64(usage.Total)
}

func (idx *Index) SetStatus(status IndexStatus) {
	idx.mu.Lock()
	idx.Status = status
	switch status {
	case INDEXING:
		idx.runningScannerCount++
	case READY, UNAVAILABLE:
		idx.runningScannerCount = 0
	}
	idx.mu.Unlock()
	idx.SendSourceUpdateEvent()
}
