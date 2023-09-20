package main

import (
	"context"
	"fmt"
	"main/internal"
	config "main/internal"
	logger "main/internal"
	storage "main/internal"
	"main/internal/kafka/pipeline"
	kafka_test "main/internal/kafka/tests"
	models "main/internal/lib/api/model/user"
	mwLogger "main/internal/middleware/logger"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/brianvoe/gofakeit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func FakeUserGenerator(n int) []models.User {
	ret := make([]models.User, n)
	for i := 0; i < n; i++ {
		u := gofakeit.Person()
		user := models.New()
		user.Name = u.FirstName
		user.Surname = u.LastName
		user.Patronymic = "Sanich"
		user.Sex = u.Gender
		user.Nationality = u.Address.Country
		user.Age = rand.Intn(100)
		ret[i] = user
	}
	return ret
}



func main() {
	//config :cleanENV
	cfg, err := config.LoadConfig()
	fmt.Println(cfg, err)
	log := logger.SetupLogger(cfg.Env)

	store, err := storage.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", internal.Err(err))
		os.Exit(1);
	} else {
		log.Info("DB init success")
	}
	// a := FakeUserGenerator(30)
	// for _, v := range a {
	// 	fmt.Println("user", v)
	// }
	// b, err := store.SaveUser(log, a...)
	// fmt.Println(b, err)
	_ = store
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	//router.Get("/user/{name}-{surname}-{patr}-{age}-{sex}-{nationality}")
	// test, err := enrichment.Test()
	// if err != nil {
	// 	log.Error("failed to enrich", internal.Err(err))
	// } else {
	// 	fmt.Println(models.User(*test))
	// }

	// k, err := url.Parse("users/?age=lt~50&sex=male&page=1&perpage=20")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// for _, v := range k.Query() {
	// 	for k := range v {
	// 		fmt.Println(strings.Split(v[k], "~"))
	// 	}
	// }
	// fmt.Println(store.GetUsers(k.Query()))
	kafka_test.Populate(100, log, cfg)
	ctx:=context.Background()
	p:=pipeline.New(&ctx,log,cfg,store)
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL)

	// go routine for getting signals asynchronously
	go func() {
		sig := <-signals
		log.Info("%s Got signal: %w", "main",sig)
		os.Exit(1)
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		defer wg.Done()
		p.Start()
	}()
	wg.Wait()
	
	// router.Post("/users/",saveuser.New(log,store))
	// router.Delete("/users/{id}",deleteuser.New(log,store))
	// router.Patch("/users/{id}",updateuser.New(log,store))

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
}
