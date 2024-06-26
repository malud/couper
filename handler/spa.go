package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/coupergateway/couper/config"
	"github.com/coupergateway/couper/config/runtime/server"
	"github.com/coupergateway/couper/errors"
	"github.com/coupergateway/couper/server/writer"
)

var (
	_ http.Handler = &Spa{}
)

type Spa struct {
	bootstrapContent []byte
	bootstrapModTime time.Time
	bootstrapOnce    sync.Once
	bootstrapCType   string
	config           *config.Spa
	modifier         []hcl.Body
	srvOptions       *server.Options
}

func NewSpa(ctx *hcl.EvalContext, config *config.Spa, srvOpts *server.Options, modifier []hcl.Body) (*Spa, error) {
	var err error
	if config.BootstrapFile, err = filepath.Abs(config.BootstrapFile); err != nil {
		return nil, err
	}

	spa := &Spa{
		config:     config,
		modifier:   modifier,
		srvOptions: srvOpts,
	}

	file, err := os.Open(config.BootstrapFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	spa.bootstrapModTime = fileInfo.ModTime()

	if config.BootstrapData == nil {
		return spa, nil
	} else if v, diags := config.BootstrapData.Value(ctx); v.IsNull() || diags.HasErrors() {
		if diags.HasErrors() {
			return nil, diags
		}
		return spa, nil
	}

	err = spa.replaceBootstrapData(ctx, file)

	return spa, err
}

func (s *Spa) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var content io.ReadSeeker
	var modTime time.Time

	if r, ok := rw.(*writer.Response); ok {
		r.AddModifier(s.modifier...)
	}

	if l := len(s.bootstrapContent); l > 0 {
		setLastModified(rw, s.bootstrapModTime)
		rw.Header().Set("Content-Type", s.bootstrapCType)
		rw.Header().Set("Content-Length", strconv.Itoa(l))
		rw.WriteHeader(http.StatusOK) // required for ResponseWriter
		_, _ = rw.Write(s.bootstrapContent)
		return
	}

	file, err := os.Open(s.config.BootstrapFile)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			s.srvOptions.ServerErrTpl.WithError(errors.RouteNotFound).ServeHTTP(rw, req)
			return
		}

		s.srvOptions.ServerErrTpl.WithError(errors.Configuration).ServeHTTP(rw, req)
		return
	}
	content = file
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil || fileInfo.IsDir() {
		s.srvOptions.ServerErrTpl.WithError(errors.Configuration).ServeHTTP(rw, req)
		return
	}

	modTime = fileInfo.ModTime()
	http.ServeContent(rw, req, s.config.BootstrapFile, modTime, content)
}

func (s *Spa) replaceBootstrapData(ctx *hcl.EvalContext, reader io.ReadCloser) error {
	if s.config.BootstrapData == nil {
		return nil
	}

	b, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	val, diags := s.config.BootstrapData.Value(ctx)
	if diags.HasErrors() {
		return diags
	}

	if !val.Type().IsObjectType() {
		r := s.config.BootstrapData.Range()
		return &hcl.Diagnostic{
			Detail:   "bootstrap_data must be an object type",
			Severity: hcl.DiagError,
			Subject:  &r,
			Summary:  "configuration error",
		}
	}

	data, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		return err
	}

	escapedData := &bytes.Buffer{}
	json.HTMLEscape(escapedData, data)

	const defaultName = "__BOOTSTRAP_DATA__"
	bootstrapName := s.config.BootStrapDataName
	if bootstrapName == "" {
		bootstrapName = defaultName
	}
	s.bootstrapContent = bytes.Replace(b, []byte(bootstrapName), escapedData.Bytes(), 1)

	s.bootstrapCType = mime.TypeByExtension(filepath.Ext(s.config.BootstrapFile))
	if s.bootstrapCType == "" {
		// read a chunk to decide between utf-8 text and binary
		var buf [512]byte
		n, _ := io.ReadFull(bytes.NewBuffer(b), buf[:])
		s.bootstrapCType = http.DetectContentType(buf[:n])
	}

	return nil
}

func (s *Spa) String() string {
	return "spa"
}

func setLastModified(w http.ResponseWriter, modtime time.Time) {
	if !modtime.IsZero() {
		w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	}
}
