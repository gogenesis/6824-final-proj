package tree

import "time"

// A bare implementation of the Node interface.
// Do not use this class directly; it exists to reduce code duplication between File and Directory.
type Inode struct {
	// TODO
}

func (in *Inode) Name() string {
	panic("TODO")
}

func (in *Inode) RelativePath() string {
	panic("TODO")
}

func (in *Inode) CreateTime() time.Time {
	panic("TODO")
}

func (in *Inode) Tree() Tree {
	panic("TODO")
}

func (in *Inode) Parent() Directory {
	panic("TODO")
}

func (in *Inode) Delete() {
	panic("TODO")
}

func (in *Inode) Rename(newName string) {
	panic("TODO")
}

func (in *Inode) Copy(newName string) {
	panic("TODO")
}

func (in *Inode) Open(mode OpenMode) {
	panic("TODO")
}

func (in *Inode) Close() {
	panic("TODO")
}

func (in *Inode) readLock() {
	panic("TODO")
}

func (in *Inode) readUnlock() {
	panic("TODO")
}

func (in *Inode) writeLock() {
	panic("TODO")
}

func (in *Inode) writeUnlock() {
	panic("TODO")
}
