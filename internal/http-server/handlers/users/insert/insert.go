package insertusers

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"usergenerator/internal"
	models "usergenerator/internal/lib/api/model/user"
	resp "usergenerator/internal/lib/api/response"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserInserter

type UserInserter interface {
	InsertUsers(users ...models.User) ([]int64, error)
}

type InsertRequest struct {
	data []models.User
}


/**
Insert users into DB in bulk.
POST request must conatins body, with array of users in JSON format
**/
func New(log *slog.Logger, urlSaver UserInserter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.insert.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req InsertRequest

		err := render.DecodeJSON(r.Body, &req.data)
		if errors.Is(err, io.EOF) {
			
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", internal.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req.data[0]))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", internal.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		id, err := urlSaver.InsertUsers(req.data...)
		if err != nil {
			log.Error("failed to insert users", internal.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to insert users"))

			return
		}

		log.Info("user inserted. FirtsID:", slog.Int64("id", id[0]))

		render.JSON(w, r, id)
	}
}
