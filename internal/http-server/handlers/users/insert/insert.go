package saveusers

import (
	"errors"
	"io"
	"log/slog"
	"main/internal"
	models "main/internal/lib/api/model/user"
	resp "main/internal/lib/api/response"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type SaveRequest struct {
	data []models.User
}

type UserSaver interface {
	SaveUser(users ...models.User) ([]int64, error)
}

func New(log *slog.Logger, urlSaver UserSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req SaveRequest

		err := render.DecodeJSON(r.Body, &req.data)
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

		log.Info("request body decoded", slog.Any("request", req.data[0]))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", internal.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		id, err := urlSaver.SaveUser(req.data...)
		if err != nil {
			log.Error("failed to add users", internal.Err(err))

			render.JSON(w, r, resp.Error("failed to add users"))

			return
		}

		log.Info("user added", slog.Int64("id", id[0]))

		render.JSON(w, r, id)
	}
}
