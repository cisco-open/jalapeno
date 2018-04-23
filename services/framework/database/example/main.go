package main

import (
	"wwwin-github.cisco.com/spa-ie/voltron/services/framework/database"
	"wwwin-github.cisco.com/spa-ie/voltron/services/framework/log"
)

var (
	routers     []database.Router
	prefixes    []database.Prefix
	linkEdges   []database.LinkEdge
	prefixEdges []database.PrefixEdge
)

func main() {
	cfg := database.ArangoConfig{
		URL:      "http://127.0.0.1:8529",
		User:     "root",
		Password: "voltron",
		Database: "voltron",
	}
	db, err := database.NewArango(cfg)
	if err != nil {
		log.WithError(err).Error("Failed to connect to db")
	}

	populateRouters()
	populateLinkEdges()
	populatePrefixes()
	populatePrefixEdges()

	for _, r := range routers {
		err := db.Upsert(&r)
		if err != nil {
			log.WithError(err).Error("Upserting routers")
		}
	}

	for _, l := range linkEdges {
		err := db.Upsert(&l)
		if err != nil {
			log.WithError(err).Error("Upserting linkEdges")
		}
	}

	for _, p := range prefixes {
		err := db.Upsert(&p)
		if err != nil {
			log.WithError(err).Error("Upserting prefixes")
		}
	}

	for _, p := range prefixEdges {
		err := db.Upsert(&p)
		if err != nil {
			log.WithError(err).Error("Upserting prefixEdges")
		}
	}
}

func populateRouters() {
	routers = []database.Router{
		database.Router{
			BGPID: "NYC",
			ASN:   "1",
		},
		database.Router{
			BGPID: "RTP",
			ASN:   "1",
		},
		database.Router{
			BGPID: "BOS",
			ASN:   "1",
		},
		database.Router{
			BGPID: "SEA",
			ASN:   "2",
		},
		database.Router{
			BGPID: "SFO",
			ASN:   "2",
		},
		database.Router{
			BGPID: "LA",
			ASN:   "3",
		},
	}
}

func populateLinkEdges() {
	linkEdges = []database.LinkEdge{
		database.LinkEdge{
			From:   "Routers/NYC_1",
			To:     "Routers/RTP_1",
			FromIP: "NYC_1",
			ToIP:   "RTP_1",
			Label:  "1",
		},
		database.LinkEdge{
			From:   "Routers/RTP_1",
			To:     "Routers/NYC_1",
			FromIP: "RTP_1",
			ToIP:   "NYC_1",
			Label:  "1",
		},
		database.LinkEdge{
			From:   "Routers/NYC_1",
			To:     "Routers/BOS_1",
			FromIP: "NYC_1",
			ToIP:   "BOS_1",
			Label:  "2",
		},
		database.LinkEdge{
			From:   "Routers/BOS_1",
			To:     "Routers/NYC_1",
			FromIP: "BOS_1",
			ToIP:   "NYC_1",
			Label:  "2",
		},
		database.LinkEdge{
			From:   "Routers/RTP_1",
			To:     "Routers/BOS_1",
			FromIP: "RTP_1",
			ToIP:   "BOS_1",
			Label:  "3",
		},
		database.LinkEdge{
			From:   "Routers/BOS_1",
			To:     "Routers/RTP_1",
			FromIP: "BOS_1",
			ToIP:   "RTP_1",
			Label:  "3",
		},
		database.LinkEdge{
			From:   "Routers/RTP_1",
			To:     "Routers/SEA_2",
			FromIP: "RTP_1",
			ToIP:   "SEA_2",
			Label:  "4",
		},
		database.LinkEdge{
			From:   "Routers/SEA_2",
			To:     "Routers/RTP_1",
			FromIP: "SEA_2",
			ToIP:   "RTP_1",
			Label:  "4",
		},
		database.LinkEdge{
			From:   "Routers/RTP_1",
			To:     "Routers/SFO_2",
			FromIP: "RTP_1",
			ToIP:   "SFO_2",
			Label:  "5",
		},
		database.LinkEdge{
			From:   "Routers/SFO_2",
			To:     "Routers/RTP_1",
			FromIP: "SFO_2",
			ToIP:   "RTP_1",
			Label:  "5",
		},
		database.LinkEdge{
			From:   "Routers/RTP_1",
			To:     "Routers/LA_3",
			FromIP: "RTP_1",
			ToIP:   "LA_3",
			Label:  "6",
		},
		database.LinkEdge{
			From:   "Routers/LA_3",
			To:     "Routers/RTP_1",
			FromIP: "LA_3",
			ToIP:   "RTP_1",
			Label:  "6",
		},
		database.LinkEdge{
			From:   "Routers/BOS_1",
			To:     "Routers/SEA_2",
			FromIP: "BOS_1",
			ToIP:   "SEA_2",
			Label:  "7",
		},
		database.LinkEdge{
			From:   "Routers/SEA_2",
			To:     "Routers/BOS_1",
			FromIP: "SEA_2",
			ToIP:   "BOS_1",
			Label:  "7",
		},
		database.LinkEdge{
			From:   "Routers/BOS_1",
			To:     "Routers/SFO_2",
			FromIP: "BOS_1",
			ToIP:   "SFO_2",
			Label:  "8",
		},
		database.LinkEdge{
			From:   "Routers/SFO_2",
			To:     "Routers/BOS_1",
			FromIP: "SFO_2",
			ToIP:   "BOS_1",
			Label:  "8",
		},
		database.LinkEdge{
			From:   "Routers/BOS_1",
			To:     "Routers/LA_3",
			FromIP: "BOS_1",
			ToIP:   "LA_3",
			Label:  "9",
		},
		database.LinkEdge{
			From:   "Routers/LA_3",
			To:     "Routers/BOS_1",
			FromIP: "LA_3",
			ToIP:   "BOS_1",
			Label:  "9",
		},
	}
}

func populatePrefixes() {
	prefixes = []database.Prefix{
		database.Prefix{
			Prefix: "10.86.204.0",
			Length: 24,
		},
		database.Prefix{
			Prefix: "192.168.0.0",
			Length: 16,
		},
		database.Prefix{
			Prefix: "10.1.1.0",
			Length: 24,
		},
	}
}

func populatePrefixEdges() {
	prefixEdges = []database.PrefixEdge{
		database.PrefixEdge{
			From:   "Routers/SEA_2",
			To:     "Prefixes/10.86.204.0_24",
			Labels: []string{"10"},
		},
		database.PrefixEdge{
			From:   "Routers/SEA_2",
			To:     "Prefixes/192.168.0.0_16",
			Labels: []string{"11"},
		},
		database.PrefixEdge{
			From:   "Routers/SFO_2",
			To:     "Prefixes/10.86.204.0_24",
			Labels: []string{"12"},
		},
		database.PrefixEdge{
			From:   "Routers/SFO_2",
			To:     "Prefixes/192.168.0.0_16",
			Labels: []string{"13"},
		},
		database.PrefixEdge{
			From:   "Routers/SFO_2",
			To:     "Prefixes/10.1.1.0_24",
			Labels: []string{"14"},
		},
		database.PrefixEdge{
			From:   "Routers/LA_3",
			To:     "Prefixes/192.168.0.0_16",
			Labels: []string{"15"},
		},
		database.PrefixEdge{
			From:   "Routers/LA_3",
			To:     "Prefixes/10.1.1.0_24",
			Labels: []string{"16"},
		},
	}
}
