package announcer

import (
	"context"
	"time"

	"github.com/cenkalti/rain/internal/logger"
	"github.com/cenkalti/rain/torrent/internal/tracker"
)

type StopAnnouncer struct {
	log      logger.Logger
	timeout  time.Duration
	trackers []tracker.Tracker
	torrent  tracker.Torrent
	resultC  chan struct{}
	closeC   chan struct{}
	doneC    chan struct{}
}

func NewStopAnnouncer(trackers []tracker.Tracker, tra tracker.Torrent, timeout time.Duration, resultC chan struct{}, l logger.Logger) *StopAnnouncer {
	return &StopAnnouncer{
		log:      l,
		timeout:  timeout,
		trackers: trackers,
		torrent:  tra,
		resultC:  resultC,
		closeC:   make(chan struct{}),
		doneC:    make(chan struct{}),
	}
}

func (a *StopAnnouncer) Close() {
	close(a.closeC)
	<-a.doneC
}

func (a *StopAnnouncer) Run() {
	defer close(a.doneC)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(a.timeout))
	go func() {
		<-a.closeC
		cancel()
	}()

	doneC := make(chan struct{})
	for _, trk := range a.trackers {
		go func(trk tracker.Tracker) {
			callAnnounce(ctx, trk, a.torrent, tracker.EventStopped, 0, a.log)
			doneC <- struct{}{}
		}(trk)
	}
	for range a.trackers {
		<-doneC
	}
	select {
	case a.resultC <- struct{}{}:
	case <-a.closeC:
	}
}