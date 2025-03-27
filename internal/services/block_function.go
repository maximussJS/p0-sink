package services

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	fx_utils "p0-sink/internal/utils/fx"
	"rogchap.com/v8go"
)

type IBlockFunctionService interface {
	ApplyFunction(block []byte) ([]byte, error)
}

type blockFunctionServiceParams struct {
	fx.In

	StreamConfig IStreamConfig
}

type blockFunctionService struct {
	v8Isolate    *v8go.Isolate
	v8Global     *v8go.ObjectTemplate
	v8Ctx        *v8go.Context
	streamConfig IStreamConfig
	functionCode string
}

func FxBlockFunctionService() fx.Option {
	return fx_utils.AsProvider(newBlockFunctionService, new(IBlockFunctionService))
}

func newBlockFunctionService(lc fx.Lifecycle, params blockFunctionServiceParams) IBlockFunctionService {
	isolate := v8go.NewIsolate()
	global := v8go.NewObjectTemplate(isolate)
	v8ctx := v8go.NewContext(isolate, global)

	bs := &blockFunctionService{
		v8Isolate:    isolate,
		v8Global:     global,
		v8Ctx:        v8ctx,
		streamConfig: params.StreamConfig,
	}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			if params.StreamConfig.FunctionEnabled() {
				functionCode, err := params.StreamConfig.FunctionCode()

				if err != nil {
					return fmt.Errorf("block function service get function code error: %w", err)
				}

				bs.functionCode = string(functionCode)

				_, err = bs.v8Ctx.RunScript(bs.functionCode, "main.js")
				if err != nil {
					return fmt.Errorf("block function service run script error: %w", err)
				}

				return nil
			}

			return nil
		},
	})

	return bs
}

func (s *blockFunctionService) ApplyFunction(block []byte) ([]byte, error) {
	blockVal, err := s.jsonToV8Value(s.v8Ctx, string(block))
	if err != nil {
		return nil, fmt.Errorf("block function service json to v8 value error: %w", err)
	}

	blockMeta := "{}"
	metaVal, _ := s.jsonToV8Value(s.v8Ctx, blockMeta)

	val, err := s.v8Ctx.Global().Get("main")
	if err != nil {
		return nil, fmt.Errorf("block function service get main function error: %w", err)
	}

	mainFunc, err := val.AsFunction()

	if err != nil {
		return nil, fmt.Errorf("block function service main is not a function: %w", err)
	}

	result, err := mainFunc.Call(v8go.Undefined(s.v8Isolate), blockVal, metaVal)
	if err != nil {
		return nil, fmt.Errorf("block function service call main function error: %w", err)
	}

	json, err := result.MarshalJSON()

	if err != nil {
		return nil, fmt.Errorf("block function service marshal json error: %w", err)
	}

	return json, nil
}

func (s *blockFunctionService) jsonToV8Value(ctx *v8go.Context, jsonData string) (*v8go.Value, error) {
	script := fmt.Sprintf("JSON.parse(%q)", jsonData)
	return ctx.RunScript(script, "json_parser.js")
}
