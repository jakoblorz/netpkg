# netpkg
Send a request to netpkg to invoke a command remotely. Hook applications to the network to automate workflows that previously were only automatable with custom solutions.

Usage
=========

```
Usage of netpkg:
  -c string
        name of the program to execute (default "sh")
  -h string
        specify the host (default "0.0.0.0")
  -p int
        specify the port (default 8000)
  -t string
        secure the api with a token; set 'n' if no token is required
```

- Additional arguments will be passed to the defined command (`stdin`):
    ```bash
    netpkg -c echo hello world # will output "hello world" when invoked
    netpkg -c docker run hello-world # will run the "hello-world" image when invoked
    ```

- Output will be piped back (`stdout`):
    ```bash
    netpkg -c docker run hello-world

    # when invoked:

    # Hello from Docker!
    # This message shows that your installation appears to be working correctly.

    # To generate this message, Docker took the following steps:
    #  1. The Docker client contacted the Docker daemon.
    #  2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    #     (amd64)
    #  3. The Docker daemon created a new container from that image which runs the
    #     executable that produces the output you are currently reading.
    #  4. The Docker daemon streamed that output to the Docker client, which sent it
    #     to your terminal.

    # To try something more ambitious, you can run an Ubuntu container with:
    #  $ docker run -it ubuntu bash

    # Share images, automate workflows, and more with a free Docker ID:
    #  https://cloud.docker.com/

    # For more examples and ideas, visit:
    #  https://docs.docker.com/engine/userguide/

    ```

Examples
=========

**You want to run the CD for your CI/CD Pipeline...**

...but you need to have it done now - just script it!

1. Continuous Deployment Script
    ```bash
    # continuous.sh (on the target server)

    # pull new changes from production branch and build' em
    git pull origin production && docker build . -t awesome-application

    # switch out the running application
    docker stop $(docker ps -q --filter name=awesome-application)
    docker run --name awesome-application -d awesome-application
    ```

2. Start `netpkg` to have the script invoked
    ```bash
    netpkg -t awesome -p 4312 -c sh ./continuous.sh
    ```

3. Configure your CI to trigger the `netpkg` endpoint

    - Send a request via http `curl -s http://your-server:4312?token=awesome`
    - Send a request via tcp `echo awesome | nc -N your-server 4312`
    - Sample configuration for GitLab-CI:
    ```yaml
    "Trigger Cloud Rebuild":
      image: alpine:3.8
      stage: deploy
      only: 
        - production
      script:
        - apk add --no-cache curl 
        - curl -s http://your-server:4312?token=awesome

    ```

Getting Help
=========
You can [file an issue](https://github.com/jakoblorz/netpkg/issues/new) to ask questions, request features, 
or ask for help.

Licensing
=========
netpkg is licensed under the MIT License. See
[LICENSE](https://github.com/jakoblorz/netpkg/blob/master/LICENSE) for the full
license text.