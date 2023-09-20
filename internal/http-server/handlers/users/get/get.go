package getusers

import (
	"fmt"
	"log/slog"
	"main/internal"
	models "main/internal/lib/api/model/user"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type OP int

const (
	EQ = iota
	NEQ
	MOR
	LES
)

type Param[T comparable] struct {
	Val T
	op  OP
}

type UserSearchParams struct {
	ID         Param[int]
	Name       Param[string]
	Surname    Param[string]
	Patronymic Param[string]
	Age        Param[int]
	Sex        Param[string]
	Page       int
	PerPage    int
}

type UserGetter interface {
	GetUsers(u UserSearchParams) ([]models.User, error)
}

func ParseQuery() (UserSearchParams, error) {
	return UserSearchParams{}, nil
}

func New(log *slog.Logger, userGetter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.getter.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		a := r.URL.Query()
		for k, v := range a {
			fmt.Println(k, v)
		}
		//err := render.DecodeJSON(r.Body, &req)
		// if errors.Is(err, io.EOF) {
		// 	// Такую ошибку встретим, если получили запрос с пустым телом.
		// 	// Обработаем её отдельно
		// 	log.Error("request body is empty")

		// 	render.JSON(w, r, resp.Error("empty request"))

		// 	return
		// }
		// if err != nil {
		// 	log.Error("failed to decode request body", internal.Err(err))

		// 	render.JSON(w, r, resp.Error("failed to decode request"))

		// 	return
		// }

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", internal.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}
		u, err := ParseQuery()
		id, err := userGetter.GetUsers(u)
		if err != nil {
			log.Error("failed to delete user", internal.Err(err))

			render.JSON(w, r, resp.Error("failed to delete user"))

			return
		}

		log.Info("user updated")

		render.JSON(w, r, id)
	}
}
