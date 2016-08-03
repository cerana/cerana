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

If you prefer you can override the Jenkins UID and GID to match yours. This simplifies accessing the directories shared with the container. This also helps maintain job configurations outside of Jenkins.

```
docker build --rm=true --tag cerana-jenkins:1 \
  --build-arg UID=`id -u` --build-arg GID=`id -g` .
```

To run the Docker image *cd* to the directory where you'd like the Jenkins home directory and the Nix store to reside and create directories to be shared. If you are running Jenkins as your user then the permissions do not need to be modified (i.e. do not use `-m 777`).

```
mkdir -p -m 777 ${PWD}/cerana/{jenkins_home,nix}
```

Then to run Jenkins:

```
docker run -p 8080:8080 -p 50000:50000 \
  -v ${PWD}/cerana/jenkins_home:/home/cerana \
  -v ${PWD}/cerana/nix:/nix \
  --name cerana-jenkins \
  cerana-jenkins:1
```

When running, the Jenkins server can be accessed as http://localhost:8080. **NOTE:** The first time you run Jenkins, a default administrator key will be displayed in the console output and Jenkins will prompt for this key. Once the key has been entered the key is no longer needed and Jenkins will prompt to create the admin user.

Installing Plugins
------------------

The first time Jenkins runs it will prompt to install plugins. It's recommended the default list be used. Plugins can be added or removed later.

Daemon Mode
-----------

In most cases you will want to run Jenkins in the background. e.g.:

```
docker run -d -p 8080:8080 -p 50000:50000 \
  -v ${PWD}/cerana/jenkins_home:/home/cerana \
  -v ${PWD}/cerana/nix:/nix \
  --name cerana-jenkins \
  cerana-jenkins:1
```

Because the console output won't be visible it's necessary to get the default administrator key from the log file.

```
docker logs cerana-jenkins
```

Installing the Default Jobs
---------------------------

If you want you can now install the default jobs into the Jenkins home directory. Simply copy the default jobs from the `cerana/build` directory to the Jenkins `jobs` directory.

**NOTE:** It's necessary to do this step after starting Jenkins because the job may require one or more of the plugins which are installed the first time Jenkins is run.

```
cp -r <ceranaGitDirectory>/build/jobs/* ${PWD}/cerana/jenkins_home/.jenkins/jobs/
```

To activate the jobs you need to tell Jenkins to reload the configuration files (Manage Jenkins -> Reload Configuration from Disk).

Connect to a Running Jenkins Container
--------------------------------------

At times you may want to check the state of a running container. Here's an example for how to connect:

```
docker exec -it cerana-jenkins /bin/bash -i -l
```

NOTE: The `-l` is the letter `l` -- not the number `1`.

If you started in daemon mode using the above example replace `<containerId>` with `cerana-jenkins`. Otherwise you'll need to determine what the id is using `docker ps`.

Starting in Console Mode
------------------------

If you prefer you can instead use the container in console mode. The command is very similar to running in server mode.

```
docker run -p 8080:8080 -p 50000:50000 \
  -v ${PWD}/cerana/jenkins_home:/home/cerana \
  -v ${PWD}/cerana/nix:/nix \
  -it cerana-jenkins:1 /bin/bash
```

Once the container has started and you see the prompt you will need to initialize the Nix environment.

```
. ~/.nix-profile/etc/profile.d/nix.sh
```

All of the nix commands are now available from the command line.

### The Default Jobs

#### build-cerana

This job builds any time it detects a push to the `ceranaos` branch of [cerana/nixpkgs](https://github.com/cerana/nixpkgs).

#### cerana-nixpkgs

This job is designed to build a local clone of [cerana/nixpkgs](https://github.com/cerana/nixpkgs). It will build the active branch. **NOTE:** More information on how to use this job as part of your development workflow is provided in the job description. If you've been following the install instructions then you can read the job description [here](http://localhost:8080/job/cerana-nixpkgs/).

#### collect-garbage

As the number of different builds increases, the Nix store (`/nix/store`) will eventually grow to occupy all available disk space. To help mitigate this problem the `collect-garbage` job is designed to run the [Nix garbage collection](https://nixos.org/releases/nixos/14.12/nixos-14.12.374.61adf9e/manual/sec-nix-gc.html) periodically. This first checks to see if the size of the store has grown beyond a configurable threshold. The job then triggers the `build-cerana` job to ensure the Nix store is current.

**NOTE:** If you have a slow network connection you may encounter one or more timeouts when the job initially runs. This is because some fairly large components are downloaded from the repository and Jenkins is configured to timeout after 10 minutes by default and *Nix* may also timeout. Because of this it may be necessary to restart the job one or more times. Once the first build finishes this will no longer be a problem.

Accessing Your Build Output
---------------------------

The build output (e.g. kernel image and ram disk image) can be downloaded using the Jenkins web interface. If you've been using the above examples the URL is: http://localhost:8080/job/<job>/ws/result/

If you've been following the above instructions then the build output can also be accessed directly in `cerana/jenkins_home/.jenkins/workspace/<job>/artifacts`. This is especially handy when using the `cerana-nixpkgs` job.

Testing Using Virtual Machines
------------------------------

Testing Cerana clusters often requires two or more nodes in order to verify the interaction of different Cerana components. Using virtual machines is the easiest and quickest way to setup multi-node systems for testing. Helper scripts for testing using virtual machines in a Linux environment are provided in the `scripts` directory. Their use is described in this [README](scripts/README.md).
