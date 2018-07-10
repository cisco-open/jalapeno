/* SPDX-License-Identifier: GPL-2.0 */
#ifndef __XDP__
#define __XDP__

#include "utils.h"

int xdp_parse(int *argc, char ***argv, struct iplink_req *req, __u32 ifindex,
	      bool generic, bool drv, bool offload);
void xdp_dump(FILE *fp, struct rtattr *tb, bool link, bool details);

#endif /* __XDP__ */
