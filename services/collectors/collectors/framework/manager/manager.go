package manager

import (
	"fmt"
	"time"

	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/database"
	"wwwin-github.cisco.com/spa-ie/voltron-redux/framework/log"
)

var (
	StatusRunning = "Running"
	StatusDown    = "Down"
)

type Config struct {
	Interval string `desc:"Time between manager looking at all Services"`
}

type Manager struct {
	db       database.Database
	interval time.Duration

	quit chan struct{}
}

func NewConfig() Config {
	return Config{Interval: "5s"}
}

func NewManager(cfg Config, db database.Database) (*Manager, error) {
	t, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		return nil, err
	}
	return &Manager{
		db:       db,
		interval: t,
		quit:     make(chan struct{}),
	}, nil

}

func (m *Manager) Start() error {
	ticker := time.NewTicker(m.interval)
	for {
		select {
		case <-m.quit:
			log.Info("Voltron Framework Manager Shutdown")
			return nil
		case <-ticker.C:
			log.Debug("Scanning the services")
			m.ProcessControllers()
		}
	}
}

func (m *Manager) ProcessControllers() error {
	q := "FOR c in Collectors return c"
	cols, err := m.db.Query(q, nil, database.Collector{})
	if err != nil {
		return err
	}

	for _, c := range cols {
		col := c.(*database.Collector)
		timeout, err := time.ParseDuration(col.Timeout)
		if err != nil {
			return fmt.Errorf("Failed to parse duration for %q", col.Name)
		}

		last, err := time.Parse(time.RFC3339, col.LastHeartbeat)
		if err != nil {
			return fmt.Errorf("Failed to parse LastHeartbeat: %q", col.Name)
		}

		duration := time.Since(last)

		change := false
		var status string
		if duration > timeout && col.Status != StatusDown {
			change = true
			status = StatusDown
		} else if duration < timeout && col.Status != StatusRunning {
			change = true
			status = StatusRunning
		}

		if change {
			col.Status = status
			err := m.db.Update(col)
			if err != nil {
				return fmt.Errorf("Failed to update %q to %q", col.Name, status)
			}
			log.Infof("Setting Collector %q status to %q", col.Name, status)
		}
	}
	return nil
}

func (m *Manager) Stop() {
	close(m.quit)
}
