# OpenZFS Resiliency Gaps

This is a summary of factors that may affect the reliability of each major OpenZFS platform.

## illumos
FMA on illumos is able to respond to a large set of possible hardware faults. Resources and devices being represented in complete trees makes managing devices much more consistent on other platforms.

FMA functionality on illumos:
* Acts on resource topology (fmtopo)
* Device drivers and userspace components emit telemetry as ereports, events are logged
* Flow of ereports to list of suspect resources via diagnosis engines for each subsystem
* Diagnoses are persistent (resources used in diagnoses can be acquitted to remove them from the suspect list)
* Agents handle responses to diagnoses
* Uses device retire store to disable faulty devices

ZFS-specific FMA functionality on illumos:
* zfs-diagnosis: sees events from zfs, sd
* zfs-retire: changes vdev state in response to events, onlines hot spares
* Other nonstandard agents exist to better detect removed or flaky storage hardware
Other storage fault management:
* sd failfast: Fails certain types of I/O including retries when device is suspect to prevent latency cascades/undesirable userspace hangs

## FreeBSD
FreeBSD has unified device event reporting in the form of devctl(4), though its current scope is limited. There is no central automated fault management.

ZFS events
* ZFS events available as notifications through devctl(4) transport, handled by devd(8) which is responsible for policy
* Existing devd rules are mostly logging actions
* Incomplete zfsd project for FreeBSD 10 (stalled?)
Other storage fault management:
* CAM provides fairly reliable timeout handling, no automated actions taken
* No devctl or other events reported from CAM other than device add/remove

## Linux
Linux has events available from some subsystems, but no consistent interface for these. There is no central automated fault management.

ZFS event daemon (ZED) on Linux:
* Event loop driven by /dev/zfs ioctls (ZFS_IOC_EVENTS_NEXT)
* zedlets as agents to respond to events: change vdev state, log to syslog or email or SNMP, online hot spares

Other storage fault management:
* scsi failfast: Disables retries when device is suspect due to I/O timeouts
* No device driver level events available

Other major functional gaps for ZFS:
* ZFS delegation
 - `zfs allow` for user administrative is supported and can be applied to user namespaces where UIDs have been mapped to a different range
 - `zoned` property to delegate datasets to a container is unsupported since Linux containers are just a collection of namespaces. Solution may require creating a new namespace for ZFS datasets.
  - https://github.com/zfsonlinux/zfs/issues/3159

# Bad stuff that can happen
## Software faults

Types of data corruption:
* Block pointer corruption
  - Usually invalid vdev number in DVA, seems to occur equally often on all platforms as a side effect of other bugs
* ZAP object corruption
  - Linux: Heavy spill block use triggered some issues previously.
* Space accounting: Spacemap out of sync with reachable block pointers
  - Possible to repair automatically in theory (offline)
* vdev label corruption: Out of sync vdev labels on pool component devices
  - Usually requires manual intervention

Full ZFS Test Suite based on STF only available on illumos; `ztest` available for all ports.

http://open-zfs.org/wiki/Platform_code_differences is handy for spotting what failures might occur differently on each platform.

Bugs:

* illumos:
* FreeBSD:
* Linux
 - Recently addressed several issues with SA use causing spill block corruption.
 - History of deadlocks from incorrectly invoking non-reentrant ARC code, known issues have been addressed.
 - May not be able to recover if first two labels corrupted.

## Hardware faults
* Single drive experiencing temporary high latency
  - illumos: short I/O timeouts can lead to failfast condition, limiting retries (good) and causing ZFS to accumulate I/O errors. Latency tracking is possible, though doesn’t exist upstream outside of timeouts so no solution for persistent device that’s just slow relative to other pool members
  - FreeBSD: CAM has I/O latency tracking but no automated response
  - Linux: failfast can temporarily limit retries to a suspect device after I/O timeouts. Allows ZFS to detect timeouts as I/O errors, similar to on illumos
* Single drive is unresponsive
  - illumos: sd gives up after a number of retries; the device will be retired eventually, at which point ZFS will also offline the vdev. Userspace tools accessing the device will hang until device retired. However, accessing the device tree can cause rescans by the controller which is not covered by the retire, and block whatever triggered the scan (fmd itself, in some cases).
  - FreeBSD: CAM differentiates between ‘selection’ and command timeouts, though this seems to be specific to whether the HBA thinks the target exists. Commands are still subject to the same retry count and there’s no way to disable I/O to a device. 
  - Linux: Cumulative I/O errors should cause ZFS to take the associated vdev offline. Block devices can be offlined (manually) through sysfs device state. Doesn’t appear most I/O or device enumeration paths lead to automatic rescan.
* Multiple drives are experiencing temporary high latency
  - illumos: same as in 1a.
  - FreeBSD: same as in 1b.
  - Linux:
* Multiple drives are unresponsive
  - illumos: cascading latency is possible while multiple devices have outstanding I/O before retire takes place. Retire itself may hang while probe I/Os time out.
  - FreeBSD: Individual timeouts can lead to cascading latency. No retire mechanism in place, but fewer rescans too.
  - Linux: Individual timeouts can lead to cascading latency. No retire mechanism in place, but fewer rescans too. Rescans don’t block device enumeration.
* Drive or controller fails I/O
  - All: I/O errors reported back to consumer (ZFS), ZFS bumps error count
* Drive disconnects
  - illumos: sysevent leads to device removal and ZFS offlining
  - FreeBSD: devd receives device removed event. ZFS isn’t consuming these events directly but I/O will fail.
  - Linux: I/O will fail.
* Bus errors affecting storage controller
  - illumos: FMA may retire the controller if rate of errors is too high. If multipathing is in use, failover will occur.
  - FreeBSD: No automated action. pciconf(8) can track PCI bus errors
  - Linux: https://www.kernel.org/doc/Documentation/PCI/pci-error-recovery.txt (out of date, API details have changed slightly) defines notification callback for drivers to attempt to recover from errors. No persistent record of errors and no userspace notification.
* Data corruption due to uncorrected bit flips in buffer/transport
  - All: Not detected by storage subsystems, detected by ZFS on read
* Data corruption due to phantom or misdirected writes in drive
  - All: Not detected by storage subsystems, detected by ZFS on read


Thanks to Matt Ahrens and Richard Yao for feedback on a draft version of this document.
