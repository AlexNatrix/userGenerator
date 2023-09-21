package getusers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	getusers "usergenerator/internal/http-server/handlers/users/get"
	"usergenerator/internal/http-server/handlers/users/get/mocks"
	models "usergenerator/internal/lib/api/model/user"
	"usergenerator/internal/middleware/slogdiscard"

	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {
	cases := []struct {
		name   string
		user   models.User  
		url    map[string][]string
		respError string
		mockError error
	}{
		{
		name :  "case1",
		user  : models.NewUser(),
		url   : make(map[string][]string),

		},
		{
				name :  "case2",
		user  : models.User{BaseUser:
							&models.BaseUser{Name:"",Surname: "",Patronymic: ""},
							Enrichment: 
							&models.Enrichment{Age:0,Sex: "",Nationality: ""},
						},
		url   : make(map[string][]string),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			userGetterMock := mocks.NewUserGetter(t)
			if tc.respError == "" || tc.mockError != nil {
				userGetterMock.On(
					"GetUsers", tc.url).
					Return([]models.User{tc.user}, tc.mockError).
					Once()

			}
			handler := getusers.New(slogdiscard.NewDiscardLogger(), userGetterMock,nil)
			input := fmt.Sprintf(`%s`, tc.user)
			req, err := http.NewRequest(http.MethodGet ,"/users", bytes.NewReader([]byte(input)))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			resp:=[]models.User{models.NewUser()}

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

		})
	}
}