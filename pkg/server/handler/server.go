package handler

import (
	"net/http"
	"os"

	"github.com/acorn-io/brent/pkg/accesscontrol"
	"github.com/acorn-io/brent/pkg/builtin"
	handlers2 "github.com/acorn-io/brent/pkg/handlers"
	parse2 "github.com/acorn-io/brent/pkg/parse"
	"github.com/acorn-io/brent/pkg/subscribe"
	types2 "github.com/acorn-io/brent/pkg/types"
	writer2 "github.com/acorn-io/brent/pkg/writer"
	"github.com/acorn-io/schemer/validation"
	"golang.org/x/exp/slices"
)

type RequestHandler interface {
	http.Handler

	GetSchemas() *types2.APISchemas
	Handle(apiOp *types2.APIRequest)
}

type Server struct {
	ResponseWriters map[string]types2.ResponseWriter
	Schemas         *types2.APISchemas
	AccessControl   types2.AccessControl
	Parser          parse2.Parser
	URLParser       parse2.URLParser
}

func defaultAPIServer() *Server {
	s := &Server{
		Schemas: types2.EmptyAPISchemas().MustAddSchemas(builtin.Schemas),
		ResponseWriters: map[string]types2.ResponseWriter{
			"json": &writer2.GzipWriter{
				ResponseWriter: &writer2.EncodingResponseWriter{
					ContentType: "application/json",
					Encoder:     types2.JSONEncoder,
				},
			},
			"jsonl": &writer2.GzipWriter{
				ResponseWriter: &writer2.EncodingResponseWriter{
					ContentType: "application/jsonl",
					Encoder:     types2.JSONLinesEncoder,
				},
			},
			"html": &writer2.GzipWriter{
				ResponseWriter: &writer2.HTMLResponseWriter{
					EncodingResponseWriter: writer2.EncodingResponseWriter{
						Encoder:     types2.JSONEncoder,
						ContentType: "application/json",
					},
				},
			},
			"yaml": &writer2.GzipWriter{
				ResponseWriter: &writer2.EncodingResponseWriter{
					ContentType: "application/yaml",
					Encoder:     types2.YAMLEncoder,
				},
			},
		},
		AccessControl: &accesscontrol.SchemaBasedAccess{},
		Parser:        parse2.Parse,
		URLParser:     parse2.MuxURLParser,
	}

	subscribe.Register(s.Schemas, subscribe.DefaultGetter, os.Getenv("SERVER_VERSION"))
	return s
}

func (s *Server) setDefaults(ctx *types2.APIRequest) {
	if ctx.ResponseWriter == nil {
		ctx.ResponseWriter = s.ResponseWriters[ctx.ResponseFormat]
		if ctx.ResponseWriter == nil {
			ctx.ResponseWriter = s.ResponseWriters["json"]
		}
	}

	if ctx.ErrorHandler == nil {
		ctx.ErrorHandler = handlers2.ErrorHandler
	}

	ctx.AccessControl = s.AccessControl

	if ctx.Schemas == nil {
		ctx.Schemas = s.Schemas
	}
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.Handle(&types2.APIRequest{
		Request:  req,
		Response: rw,
	})
}

func (s *Server) Handle(apiOp *types2.APIRequest) {
	s.handle(apiOp, s.Parser)
}

func (s *Server) handle(apiOp *types2.APIRequest, parser parse2.Parser) {
	if apiOp.Schemas == nil {
		apiOp.Schemas = s.Schemas
	}

	if err := parser(apiOp, parse2.MuxURLParser); err != nil {
		// ensure defaults set so writer is assigned
		s.setDefaults(apiOp)
		apiOp.WriteError(err)
		return
	}

	s.setDefaults(apiOp)

	var cloned *types2.APISchemas
	for id, schema := range apiOp.Schemas.Schemas {
		if schema.RequestModifier == nil {
			continue
		}

		if cloned == nil {
			cloned = apiOp.Schemas.ShallowCopy()
		}

		schema := schema.DeepCopy()
		schema = schema.RequestModifier(apiOp, schema)
		cloned.Schemas[id] = schema
	}

	if cloned != nil {
		apiOp.Schemas = cloned
	}

	if apiOp.Schema != nil && apiOp.Schema.RequestModifier != nil {
		apiOp.Schema = apiOp.Schema.RequestModifier(apiOp, apiOp.Schema)
	}

	var code int
	var data interface{}
	var err error
	if code, data, err = s.handleOp(apiOp); err != nil {
		apiOp.WriteError(err)
	} else if obj, ok := data.(types2.APIObject); ok {
		apiOp.WriteResponse(code, obj)
	} else if list, ok := data.(types2.APIObjectList); ok {
		apiOp.WriteResponseList(code, list)
	} else if code > http.StatusOK {
		if code == http.StatusNotFound && slices.Contains([]string{"system:unauthenticated", "system:cattle:error"}, apiOp.GetUser()) {
			apiOp.Response.WriteHeader(http.StatusUnauthorized)
		} else {
			apiOp.Response.WriteHeader(code)
		}
	}
}

func (s *Server) handleOp(apiOp *types2.APIRequest) (int, interface{}, error) {
	if err := checkCSRF(apiOp); err != nil {
		return 0, nil, err
	}

	if apiOp.Schema == nil {
		return http.StatusNotFound, nil, nil
	}

	action, err := validateAction(apiOp)
	if err != nil {
		return 0, nil, err
	}

	if action != nil {
		return http.StatusOK, nil, handleAction(apiOp)
	}

	switch apiOp.Method {
	case http.MethodGet:
		if apiOp.Name == "" {
			data, err := handleList(apiOp, apiOp.Schema.ListHandler, handlers2.ListHandler)
			return http.StatusOK, data, err
		}
		data, err := handle(apiOp, apiOp.Schema.ByIDHandler, handlers2.ByIDHandler)
		return http.StatusOK, data, err
	case http.MethodPatch:
		fallthrough
	case http.MethodPut:
		data, err := handle(apiOp, apiOp.Schema.UpdateHandler, handlers2.UpdateHandler)
		return http.StatusOK, data, err
	case http.MethodPost:
		data, err := handle(apiOp, apiOp.Schema.CreateHandler, handlers2.CreateHandler)
		return http.StatusCreated, data, err
	case http.MethodDelete:
		data, err := handle(apiOp, apiOp.Schema.DeleteHandler, handlers2.DeleteHandler)
		return http.StatusOK, data, err
	}

	return http.StatusNotFound, nil, nil
}

func handleList(apiOp *types2.APIRequest, custom types2.RequestListHandler, handler types2.RequestListHandler) (types2.APIObjectList, error) {
	if custom != nil {
		return custom(apiOp)
	}
	return handler(apiOp)
}

func handle(apiOp *types2.APIRequest, custom types2.RequestHandler, handler types2.RequestHandler) (types2.APIObject, error) {
	if custom != nil {
		return custom(apiOp)
	}
	return handler(apiOp)
}

func handleAction(context *types2.APIRequest) error {
	if err := context.AccessControl.CanAction(context, context.Schema, context.Action); err != nil {
		return err
	}
	if handler, ok := context.Schema.ActionHandlers[context.Action]; ok {
		handler.ServeHTTP(context.Response, context.Request)
		return validation.ErrComplete
	}
	return nil
}

func (s *Server) CustomAPIUIResponseWriter(cssURL, jsURL, version writer2.StringGetter) {
	wi, ok := s.ResponseWriters["html"]
	if !ok {
		return
	}
	gw, ok := wi.(*writer2.GzipWriter)
	if !ok {
		return
	}

	w, ok := gw.ResponseWriter.(*writer2.HTMLResponseWriter)
	if !ok {
		return
	}
	w.CSSURL = cssURL
	w.JSURL = jsURL
	w.APIUIVersion = version
}
