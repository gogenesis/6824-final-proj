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
   //
   // your code looks beautiful - wonderful symmetric documentation up front
   // you rock 
   //
   // it is late, I want to collect my thoughts again in the morning...
   // but is the progression we are going for:
   //
   // add test(s) for open
   // for now, set this inode as open while locked and return
   // observe test(s) pass
   // add network partition verison of test
   // observe test(s) pass
   // add lossy network version of test
   // observe test(s) pass
   //
   // (repeat for other functions)
   //
   // then when adapting to raft version implementation
   // run test(s) for open and observe crash
   // start open on raft
   // wait for leadership loss or completed apply msg
   // if open was replicated, then and only then set this inode as open
   // else retry... and or eventually timeout / crash?
   // observe test(s) pass
   //
   // (repeat for other functions)
   // 
   // @taylor
   // 
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
