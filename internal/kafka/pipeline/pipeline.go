package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"usergenerator/internal"
	"usergenerator/internal/kafka/enrichment"
	models "usergenerator/internal/lib/api/model/user"
	storage "usergenerator/storage"

	"github.com/segmentio/kafka-go"
)

type PipelineFIO struct {
	In      chan kafka.Message
	Out     chan models.BaseUser
	Failed  chan kafka.Message
	DBqueue chan models.User
	DB      *storage.Storage
	CFG     internal.Config
	Ctx     *context.Context
	logger  *slog.Logger
	stats stats
}

type stats struct{
	in int
	out int
	fail int
	miss int
}

func New(ctx *context.Context, logger *slog.Logger, cfg internal.Config, db *storage.Storage) *PipelineFIO {
	return &PipelineFIO{
		In:     make(chan kafka.Message,10),
		Out:    make(chan models.BaseUser,10),
		Failed: make(chan kafka.Message,10),
		DB:     db,
		CFG:    cfg,
		Ctx:    ctx,
		logger: logger,
		stats: stats{0,0,0,0},
	}
}

func (p *PipelineFIO) Start() {


	op := "kafka.Pipeline"
	var quit chan struct{}

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)
	_, cancel := context.WithCancel(*p.Ctx)
	// go routine for getting signals asynchronously
	go func() {
		sig := <-signals
		p.logger.Info(fmt.Sprintf("%s Got signal: %v", op, sig))
		quit <- struct{}{}
		cancel()
	}()
	batch := make([]models.User, 0)
	var msg kafka.Message
	var user models.BaseUser
	pool := make(chan struct{}, 3)
	for {
		select {
		case <-p.In:
			p.stats.in++
			msg = <-p.In
			err := json.Unmarshal(msg.Value, &user)
			if err != nil || !user.Validate() {
				if err != nil {
					p.logger.Info(fmt.Sprintf("%s :%s", op, err))
				} else {
					p.logger.Info(fmt.Sprintf("%s :%s [%v]", op, "validation",user))
				}
				p.stats.fail++
				p.Failed <- msg
			} else {
				p.stats.out++
				p.Out <- user
				p.logger.Info(fmt.Sprintf("%s: proceed messsage %s", op, user))
			}
		case <-p.Out:
			select {
			case pool <- struct{}{}:
				go func() {
					u, err := enrichment.Enrichment(<-p.Out, p.CFG.EnrichmentURLS, p.logger)
					if err != nil {
						p.stats.miss++
						p.logger.Info(fmt.Sprintf("%s: enrichment error %s", op, err))
					} else {
						batch = append(batch, *u)
						p.logger.Info(fmt.Sprintf("%s: add to batch: %s size:%d", op, user, len(batch)))
					}
					<-pool
				}()
			}
			if len(batch) > 10 {
				//fmt.Println(batch)
				data, err := p.DB.InsertUsers(batch...)
				if err != nil {
					p.logger.Error(fmt.Sprintf("%s: failed to batch.Check logs for more info %s", op, err))
				} else {
					p.logger.Info(fmt.Sprintf("%s: batch complete %v", op, data))
					batch = make([]models.User, 0)
				}
			}
		case <-quit:
			close(p.In)
			close(p.Out)
			close(p.Failed)
			p.logger.Warn(fmt.Sprintf("%s :quit", op))
			break
		default:
			p.logger.Info(fmt.Sprintf("%s:Pulse, stats in:%d, fail:%d,out:%d,miss:%d", op,p.stats.in,p.stats.fail,p.stats.out,p.stats.miss))
			time.Sleep(5 * time.Second)
		}
	}
}
