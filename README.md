# 6824-final-proj
David and Taylor's 6.824 Final Project - Simple Synchronized Distributed Filesystem

The purpose of this project is to solve a hard distributed systems problem.

We implement a tree metadata structure across Raft, a replicated state machine written in Golang. We expose a client API that provides a subset of 64-bit POSIX file operations, and implment a simple FUSE driver to allow any application to access the filesystem through a standard Linux mount point.

-Priorities
...

