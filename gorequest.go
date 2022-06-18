// Package gorequest inspired by Nodejs SuperAgent provides easy-way to write http client
package gorequest

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptrace"
	"net/textproto"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/wklken/gorequest/internal/json"
	"golang.org/x/net/publicsuffix"
)

type (
	Request  *http.Request
	Response *http.Response
)

// A SuperAgent is a object storing all request data for client.
type SuperAgent struct {
	Url                  string
	Method               string
	Header               http.Header
	TargetType           string
	ForceType            string
	Data                 map[string]interface{}
	SliceData            []interface{}
	FormData             url.Values
	QueryData            url.Values
	FileData             []File
	BounceToRawString    bool
	RawString            string
	Client               *http.Client
	Transport            *http.Transport
	Cookies              []*http.Cookie
	Errors               []error
	BasicAuth            basicAuth
	Debug                bool
	CurlCommand          bool
	logger               Logger
	Retryable            superAgentRetryable
	DoNotClearSuperAgent bool
	isClone              bool
	ctx                  context.Context
	trace                *httptrace.ClientTrace
	Stats                Stats
	isMock               bool
}

var DisableTransportSwap = false

func newHttpClient() *http.Client {
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)
	return &http.Client{Jar: jar}
}

// New used to create a new SuperAgent object.
func New() *SuperAgent {
	debug := os.Getenv("GOREQUEST_DEBUG") == "1"

	s := &SuperAgent{
		TargetType:        TypeJSON,
		Data:              make(map[string]interface{}),
		Header:            http.Header{},
		RawString:         "",
		SliceData:         []interface{}{},
		FormData:          url.Values{},
		QueryData:         url.Values{},
		FileData:          make([]File, 0),
		BounceToRawString: false,
		Client:            newHttpClient(),
		Transport:         &http.Transport{},
		Cookies:           make([]*http.Cookie, 0),
		Errors:            nil,
		BasicAuth:         basicAuth{},
		Debug:             debug,
		CurlCommand:       false,
		logger:            log.New(os.Stderr, "[gorequest]", log.LstdFlags),
		isClone:           false,
		ctx:               nil,
		trace:             nil,
		Stats:             Stats{},
		isMock:            false,
	}
	// disable keep alives by default, see this issue https://github.com/parnurzeal/gorequest/issues/75
	s.Transport.DisableKeepAlives = true
	return s
}

// Clone returns a copy of this superAgent. Useful if you want to reuse the client/settings
// concurrently.
// Note: This does a shallow copy of the parent. So you will need to be
// careful of Data provided
// Note: It also directly re-uses the client and transport. If you modify the Timeout,
// or RedirectPolicy on a clone, the clone will have a new http.client. It is recommended
// that the base request set your timeout and redirect polices, and no modification of
// the client or transport happen after cloning.
// Note: DoNotClearSuperAgent is forced to "true" after Clone
func (s *SuperAgent) Clone() *SuperAgent {
	clone := &SuperAgent{
		Url:                  s.Url,
		Method:               s.Method,
		Header:               http.Header(cloneMapArray(s.Header)),
		TargetType:           s.TargetType,
		ForceType:            s.ForceType,
		Data:                 shallowCopyData(s.Data),
		SliceData:            shallowCopyDataSlice(s.SliceData),
		FormData:             url.Values(cloneMapArray(s.FormData)),
		QueryData:            url.Values(cloneMapArray(s.QueryData)),
		FileData:             shallowCopyFileArray(s.FileData),
		BounceToRawString:    s.BounceToRawString,
		RawString:            s.RawString,
		Client:               s.Client,
		Transport:            s.Transport,
		Cookies:              shallowCopyCookies(s.Cookies),
		Errors:               shallowCopyErrors(s.Errors),
		BasicAuth:            s.BasicAuth,
		Debug:                s.Debug,
		CurlCommand:          s.CurlCommand,
		logger:               s.logger, // thread safe.. anyway
		Retryable:            copyRetryable(s.Retryable),
		DoNotClearSuperAgent: true,
		isClone:              true,
		ctx:                  s.ctx,
		trace:                s.trace,
		Stats:                copyStats(s.Stats),
		isMock:               s.isMock,
	}
	return clone
}

func (s *SuperAgent) Context(ctx context.Context) *SuperAgent {
	s.ctx = ctx
	return s
}

// SetDoNotClearSuperAgent enable the DoNotClear mode for not clearing super agent and reuse for the next request.
func (s *SuperAgent) SetDoNotClearSuperAgent(enable bool) *SuperAgent {
	s.DoNotClearSuperAgent = enable
	return s
}

// DisableCompression disable the compression of http.Client.
func (s *SuperAgent) DisableCompression() *SuperAgent {
	s.Transport.DisableCompression = true
	return s
}

// ClearSuperAgent clear SuperAgent data for another new request.
func (s *SuperAgent) ClearSuperAgent() {
	if s.DoNotClearSuperAgent {
		return
	}
	s.Url = ""
	s.Method = ""
	s.Header = http.Header{}
	s.Data = make(map[string]interface{})
	s.SliceData = []interface{}{}
	s.FormData = url.Values{}
	s.QueryData = url.Values{}
	s.FileData = make([]File, 0)
	s.BounceToRawString = false
	s.RawString = ""
	s.ForceType = ""
	s.TargetType = TypeJSON
	s.Cookies = make([]*http.Cookie, 0)
	s.Errors = nil
	s.ctx = nil
	s.Stats = Stats{}
}

// CustomMethod is just a wrapper to initialize SuperAgent instance by method string.
func (s *SuperAgent) CustomMethod(method, targetUrl string) *SuperAgent {
	switch method {
	case POST:
		return s.Post(targetUrl)
	case GET:
		return s.Get(targetUrl)
	case HEAD:
		return s.Head(targetUrl)
	case PUT:
		return s.Put(targetUrl)
	case DELETE:
		return s.Delete(targetUrl)
	case PATCH:
		return s.Patch(targetUrl)
	case OPTIONS:
		return s.Options(targetUrl)
	default:
		s.ClearSuperAgent()
		s.Method = method
		s.Url = targetUrl
		s.Errors = nil
		return s
	}
}

func (s *SuperAgent) Get(targetUrl string) *SuperAgent {
	s.ClearSuperAgent()
	s.Method = GET
	s.Url = targetUrl
	s.Errors = nil
	s.TargetType = ""
	return s
}

func (s *SuperAgent) Post(targetUrl string) *SuperAgent {
	s.ClearSuperAgent()
	s.Method = POST
	s.Url = targetUrl
	s.Errors = nil
	return s
}

func (s *SuperAgent) Head(targetUrl string) *SuperAgent {
	s.ClearSuperAgent()
	s.Method = HEAD
	s.Url = targetUrl
	s.Errors = nil
	return s
}

func (s *SuperAgent) Put(targetUrl string) *SuperAgent {
	s.ClearSuperAgent()
	s.Method = PUT
	s.Url = targetUrl
	s.Errors = nil
	return s
}

func (s *SuperAgent) Delete(targetUrl string) *SuperAgent {
	s.ClearSuperAgent()
	s.Method = DELETE
	s.Url = targetUrl
	s.Errors = nil
	return s
}

func (s *SuperAgent) Patch(targetUrl string) *SuperAgent {
	s.ClearSuperAgent()
	s.Method = PATCH
	s.Url = targetUrl
	s.Errors = nil
	return s
}

func (s *SuperAgent) Options(targetUrl string) *SuperAgent {
	s.ClearSuperAgent()
	s.Method = OPTIONS
	s.Url = targetUrl
	s.Errors = nil
	return s
}

// Type is a convenience function to specify the data type to send.
// For example, to send data as `application/x-www-form-urlencoded` :
//
//    gorequest.New().
//      Post("/recipe").
//      Type("form").
//      Send(`{ "name": "egg benedict", "category": "brunch" }`).
//      End()
//
// This will POST the body "name=egg benedict&category=brunch" to url /recipe
//
// GoRequest supports
//
//    "text/html" uses "html"
//    "application/json" uses "json"
//    "application/xml" uses "xml"
//    "text/plain" uses "text"
//    "application/x-www-form-urlencoded" uses "urlencoded", "form" or "form-data"
//
func (s *SuperAgent) Type(typeStr string) *SuperAgent {
	if _, ok := Types[typeStr]; ok {
		s.ForceType = typeStr
	} else {
		s.Errors = append(s.Errors, fmt.Errorf("type func: incorrect type \"%s\"", typeStr))
	}
	return s
}

// Query function accepts either json string or strings which will form a query-string in url of GET method or body of POST method.
// For example, making "/search?query=bicycle&size=50x50&weight=20kg" using GET method:
//
//      gorequest.New().
//        Get("/search").
//        Query(`{ query: 'bicycle' }`).
//        Query(`{ size: '50x50' }`).
//        Query(`{ weight: '20kg' }`).
//        End()
//
// Or you can put multiple json values:
//
//      gorequest.New().
//        Get("/search").
//        Query(`{ query: 'bicycle', size: '50x50', weight: '20kg' }`).
//        End()
//
// Strings are also acceptable:
//
//      gorequest.New().
//        Get("/search").
//        Query("query=bicycle&size=50x50").
//        Query("weight=20kg").
//        End()
//
// Or even Mixed! :)
//
//      gorequest.New().
//        Get("/search").
//        Query("query=bicycle").
//        Query(`{ size: '50x50', weight:'20kg' }`).
//        End()
//
func (s *SuperAgent) Query(content interface{}) *SuperAgent {
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		s.queryString(v.String())
	case reflect.Struct:
		s.queryStruct(v.Interface())
	case reflect.Map:
		s.queryMap(v.Interface())
	default:
	}
	return s
}

func (s *SuperAgent) queryStruct(content interface{}) *SuperAgent {
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
				// k = strings.ToLower(k)
				var queryVal string
				switch t := v.(type) {
				case string:
					queryVal = t
				case float64:
					queryVal = strconv.FormatFloat(t, 'f', -1, 64)
				case time.Time:
					queryVal = t.Format(time.RFC3339)
				case stdjson.Number:
					queryVal = string(t)
				default:
					j, err := json.Marshal(v)
					if err != nil {
						continue
					}
					queryVal = BytesToString(j)
				}
				s.QueryData.Add(k, queryVal)
			}
		}
	}
	return s
}

func (s *SuperAgent) queryString(content string) *SuperAgent {
	var val map[string]string
	if err := json.Unmarshal(StringToBytes(content), &val); err == nil {
		for k, v := range val {
			s.QueryData.Add(k, v)
		}
	} else {
		if queryData, err := url.ParseQuery(content); err == nil {
			for k, queryValues := range queryData {
				for _, queryValue := range queryValues {
					s.QueryData.Add(k, queryValue)
				}
			}
		} else {
			s.Errors = append(s.Errors, err)
		}
		// TODO: need to check correct format of 'field=val&field=val&...'
	}
	return s
}

func (s *SuperAgent) queryMap(content interface{}) *SuperAgent {
	return s.queryStruct(content)
}

// Param as Go conventions accepts ; as a synonym for &. (https://github.com/golang/go/issues/2210)
// Thus, Query won't accept ; in a query string if we provide something like fields=f1;f2;f3
// This Param is then created as an alternative method to solve this.
func (s *SuperAgent) Param(key string, value string) *SuperAgent {
	s.QueryData.Add(key, value)
	return s
}

// Send function accepts either json string or query strings which is usually used to assign data to POST or PUT method.
// Without specifying any type, if you give Send with json data, you are doing requesting in json format:
//
//      gorequest.New().
//        Post("/search").
//        Send(`{ query: 'foo' }`).
//        End()
//
// While if you use at least one of query string, GoRequest understands and automatically set the Content-Type to `application/x-www-form-urlencoded`
//
//      gorequest.New().
//        Post("/search").
//        Send("query=foo").
//        End()
//
// So, if you want to strictly send json format, you need to use Type func to set it as `json` (Please see more details in Type function).
// You can also do multiple chain of Send:
//
//      gorequest.New().
//        Post("/search").
//        Send("query=bicycle&size=50x50").
//        Send(`{ wheel: '4'}`).
//        End()
//
// From v0.2.0, Send function provide another convenience way to work with Struct type. You can mix and match it with json and query string:
//
//      type BrowserVersionSupport struct {
//        Chrome string
//        Firefox string
//      }
//      ver := BrowserVersionSupport{ Chrome: "37.0.2041.6", Firefox: "30.0" }
//      gorequest.New().
//        Post("/update_version").
//        Send(ver).
//        Send(`{"Safari":"5.1.10"}`).
//        End()
//
// If you have set Type to text or Content-Type to text/plain, content will be sent as raw string in body instead of form
//
//      gorequest.New().
//        Post("/greet").
//        Type("text").
//        Send("hello world").
//        End()
//
func (s *SuperAgent) Send(content interface{}) *SuperAgent {
	// TODO: add normal text mode or other mode to Send func
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		s.SendString(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // includes rune
		s.SendString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // includes byte
		s.SendString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float64:
		s.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Float32:
		s.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 32))
	case reflect.Bool:
		s.SendString(strconv.FormatBool(v.Bool()))
	case reflect.Struct:
		s.SendStruct(v.Interface())
	case reflect.Slice:
		s.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Array:
		s.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Ptr:
		s.Send(v.Elem().Interface())
	case reflect.Map:
		s.SendMap(v.Interface())
	default:
		// TODO: leave default for handling other types in the future, such as complex numbers, (nested) maps, etc
		return s
	}
	return s
}

// SendSlice (similar to SendString) returns SuperAgent's itself for any next chain and takes content []interface{} as a parameter.
// Its duty is to append slice of interface{} into s.SliceData ([]interface{}) which later changes into json array in the End() func.
func (s *SuperAgent) SendSlice(content []interface{}) *SuperAgent {
	s.SliceData = append(s.SliceData, content...)
	return s
}

func (s *SuperAgent) SendMap(content interface{}) *SuperAgent {
	return s.SendStruct(content)
}

// SendStruct (similar to SendString) returns SuperAgent's itself for any next chain and takes content interface{} as a parameter.
// Its duty is to transform interface{} (implicitly always a struct) into s.Data (map[string]interface{}) which later changes into appropriate format such as json, form, text, etc. in the End() func.
func (s *SuperAgent) SendStruct(content interface{}) *SuperAgent {
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
				s.Data[k] = v
			}
		}
	}
	return s
}

// SendString returns SuperAgent's itself for any next chain and takes content string as a parameter.
// Its duty is to transform String into s.Data (map[string]interface{}) which later changes into appropriate format such as json, form, text, etc. in the End func.
// Send implicitly uses SendString and you should use Send instead of this.
func (s *SuperAgent) SendString(content string) *SuperAgent {
	if !s.BounceToRawString {
		var val interface{}
		d := json.NewDecoder(strings.NewReader(content))
		d.UseNumber()
		if err := d.Decode(&val); err == nil {
			switch v := reflect.ValueOf(val); v.Kind() {
			case reflect.Map:
				for k, v := range val.(map[string]interface{}) {
					s.Data[k] = v
				}
				// NOTE: if SendString(`{}`), will come into this case, but set nothing into s.Data
				if len(s.Data) == 0 {
					s.BounceToRawString = true
				}
			// add to SliceData
			case reflect.Slice:
				s.SendSlice(val.([]interface{}))
			// bounce to raw string if it is array json, or others
			default:
				s.BounceToRawString = true
			}
		} else if formData, err := url.ParseQuery(content); err == nil {
			for k, formValues := range formData {
				for _, formValue := range formValues {
					// make it array if already have key
					if val, ok := s.Data[k]; ok {
						var strArray []string
						strArray = append(strArray, formValue)
						// check if previous data is one string or array
						switch oldValue := val.(type) {
						case []string:
							strArray = append(strArray, oldValue...)
						case string:
							strArray = append(strArray, oldValue)
						}
						s.Data[k] = strArray
					} else {
						// make it just string if does not already have same key
						s.Data[k] = formValue
					}
				}
			}
			s.TargetType = TypeForm
		} else {
			s.BounceToRawString = true
		}
	}
	// Dump all contents to RawString in case in the end user doesn't want json or form.
	s.RawString += content
	return s
}

// End is the most important function that you need to call when ending the chain. The request won't proceed without calling it.
// End function returns Response which matches the structure of Response type in Golang's http package (but without Body data). The body data itself returns as a string in a 2nd return value.
// Lastly but worth noticing, error array (NOTE: not just single error value) is returned as a 3rd value and nil otherwise.
//
// For example:
//
//    resp, body, errs := gorequest.New().Get("http://www.google.com").End()
//    if errs != nil {
//      fmt.Println(errs)
//    }
//    fmt.Println(resp, body)
//
// Moreover, End function also supports callback which you can put as a parameter.
// This extends the flexibility and makes GoRequest fun and clean! You can use GoRequest in whatever style you love!
//
// For example:
//
//    func printBody(resp gorequest.Response, body string, errs []error){
//      fmt.Println(resp.Status)
//    }
//    gorequest.New().Get("http://www..google.com").End(printBody)
//
func (s *SuperAgent) End(callback ...func(response Response, body string, errs []error)) (Response, string, []error) {
	var bytesCallback []func(response Response, body []byte, errs []error)
	if len(callback) > 0 {
		bytesCallback = []func(response Response, body []byte, errs []error){
			func(response Response, body []byte, errs []error) {
				callback[0](response, BytesToString(body), errs)
			},
		}
	}

	resp, body, errs := s.EndBytes(bytesCallback...)
	bodyString := BytesToString(body)

	return resp, bodyString, errs
}

// EndBytes should be used when you want the body as bytes. The callbacks work the same way as with `End`, except that a byte array is used instead of a string.
func (s *SuperAgent) EndBytes(
	callback ...func(response Response, body []byte, errs []error),
) (Response, []byte, []error) {
	var (
		errs []error
		resp Response
		body []byte
	)

	for {
		resp, body, errs = s.getResponseBytes()
		// if errs != nil {
		// 	return nil, nil, errs
		// }
		if !s.shouldRetry(resp, len(errs) > 0) {
			if resp != nil {
				resp.Header.Set("Retry-Count", strconv.Itoa(s.Retryable.Attempt))
			}
			break
		}

		s.Errors = nil
	}

	if len(callback) != 0 {
		if resp == nil {
			callback[0](nil, body, s.Errors)
		} else {
			respCallback := *resp
			callback[0](&respCallback, body, s.Errors)
		}
	}
	return resp, body, errs
}

// EndStruct should be used when you want the body as a struct. The callbacks work the same way as with `End`, except that a struct is used instead of a string.
func (s *SuperAgent) EndStruct(
	v interface{},
	callback ...func(response Response, v interface{}, body []byte, errs []error),
) (Response, []byte, []error) {
	resp, body, errs := s.EndBytes()
	if errs != nil {
		return nil, body, errs
	}

	err := json.Unmarshal(body, &v)
	if err != nil {
		respContentType := filterFlags(resp.Header.Get("Content-Type"))
		if respContentType != MIMEJSON {
			s.Errors = append(
				s.Errors,
				fmt.Errorf(
					"response content-type is %s not application/json, so can't be json decoded: %w",
					respContentType,
					err,
				),
			)
		} else {
			s.Errors = append(s.Errors, fmt.Errorf("response body json decode fail: %w", err))
		}

		return resp, body, s.Errors
	}
	respCallback := *resp
	if len(callback) != 0 {
		callback[0](&respCallback, v, body, s.Errors)
	}
	return resp, body, nil
}

func (s *SuperAgent) getResponseBytes() (Response, []byte, []error) {
	var (
		req  *http.Request
		err  error
		resp Response
	)
	// check whether there is an error. if yes, return all errors
	if len(s.Errors) != 0 {
		return nil, nil, s.Errors
	}

	// Make Request
	req, err = s.MakeRequest()
	if err != nil {
		s.Errors = append(s.Errors, err)
		return nil, nil, s.Errors
	}

	// Set Transport
	if !DisableTransportSwap && !s.isMock {
		s.Client.Transport = s.Transport
	}

	// Log details of this request
	s.debuggingRequest(req)

	// Display CURL command line
	s.logCurlCommand(req)

	startTime := time.Now()
	// stats collect the requestBytes
	s.Stats.RequestBytes = req.ContentLength

	// Send request
	resp, err = s.Client.Do(req)
	if err != nil {
		s.Errors = append(s.Errors, err)
		return nil, nil, s.Errors
	}
	defer resp.Body.Close()

	// stats collect the RequestDuration
	s.Stats.RequestDuration = time.Since(startTime)

	// Log details of this response
	s.debuggingResponse(resp)

	body, err := ioutil.ReadAll(resp.Body)
	// Reset resp.Body so it can be use again
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, []error{err}
	}

	// stats collect the responseBytes
	s.Stats.ResponseBytes = int64(len(body))
	return resp, body, nil
}

func (s *SuperAgent) MakeRequest() (*http.Request, error) {
	var (
		req           *http.Request
		contentType   string // This is only set when the request body content is non-empty.
		contentReader io.Reader
		err           error
	)

	// check if there is forced type
	switch s.ForceType {
	case TypeJSON, TypeForm, TypeXML, TypeText, TypeMultipart:
		s.TargetType = s.ForceType
		// If forceType is not set, check whether user set Content-Type header.
		// If yes, also bounce to the correct supported TargetType automatically.
	default:
		contentType := s.Header.Get("Content-Type")
		for k, v := range Types {
			if contentType == v {
				s.TargetType = k
			}
		}
	}

	// if slice and map get mixed, let's bounce to raw string
	if len(s.Data) != 0 && len(s.SliceData) != 0 {
		s.BounceToRawString = true
	}

	if s.Method == "" {
		return nil, fmt.Errorf("no method specified")
	}

	// !!! Important Note !!!
	//
	// Throughout this region, contentReader and contentType are only set when
	// the contents will be non-empty.
	// This is done avoid ever sending a non-nil request body with nil contents
	// to http.NewRequest, because it contains logic which depended on
	// whether or not the body is "nil".
	//
	// See PR #136 for more information:
	//
	//     https://github.com/parnurzeal/gorequest/pull/136
	//
	switch s.TargetType {
	case TypeJSON:
		// If-case to give support to json array. we check if
		// 1) Map only: send it as json map from s.Data
		// 2) Array or Mix of map & array or others: send it as raw string from s.RawString
		var contentJson []byte
		if s.BounceToRawString {
			contentJson = StringToBytes(s.RawString)
		} else if len(s.Data) != 0 {
			contentJson, _ = json.Marshal(s.Data)
		} else if len(s.SliceData) != 0 {
			contentJson, _ = json.Marshal(s.SliceData)
		}
		if contentJson != nil {
			contentReader = bytes.NewReader(contentJson)
			contentType = "application/json"
		}
	case TypeForm, TypeFormData, TypeUrlencoded:
		var contentForm []byte
		if s.BounceToRawString || len(s.SliceData) != 0 {
			contentForm = StringToBytes(s.RawString)
		} else {
			formData := changeMapToURLValues(s.Data)
			contentForm = StringToBytes(formData.Encode())
		}
		if len(contentForm) != 0 {
			contentReader = bytes.NewReader(contentForm)
			contentType = "application/x-www-form-urlencoded"
		}
	case TypeText:
		if len(s.RawString) != 0 {
			contentReader = strings.NewReader(s.RawString)
			contentType = "text/plain"
		}
	case TypeXML:
		if len(s.RawString) != 0 {
			contentReader = strings.NewReader(s.RawString)
			contentType = "application/xml"
		}
	case TypeMultipart:
		var (
			buf = &bytes.Buffer{}
			mw  = multipart.NewWriter(buf)
		)

		if s.BounceToRawString {
			fieldName := s.Header.Get("data_fieldname")
			if fieldName == "" {
				fieldName = "data"
			}
			fw, _ := mw.CreateFormField(fieldName)
			fw.Write(StringToBytes(s.RawString))
			contentReader = buf
		}

		if len(s.Data) != 0 {
			formData := changeMapToURLValues(s.Data)
			for key, values := range formData {
				for _, value := range values {
					fw, _ := mw.CreateFormField(key)
					fw.Write(StringToBytes(value))
				}
			}
			contentReader = buf
		}

		if len(s.SliceData) != 0 {
			fieldName := s.Header.Get("json_fieldname")
			if fieldName == "" {
				fieldName = "data"
			}
			// copied from CreateFormField() in mime/multipart/writer.go
			h := make(textproto.MIMEHeader)
			fieldName = strings.Replace(strings.Replace(fieldName, "\\", "\\\\", -1), `"`, "\\\"", -1)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"`, fieldName))
			h.Set("Content-Type", "application/json")
			fw, _ := mw.CreatePart(h)
			contentJson, err := json.Marshal(s.SliceData)
			if err != nil {
				return nil, err
			}
			fw.Write(contentJson)
			contentReader = buf
		}

		// add the files
		if len(s.FileData) != 0 {
			for _, file := range s.FileData {
				fw, _ := CreateFormFile(mw, file.Fieldname, file.Filename, file.MimeType)
				fw.Write(file.Data)
			}
			contentReader = buf
		}

		// close before call to FormDataContentType ! otherwise, it's not valid multipart
		mw.Close()

		if contentReader != nil {
			contentType = mw.FormDataContentType()
		}
	case "":
		contentType = ""
		contentReader = nil
	default:
		// let's return an error instead of an nil pointer exception here
		return nil, fmt.Errorf("TargetType '%s' could not be determined", s.TargetType)
	}

	if req, err = http.NewRequest(s.Method, s.Url, contentReader); err != nil {
		return nil, err
	}

	if s.ctx != nil {
		req = req.WithContext(s.ctx)
	}
	if s.trace != nil {
		clientTraceCtx := httptrace.WithClientTrace(req.Context(), s.trace)
		req = req.WithContext(clientTraceCtx)
	}

	for k, values := range s.Header {
		for _, v := range values {
			req.Header.Add(k, v)
		}

		// Setting the Host header is a special case, see this issue: https://github.com/golang/go/issues/7682
		if strings.EqualFold(k, "Host") {
			req.Host = values[0]
		}
	}

	// https://github.com/parnurzeal/gorequest/issues/164
	// Don't infer the content type header if an override is already provided.
	if len(contentType) != 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Add all query string from Query func
	q := req.URL.Query()
	for k, v := range s.QueryData {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	req.URL.RawQuery = q.Encode()

	// Add basic auth
	if s.BasicAuth.Username != "" || s.BasicAuth.Password != "" {
		req.SetBasicAuth(s.BasicAuth.Username, s.BasicAuth.Password)
	}

	// Add cookies
	for _, cookie := range s.Cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

// we don't want to mess up other clones when we modify the client..
// so unfortunately we need to create a new client
func (s *SuperAgent) safeModifyHttpClient() {
	if !s.isClone {
		return
	}
	oldClient := s.Client
	s.Client = &http.Client{}
	s.Client.Jar = oldClient.Jar
	s.Client.Transport = oldClient.Transport
	s.Client.Timeout = oldClient.Timeout
	s.Client.CheckRedirect = oldClient.CheckRedirect
}

// does a shallow clone of the transport
func (s *SuperAgent) safeModifyTransport() {
	if !s.isClone {
		return
	}
	oldTransport := s.Transport
	s.Transport = &http.Transport{
		Proxy:                 oldTransport.Proxy,
		DialContext:           oldTransport.DialContext,
		Dial:                  oldTransport.Dial,
		DialTLSContext:        oldTransport.DialTLSContext,
		DialTLS:               oldTransport.DialTLS,
		TLSClientConfig:       oldTransport.TLSClientConfig,
		TLSHandshakeTimeout:   oldTransport.TLSHandshakeTimeout,
		DisableKeepAlives:     oldTransport.DisableKeepAlives,
		DisableCompression:    oldTransport.DisableCompression,
		MaxIdleConns:          oldTransport.MaxIdleConns,
		MaxIdleConnsPerHost:   oldTransport.MaxIdleConnsPerHost,
		MaxConnsPerHost:       oldTransport.MaxConnsPerHost,
		IdleConnTimeout:       oldTransport.IdleConnTimeout,
		ResponseHeaderTimeout: oldTransport.ResponseHeaderTimeout,
		ExpectContinueTimeout: oldTransport.ExpectContinueTimeout,
		TLSNextProto:          oldTransport.TLSNextProto,
		ProxyConnectHeader:    oldTransport.ProxyConnectHeader,

		// new from go 1.16
		GetProxyConnectHeader: oldTransport.GetProxyConnectHeader,

		MaxResponseHeaderBytes: oldTransport.MaxResponseHeaderBytes,
		WriteBufferSize:        oldTransport.WriteBufferSize,
		ReadBufferSize:         oldTransport.ReadBufferSize,
		ForceAttemptHTTP2:      oldTransport.ForceAttemptHTTP2,
	}
}
