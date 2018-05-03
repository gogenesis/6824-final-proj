package memoryFS

import "ad"

// An abstraction of a File or a Directory.
// Note that Node is implemented by *File and *Directory,
// NOT by File or Directory.
type Node interface {
	// The name of this Node.
	Name() string

	// Delete this Node.
	// success == true iff err == nil.
	Delete() (success bool, err error)
}

// An "abstract class" to hold shared implementations of the functions in Node.
// Like File and Directory, *Inode implements Node but Inode (no pointer) does not.
type Inode struct {
	name string
}

func (in *Inode) Name() string {
	return in.name
}

func (in *Inode) Delete() (success bool, err error) {
	ad.Debug(ad.TRACE, "Deleting %v", in.Name())
	// nothing else necessary
	return
}
