package fsraft

import (
	"labrpc"
	"raft"
)

// A file server built on Raft.
// This is similar to the kvserver, but stores a Tree of files (see package "tree") instead of a map[string]string.
// This filesystem is linearizable; a Read() that begins after a Write() finishes is guaranteed to return the new data.
// Many enumerations are referenced in this file, like OpenMode and CloseError. So as to not clutter this file,
// those are stored in enums.go.
type FileServer struct {
	// TODO
}

// Start a FileServer.
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to form the fault-tolerant file service.
// me is the index of the current server in servers[].
func StartFileServer(servers []*labrpc.ClientEnd, me int, persister *raft.Persister, maxRaftState int) *FileServer {
	panic("TODO")
}

// Kill a FileServer.
func (fs *FileServer) Kill() {
	panic("TODO")
}

// Get a pointer to this server's Raft instance. FOR TESTING PURPOSES ONLY.
func (fs *FileServer) raft() *raft.Raft {
	panic("TODO")
}

// TODO copy in stuff from David's lab 3 when Taylor is done with lab 3B
