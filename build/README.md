Automating Cerana Builds Using Jenkins in a Docker Container.
=============================================================

Building Cerana requires an install of the [Nix Package Manager](https://nixos.org/nix/) and can be a repetitive task during development. To help with this a [Dockerfile](https://docs.docker.com/engine/reference/builder/) and support scripts are provided which use [Jenkins](https://jenkins.io/) to automate the build process. These can also be used to setup a nightly build server.

The Jenkins Master Container
----------------------------

**NOTE:** Building the image requires a version of Docker which supports the *ARG* instruction. It's recommended that the [latest version of Docker](https://docs.docker.com/engine/installation/) be installed.

To build the Docker image:

```
docker build --rm=true --tag cerana-jenkins:1 .
```

The Dockerfile installs Jenkins and configures it to run as the "cerana" user rather than the typical "jenkins" user. The Jenkins home directory is setup to be a volume which can be accessed from outside the running container. This is handy for both using pre-configured jobs and to retrieve build results.

To run the Docker image *cd* to the directory where you'd like the Jenkins home directory to reside and:

```
mkdir -m 777 ${PWD}/cerana
mkdir -m 777 ${PWD}/cerana/nix
docker run -p 8080:8080 -p 50000:50000 \
  -v ${PWD}/cerana:/home/cerana/.jenkins \
  -v ${PWD}/cerana/nix:/nix \
  cerana-jenkins:1
```

When running, the Jenkins server can be accessed as http://localhost:8080. **NOTE:** The first time you run Jenkins, a default administrator key will be displayed in the console output and Jenkins will prompt for this key. Once the key has been entered the key is no longer needed and Jenkins will prompt to create the admin user.

Console Mode
------------

If you prefer you can instead use the container in console mode. The command is very similar to running in server mode.

```
docker run -p 8080:8080 -p 50000:50000 \
  -v ${PWD}/cerana:/home/cerana/.jenkins \
  -v ${PWD}/cerana/nix:/nix \
  -it cerana-jenkins:1 /bin/bash
```

Once the container has started and you see the prompt you will need to initialize the Nix environment.

```
. ~/.nix-profile/etc/profile.d/nix.sh
```

All of the nix commands are now available from the command line.

Running the Build
-----------------

TBD -- a default job is being created and packaged in a tar file which can then be unpacked into the Jenkins home directory. This job will pull Cerana from github and then start the Nix package manager to run the build.

Accessing Your Build Output
---------------------------

TBD
