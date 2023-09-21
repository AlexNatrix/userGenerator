package getusers

import (
	"fmt"
	"log/slog"
	"net/http"
	"usergenerator/internal"
	"usergenerator/internal/cache"
	models "usergenerator/internal/lib/api/model/user"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	redisCache "github.com/go-redis/cache/v9"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserGetter
type UserGetter interface {
	GetUsers(userQuery map[string][]string) ([]models.User, error)
}

/**
Fetch users from DB, 
Query must be in format users/?{param}={op}~{value}...&page={n}&per_page={m},
where op one of operators:
lt="less than",gt="greater than",eq="equal",neq="not equal",
iff operator doesnt exists, than by default eq operator will be used.
Example:users/?name=vitalya&age=lt~50&surname=gt~b
if page missed, than page will be equal to 1.
if per_page is missed, per_page will be equal to 100.
**/
func New(log *slog.Logger, userGetter UserGetter,c *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.getter.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		id, err := userGetter.GetUsers(r.URL.Query())
		if err != nil {
			log.Error("failed to fetch users", internal.Err(err))

			render.JSON(w, r, fmt.Errorf("failed to fetch user"))

			return
		}
		if c!=nil{
			key:=r.URL.String()
			item:=&redisCache.Item{
				Ctx:   c.CTX,
				Key:   key,
				Value: id,
				TTL:   c.TTL,
			}
			err = c.Exmpl.Set(item)
			if err != nil {
				log.Error("redis failed caching ", internal.Err(err))
			}
		}
		log.Info("fetched users")

		render.JSON(w, r, id)
	}
}



