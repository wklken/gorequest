package gorequest

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

func cloneMapArray(old map[string][]string) map[string][]string {
	newMap := make(map[string][]string, len(old))
	for k, vals := range old {
		newMap[k] = make([]string, len(vals))
		for i := range vals {
			newMap[k][i] = vals[i]
		}
	}
	return newMap
}

func shallowCopyData(old map[string]any) map[string]any {
	if old == nil {
		return nil
	}
	newData := make(map[string]any, len(old))
	for k, val := range old {
		newData[k] = val
	}
	return newData
}

func shallowCopyDataSlice(old []any) []any {
	if old == nil {
		return nil
	}
	newData := make([]any, len(old))
	copy(newData, old)
	return newData
}

func shallowCopyFileArray(old []File) []File {
	if old == nil {
		return nil
	}
	newData := make([]File, len(old))
	copy(newData, old)
	return newData
}

func shallowCopyErrors(old []error) []error {
	if old == nil {
		return nil
	}
	newData := make([]error, len(old))
	copy(newData, old)
	return newData
}

func statusesContains(statuses []int, respStatus int) bool {
	return slices.Contains(statuses, respStatus)
}

func makeSliceOfReflectValue(v reflect.Value) (slice []any) {
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return slice
	}

	slice = make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		slice[i] = v.Index(i).Interface()
	}

	return slice
}

func changeMapToURLValues(data map[string]any) url.Values {
	var newUrlValues = url.Values{}
	for k, v := range data {
		switch val := v.(type) {
		case string:
			newUrlValues.Add(k, val)
		case bool:
			newUrlValues.Add(k, strconv.FormatBool(val))
		// if a number, change to string
		// json.Number used to protect against a wrong (for GoRequest) default conversion
		// which always converts number to float64.
		// This type is caused by using Decoder.UseNumber()
		case json.Number:
			newUrlValues.Add(k, val.String())
		case int:
			newUrlValues.Add(k, strconv.FormatInt(int64(val), 10))
		// TODO add all other int-Types (int8, int16, ...)
		case float64:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		case float32:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		// following slices are mostly needed for tests
		case []string:
			for _, element := range val {
				newUrlValues.Add(k, element)
			}
		case []int:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatInt(int64(element), 10))
			}
		case []bool:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatBool(element))
			}
		case []float64:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		case []float32:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		// these slices are used in practice like sending a struct
		case []any:
			if len(val) <= 0 {
				continue
			}

			switch val[0].(type) {
			case string:
				for _, element := range val {
					newUrlValues.Add(k, element.(string))
				}
			case bool:
				for _, element := range val {
					newUrlValues.Add(k, strconv.FormatBool(element.(bool)))
				}
			case json.Number:
				for _, element := range val {
					newUrlValues.Add(k, element.(json.Number).String())
				}
			}
		default:
			// TODO add ptr, arrays, ...
		}
	}
	return newUrlValues
}

// ===========================================================

// Copyright 2020 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ===========================================================
// copy from gin/util.go of https://github.com/gin-gonic/gin
// MIT License
func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// CreateFormFile is a convenience wrapper around CreatePart. It creates
// a new form-data header with the provided field name and file name.
func CreateFormFile(w *multipart.Writer, fieldname, filename string, contenttype string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", contenttype)
	return w.CreatePart(h)
}
