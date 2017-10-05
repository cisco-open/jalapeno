package main

import (
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/arango"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
)

var (
	routers     []arango.Router
	prefixes    []arango.Prefix
	linkEdges   []arango.LinkEdge
	prefixEdges []arango.PrefixEdge
)

func main() {
	cfg := arango.ArangoConfig{
		URL:      "http://127.0.0.1:8529",
		User:     "root",
		Password: "voltron",
		Database: "graphTest",
	}
	db, err := arango.New(cfg)
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
	routers = []arango.Router{
		arango.Router{
			BGPID: "NYC",
			ASN:   "1",
		},
		arango.Router{
			BGPID: "RTP",
			ASN:   "1",
		},
		arango.Router{
			BGPID: "BOS",
			ASN:   "1",
		},
		arango.Router{
			BGPID: "SEA",
			ASN:   "2",
		},
		arango.Router{
			BGPID: "SFO",
			ASN:   "2",
		},
		arango.Router{
			BGPID: "LA",
			ASN:   "3",
		},
	}
}

func populateLinkEdges() {
	linkEdges = []arango.LinkEdge{
		arango.LinkEdge{
			From:        "Routers/NYC_1",
			To:          "Routers/RTP_1",
			FromIP:      "NYC_1",
			ToIP:        "RTP_1",
			Latency:     10,
			Utilization: 25,
			Label:       "1",
		},
		arango.LinkEdge{
			From:        "Routers/RTP_1",
			To:          "Routers/NYC_1",
			FromIP:      "RTP_1",
			ToIP:        "NYC_1",
			Latency:     10,
			Utilization: 25,
			Label:       "1",
		},
		arango.LinkEdge{
			From:        "Routers/NYC_1",
			To:          "Routers/BOS_1",
			FromIP:      "NYC_1",
			ToIP:        "BOS_1",
			Latency:     10,
			Utilization: 25,
			Label:       "2",
		},
		arango.LinkEdge{
			From:        "Routers/BOS_1",
			To:          "Routers/NYC_1",
			FromIP:      "BOS_1",
			ToIP:        "NYC_1",
			Latency:     10,
			Utilization: 25,
			Label:       "2",
		},
		arango.LinkEdge{
			From:        "Routers/RTP_1",
			To:          "Routers/BOS_1",
			FromIP:      "RTP_1",
			ToIP:        "BOS_1",
			Latency:     10,
			Utilization: 10,
			Label:       "3",
		},
		arango.LinkEdge{
			From:        "Routers/BOS_1",
			To:          "Routers/RTP_1",
			FromIP:      "BOS_1",
			ToIP:        "RTP_1",
			Latency:     10,
			Utilization: 10,
			Label:       "3",
		},
		arango.LinkEdge{
			From:        "Routers/RTP_1",
			To:          "Routers/SEA_2",
			FromIP:      "RTP_1",
			ToIP:        "SEA_2",
			Latency:     25,
			Utilization: 10,
			Label:       "4",
		},
		arango.LinkEdge{
			From:        "Routers/SEA_2",
			To:          "Routers/RTP_1",
			FromIP:      "SEA_2",
			ToIP:        "RTP_1",
			Latency:     25,
			Utilization: 10,
			Label:       "4",
		},
		arango.LinkEdge{
			From:        "Routers/RTP_1",
			To:          "Routers/SFO_2",
			FromIP:      "RTP_1",
			ToIP:        "SFO_2",
			Latency:     10,
			Utilization: 40,
			Label:       "5",
		},
		arango.LinkEdge{
			From:        "Routers/SFO_2",
			To:          "Routers/RTP_1",
			FromIP:      "SFO_2",
			ToIP:        "RTP_1",
			Latency:     10,
			Utilization: 40,
			Label:       "5",
		},
		arango.LinkEdge{
			From:        "Routers/RTP_1",
			To:          "Routers/LA_3",
			FromIP:      "RTP_1",
			ToIP:        "LA_3",
			Latency:     15,
			Utilization: 45,
			Label:       "6",
		},
		arango.LinkEdge{
			From:        "Routers/LA_3",
			To:          "Routers/RTP_1",
			FromIP:      "LA_3",
			ToIP:        "RTP_1",
			Latency:     15,
			Utilization: 45,
			Label:       "6",
		},
		arango.LinkEdge{
			From:        "Routers/BOS_1",
			To:          "Routers/SEA_2",
			FromIP:      "BOS_1",
			ToIP:        "SEA_2",
			Latency:     15,
			Utilization: 40,
			Label:       "7",
		},
		arango.LinkEdge{
			From:        "Routers/SEA_2",
			To:          "Routers/BOS_1",
			FromIP:      "SEA_2",
			ToIP:        "BOS_1",
			Latency:     15,
			Utilization: 40,
			Label:       "7",
		},
		arango.LinkEdge{
			From:        "Routers/BOS_1",
			To:          "Routers/SFO_2",
			FromIP:      "BOS_1",
			ToIP:        "SFO_2",
			Latency:     15,
			Utilization: 35,
			Label:       "8",
		},
		arango.LinkEdge{
			From:        "Routers/SFO_2",
			To:          "Routers/BOS_1",
			FromIP:      "SFO_2",
			ToIP:        "BOS_1",
			Latency:     15,
			Utilization: 35,
			Label:       "8",
		},
		arango.LinkEdge{
			From:        "Routers/BOS_1",
			To:          "Routers/LA_3",
			FromIP:      "BOS_1",
			ToIP:        "LA_3",
			Latency:     25,
			Utilization: 10,
			Label:       "9",
		},
		arango.LinkEdge{
			From:        "Routers/LA_3",
			To:          "Routers/BOS_1",
			FromIP:      "LA_3",
			ToIP:        "BOS_1",
			Latency:     25,
			Utilization: 10,
			Label:       "9",
		},
	}
}

func populatePrefixes() {
	prefixes = []arango.Prefix{
		arango.Prefix{
			Prefix: "10.86.204.0",
			Length: 24,
		},
		arango.Prefix{
			Prefix: "192.168.0.0",
			Length: 16,
		},
		arango.Prefix{
			Prefix: "10.1.1.0",
			Length: 24,
		},
	}
}

func populatePrefixEdges() {
	prefixEdges = []arango.PrefixEdge{
		arango.PrefixEdge{
			From:        "Routers/SEA_2",
			To:          "Prefixes/10.86.204.0_24",
			Latency:     10,
			Utilization: 40,
			Labels:      []string{"10"},
		},
		arango.PrefixEdge{
			From:        "Routers/SEA_2",
			To:          "Prefixes/192.168.0.0_16",
			Latency:     15,
			Utilization: 60,
			Labels:      []string{"11"},
		},
		arango.PrefixEdge{
			From:        "Routers/SFO_2",
			To:          "Prefixes/10.86.204.0_24",
			Latency:     15,
			Utilization: 50,
			Labels:      []string{"12"},
		},
		arango.PrefixEdge{
			From:        "Routers/SFO_2",
			To:          "Prefixes/192.168.0.0_16",
			Latency:     10,
			Utilization: 40,
			Labels:      []string{"13"},
		},
		arango.PrefixEdge{
			From:        "Routers/SFO_2",
			To:          "Prefixes/10.1.1.0_24",
			Latency:     15,
			Utilization: 50,
			Labels:      []string{"14"},
		},
		arango.PrefixEdge{
			From:        "Routers/LA_3",
			To:          "Prefixes/192.168.0.0_16",
			Latency:     15,
			Utilization: 60,
			Labels:      []string{"15"},
		},
		arango.PrefixEdge{
			From:        "Routers/LA_3",
			To:          "Prefixes/10.1.1.0_24",
			Latency:     10,
			Utilization: 40,
			Labels:      []string{"16"},
		},
	}
}
