#!/usr/bin/env bash
gcc -o TCPEchoClient TCPEchoClient.c DieWithError.c
gcc -o TCPEchoServer TCPEchoServer.c DieWithError.c CreateTCPServerSocket.c AcceptTCPConnection.c HandleTCPClient.c 
