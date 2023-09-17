package main

import (
	"context"
	"fmt"
	"main/internal"
	config "main/internal"
	logger "main/internal"
	storage "main/internal"
	"main/internal/kafka/consumer"
	models "main/internal/lib/api/model/user"
	"main/internal/lib/enrichment"
	mwLogger "main/internal/middleware/logger"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)




func main() {
	//config :cleanENV
	cfg,err := config.LoadConfig()
	fmt.Println(cfg,err)
	log:= logger.SetupLogger(cfg.Env)

	store, err := storage.New(cfg.StoragePath)
	if err!=nil{
		log.Error("failed to init storage",internal.Err(err))
	}else{
		log.Info("DB init success")
	}
	_=store
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	test,err:=enrichment.Test()
	if err!=nil{
		log.Error("failed to init storage",internal.Err(err))
	}else{
		fmt.Println(models.User(*test))
	}
	// fmt.Println("COnsuming")
	// consumer.Produce(context.Background())
	// fmt.Println("COnsuming")
	// consumer.Consume()
	var prod func (pinger <-chan int, ponger chan<- int);
	var cons  func(pinger chan<- int, ponger <-chan int) ;
	// The pinger prints a ping and waits for a pong
	prod=func (pinger <-chan int, ponger chan<- int) {
			for {
				<-pinger
				fmt.Println("ping")
				consumer.Produce(context.Background())
				time.Sleep(time.Second)
				ponger <- 1
			}
		}

// The ponger prints a pong and waits for a ping
	cons=func(pinger chan<- int, ponger <-chan int) {
    for {
        <-ponger
        fmt.Println("pong")
		for i:=0;i<10;i++{
			consumer.Consume()
		}
        time.Sleep(time.Second)
        pinger <- 1
    	}
	}

	ping := make(chan int)
    pong := make(chan int)

    go prod(ping, pong)
    go cons(ping, pong)

    // The main goroutine starts the ping/pong by sending into the ping channel
    ping <- 1

    for {
        // Block the main thread until an interrupt
        time.Sleep(time.Second)
    }
	// router.Post("/users/",saveuser.New(log,store))
	// router.Delete("/users/{id}",deleteuser.New(log,store))
	// router.Put("/users/{id}",updateuser.New(log,store))
	// router.Get("/user/{name}-{surname}-{patr}-{age}-{sex}-{nationality}")


	// log.Info("Starting server...",slog.String("address",cfg.Address))


	// srv := &http.Server{
	// 	Addr:  cfg.Address,
	// 	Handler: router,
	// 	ReadTimeout: cfg.Timeout,
	// 	WriteTimeout: cfg.Timeout,
	// 	IdleTimeout: cfg.IdleTimeout,
	// }

	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil {
	// 		log.Error("failed to start server")
	// 	}
	// }()

	
	// id ,err := store.DeleteUser(2)
	// if err!=nil{
	// 	log.Error("failed to save user ",
	// 	slog.M{
	// 		"error": err.Error(),
	// 	})
	// 	os.Exit(1)
	// }
	// fmt.Println(id)
	// var patr string ="Huesosovich"
	// testUser:= config.GenUser{ 
	// 	Name:"Dmitriy",
	// 	Surname: "Ushakov",
	// 	Patronymic: &patr,
	// 	Age:42,
	// 	Sex:"male",
	// 	Nationality: []config.UserCountry{{CountryID:"RU",Probability:0.419}},
	// }
	// err = store.UpdateUser(7,testUser)
	// if err!=nil{
	// 	log.Error("failed to save user ",
	// 	slog.M{
	// 		"error": err.Error(),
	// 	})
	// 	os.Exit(1)
	// }
	// _=id

	// var patr string
	// testUser:= config.GenUser{ 
	// 	Name:"Dmitriy",
	// 	Surname: "Ushakov",
	// 	Patronymic: &patr,
	// 	Age:42,
	// 	Sex:"male",
	// 	Nationality: []config.UserCountry{{CountryID:"UA",Probability:0.419}},
	// }
	// id ,err := store.SaveUser(testUser)
	// if err!=nil{
	// 	log.Error("failed to save user ",
	// 	slog.M{
	// 		"error": err.Error(),
	// 	})
	// 	os.Exit(1)
	// }
	// _=id

	//logger : slog

	//storage : pg

	//router : chi, render

	//run :
}

