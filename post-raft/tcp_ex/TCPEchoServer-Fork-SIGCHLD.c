#include "TCPEchoServer.h"  /* TCP echo server includes */
#include <sys/wait.h>       /* for waitpid() */
#include <signal.h>         /* for sigaction() */

void ChildExitSignalHandler();     /* Function to clean up zombie child processes */

/* Global so accessable by SIGCHLD signal handler */
unsigned int childProcCount = 0;   /* Number of child processes */

int main(int argc, char *argv[])
{
    int servSock;                    /* Socket descriptor for server */
    int clntSock;                    /* Socket descriptor for client */
    unsigned short echoServPort;     /* Server port */
    pid_t processID;                 /* Process ID from fork() */
    struct sigaction myAction;       /* Signal handler specification structure */
 
    if (argc != 2)     /* Test for correct number of arguments */
    {
        fprintf(stderr, "Usage:  %s <Server Port>\n", argv[0]);
        exit(1);
    }

    echoServPort = atoi(argv[1]);  /* First arg:  local port */

    servSock = CreateTCPServerSocket(echoServPort);

    /* Set ChildExitSignalHandler() as handler function */
    myAction.sa_handler =  ChildExitSignalHandler;
    if (sigfillset(&myAction.sa_mask) < 0)   /* mask all signals */
        DieWithError("sigfillset() failed");
    /* SA_RESTART causes interrupted system calls to be restarted */
    myAction.sa_flags = SA_RESTART;

    /* Set signal disposition for child-termination signals */
    if (sigaction(SIGCHLD, &myAction, 0) < 0)
        DieWithError("sigaction() failed");

    for (;;) /* run forever */
    {
	clntSock = AcceptTCPConnection(servSock);
        /* Fork child process and report any errors */
        if ((processID = fork()) < 0)
            DieWithError("fork() failed");
        else if (processID == 0)  /* If this is the child process */
        {
            close(servSock);   /* Child closes parent socket file descriptor */
            HandleTCPClient(clntSock);
            exit(0);              /* Child process done */
        }

	printf("with child process: %d\n", (int) processID);
        close(clntSock);       /* Parent closes child socket descriptor */
        childProcCount++;      /* Increment number of outstanding child processes */
    }
    /* NOT REACHED */
}

void ChildExitSignalHandler()
{
    pid_t processID;           /* Process ID from fork() */

    while (childProcCount) /* Clean up all zombies */
    {
	processID = waitpid((pid_t) -1, NULL, WNOHANG);  /* Non-blocking wait */
	if (processID < 0)  /* waitpid() error? */
	    DieWithError("waitpid() failed");
	else if (processID == 0)  /* No child to wait on */
	    break;
	else
	    childProcCount--;  /* Cleaned up after a child */
    }
}
