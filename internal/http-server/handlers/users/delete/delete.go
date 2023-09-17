package deleteuser

import (
	"errors"
	"io"
	"log/slog"
	"main/internal"
	resp "main/internal/lib/api/response"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)



type DeleteRequest struct {
	Id int64 `json:"id"`
}

type UserDeleter interface {
	DeleteUser(userID int64) (int64, error)
}



func New(log *slog.Logger, userDeleter UserDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req DeleteRequest

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", internal.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", internal.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}


		id,err := userDeleter.DeleteUser(req.Id)
		if err != nil {
			log.Error("failed to delete user", internal.Err(err))

			render.JSON(w, r, resp.Error("failed to delete user"))

			return
		}

		log.Info("user updated")

		render.JSON(w, r, id)
	}
}


