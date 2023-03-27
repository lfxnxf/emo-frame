package proxy

import (
	"github.com/lfxnxf/emo-frame/inits"
	"github.com/lfxnxf/emo-frame/resource/sql"
)

type SQL struct {
	name []string
}

func InitSQL(name ...string) *SQL {
	if len(name) == 0 {
		return nil
	}
	return &SQL{name}
}

func (s *SQL) Master(name ...string) *sql.Client {
	var gName string
	if len(name) == 0 {
		gName = s.name[0]
	} else {
		gName = name[0]
	}
	return inits.SQLClient(gName).Master()
}

func (s *SQL) MasterNoLog(name ...string) *sql.Client {
	var gName string
	if len(name) == 0 {
		gName = s.name[0]
	} else {
		gName = name[0]
	}
	return inits.SQLClient(gName).MasterNoLog()
}

func (s *SQL) Slave(name ...string) *sql.Client {
	var gName string
	if len(name) == 0 {
		gName = s.name[0]
	} else {
		gName = name[0]
	}
	return inits.SQLClient(gName).Slave()
}

func (s *SQL) SlaveNoLog(name ...string) *sql.Client {
	var gName string
	if len(name) == 0 {
		gName = s.name[0]
	} else {
		gName = name[0]
	}
	return inits.SQLClient(gName).SlaveNoLog()
}
