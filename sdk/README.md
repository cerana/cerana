Mistify-OS SDK
==============

Mistify-OS can be used to build Mistify-OS. This is a fully functional Mistify-OS with the addition of tools needed to build Mistify-OS or even the SDK version of Mistify-OS. The SDK version also adds tools needed to manage multiple users making it possible for multiple users to use their existing credentials for sites such as http://github.com share the same Mistify-OS environment for development.

**NOTE:** This document assumes the reader is familiar with Linux system administration.

Building the Mistify-OS SDK Variant
-----------------------------------

Building the SDK variant of the Mistify-OS can be as simple as:

```
./buildmistify --variant sdk
```

### If You've Already Been Building Mistify-OS

**NOTE:** The introduction of Mistify-OS SDK version required a change to the toolchain in order to support *multilib* for building both 64 and 32 bit binaries. [Crosstool-NG](http://crosstool-ng.org/) doesn't yet implement a complete *multilib* solution. Because of this the SDK requires using the [mistifyio](https://github.com/mistifyio) fork of the toolchain located at: [mistifyio/crosstool-ng](https://github.com/mistifyio/crosstool-ng). This is now the default for the *buildmistify* script and applies to both the *sdk* and *base* variations of Mistify-OS. **If you've been building prior versions of Mistify-OS you will need to specify the toolchain URL, version and a new location.** e.g:

```
./buildmistify --variant sdk \
--toolchaindir <toolchaindir> \
--tcuri git@github.com:mistifyio/crosstool-ng.git \
--toolchainversion glibc-multilib-sdk
```

If you want to continue using the same location for the toolchain then simply remove the toolchain directory.

Runtime System Requirements
---------------------------

Building Mistify-OS and especially the SDK version of Mistify-OS is very resource intensive compared to just running Mistify-OS. This is because of supporting full development toolchains. Like with the base variant the SDK root file system is RAM based but is much larger also because of including a complete toolchain and supporting utilities. The minimum system requirements for running the SDK version of Mistiy-OS are listed below.

-	6GB RAM 1GB of which is needed for the root file system RAM disk.
-	25GB of available harddrive space. This is sufficient to support a single build. Each additional build which shares toolchains and downloads requires an additional 10GB.
-	A network configuration supporting access to the Internet which is needed for downloading packages to be built.
-	And of course, anything needed to run Mistify-OS itself.

Booting the Mistify-OS SDK
--------------------------

Nightly builds of the SDK are available for download from: [SDK Artifacts](http://omniti-mystify-artifacts.s3.amazonaws.com/index.html?prefix=jobs/SDK-Container-Build/). These include a kernel and *initrd* image as well as an ISO image. These can be booted in a manner identical to the standard version of Mistify-OS. Therefore the methods are not described here except to note the kernel command line needs to be adjusted to accommodate the larger RAM disk requirements.

These images can be used to boot either in a virtual machine or on actual hardware.

Once booted the SDK version identifies itself on the default console by displaying the banner string "Welcome to Mistify-OS -- SDK".

Making Space
------------

A build of Mistify-OS can require as much as 25GB of disk space. Mistify-OS by default allocates only 4GB to user home directories. Since users will likely be building within their home directory, it's recommended the ZFS quota for */mistify/private* be increased to at least 25GB per user. Using pre-built toolchains and sharing other directories can reduce this amount. **NOTE:** If a user will be maintaining more than one build then this space needs to be increased accordingly.

For example, to allocate space for two users to do complete builds or for one user to manage two separate builds:

```bash
zfs set quota=50GB /mistify/private
```

Managing users
--------------

Once the SDK version is running users can be created. The SDK includes the *adduser* utility. It's recommended the user's *ssh* credentials (*~/.ssh*) be copied to the user's home directory.

### SUDO

Some users may need root access. This can be setup creating a configuration file for the user in */etc/suders.d* file. Here's an example for full access with no password:

```
<user> ALL=(ALL) NOPASSWD:ALL
```

**NOTE:** Having root access is not needed for building Mistify-OS. Configuring *sudo* is recommended only for users who may also need to administer the SDK version of Mistiy-OS. For example to install additional tools.

Local Installs
==============

In the interest of keeping Mistify-OS SDK build times and RAM disk size to a minimum and different developers will have different requirements, many tools which may be of use to a developer are not included in the standard SDK. Instead, the SDK is capable of building many of those tools. To make these persistent across reboots mounting and installing into */usr/local* is recommended. One method of doing this is to simply create a symlink pointing to a persistent directory under */mistify/private*. Here's and example of how to do this:

```bash
sudo mkdir -p /mistify/private/usr/local
sudo ln -s /mistify/private/usr/local /usr/local
```

**NOTE:** Using this method the symbolic link is not persistent across reboots and needs to be restored on boot. An additional script in */etc/pre-init.d* can be used to do this.

Another method is to add a script to */etc/pre-init.d* which uses *aufs* to mount the directory. The existing script */etc/pre-init.d/mount-aufs.sh* can be used as an example of how to do this.

Installing *clang*
------------------

By default *gcc* is included in the SDK and is used to bootstrap a Mistify-OS build. An alternate compiler named *clang* is sometimes useful in some debugging scenarios. This section describes how to install *clang* in the local environment (as a normal user with *sudo* enabled).

```bash
wget http://llvm.org/releases/3.7.0/llvm-3.7.0.src.tar.xz  
wget http://llvm.org/releases/3.7.0/cfe-3.7.0.src.tar.xz  
wget http://llvm.org/releases/3.7.0/compiler-rt-3.7.0.src.tar.xz
tar xf llvm-3.7.0.src.tar.xz
mv llvm-3.7.0.src llvm-3.7.0
cd llvm-3.7.0
tar xf ../cfe-3.7.0.src.tar.xz -C tools
tar xf ../compiler-rt-3.7.0.src.tar.xz -C projects
mv tools/cfe-3.7.0.src tools/clang
mv projects/compiler-rt-3.7.0.src projects/compiler-rt
mkdir build
cd build
../configure --prefix=/usr/local --enable-shared
make
sudo make install
```
