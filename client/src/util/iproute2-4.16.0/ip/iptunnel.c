/*
 * iptunnel.c	       "ip tunnel"
 *
 *		This program is free software; you can redistribute it and/or
 *		modify it under the terms of the GNU General Public License
 *		as published by the Free Software Foundation; either version
 *		2 of the License, or (at your option) any later version.
 *
 * Authors:	Alexey Kuznetsov, <kuznet@ms2.inr.ac.ru>
 *
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <sys/ioctl.h>
#include <net/if.h>
#include <net/if_arp.h>
#include <linux/ip.h>
#include <linux/if_tunnel.h>

#include "rt_names.h"
#include "utils.h"
#include "ip_common.h"
#include "tunnel.h"

static void usage(void) __attribute__((noreturn));

static void usage(void)
{
	fprintf(stderr, "Usage: ip tunnel { add | change | del | show | prl | 6rd } [ NAME ]\n");
	fprintf(stderr, "          [ mode { ipip | gre | sit | isatap | vti } ] [ remote ADDR ] [ local ADDR ]\n");
	fprintf(stderr, "          [ [i|o]seq ] [ [i|o]key KEY ] [ [i|o]csum ]\n");
	fprintf(stderr, "          [ prl-default ADDR ] [ prl-nodefault ADDR ] [ prl-delete ADDR ]\n");
	fprintf(stderr, "          [ 6rd-prefix ADDR ] [ 6rd-relay_prefix ADDR ] [ 6rd-reset ]\n");
	fprintf(stderr, "          [ ttl TTL ] [ tos TOS ] [ [no]pmtudisc ] [ dev PHYS_DEV ]\n");
	fprintf(stderr, "\n");
	fprintf(stderr, "Where: NAME := STRING\n");
	fprintf(stderr, "       ADDR := { IP_ADDRESS | any }\n");
	fprintf(stderr, "       TOS  := { STRING | 00..ff | inherit | inherit/STRING | inherit/00..ff }\n");
	fprintf(stderr, "       TTL  := { 1..255 | inherit }\n");
	fprintf(stderr, "       KEY  := { DOTTED_QUAD | NUMBER }\n");
	exit(-1);
}

static void set_tunnel_proto(struct ip_tunnel_parm *p, int proto)
{
	if (p->iph.protocol && p->iph.protocol != proto) {
		fprintf(stderr,
			"You managed to ask for more than one tunnel mode.\n");
		exit(-1);
	}
	p->iph.protocol = proto;
}

static int parse_args(int argc, char **argv, int cmd, struct ip_tunnel_parm *p)
{
	int count = 0;
	const char *medium = NULL;
	int isatap = 0;

	memset(p, 0, sizeof(*p));
	p->iph.version = 4;
	p->iph.ihl = 5;
#ifndef IP_DF
#define IP_DF		0x4000		/* Flag: "Don't Fragment"	*/
#endif
	p->iph.frag_off = htons(IP_DF);

	while (argc > 0) {
		if (strcmp(*argv, "mode") == 0) {
			NEXT_ARG();
			if (strcmp(*argv, "ipip") == 0 ||
			    strcmp(*argv, "ip/ip") == 0) {
				set_tunnel_proto(p, IPPROTO_IPIP);
			} else if (strcmp(*argv, "gre") == 0 ||
				   strcmp(*argv, "gre/ip") == 0) {
				set_tunnel_proto(p, IPPROTO_GRE);
			} else if (strcmp(*argv, "sit") == 0 ||
				   strcmp(*argv, "ipv6/ip") == 0) {
				set_tunnel_proto(p, IPPROTO_IPV6);
			} else if (strcmp(*argv, "isatap") == 0) {
				set_tunnel_proto(p, IPPROTO_IPV6);
				isatap++;
			} else if (strcmp(*argv, "vti") == 0) {
				set_tunnel_proto(p, IPPROTO_IPIP);
				p->i_flags |= VTI_ISVTI;
			} else {
				fprintf(stderr,
					"Unknown tunnel mode \"%s\"\n", *argv);
				exit(-1);
			}
		} else if (strcmp(*argv, "key") == 0) {
			NEXT_ARG();
			p->i_flags |= GRE_KEY;
			p->o_flags |= GRE_KEY;
			p->i_key = p->o_key = tnl_parse_key("key", *argv);
		} else if (strcmp(*argv, "ikey") == 0) {
			NEXT_ARG();
			p->i_flags |= GRE_KEY;
			p->i_key = tnl_parse_key("ikey", *argv);
		} else if (strcmp(*argv, "okey") == 0) {
			NEXT_ARG();
			p->o_flags |= GRE_KEY;
			p->o_key = tnl_parse_key("okey", *argv);
		} else if (strcmp(*argv, "seq") == 0) {
			p->i_flags |= GRE_SEQ;
			p->o_flags |= GRE_SEQ;
		} else if (strcmp(*argv, "iseq") == 0) {
			p->i_flags |= GRE_SEQ;
		} else if (strcmp(*argv, "oseq") == 0) {
			p->o_flags |= GRE_SEQ;
		} else if (strcmp(*argv, "csum") == 0) {
			p->i_flags |= GRE_CSUM;
			p->o_flags |= GRE_CSUM;
		} else if (strcmp(*argv, "icsum") == 0) {
			p->i_flags |= GRE_CSUM;
		} else if (strcmp(*argv, "ocsum") == 0) {
			p->o_flags |= GRE_CSUM;
		} else if (strcmp(*argv, "nopmtudisc") == 0) {
			p->iph.frag_off = 0;
		} else if (strcmp(*argv, "pmtudisc") == 0) {
			p->iph.frag_off = htons(IP_DF);
		} else if (strcmp(*argv, "remote") == 0) {
			NEXT_ARG();
			p->iph.daddr = get_addr32(*argv);
		} else if (strcmp(*argv, "local") == 0) {
			NEXT_ARG();
			p->iph.saddr = get_addr32(*argv);
		} else if (strcmp(*argv, "dev") == 0) {
			NEXT_ARG();
			medium = *argv;
		} else if (strcmp(*argv, "ttl") == 0 ||
			   strcmp(*argv, "hoplimit") == 0 ||
			   strcmp(*argv, "hlim") == 0) {
			__u8 uval;

			NEXT_ARG();
			if (strcmp(*argv, "inherit") != 0) {
				if (get_u8(&uval, *argv, 0))
					invarg("invalid TTL\n", *argv);
				p->iph.ttl = uval;
			}
		} else if (strcmp(*argv, "tos") == 0 ||
			   strcmp(*argv, "tclass") == 0 ||
			   matches(*argv, "dsfield") == 0) {
			char *dsfield;
			__u32 uval;

			NEXT_ARG();
			dsfield = *argv;
			strsep(&dsfield, "/");
			if (strcmp(*argv, "inherit") != 0) {
				dsfield = *argv;
				p->iph.tos = 0;
			} else
				p->iph.tos = 1;
			if (dsfield) {
				if (rtnl_dsfield_a2n(&uval, dsfield))
					invarg("bad TOS value", *argv);
				p->iph.tos |= uval;
			}
		} else {
			if (strcmp(*argv, "name") == 0)
				NEXT_ARG();
			else if (matches(*argv, "help") == 0)
				usage();

			if (p->name[0])
				duparg2("name", *argv);
			if (get_ifname(p->name, *argv))
				invarg("\"name\" not a valid ifname", *argv);
			if (cmd == SIOCCHGTUNNEL && count == 0) {
				struct ip_tunnel_parm old_p = {};

				if (tnl_get_ioctl(*argv, &old_p))
					return -1;
				*p = old_p;
			}
		}
		count++;
		argc--; argv++;
	}


	if (p->iph.protocol == 0) {
		if (memcmp(p->name, "gre", 3) == 0)
			p->iph.protocol = IPPROTO_GRE;
		else if (memcmp(p->name, "ipip", 4) == 0)
			p->iph.protocol = IPPROTO_IPIP;
		else if (memcmp(p->name, "sit", 3) == 0)
			p->iph.protocol = IPPROTO_IPV6;
		else if (memcmp(p->name, "isatap", 6) == 0) {
			p->iph.protocol = IPPROTO_IPV6;
			isatap++;
		} else if (memcmp(p->name, "vti", 3) == 0) {
			p->iph.protocol = IPPROTO_IPIP;
			p->i_flags |= VTI_ISVTI;
		}
	}

	if ((p->i_flags & GRE_KEY) || (p->o_flags & GRE_KEY)) {
		if (!(p->i_flags & VTI_ISVTI) &&
		    (p->iph.protocol != IPPROTO_GRE)) {
			fprintf(stderr, "Keys are not allowed with ipip and sit tunnels\n");
			return -1;
		}
	}

	if (medium) {
		p->link = ll_name_to_index(medium);
		if (p->link == 0) {
			fprintf(stderr, "Cannot find device \"%s\"\n", medium);
			return -1;
		}
	}

	if (p->i_key == 0 && IN_MULTICAST(ntohl(p->iph.daddr))) {
		p->i_key = p->iph.daddr;
		p->i_flags |= GRE_KEY;
	}
	if (p->o_key == 0 && IN_MULTICAST(ntohl(p->iph.daddr))) {
		p->o_key = p->iph.daddr;
		p->o_flags |= GRE_KEY;
	}
	if (IN_MULTICAST(ntohl(p->iph.daddr)) && !p->iph.saddr) {
		fprintf(stderr, "A broadcast tunnel requires a source address\n");
		return -1;
	}
	if (isatap)
		p->i_flags |= SIT_ISATAP;

	return 0;
}

static const char *tnl_defname(const struct ip_tunnel_parm *p)
{
	switch (p->iph.protocol) {
	case IPPROTO_IPIP:
		if (p->i_flags & VTI_ISVTI)
			return "ip_vti0";
		else
			return "tunl0";
	case IPPROTO_GRE:
		return "gre0";
	case IPPROTO_IPV6:
		return "sit0";
	}
	return NULL;
}

static int do_add(int cmd, int argc, char **argv)
{
	struct ip_tunnel_parm p;
	const char *basedev;

	if (parse_args(argc, argv, cmd, &p) < 0)
		return -1;

	if (p.iph.ttl && p.iph.frag_off == 0) {
		fprintf(stderr, "ttl != 0 and nopmtudisc are incompatible\n");
		return -1;
	}

	basedev = tnl_defname(&p);
	if (!basedev) {
		fprintf(stderr,
			"cannot determine tunnel mode (ipip, gre, vti or sit)\n");
		return -1;
	}

	return tnl_add_ioctl(cmd, basedev, p.name, &p);
}

static int do_del(int argc, char **argv)
{
	struct ip_tunnel_parm p;

	if (parse_args(argc, argv, SIOCDELTUNNEL, &p) < 0)
		return -1;

	return tnl_del_ioctl(tnl_defname(&p) ? : p.name, p.name, &p);
}

static void print_tunnel(struct ip_tunnel_parm *p)
{
	struct ip_tunnel_6rd ip6rd = {};
	char s1[1024];
	char s2[1024];

	/* Do not use format_host() for local addr,
	 * symbolic name will not be useful.
	 */
	printf("%s: %s/ip remote %s local %s",
	       p->name,
	       tnl_strproto(p->iph.protocol),
	       p->iph.daddr ? format_host_r(AF_INET, 4, &p->iph.daddr, s1, sizeof(s1)) : "any",
	       p->iph.saddr ? rt_addr_n2a_r(AF_INET, 4, &p->iph.saddr, s2, sizeof(s2)) : "any");

	if (p->iph.protocol == IPPROTO_IPV6 && (p->i_flags & SIT_ISATAP)) {
		struct ip_tunnel_prl prl[16] = {};
		int i;

		prl[0].datalen = sizeof(prl) - sizeof(prl[0]);
		prl[0].addr = htonl(INADDR_ANY);

		if (!tnl_prl_ioctl(SIOCGETPRL, p->name, prl))
			for (i = 1; i < ARRAY_SIZE(prl); i++) {
				if (prl[i].addr != htonl(INADDR_ANY)) {
					printf(" %s %s ",
					       (prl[i].flags & PRL_DEFAULT) ? "pdr" : "pr",
					       format_host(AF_INET, 4, &prl[i].addr));
				}
			}
	}

	if (p->link) {
		const char *n = ll_index_to_name(p->link);

		if (n)
			printf(" dev %s", n);
	}

	if (p->iph.ttl)
		printf(" ttl %u", p->iph.ttl);
	else
		printf(" ttl inherit");

	if (p->iph.tos) {
		SPRINT_BUF(b1);
		printf(" tos");
		if (p->iph.tos & 1)
			printf(" inherit");
		if (p->iph.tos & ~1)
			printf("%c%s ", p->iph.tos & 1 ? '/' : ' ',
			       rtnl_dsfield_n2a(p->iph.tos & ~1, b1, sizeof(b1)));
	}

	if (!(p->iph.frag_off & htons(IP_DF)))
		printf(" nopmtudisc");

	if (p->iph.protocol == IPPROTO_IPV6 && !tnl_ioctl_get_6rd(p->name, &ip6rd) && ip6rd.prefixlen) {
		printf(" 6rd-prefix %s/%u",
		       inet_ntop(AF_INET6, &ip6rd.prefix, s1, sizeof(s1)),
		       ip6rd.prefixlen);
		if (ip6rd.relay_prefix) {
			printf(" 6rd-relay_prefix %s/%u",
			       format_host(AF_INET, 4, &ip6rd.relay_prefix),
			       ip6rd.relay_prefixlen);
		}
	}

	if ((p->i_flags & GRE_KEY) && (p->o_flags & GRE_KEY) && p->o_key == p->i_key)
		printf(" key %u", ntohl(p->i_key));
	else if ((p->i_flags | p->o_flags) & GRE_KEY) {
		if (p->i_flags & GRE_KEY)
			printf(" ikey %u", ntohl(p->i_key));
		if (p->o_flags & GRE_KEY)
			printf(" okey %u", ntohl(p->o_key));
	}

	if (p->i_flags & GRE_SEQ)
		printf("%s  Drop packets out of sequence.", _SL_);
	if (p->i_flags & GRE_CSUM)
		printf("%s  Checksum in received packet is required.", _SL_);
	if (p->o_flags & GRE_SEQ)
		printf("%s  Sequence packets on output.", _SL_);
	if (p->o_flags & GRE_CSUM)
		printf("%s  Checksum output packets.", _SL_);
}

static int do_tunnels_list(struct ip_tunnel_parm *p)
{
	char buf[512];
	int err = -1;
	FILE *fp = fopen("/proc/net/dev", "r");

	if (fp == NULL) {
		perror("fopen");
		return -1;
	}

	/* skip header lines */
	if (!fgets(buf, sizeof(buf), fp) ||
	    !fgets(buf, sizeof(buf), fp)) {
		fprintf(stderr, "/proc/net/dev read error\n");
		goto end;
	}

	while (fgets(buf, sizeof(buf), fp) != NULL) {
		char name[IFNAMSIZ];
		int index, type;
		struct ip_tunnel_parm p1 = {};
		char *ptr;

		buf[sizeof(buf) - 1] = 0;
		ptr = strchr(buf, ':');
		if (ptr == NULL ||
		    (*ptr++ = 0, sscanf(buf, "%s", name) != 1)) {
			fprintf(stderr, "Wrong format for /proc/net/dev. Giving up.\n");
			goto end;
		}
		if (p->name[0] && strcmp(p->name, name))
			continue;
		index = ll_name_to_index(name);
		if (index == 0)
			continue;
		type = ll_index_to_type(index);
		if (type == -1) {
			fprintf(stderr, "Failed to get type of \"%s\"\n", name);
			continue;
		}
		if (type != ARPHRD_TUNNEL && type != ARPHRD_IPGRE && type != ARPHRD_SIT)
			continue;
		if (tnl_get_ioctl(name, &p1))
			continue;
		if ((p->link && p1.link != p->link) ||
		    (p->name[0] && strcmp(p1.name, p->name)) ||
		    (p->iph.daddr && p1.iph.daddr != p->iph.daddr) ||
		    (p->iph.saddr && p1.iph.saddr != p->iph.saddr) ||
		    (p->i_key && p1.i_key != p->i_key))
			continue;
		print_tunnel(&p1);
		if (show_stats)
			tnl_print_stats(ptr);
		printf("\n");
	}
	err = 0;
 end:
	fclose(fp);
	return err;
}

static int do_show(int argc, char **argv)
{
	struct ip_tunnel_parm p;
	const char *basedev;

	ll_init_map(&rth);
	if (parse_args(argc, argv, SIOCGETTUNNEL, &p) < 0)
		return -1;

	basedev = tnl_defname(&p);
	if (!basedev)
		return do_tunnels_list(&p);

	if (tnl_get_ioctl(p.name[0] ? p.name : basedev, &p))
		return -1;

	print_tunnel(&p);
	printf("\n");
	return 0;
}

static int do_prl(int argc, char **argv)
{
	struct ip_tunnel_prl p = {};
	int count = 0;
	int cmd = 0;
	const char *medium = NULL;

	while (argc > 0) {
		if (strcmp(*argv, "prl-default") == 0) {
			NEXT_ARG();
			cmd = SIOCADDPRL;
			p.addr = get_addr32(*argv);
			p.flags |= PRL_DEFAULT;
			count++;
		} else if (strcmp(*argv, "prl-nodefault") == 0) {
			NEXT_ARG();
			cmd = SIOCADDPRL;
			p.addr = get_addr32(*argv);
			count++;
		} else if (strcmp(*argv, "prl-delete") == 0) {
			NEXT_ARG();
			cmd = SIOCDELPRL;
			p.addr = get_addr32(*argv);
			count++;
		} else if (strcmp(*argv, "dev") == 0) {
			NEXT_ARG();
			if (check_ifname(*argv))
				invarg("\"dev\" not a valid ifname", *argv);
			medium = *argv;
		} else {
			fprintf(stderr,
				"Invalid PRL parameter \"%s\"\n", *argv);
			exit(-1);
		}
		if (count > 1) {
			fprintf(stderr,
				"One PRL entry at a time\n");
			exit(-1);
		}
		argc--; argv++;
	}
	if (!medium) {
		fprintf(stderr, "Must specify device\n");
		exit(-1);
	}

	return tnl_prl_ioctl(cmd, medium, &p);
}

static int do_6rd(int argc, char **argv)
{
	struct ip_tunnel_6rd ip6rd = {};
	int cmd = 0;
	const char *medium = NULL;
	inet_prefix prefix;

	while (argc > 0) {
		if (strcmp(*argv, "6rd-prefix") == 0) {
			NEXT_ARG();
			if (get_prefix(&prefix, *argv, AF_INET6))
				invarg("invalid 6rd_prefix\n", *argv);
			cmd = SIOCADD6RD;
			memcpy(&ip6rd.prefix, prefix.data, 16);
			ip6rd.prefixlen = prefix.bitlen;
		} else if (strcmp(*argv, "6rd-relay_prefix") == 0) {
			NEXT_ARG();
			if (get_prefix(&prefix, *argv, AF_INET))
				invarg("invalid 6rd-relay_prefix\n", *argv);
			cmd = SIOCADD6RD;
			memcpy(&ip6rd.relay_prefix, prefix.data, 4);
			ip6rd.relay_prefixlen = prefix.bitlen;
		} else if (strcmp(*argv, "6rd-reset") == 0) {
			cmd = SIOCDEL6RD;
		} else if (strcmp(*argv, "dev") == 0) {
			NEXT_ARG();
			if (check_ifname(*argv))
				invarg("\"dev\" not a valid ifname", *argv);
			medium = *argv;
		} else {
			fprintf(stderr,
				"Invalid 6RD parameter \"%s\"\n", *argv);
			exit(-1);
		}
		argc--; argv++;
	}
	if (!medium) {
		fprintf(stderr, "Must specify device\n");
		exit(-1);
	}

	return tnl_6rd_ioctl(cmd, medium, &ip6rd);
}

static int tunnel_mode_is_ipv6(char *tunnel_mode)
{
	static const char * const ipv6_modes[] = {
		"ipv6/ipv6", "ip6ip6",
		"vti6",
		"ip/ipv6", "ipv4/ipv6", "ipip6", "ip4ip6",
		"ip6gre", "gre/ipv6",
		"any/ipv6", "any"
	};
	int i;

	for (i = 0; i < ARRAY_SIZE(ipv6_modes); i++) {
		if (strcmp(ipv6_modes[i], tunnel_mode) == 0)
			return 1;
	}
	return 0;
}

int do_iptunnel(int argc, char **argv)
{
	int i;

	for (i = 0; i < argc - 1; i++) {
		if (strcmp(argv[i], "mode") == 0) {
			if (tunnel_mode_is_ipv6(argv[i + 1]))
				preferred_family = AF_INET6;
			break;
		}
	}
	switch (preferred_family) {
	case AF_UNSPEC:
		preferred_family = AF_INET;
		break;
	case AF_INET:
		break;
	/*
	 * This is silly enough but we have no easy way to make it
	 * protocol-independent because of unarranged structure between
	 * IPv4 and IPv6.
	 */
	case AF_INET6:
		return do_ip6tunnel(argc, argv);
	default:
		fprintf(stderr, "Unsupported protocol family: %d\n", preferred_family);
		exit(-1);
	}

	if (argc > 0) {
		if (matches(*argv, "add") == 0)
			return do_add(SIOCADDTUNNEL, argc - 1, argv + 1);
		if (matches(*argv, "change") == 0)
			return do_add(SIOCCHGTUNNEL, argc - 1, argv + 1);
		if (matches(*argv, "delete") == 0)
			return do_del(argc - 1, argv + 1);
		if (matches(*argv, "show") == 0 ||
		    matches(*argv, "lst") == 0 ||
		    matches(*argv, "list") == 0)
			return do_show(argc - 1, argv + 1);
		if (matches(*argv, "prl") == 0)
			return do_prl(argc - 1, argv + 1);
		if (matches(*argv, "6rd") == 0)
			return do_6rd(argc - 1, argv + 1);
		if (matches(*argv, "help") == 0)
			usage();
	} else
		return do_show(0, NULL);

	fprintf(stderr, "Command \"%s\" is unknown, try \"ip tunnel help\"\n", *argv);
	exit(-1);
}
