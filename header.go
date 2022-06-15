package gorequest

import (
	"bytes"
	"encoding/json"
	"reflect"

	"github.com/spf13/cast"
)

// Set is used for setting header fields,
// this will overwrite the existed values of Header through AppendHeader().
// Example. To set `Accept` as `application/json`
//
//    gorequest.New().
//      Post("https://httpbin.org/post").
//      Set("Accept", "application/json").
//      End()
func (s *SuperAgent) Set(param string, value string) *SuperAgent {
	s.Header.Set(param, value)
	return s
}

// SetHeaders is used for setting all your headers with the use of a map or a struct.
// It uses AppendHeader() method so it allows for multiple values of the same header
// Example. To set the following struct as headers, simply do
//
//    headers := apiHeaders{Accept: "application/json", Content-Type: "text/html", X-Frame-Options: "deny"}
//    gorequest.New().
//      Post("apiEndPoint").
//      Set(headers).
//      End()
func (s *SuperAgent) SetHeaders(headers interface{}) *SuperAgent {
	switch v := reflect.ValueOf(headers); v.Kind() {
	case reflect.Struct:
		s.setHeadersStruct(v.Interface())
	case reflect.Map:
		s.setHeadersMap(v.Interface())
	default:
		return s
	}
	return s
}

func (s *SuperAgent) setHeadersMap(content interface{}) *SuperAgent {
	return s.setHeadersStruct(content)
}

// SendStruct (similar to SendString) returns SuperAgent's itself for any next chain and takes content interface{} as a parameter.
// Its duty is to transform interface{} (implicitly always a struct) into s.Data (map[string]interface{}) which later changes into appropriate format such as json, form, text, etc. in the End() func.
func (s *SuperAgent) setHeadersStruct(content interface{}) *SuperAgent {
	if marshalContent, err := json.Marshal(content); err != nil {
		s.Errors = append(s.Errors, err)
	} else {
		var val map[string]interface{}
		d := json.NewDecoder(bytes.NewBuffer(marshalContent))
		d.UseNumber()
		if err := d.Decode(&val); err != nil {
			s.Errors = append(s.Errors, err)
		} else {
			for k, v := range val {
				strValue, err := cast.ToStringE(v)
				if err != nil {
					// TODO: log err?
					continue
				}

				s.AppendHeader(k, strValue)
			}
		}
	}
	return s
}

// AppendHeader is used for setting headers with multiple values,
// Example. To set `Accept` as `application/json, text/plain`
//
//    gorequest.New().
//      Post("https://httpbin.org/post").
//      AppendHeader("Accept", "application/json").
//      AppendHeader("Accept", "text/plain").
//      End()
func (s *SuperAgent) AppendHeader(param string, value string) *SuperAgent {
	s.Header.Add(param, value)
	return s
}
