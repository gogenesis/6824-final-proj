#include <stdio.h>      /* for fprintf() */
#include <netdb.h>      /* for gethostbyname() */
#include <stdlib.h>     /* for exit() */

unsigned long ResolveName(char name[])
{
    struct hostent *host;            /* Structure containing host information */

    if ((host = gethostbyname(name)) == NULL)
    {
        fprintf(stderr, "gethostbyname() failed");
        exit(1);
    }

    /* Return the binary, network byte ordered address */
    return *((unsigned long *) host->h_addr_list[0]);
}
