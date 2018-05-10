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

	// The parent of this Node.
	// Nil if and only if this is the root directory.
	Parent() *Directory
}

// An "abstract class" to hold shared implementations of the functions in Node.
// Like File and Directory, *Inode implements Node but Inode (no pointer) does not.
type Inode struct {
	name   string
	parent *Directory
}

func (in *Inode) Name() string {
	return in.name
}

func (in *Inode) Delete() (success bool, err error) {
	ad.Debug(ad.TRACE, "Deleting %v", in.Name())
	// If this was the root directory, this would have been stopped earlier
	ad.Assert(in.parent != nil)
	delete(in.Parent().children, in.Name())
	return
}

func (in *Inode) Parent() *Directory {
	return in.parent
}
