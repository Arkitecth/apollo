package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Arkitecth/apollo/validator"
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil

}
func (app *application) readParamID(r *http.Request, key string) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName(key), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil

}

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError

		var unmarshalTypeError *json.UnmarshalTypeError

		var invalidUnmarshalError *json.InvalidUnmarshalError

		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed json at (character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly formed json")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains invalid field of type %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains invalid field at (character %d)", unmarshalTypeError.Offset)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not exceed limit of %d bytes", maxBytesError.Limit)

		case errors.Is(err, io.EOF):
			return fmt.Errorf("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown field %q", field)
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("body must have only one json body")
	}

	return nil
}

func (app *application) uploadS3(r *http.Request, key string) (string, error) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile(key)
	if err != nil {
		return "", err
	}
	dstCopy, err := os.Create(handler.Filename)
	if err != nil {
		return "", err
	}

	io.Copy(dstCopy, file)

	cfg, err := aws_config.LoadDefaultConfig(r.Context())
	if err != nil {
		return "", err
	}

	client := s3.NewFromConfig(cfg)
	bucketName := "apollomusicplayer"
	_, err = client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(handler.Filename),
		Body:   file,
	})
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, cfg.Region, handler.Filename)

	return url, nil

}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.Add(key, "must be an integer value")
		return defaultValue
	}
	return i
}

func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		fn()
	}()

}
