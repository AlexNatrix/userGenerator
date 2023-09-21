package main

import (
	"context"
	"fmt"
	"log/slog"
	"usergenerator/internal"
	config "usergenerator/internal"
	logger "usergenerator/internal"
	"usergenerator/internal/cache"
	deleteuser "usergenerator/internal/http-server/handlers/users/delete"
	getusers "usergenerator/internal/http-server/handlers/users/get"
	insertusers "usergenerator/internal/http-server/handlers/users/insert"
	updateuser "usergenerator/internal/http-server/handlers/users/update"
	"usergenerator/internal/kafka/consumer"
	"usergenerator/internal/kafka/pipeline"
	"usergenerator/internal/kafka/producer"
	mwLogger "usergenerator/internal/middleware/logger"
	storage "usergenerator/storage"

	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)



func main() {
	//config :cleanENV
	cfg, err := config.LoadConfig()
	fmt.Println(cfg, err)
	log := logger.SetupLogger(cfg.Env)

	store, err := storage.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", internal.Err(err))
		os.Exit(1)
	} else {
		log.Info("DB init success")
	}
	ctx:=context.Background()

	cache,err:=cache.New(cfg,ctx)
	if err!=nil{
		log.Error("failed to init cache", internal.Err(err))
	}

	_ = store
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	
	router.Post("/users", insertusers.New(log, store))
	router.Delete("/users", deleteuser.New(log, store))
	router.Patch("/users", updateuser.New(log, store))
	router.Get("/users", cache.CacheHandler(getusers.New(log, store, cache)))
	log.Info("Starting server...", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	p := pipeline.New(&ctx, log, cfg, store)

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)


	// go routine for getting signals asynchronously
	go func() {
		sig := <-signals
		log.Info(fmt.Sprintf("%s Got signal: %v", "main",sig))
		os.Exit(1)
	}()
	var wg sync.WaitGroup
	

	
	go func() {
		wg.Add(1)
		defer wg.Done()
		go consumer.Consumer(p.Ctx, p.CFG, log, p.In)
		go producer.Produce(p.Ctx, p.CFG, log, p.Failed)
		go p.Start()
	}()
	srv.ListenAndServe()
	wg.Wait()

}
