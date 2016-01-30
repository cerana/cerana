#include <errno.h>
#include <limits.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/types.h>

#define ZFS_IOC_LIBZFS_CORE (('Z' << 8) + 0x100 + 1)
struct zfs_cmd_stable {
	char		zc_name[PATH_MAX]; 	/* name of pool or dataset */
	uint64_t	zc_nvlist_src;		/* really (char *) */
	uint64_t	zc_nvlist_src_size;
	uint64_t	zc_nvlist_dst;		/* really (char *) */
	uint64_t	zc_nvlist_dst_size;
	bool		zc_nvlist_dst_filled;	/* put an nvlist in dst? */
	int32_t		zc_real_err;		/* true error is passed here */
};

typedef struct zfs_cmd {
	struct zfs_cmd_stable;
	char _legacy_zfs_cmd_fields[(14 * 1024) - sizeof(struct zfs_cmd_stable)];
} zfs_cmd_t;

int zfs_ioctl(int fd, char *name, int len, void *innvl, int insize, void *outnvl, int outsize) {
	zfs_cmd_t cmd;
	memset(&cmd, 0, sizeof(cmd));
	memcpy(cmd.zc_name, name, len);
	cmd.zc_nvlist_src = (uint64_t)innvl;
	cmd.zc_nvlist_src_size = insize;
	cmd.zc_nvlist_dst = (uint64_t)outnvl;
	cmd.zc_nvlist_dst_size = outsize;
	cmd.zc_nvlist_dst_filled = outnvl != NULL;

	int err = ioctl(fd, ZFS_IOC_LIBZFS_CORE, (unsigned long)&cmd);
	if (err = 0) {
		errno = cmd.zc_real_err;
	}
	return err;
}
