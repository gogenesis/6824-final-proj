package memoryFS

import (
	"ad"
	"filesystem"
	"path"
	"strings"
)

// An in-memory file system.
// Files here are stored in memory in Go.
type MemoryFS struct {
	activeFDs           map[int]*File // A map from active file descriptors to open files.
	smallestAvailableFD int           // The smallest positive number that is not 0, 1, 2, or an active file descriptor.
	// (0, 1, and 2 are banned because they are reserved for stdin, stdout, and stderr)
	rootDir Directory
}

// Create an empty in-memory FileSystem rooted at "/".
func CreateEmptyMemoryFS() MemoryFS {
	mfs := MemoryFS{
		activeFDs:           make(map[int]*File), //opened FDs ...
		smallestAvailableFD: 3,
		rootDir:             Directory{},
	}
	mfs.rootDir.inode = Inode{
		name: "",
	}
	mfs.rootDir.children = make(map[string]Node)
	return mfs
}

// Operations from FileSystem =================================================

// See the spec for FileSystem::Mkdir.
func (mfs *MemoryFS) Mkdir(filePath string) (success bool, err error) {
	ad.Debug(ad.TRACE, "Starting Mkdir(%v)", filePath)
	success = false // in case we return early, set it here
	currentDir, _, newDirName, existence := mfs.followPath(filePath)
	switch existence {
	case NodeExists:
		err = filesystem.AlreadyExists
	case ParentExistsButNodeDoesNot:
		// true for directory instead of file
		currentDir.createChild(newDirName, true)
		success = true
	case ParentDoesNotExist:
		err = filesystem.NotFound
	}

	ad.Debug(ad.TRACE, "Done with Mkdir(%v), returning (%t, %v)", filePath, success, err)
	return // Needed for compilation
}

// See the spec for FileSystem::Open.
func (mfs *MemoryFS) Open(filePath string, mode filesystem.OpenMode, flags filesystem.OpenFlags) (fileDescriptor int, err error) {
	ad.Debug(ad.TRACE, "Starting Open(%v, %v, %v)", filePath, mode.String(), flags)
	// We have to have the "Done with Open" debug on every return, we can't defer it, because parameters to a deferred
	// function are evaluated at defer time, not at call time.
	fileDescriptor = -1 // in case we return early, set it here

	currentDir, node, fileName, existence := mfs.followPath(filePath)

	switch existence {
	case NodeExists:
		// proceed as normal

	case ParentExistsButNodeDoesNot:
		if filesystem.FlagIsSet(flags, filesystem.Create) {
			currentDir.CreateFile(fileName)
			// Set node here because node was set to nil above because it didn't exist
			node = currentDir.GetChildNamed(fileName)
		} else {
			err = filesystem.NotFound
			ad.Debug(ad.TRACE, "Done with Open(%v, %v, %v), returning (%v, %v)", filePath, mode.String(), flags, fileDescriptor, err)
			return
		}

	case ParentDoesNotExist:
		err = filesystem.NotFound
		ad.Debug(ad.TRACE, "Done with Open(%v, %v, %v), returning (%v, %v)", filePath, mode.String(), flags, fileDescriptor, err)
		return
	}

	file, isFile := node.(*File)
	ad.Assert(node != nil)
	if !isFile {
		err = filesystem.IsDirectory
		ad.Debug(ad.TRACE, "Done with Open(%v, %v, %v), returning (%v, %v)", filePath, mode.String(), flags, fileDescriptor, err)
		return
	}

	errFromFile := file.Open(mode, flags)
	if errFromFile != nil {
		fileDescriptor = -1
		err = errFromFile
		ad.Debug(ad.TRACE, "Done with Open(%v, %v, %v), returning (%v, %v)", filePath, mode.String(), flags, fileDescriptor, err)
		return
	}

	// and now to assign it to a file descriptor
	fileDescriptor = mfs.smallestAvailableFD
	mfs.activeFDs[fileDescriptor] = file

	// Maintain the invariant of smallestAvailableFD.
	for {
		_, fdIsActive := mfs.activeFDs[mfs.smallestAvailableFD]
		if fdIsActive {
			mfs.smallestAvailableFD++
			// + 3 for the reserved FDs 0, 1, and 2.
		} else if mfs.smallestAvailableFD > filesystem.MaxActiveFDs+3 {
			// Rewind the operation
			mfs.smallestAvailableFD = fileDescriptor
			fileDescriptor = -1
			err = filesystem.TooManyFDsOpen
			ad.Debug(ad.TRACE, "Done with Open(%v, %v, %v), returning (%v, %v)", filePath, mode.String(), flags, fileDescriptor, err)
			return
		} else {
			// we've found our new smallestAvailableFD
			break
		}
	}

	ad.Debug(ad.TRACE, "Done with Open(%v, %v, %v), returning (%v, %v)", filePath, mode.String(), flags, fileDescriptor, err)
	return // this is necessary for compilation, idk why
}

// See the spec for FileSystem::Close.
func (mfs *MemoryFS) Close(fileDescriptor int) (success bool, err error) {
	ad.Debug(ad.TRACE, "Closing FD %v", fileDescriptor)
	file, fdIsActive := mfs.activeFDs[fileDescriptor]
	if !fdIsActive {
		success = false
		err = filesystem.InactiveFD
		ad.Debug(ad.TRACE, "Done closing FD %v, returning (%t, %v)", fileDescriptor, success, err)
		return
	}

	success, err = file.Close()

	if success {
		ad.Assert(err == nil)
		delete(mfs.activeFDs, fileDescriptor)
	}
	// Maintain the invariant that smallestAvailableFD is actually the smallest
	if fileDescriptor < mfs.smallestAvailableFD {
		mfs.smallestAvailableFD = fileDescriptor
	}
	ad.Debug(ad.TRACE, "Done closing FD %v, returning (%t, %v)", fileDescriptor, success, err)
	return
}

// See the spec for FileSystem::Seek.
func (mfs *MemoryFS) Seek(fileDescriptor int, offset int, base filesystem.SeekMode) (newPosition int, err error) {
	file, fdIsActive := mfs.activeFDs[fileDescriptor]
	if !fdIsActive {
		return -1, filesystem.InactiveFD
	}
	newPosition, err = file.Seek(offset, base)
	ad.Debug(ad.TRACE, "FD %d seek complete - offset now %d", fileDescriptor, newPosition)
	return
}

// See the spec for FileSystem::Read.
func (mfs *MemoryFS) Read(fileDescriptor int, numBytes int) (bytesRead int, data []byte, err error) {
	file, fdIsActive := mfs.activeFDs[fileDescriptor]
	if !fdIsActive {
		return -1, make([]byte, 0), filesystem.InactiveFD
	}
	return file.Read(numBytes)
}

// See the spec for FileSystem::Write.
func (mfs *MemoryFS) Write(fileDescriptor int, numBytes int, data []byte) (bytesWritten int, err error) {
	file, fdIsActive := mfs.activeFDs[fileDescriptor]
	if !fdIsActive {
		return -1, filesystem.InactiveFD
	}
	return file.Write(numBytes, data)
}

// See the spec for FileSystem::Delete.
func (mfs *MemoryFS) Delete(filePath string) (success bool, err error) {
	ad.Debug(ad.TRACE, "Starting Delete(%v)", filePath)
	defer ad.Debug(ad.TRACE, "Done with Delete(%v)", filePath)

	currentDir, node, nodeName, existence := mfs.followPath(filePath)
	ad.Debug(ad.TRACE, "Got currentDir=%+v, node=%+v, nodeName=%v, existence=%v", currentDir, node, nodeName, existence)

	switch existence {
	case NodeExists:
		// proceed as normal
	case ParentExistsButNodeDoesNot:
		fallthrough
	case ParentDoesNotExist:
		return false, filesystem.NotFound
	}

	dir, nodeIsDirectory := node.(*Directory)
	if nodeIsDirectory && len(dir.children) > 0 {
		return false, filesystem.DirectoryNotEmpty
	}

	node.Delete()
	return true, nil
}

// Private helper methods =====================================================

// Follow a path.
// Assuming the path points to a valid Node, returns that Node, its parent, and NodeExists.
// If the parent exists and is a Directory but it has no child with the specified name, then node=nil and existence=ParentExistsButNodeDoesNot
// If the parent does not exist, parent is a File (not a Directory),  or the path is not well-formed,
// returns parentDir=nil, node=nil, and existence=ParentDoesNotExist.
// Regardless of existence, nodeName is the component of the path after the final "/".
func (mfs *MemoryFS) followPath(filePath string) (parentDir *Directory, node Node, nodeName string, existence followPathResult) {
	ad.Debug(ad.TRACE, "Following path %v", filePath)
	dirPath, nodeName := path.Split(filePath)
	if string(filePath[0]) != "/" {
		return nil, nil, nodeName, ParentDoesNotExist
	}
	dirs := strings.Split(dirPath, "/")
	dirs = dirs[1:] // Remove the empty string before the initial "/"

	currentDir := &mfs.rootDir
	// - 1 to get to the parent, we're not at the child yet
	for _, dir := range dirs[:len(dirs)-1] {
		ad.Debug(ad.TRACE, "currentDir=%+v", currentDir)
		if currentDir.HasChildNamed(dir) {
			child := currentDir.GetChildNamed(dir)
			childDir, childIsDirectory := child.(*Directory)
			// if the child is a file but the path expects it to be a directory because there are more path components
			if !childIsDirectory {
				ad.Debug(ad.TRACE, "Child named %v is not a directory", dir)
				return nil, nil, nodeName, ParentDoesNotExist
			}
			currentDir = childDir
		} else {
			ad.Debug(ad.TRACE, "Child named %v does not exist", dir)
			return nil, nil, nodeName, ParentDoesNotExist
		}
	}

	if !currentDir.HasChildNamed(nodeName) {
		ad.Debug(ad.TRACE, "Final child named %v does not exist, returning Parent exists but node does not", nodeName)
		return currentDir, nil, nodeName, ParentExistsButNodeDoesNot
	}

	ad.Debug(ad.TRACE, "Node %s exists", nodeName)
	return currentDir, currentDir.GetChildNamed(nodeName), nodeName, NodeExists
}

type followPathResult int

const (
	NodeExists followPathResult = iota
	ParentExistsButNodeDoesNot
	ParentDoesNotExist
)
