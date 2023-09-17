package updateuser

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



type UpdateRequest struct {
	Id int64 `json:"id"`
	*GenUser
}

type UserUpdater interface {
	UpdateUser(userID int64, user GenUser) error
}



func New(log *slog.Logger, userUpdater UserUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req UpdateRequest

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


		err = userUpdater.UpdateUser(req.Id ,req.GenUser)
		if err != nil {
			log.Error("failed to update user", internal.Err(err))

			render.JSON(w, r, resp.Error("failed to update user"))

			return
		}

		log.Info("user updated")

		render.JSON(w, r, true)
	}
}


