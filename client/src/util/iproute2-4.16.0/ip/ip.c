/*
 * ip.c		"ip" utility frontend.
 *
 *		This program is free software; you can redistribute it and/or
 *		modify it under the terms of the GNU General Public License
 *		as published by the Free Software Foundation; either version
 *		2 of the License, or (at your option) any later version.
 *
 * Authors:	Alexey Kuznetsov, <kuznet@ms2.inr.ac.ru>
 */

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <string.h>
#include <errno.h>

#include "SNAPSHOT.h"
#include "utils.h"
#include "ip_common.h"
#include "namespace.h"
#include "color.h"

int preferred_family = AF_UNSPEC;
int human_readable;
int use_iec;
int show_stats;
int show_details;
int oneline;
int brief;
int json;
int timestamp;
const char *_SL_;
int force;
int max_flush_loops = 10;
int batch_mode;
bool do_all;

struct rtnl_handle rth = { .fd = -1 };

static void usage(void) __attribute__((noreturn));

static void usage(void)
{
	fprintf(stderr,
"Usage: ip [ OPTIONS ] OBJECT { COMMAND | help }\n"
"       ip [ -force ] -batch filename\n"
"where  OBJECT := { link | address | addrlabel | route | rule | neigh | ntable |\n"
"                   tunnel | tuntap | maddress | mroute | mrule | monitor | xfrm |\n"
"                   netns | l2tp | fou | macsec | tcp_metrics | token | netconf | ila |\n"
"                   vrf | sr }\n"
"       OPTIONS := { -V[ersion] | -s[tatistics] | -d[etails] | -r[esolve] |\n"
"                    -h[uman-readable] | -iec |\n"
"                    -f[amily] { inet | inet6 | ipx | dnet | mpls | bridge | link } |\n"
"                    -4 | -6 | -I | -D | -B | -0 |\n"
"                    -l[oops] { maximum-addr-flush-attempts } | -br[ief] |\n"
"                    -o[neline] | -t[imestamp] | -ts[hort] | -b[atch] [filename] |\n"
"                    -rc[vbuf] [size] | -n[etns] name | -a[ll] | -c[olor]}\n");
	exit(-1);
}

static int do_help(int argc, char **argv)
{
	usage();
	return 0;
}

static const struct cmd {
	const char *cmd;
	int (*func)(int argc, char **argv);
} cmds[] = {
	{ "address",	do_ipaddr },
	{ "addrlabel",	do_ipaddrlabel },
	{ "maddress",	do_multiaddr },
	{ "route",	do_iproute },
	{ "rule",	do_iprule },
	{ "neighbor",	do_ipneigh },
	{ "neighbour",	do_ipneigh },
	{ "ntable",	do_ipntable },
	{ "ntbl",	do_ipntable },
	{ "link",	do_iplink },
	{ "l2tp",	do_ipl2tp },
	{ "fou",	do_ipfou },
	{ "ila",	do_ipila },
	{ "macsec",	do_ipmacsec },
	{ "tunnel",	do_iptunnel },
	{ "tunl",	do_iptunnel },
	{ "tuntap",	do_iptuntap },
	{ "tap",	do_iptuntap },
	{ "token",	do_iptoken },
	{ "tcpmetrics",	do_tcp_metrics },
	{ "tcp_metrics", do_tcp_metrics },
	{ "monitor",	do_ipmonitor },
	{ "xfrm",	do_xfrm },
	{ "mroute",	do_multiroute },
	{ "mrule",	do_multirule },
	{ "netns",	do_netns },
	{ "netconf",	do_ipnetconf },
	{ "vrf",	do_ipvrf},
	{ "sr",		do_seg6 },
	{ "help",	do_help },
	{ 0 }
};

static int do_cmd(const char *argv0, int argc, char **argv)
{
	const struct cmd *c;

	for (c = cmds; c->cmd; ++c) {
		if (matches(argv0, c->cmd) == 0)
			return -(c->func(argc-1, argv+1));
	}

	fprintf(stderr, "Object \"%s\" is unknown, try \"ip help\".\n", argv0);
	return EXIT_FAILURE;
}

static int batch(const char *name)
{
	char *line = NULL;
	size_t len = 0;
	int ret = EXIT_SUCCESS;
	int orig_family = preferred_family;

	batch_mode = 1;

	if (name && strcmp(name, "-") != 0) {
		if (freopen(name, "r", stdin) == NULL) {
			fprintf(stderr,
				"Cannot open file \"%s\" for reading: %s\n",
				name, strerror(errno));
			return EXIT_FAILURE;
		}
	}

	if (rtnl_open(&rth, 0) < 0) {
		fprintf(stderr, "Cannot open rtnetlink\n");
		return EXIT_FAILURE;
	}

	cmdlineno = 0;
	while (getcmdline(&line, &len, stdin) != -1) {
		char *largv[100];
		int largc;

		preferred_family = orig_family;

		largc = makeargs(line, largv, 100);
		if (largc == 0)
			continue;	/* blank line */

		if (do_cmd(largv[0], largc, largv)) {
			fprintf(stderr, "Command failed %s:%d\n",
				name, cmdlineno);
			ret = EXIT_FAILURE;
			if (!force)
				break;
		}
	}
	if (line)
		free(line);

	rtnl_close(&rth);
	return ret;
}


int main(int argc, char **argv)
{
	char *basename;
	char *batch_file = NULL;
	int color = 0;

	/* to run vrf exec without root, capabilities might be set, drop them
	 * if not needed as the first thing.
	 * execv will drop them for the child command.
	 * vrf exec requires:
	 * - cap_dac_override to create the cgroup subdir in /sys
	 * - cap_sys_admin to load the BPF program
	 * - cap_net_admin to set the socket into the cgroup
	 */
	if (argc < 3 || strcmp(argv[1], "vrf") != 0 ||
			strcmp(argv[2], "exec") != 0)
		drop_cap();

	basename = strrchr(argv[0], '/');
	if (basename == NULL)
		basename = argv[0];
	else
		basename++;

	while (argc > 1) {
		char *opt = argv[1];

		if (strcmp(opt, "--") == 0) {
			argc--; argv++;
			break;
		}
		if (opt[0] != '-')
			break;
		if (opt[1] == '-')
			opt++;
		if (matches(opt, "-loops") == 0) {
			argc--;
			argv++;
			if (argc <= 1)
				usage();
			max_flush_loops = atoi(argv[1]);
		} else if (matches(opt, "-family") == 0) {
			argc--;
			argv++;
			if (argc <= 1)
				usage();
			if (strcmp(argv[1], "help") == 0)
				usage();
			else
				preferred_family = read_family(argv[1]);
			if (preferred_family == AF_UNSPEC)
				invarg("invalid protocol family", argv[1]);
		} else if (strcmp(opt, "-4") == 0) {
			preferred_family = AF_INET;
		} else if (strcmp(opt, "-6") == 0) {
			preferred_family = AF_INET6;
		} else if (strcmp(opt, "-0") == 0) {
			preferred_family = AF_PACKET;
		} else if (strcmp(opt, "-I") == 0) {
			preferred_family = AF_IPX;
		} else if (strcmp(opt, "-D") == 0) {
			preferred_family = AF_DECnet;
		} else if (strcmp(opt, "-M") == 0) {
			preferred_family = AF_MPLS;
		} else if (strcmp(opt, "-B") == 0) {
			preferred_family = AF_BRIDGE;
		} else if (matches(opt, "-human") == 0 ||
			   matches(opt, "-human-readable") == 0) {
			++human_readable;
		} else if (matches(opt, "-iec") == 0) {
			++use_iec;
		} else if (matches(opt, "-stats") == 0 ||
			   matches(opt, "-statistics") == 0) {
			++show_stats;
		} else if (matches(opt, "-details") == 0) {
			++show_details;
		} else if (matches(opt, "-resolve") == 0) {
			++resolve_hosts;
		} else if (matches(opt, "-oneline") == 0) {
			++oneline;
		} else if (matches(opt, "-timestamp") == 0) {
			++timestamp;
		} else if (matches(opt, "-tshort") == 0) {
			++timestamp;
			++timestamp_short;
		} else if (matches(opt, "-Version") == 0) {
			printf("ip utility, iproute2-ss%s\n", SNAPSHOT);
			exit(0);
		} else if (matches(opt, "-force") == 0) {
			++force;
		} else if (matches(opt, "-batch") == 0) {
			argc--;
			argv++;
			if (argc <= 1)
				usage();
			batch_file = argv[1];
		} else if (matches(opt, "-brief") == 0) {
			++brief;
		} else if (matches(opt, "-json") == 0) {
			++json;
		} else if (matches(opt, "-rcvbuf") == 0) {
			unsigned int size;

			argc--;
			argv++;
			if (argc <= 1)
				usage();
			if (get_unsigned(&size, argv[1], 0)) {
				fprintf(stderr, "Invalid rcvbuf size '%s'\n",
					argv[1]);
				exit(-1);
			}
			rcvbuf = size;
		} else if (matches(opt, "-color") == 0) {
			++color;
		} else if (matches(opt, "-help") == 0) {
			usage();
		} else if (matches(opt, "-netns") == 0) {
			NEXT_ARG();
			if (netns_switch(argv[1]))
				exit(-1);
		} else if (matches(opt, "-all") == 0) {
			do_all = true;
		} else {
			fprintf(stderr,
				"Option \"%s\" is unknown, try \"ip -help\".\n",
				opt);
			exit(-1);
		}
		argc--;	argv++;
	}

	_SL_ = oneline ? "\\" : "\n";

	if (color && !json)
		enable_color();

	if (batch_file)
		return batch(batch_file);

	if (rtnl_open(&rth, 0) < 0)
		exit(1);

	if (strlen(basename) > 2)
		return do_cmd(basename+2, argc, argv);

	if (argc > 1)
		return do_cmd(argv[1], argc-1, argv+1);

	rtnl_close(&rth);
	usage();
}
