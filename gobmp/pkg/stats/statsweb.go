package stats

type Srv interface {
	Start()
	Stop()
}

type statsSrv struct {
}

func (s *statsSrv) Start() {

}

func (s *statsSrv) Stop() {

}

func NewStatsWebSrv() Srv {

	return &statsSrv{}
}
