Mistify-OS SDK
==============

Mistify-OS can be used to build Mistify-OS.

Building the Mistify-OS SDK Variant
-----------------------------------

Building the SDK variant of the Mistify-OS can be as simple as:

```
./buildmistify --variant sdk
```

**NOTE:** The introduction of Mistify-OS SDK version required a change to the toolchain in order to support *multilib* for building both 64 and 32 bit binaries. [Crosstool-NG](http://crosstool-ng.org/) doesn't yet implement a complete *multilib* solution. Because of this the SDK requires using the [mistifyio](https://github.com/mistifyio) fork of the toolchain located at: [mistifyio/crosstool-ng](https://github.com/mistifyio/crosstool-ng). This is now the default for the *buildmistify* script. **If you've been building prior versions of Mistify-OS you will need to specify the toolchain URL, version and a new location.** e.g:

```
./buildmistify --variant sdk
--toolchaindir <toolchaindir>
--tcuri git@github.com:mistifyio/crosstool-ng.git
--toolchainversion glibc-multilib-sdk
```

Runtime System Requirements
---------------------------

Building Mistify-OS and especially the SDK version of Mistify-OS is very resource intensive compared to just running Mistify-OS. This is because of supporting full development toolchains. Like with the base variant the SDK root file system is RAM based but is much larger also because of including a complete toolchain and supporting utilities. The minimum system requirements for running the SDK version of Mistiy-OS are listed below.

-	6GB RAM 1GB of which is for the root file system RAM disk.
-	25GB of available harddrive space. This is sufficient to support a single build. Each additional build which shares toolchains and downloads requires an additional 10GB.
-	A network configuration supporting access to the Internet which is needed for downloading packages to be built.

Booting the Mistify-OS SDK
--------------------------

Nightly builds of the SDK are available for download from:[SDK Artifacts](http://omniti-mystify-artifacts.s3.amazonaws.com/index.html?prefix=jobs/SDK-Container-Build/). These include a kernel and *initrd* image as well as an ISO image. These can be booted in a manner identical to the standard version of Mistify-OS. Therefore the methods are not described here except to note the kernel command line needs to be adjusted to accommodate the larger RAM disk requirements.

These images can be used to boot either in a virtual machine or on actual hardware.

Once booted the SDK version identifies itself on the default console by displaying the banner string "Welcome to Mistify-OS -- SDK".
