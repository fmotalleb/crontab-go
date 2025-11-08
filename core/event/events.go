package event

import (
	"fmt"
	"maps"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/generator"
)

var eg = generator.New[*config.JobEvent, abstraction.EventGenerator]()

func Build(log *logrus.Entry, cfg *config.JobEvent) abstraction.EventGenerator {
	if g, ok := eg.Get(log, cfg); ok {
		return g
	}
	err := fmt.Errorf("no event generator matched %+v", *cfg)
	log.WithError(err).Warn("event.Build: generator not found")
	return nil
}

type MetaData struct {
	Emitter string
	Extra   map[string]any
}

func NewMetaData(emitter string, extra map[string]any) *MetaData {
	var e map[string]any
	if extra != nil {
		e = maps.Clone(extra)
	} else {
		e = make(map[string]any)
	}
	e["emitter"] = emitter
	return &MetaData{
		Emitter: emitter,
		Extra:   e,
	}
}

func NewErrMetaData(emitter string, err error) *MetaData {
	return &MetaData{
		Emitter: emitter,
		Extra: map[string]any{
			"error": err.Error(),
		},
	}
}

func (m *MetaData) GetData() map[string]any {
	return m.Extra
}
