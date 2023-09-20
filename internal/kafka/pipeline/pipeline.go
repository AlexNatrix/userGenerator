package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"main/internal"
	storage "main/internal"
	"main/internal/kafka/consumer"
	"main/internal/kafka/enrichment"
	"main/internal/kafka/producer"
	models "main/internal/lib/api/model/user"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)


type PipelineFIO struct{
	In chan kafka.Message
	Out chan models.BaseUser
	Failed chan kafka.Message
	DBqueue chan models.User
	DB *internal.Storage
	CFG internal.Config
	Ctx *context.Context
	logger *slog.Logger
}

func New(ctx *context.Context,logger *slog.Logger,cfg internal.Config, db *storage.Storage ) *PipelineFIO{
	return &PipelineFIO{
		In:make(chan kafka.Message,10),
		Out: make(chan models.BaseUser,10),
		Failed: make(chan kafka.Message,10),
		DB: db,
		CFG :cfg,
		Ctx : ctx,
		logger: logger,
	}
}




func (p *PipelineFIO) Start(){
	go consumer.Consumer(p.Ctx,p.CFG,p.logger, p.In)
	go producer.Produce(p.Ctx,p.CFG,p.logger,p.Failed)
	
	op:="kafka.Pipeline"
	var quit chan struct{}

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)
	_, cancel := context.WithCancel(*p.Ctx)
	// go routine for getting signals asynchronously
	go func() {
		sig := <-signals
		p.logger.Info("%s Got signal: %w", op,sig)
		quit<-struct{}{}
		cancel()
	}()
	batch:=make([]models.User,0)
	var msg kafka.Message
	var user models.BaseUser
	pool:=make(chan struct{},3) 
	for {
		select {
		case  <- p.In:
			msg=<-p.In
			err:=json.Unmarshal(msg.Value, &user)
			if err!=nil || !user.Validate(){
				p.Failed<-msg
				if err!=nil{
					p.logger.Info(fmt.Sprintf("%s :%s",op,err))
				}else{
					p.logger.Info(fmt.Sprintf("%s :%s",op,"VALIDATION ERROR"))
				}
				
			}else{
				p.Out<-user
				p.logger.Info(fmt.Sprintf("%s: proceed messsage %s",op,user))
			}
		case <-p.Out:
			select{
			case pool<-struct{}{}:
				go func(){
					u,err:=enrichment.Enrichment(<-p.Out,p.CFG.EnrichmentURLS,p.logger)
				if err!=nil{
					p.logger.Info(fmt.Sprintf("%s: enrichment error %s",op,err))

					}else{
						p.logger.Info(fmt.Sprintf("%s: batch++ %s %d",op,user,len(batch)))
						batch=append(batch, *u)
					}
					<-pool
				}()
			}
			if len(batch)>10{
				//fmt.Println(batch)
				data,err:=p.DB.SaveUser(p.logger,batch...)
				if err!=nil{
					p.logger.Error(fmt.Sprintf("%s:failed to batch.Check logs for more info %s",op,err))
				}else{
					p.logger.Info(fmt.Sprintf("%s batch complete %v",op,data))
					batch=make([]models.User,0)
				}
			}
		case <-quit:
			close(p.In)
			close(p.Out)
			close(p.Failed)
			p.logger.Warn(fmt.Sprintf("%s :quit",op))
			break
		default:
			p.logger.Info(fmt.Sprintf("%s:Pulse",op))
			time.Sleep(5*time.Second)	
		}
	}
}



// var wg sync.WaitGroup

//     for i := 1; i <= 5; i++ {
//         wg.Add(1)

//         i := i

//         go func() {
//             defer wg.Done()
//             worker(i)
//         }()
//     }

//     wg.Wait()