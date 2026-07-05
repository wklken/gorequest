package gorequest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type File struct {
	Filename  string
	Fieldname string
	MimeType  string
	Data      []byte
}

// SendFile function works only with type "multipart". The function accepts one mandatory and up to three optional arguments. The mandatory (first) argument is the file.
// The function accepts a path to a file as string:
//
//	gorequest.New().
//	  Post("http://example.com").
//	  Type("multipart").
//	  SendFile("./example_file.ext").
//	  End()
//
// File can also be a []byte slice of a already file read by eg. os.ReadFile:
//
//	b, _ := os.ReadFile("./example_file.ext")
//	gorequest.New().
//	  Post("http://example.com").
//	  Type("multipart").
//	  SendFile(b).
//	  End()
//
// Furthermore file can also be a os.File:
//
//	f, _ := os.Open("./example_file.ext")
//	gorequest.New().
//	  Post("http://example.com").
//	  Type("multipart").
//	  SendFile(f).
//	  End()
//
// The first optional argument (second argument overall) is the filename, which will be automatically determined when file is a string (path) or a os.File.
// When file is a []byte slice, filename defaults to "filename". In all cases the automatically determined filename can be overwritten:
//
//	b, _ := os.ReadFile("./example_file.ext")
//	gorequest.New().
//	  Post("http://example.com").
//	  Type("multipart").
//	  SendFile(b, "my_custom_filename").
//	  End()
//
// The second optional argument (third argument overall) is the fieldname in the multipart/form-data request. It defaults to fileNUMBER (eg. file1), where number is ascending and starts counting at 1.
// So if you send multiple files, the fieldnames will be file1, file2, ... unless it is overwritten. If fieldname is set to "file" it will be automatically set to fileNUMBER, where number is the greatest existing number+1 unless
// a third argument skipFileNumbering is provided and true.
//
//	b, _ := os.ReadFile("./example_file.ext")
//	gorequest.New().
//	  Post("http://example.com").
//	  Type("multipart").
//	  SendFile(b, "", "my_custom_fieldname"). // filename left blank, will become "example_file.ext"
//	  End()
//
// The third optional argument (fourth argument overall) is a bool value skipFileNumbering. It defaults to "false",
// if fieldname is "file" and skipFileNumbering is set to "false", the fieldname will be automatically set to
// fileNUMBER, where number is the greatest existing number+1.
//
//	b, _ := os.ReadFile("./example_file.ext")
//	gorequest.New().
//	  Post("http://example.com").
//	  Type("multipart").
//	  SendFile(b, "filename", "my_custom_fieldname", false).
//	  End()
//
// The fourth optional argument (fifth argument overall) is the mimetype request form-data part. It defaults to "application/octet-stream".
//
//	b, _ := os.ReadFile("./example_file.ext")
//	gorequest.New().
//	  Post("http://example.com").
//	  Type("multipart").
//	  SendFile(b, "filename", "my_custom_fieldname", false, "mime_type").
//	  End()
func (s *SuperAgent) SendFile(file any, args ...any) *SuperAgent {
	if file == nil {
		s.Errors = append(s.Errors, errors.New("sendFile func: nil file"))
		return s
	}

	filename := ""
	fieldname := "file"
	skipFileNumbering := false
	fileType := "application/octet-stream"

	if len(args) >= 1 {
		argFilename := fmt.Sprintf("%v", args[0])
		if len(argFilename) > 0 {
			filename = strings.TrimSpace(argFilename)
		}
	}

	if len(args) >= 2 {
		argFieldname := fmt.Sprintf("%v", args[1])
		if len(argFieldname) > 0 {
			fieldname = strings.TrimSpace(argFieldname)
		}
	}

	if len(args) >= 3 {
		argSkipFileNumbering := reflect.ValueOf(args[2])
		if !argSkipFileNumbering.IsValid() {
			s.Errors = append(s.Errors, errors.New("the fourth SendFile method argument for skipFileNumbering cannot be nil"))
			return s
		}
		if argSkipFileNumbering.Kind() == reflect.Bool {
			skipFileNumbering = argSkipFileNumbering.Interface().(bool)
		}
	}

	if len(args) >= 4 {
		argFileType := fmt.Sprintf("%v", args[3])
		if len(argFileType) > 0 {
			fileType = strings.TrimSpace(argFileType)
		}
		if fileType == "" {
			s.Errors = append(s.Errors, errors.New("the fifth SendFile method argument for MIME type cannot be an empty string"))
			return s
		}
	}

	if (fieldname == "file" && !skipFileNumbering) || fieldname == "" {
		fieldname = "file" + strconv.Itoa(len(s.FileData)+1)
	}

	v := reflect.ValueOf(file)
	switch v.Kind() {
	case reflect.String:
		pathToFile, err := filepath.Abs(v.String())
		if err != nil {
			s.Errors = append(s.Errors, err)
			return s
		}
		if filename == "" {
			filename = filepath.Base(pathToFile)
		}
		data, err := os.ReadFile(v.String())
		if err != nil {
			s.Errors = append(s.Errors, err)
			return s
		}
		s.FileData = append(s.FileData, File{
			Filename:  filename,
			Fieldname: fieldname,
			MimeType:  fileType,
			Data:      data,
		})
	case reflect.Slice:
		if v.Type().Elem().Kind() != reflect.Uint8 {
			s.Errors = append(s.Errors, errors.New("sendFile currently only supports byte slices as file content"))
			return s
		}
		if filename == "" {
			filename = "filename"
		}
		f := File{
			Filename:  filename,
			Fieldname: fieldname,
			MimeType:  fileType,
			Data:      make([]byte, v.Len()),
		}
		for i := 0; i < v.Len(); i++ {
			f.Data[i] = byte(v.Index(i).Uint())
		}
		s.FileData = append(s.FileData, f)
	case reflect.Pointer:
		if v.IsNil() {
			s.Errors = append(s.Errors, errors.New("sendFile func: nil pointer"))
			return s
		}
		if len(args) == 1 {
			return s.SendFile(v.Elem().Interface(), args[0])
		}
		if len(args) == 2 {
			return s.SendFile(v.Elem().Interface(), args[0], args[1])
		}
		if len(args) == 3 {
			return s.SendFile(v.Elem().Interface(), args[0], args[1], args[2])
		}
		if len(args) == 4 {
			return s.SendFile(v.Elem().Interface(), args[0], args[1], args[2], args[3])
		}
		return s.SendFile(v.Elem().Interface())
	default:
		if v.Type() == reflect.TypeOf(os.File{}) {
			osfile := v.Interface().(os.File)
			if filename == "" {
				filename = filepath.Base(osfile.Name())
			}
			data, err := os.ReadFile(osfile.Name())
			if err != nil {
				s.Errors = append(s.Errors, err)
				return s
			}
			s.FileData = append(s.FileData, File{
				Filename:  filename,
				Fieldname: fieldname,
				MimeType:  fileType,
				Data:      data,
			})
			return s
		}

		s.Errors = append(s.Errors, fmt.Errorf("sendFile currently only supports either a string (path/to/file), a slice of bytes (file content itself), or a os.File"))
	}

	return s
}
