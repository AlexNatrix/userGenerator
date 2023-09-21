package updateuser

import (
	"errors"
	"fmt"
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

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserUpdater
type UserUpdater interface {
	UpdateUser(userID int64, user models.User) error
}

type UpdateRequest struct {
	Id int64 `json:"id"`
	Data models.User `json:"data"`
}


/**
Update user by ID in DB.
PATCH request must contain body, with ID:{id} and data:{user} in JSON format 
**/
func New(log *slog.Logger, userUpdater UserUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		req :=UpdateRequest{
			Id:-1,
			Data:models.NewUser(),
		}
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil && !req.Data.Validate() && req.Id==-1 {
			log.Error("failed to decode request body", internal.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}
		fmt.Println(req)
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", internal.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}


		err = userUpdater.UpdateUser(req.Id ,req.Data)
		if err != nil {
			log.Error("failed to update user", internal.Err(err))

			render.JSON(w, r, resp.Error("failed to update user"))

			return
		}

		log.Info("user updated")

		render.JSON(w, r, fmt.Sprintf("updated user with id:%d",req.Id))
	}
}


