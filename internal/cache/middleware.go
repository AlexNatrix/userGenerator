package cache

import (
	"net/http"
	models "usergenerator/internal/lib/api/model/user"

	"github.com/go-chi/render"
)



func (c *Cache) CacheHandler(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // w.Header().Set("Cache-Control", "public, max-age=3600")
        // w.Header().Set("Expires", time.Now().Add(time.Hour).Format(http.TimeFormat))
        // w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		var wanted []models.User
		key:=r.URL.String()
		err:=c.Exmpl.Get(c.CTX, key, &wanted);
		if  err == nil {
			render.JSON(w, r, wanted)
			return
		}
        // Call the next handler in the chain
        next.ServeHTTP(w, r)
    })
}