package memoryFS

import (
	"fmt"
	"fsraft"
	"math"
	"path"
	"strings"
)

// An in-memory file system.
// Files here are stored in memory in Go.
type MemoryFS struct {
	activeFDs           map[int]File // A map from active file descriptors to File objects.
	smallestAvailableFD int          // The smallest positive number that is not 0, 1, 2, or an active file descriptor.
	// (0, 1, and 2 are banned because they are reserved for stdin, stdout, and stderr)
	rootDir Directory
}

// Create an empty in-memory FileSystem rooted at "/".
func CreateEmptyMemoryFS() MemoryFS {
	// the root directory is unnamed, which is unintentional
	return MemoryFS{activeFDs: make(map[int]File), smallestAvailableFD: 3, rootDir: Directory{}}
}

// See the spec for FileSystem::Mkdir.
func (mfs *MemoryFS) Mkdir(path string) (success bool, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Open.
func (mfs *MemoryFS) Open(filePath string, mode fsraft.OpenMode, flags fsraft.OpenFlags) (fileDescriptor int, err error) {
	//fmt.Printf("Starting Open(%v, %v, %v)\n", filePath, mode.String(), flags)
	fileDescriptor = -1 // in case we return early, set it here

	dirPath, fileName := path.Split(filePath)
	dirs := strings.Split(dirPath, "/")
	if dirs[0] != "" {
		// There is something before the first "/", so this is an invalid path
		err = fsraft.InvalidPath
		return
	}
	dirs = dirs[1:]

	currentDir := mfs.rootDir
	for _, dir := range dirs {
		if currentDir.HasChildNamed(dir) {
			child := currentDir.GetChildNamed(dir)
			childDir, ok := child.(Directory)
			if !ok {
				err = fsraft.InvalidPath
				return
			}
			currentDir = childDir
		} else {
			err = fsraft.InvalidPath
			return
		}
	}

	if !currentDir.HasChildNamed(fileName) {
		// If the create flag is set
		if (flags & fsraft.Create) != 0 {
			currentDir.CreateFile(fileName)
		} else {
			err = fsraft.NotFound
			return
		}
	}

	file, isFile := currentDir.GetChildNamed(fileName).(File)
	if !isFile {
		// It's a directory
		err = fsraft.IsDirectory
		return
	}

	file.Open(mode, flags)

	// and now to assign it to a file descriptor
	fileDescriptor = mfs.smallestAvailableFD
	mfs.activeFDs[fileDescriptor] = file

	// Maintain the invariant of smallestAvailableFD
	for {
		_, fdIsActive := mfs.activeFDs[mfs.smallestAvailableFD]
		if fdIsActive {
			mfs.smallestAvailableFD++
		} else {
			break
		}
	}

	//fmt.Printf("Returning FD=%v", fileDescriptor)
	err = nil // Not sure if this is necessary? If not, just delete it
}

// See the spec for FileSystem::Close.
func (mfs *MemoryFS) Close(fileDescriptor int) (success bool, err error) {
	file, fdIsActive := mfs.activeFDs[fileDescriptor]
	if !fdIsActive {
		return false, fsraft.InvalidFD
	}
	return file.Close()
}

// See the spec for FileSystem::Seek.
func (mfs *MemoryFS) Seek(fileDescriptor int, offset int, base fsraft.SeekMode) (newPosition int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Read.
func (mfs *MemoryFS) Read(fileDescriptor int, numBytes int) (bytesRead int, data []byte, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Write.
func (mfs *MemoryFS) Write(fileDescriptor int, numBytes int, data []byte) (bytesWritten int, err error) {
	panic("TODO")
}

// See the spec for FileSystem::Delete.
func (mfs *MemoryFS) Delete(path string) (success bool, err error) {
	panic("TODO")
}
