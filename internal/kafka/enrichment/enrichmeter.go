package enrichment

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	models "main/internal/lib/api/model/user"
	"net/http"
	"net/url"
	"strings"
	"time"
)




	
var (
	s=[]string{
		"https://api.genderize.io",
		"https://api.agify.io",
		"https://api.nationalize.io",
	}
	user = models.BaseUser{Name: "Vasiliy",Surname: "Ushakov"}
)


func GetUrls(s []string) ([]url.URL,error){
	const op = "storage.Enrichment.get"
	urls:=make([]url.URL,len(s));
	for i,v := range s{
		temp,err:=url.Parse(v)
		if err!=nil{
			return nil, fmt.Errorf("%s : %w",op,err)
		}
		urls[i]=*temp
	}
	return urls,nil
}



// func Test()(*models.User,error){
// 	const op = "storage.Enrichment.test"
// 	urls,err:=GetUrls(s)
// 	if err!=nil{
// 		return nil, fmt.Errorf("%s : %w",op,err)
// 	}
// 	return Enrichment(user,urls)
// }

func Enrichment(BaseUser models.BaseUser,s []string, log *slog.Logger) (*models.User,error){
	fromURLs,err:=GetUrls(s)
	if err!=nil{
		return nil,err
	}
	const op = "storage.Enrichment"
	retval:= make([]models.Enrichment,0)
	for _,v := range fromURLs{
		var data models.Enrichment
		q:=v.Query()
		q.Set("name",BaseUser.Name)
		v.RawQuery=q.Encode()
		client := http.Client{
    		Timeout: 3 * time.Second,
		}
		resp,err:=client.Get(v.String())
		if err!=nil{
			return nil, fmt.Errorf("%s : %w",op,err)
		}
		body, err :=io.ReadAll(resp.Body)
		if err!=nil{
			return nil, fmt.Errorf("%s : %w",op,err)
		}	
		log.Info(fmt.Sprintf("%s:fetched from URL:%v [%d]bytes",op,v.String(),len(body)))
		if strings.HasPrefix(v.String(), "https://api.nationalize.io"){
			var countryArray models.CountryArray
			err=json.Unmarshal(body,&countryArray)
			if err!=nil{
				return nil, fmt.Errorf("%s : %w",op,err)
			}
			if len(countryArray.Country)<1{
				return nil, fmt.Errorf("%s : Bad nationality data (len<1) %v %s",op,countryArray,BaseUser)
			}else{
				data.Nationality=countryArray.Country[0].CountryID
			}
		}else{
			err=json.Unmarshal(body,&data)
			if err!=nil{
				return nil, fmt.Errorf("%s : %w",op,err)
			}
		}
		retval = append(retval, data)
		resp.Body.Close()	
	}
	data,err:=Merge(retval);
	if err!=nil{
		return nil,err
	}
	return &models.User{
		BaseUser: &BaseUser,
		Enrichment: data,
	},nil
	
}

func Merge(data []models.Enrichment) (*models.Enrichment,error){
	const op = "storage.Enrichment.Merge"
	var mergedData models.Enrichment;
	counter:=0
	for _, v:= range data{
		switch {
			case v.Age>0:
				mergedData.Age=v.Age
				counter++
			case v.Sex!="":
				mergedData.Sex = v.Sex
				counter++
			case v.Nationality!="":
				mergedData.Nationality = v.Nationality
				counter++
		}
	}
	if counter>=3{
		return &mergedData,nil
	}
	fmt.Println(mergedData)
		return nil, fmt.Errorf("%s : %s",op,"Enrichment data error")
}